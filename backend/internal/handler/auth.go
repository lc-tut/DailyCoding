package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	"golang.org/x/oauth2"

	"daileycoding/backend/internal/domain"
	"daileycoding/backend/internal/usecase"
)

const (
	cookieName   = "dc_session"
	stateCookie  = "dc_oauth_state"
	nonceCookie  = "dc_oauth_nonce"
	returnCookie = "dc_return_to"
	cookieMaxAge = 24 * 60 * 60 // 24 時間 (秒)
)

// OIDCVerifier は ID Token を検証し OIDCClaims を返すインターフェース。
// infrastructure/oidc.Verifier が実装する。
type OIDCVerifier interface {
	Verify(ctx context.Context, rawIDToken, expectedNonce string) (*domain.OIDCClaims, error)
}

// ProviderConfig は OAuth2 / OIDC プロバイダーの HTTP レベル設定
type ProviderConfig struct {
	Name     string
	OAuth2   *oauth2.Config
	Verifier OIDCVerifier
}

// authUsecase は AuthHandler が依存するユースケースのインターフェース
type authUsecase interface {
	ProcessLogin(ctx context.Context, provider string, claims *domain.OIDCClaims) (string, error)
}

// AuthHandler は認証フローの HTTP 層を担う。ビジネスロジックは usecase に委譲する。
type AuthHandler struct {
	usecase     authUsecase
	providers   map[string]*ProviderConfig
	frontendURL string
	bypass      bool
}

func NewAuthHandler(
	uc *usecase.AuthUsecase,
	providers []*ProviderConfig,
	frontendURL string,
	bypass bool,
) *AuthHandler {
	pm := make(map[string]*ProviderConfig, len(providers))
	for _, p := range providers {
		pm[p.Name] = p
	}
	return &AuthHandler{
		usecase:     uc,
		providers:   pm,
		frontendURL: frontendURL,
		bypass:      bypass,
	}
}

// Providers GET /api/auth/providers
func (h *AuthHandler) Providers(c *echo.Context) error {
	if h.bypass {
		return c.JSON(http.StatusOK, map[string]any{"providers": []string{}})
	}
	names := make([]string, 0, len(h.providers))
	for name := range h.providers {
		names = append(names, name)
	}
	return c.JSON(http.StatusOK, map[string]any{"providers": names})
}

// Redirect GET /api/auth/:provider — state・nonce を発行して認可画面へリダイレクト
func (h *AuthHandler) Redirect(c *echo.Context) error {
	if h.bypass {
		return c.Redirect(http.StatusFound, h.frontendURL)
	}

	p, err := h.lookupProvider(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	state, err := randomToken()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "generate state"})
	}
	nonce, err := randomToken()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "generate nonce"})
	}

	setTempCookie(c, stateCookie, state)
	setTempCookie(c, nonceCookie, nonce)

	if next := c.QueryParam("next"); isSafePath(next) {
		setTempCookie(c, returnCookie, next)
	}

	url := p.OAuth2.AuthCodeURL(state,
		oauth2.AccessTypeOnline,
		oauth2.SetAuthURLParam("nonce", nonce),
	)
	return c.Redirect(http.StatusFound, url)
}

// Callback GET /api/auth/:provider/callback — OIDC コード交換 → ID Token 検証 → ログイン処理
func (h *AuthHandler) Callback(c *echo.Context) error {
	p, err := h.lookupProvider(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	// state 検証（CSRF 対策）
	stateCookieVal, err := c.Cookie(stateCookie)
	if err != nil || stateCookieVal.Value != c.QueryParam("state") {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid state"})
	}
	deleteCookie(c, stateCookie)

	// nonce Cookie を取得・削除
	nonceCookieVal, err := c.Cookie(nonceCookie)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "missing nonce cookie"})
	}
	expectedNonce := nonceCookieVal.Value
	deleteCookie(c, nonceCookie)

	// Authorization Code → トークンセット
	token, err := p.OAuth2.Exchange(c.Request().Context(), c.QueryParam("code"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "code exchange failed: " + err.Error()})
	}

	// ID Token を取り出す
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "id_token not found in token response"})
	}

	// ID Token 検証（署名・iss・aud・exp・nonce）
	claims, err := p.Verifier.Verify(c.Request().Context(), rawIDToken, expectedNonce)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]any{"error": "id token verification failed: " + err.Error()})
	}

	// ビジネスロジック（メール許可チェック・プロビジョニング・JWT発行）
	appToken, err := h.usecase.ProcessLogin(c.Request().Context(), p.Name, claims)
	if err != nil {
		var authErr *usecase.AuthError
		if errors.As(err, &authErr) {
			return c.Redirect(http.StatusFound, h.frontendURL+"/login?error="+authErr.Code)
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}

	returnTo := "/"
	if rc, err := c.Cookie(returnCookie); err == nil && isSafePath(rc.Value) {
		returnTo = rc.Value
	}
	deleteCookie(c, returnCookie)

	setSessionCookie(c, appToken, cookieMaxAge)
	return c.Redirect(http.StatusFound, h.frontendURL+returnTo)
}

// Me GET /api/auth/me — 現在のログインユーザー情報を返す
func (h *AuthHandler) Me(c *echo.Context) error {
	if h.bypass {
		return c.JSON(http.StatusOK, map[string]any{
			"id":    "00000000-0000-0000-0000-000000000001",
			"email": "admin@linuxclub.example",
			"name":  "テスト管理者",
			"role":  "admin",
		})
	}
	user := userFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]any{"error": "not authenticated"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"id":    user.ID.String(),
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
	})
}

// Logout POST /api/auth/logout — Cookie を削除
func (h *AuthHandler) Logout(c *echo.Context) error {
	setSessionCookie(c, "", -1)
	return c.JSON(http.StatusOK, map[string]any{"ok": true})
}

// --- ヘルパー ---

func (h *AuthHandler) lookupProvider(c *echo.Context) (*ProviderConfig, error) {
	name := c.Param("provider")
	p, ok := h.providers[name]
	if !ok {
		return nil, errors.New("unknown provider: " + name)
	}
	return p, nil
}

func userFromContext(c *echo.Context) *domain.User {
	u, _ := (*c).Get("auth_user").(*domain.User)
	return u
}

func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func setSessionCookie(c *echo.Context, value string, maxAge int) {
	cookie := new(http.Cookie)
	cookie.Name = cookieName
	cookie.Value = value
	cookie.HttpOnly = true
	cookie.Path = "/"
	cookie.MaxAge = maxAge
	cookie.SameSite = http.SameSiteLaxMode
	(*c).SetCookie(cookie)
}

func setTempCookie(c *echo.Context, name, value string) {
	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = value
	cookie.HttpOnly = true
	cookie.Path = "/"
	cookie.MaxAge = 600 // 10 分
	cookie.SameSite = http.SameSiteLaxMode
	(*c).SetCookie(cookie)
}

// isSafePath はリダイレクト先が自サイト内の相対パスかどうか検証する（オープンリダイレクト対策）
func isSafePath(p string) bool {
	return len(p) > 0 && p[0] == '/' && (len(p) < 2 || p[1] != '/')
}

func deleteCookie(c *echo.Context, name string) {
	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = ""
	cookie.MaxAge = -1
	cookie.Path = "/"
	(*c).SetCookie(cookie)
}
