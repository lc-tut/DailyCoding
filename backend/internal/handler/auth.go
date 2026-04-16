package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"golang.org/x/oauth2"

	"daileycoding/backend/internal/domain"
)

const (
	cookieName    = "dc_session"
	stateCookie   = "dc_oauth_state"
	cookieMaxAge  = 24 * 60 * 60 // 24 時間 (秒)
)

type userRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpsertWithIdentity(ctx context.Context, email, name, provider, providerSub string) (*domain.User, error)
}

// ProviderConfig は OAuth2 プロバイダー設定をまとめた型
type ProviderConfig struct {
	Name     string        // "google" | "keycloak"
	OAuth2   *oauth2.Config
	UserinfoURL string     // ユーザー情報取得エンドポイント
}

// AppClaims は アプリ独自 JWT のペイロード
type AppClaims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthHandler は認証フロー全体を担う
type AuthHandler struct {
	userRepo    userRepo
	jwtSecret   []byte
	providers   map[string]*ProviderConfig
	frontendURL string
	bypass      bool // AUTH_BYPASS=true のとき真
}

func NewAuthHandler(
	repo userRepo,
	jwtSecret []byte,
	providers []*ProviderConfig,
	frontendURL string,
	bypass bool,
) *AuthHandler {
	pm := make(map[string]*ProviderConfig, len(providers))
	for _, p := range providers {
		pm[p.Name] = p
	}
	return &AuthHandler{
		userRepo:    repo,
		jwtSecret:   jwtSecret,
		providers:   pm,
		frontendURL: frontendURL,
		bypass:      bypass,
	}
}

// --- ヘルパー ---

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (h *AuthHandler) issueToken(user *domain.User) (string, error) {
	claims := AppClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Name:   user.Name,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(h.jwtSecret)
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

// --- ルートハンドラー ---

// Redirect GET /api/auth/:provider — OAuth2 認可画面へリダイレクト
func (h *AuthHandler) Redirect(c *echo.Context) error {
	providerName := c.Param("provider")
	p, ok := h.providers[providerName]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "unknown provider: " + providerName})
	}

	state, err := randomState()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "generate state"})
	}

	// state を短命 Cookie に保存（CSRF 対策）
	sc := new(http.Cookie)
	sc.Name = stateCookie
	sc.Value = state
	sc.HttpOnly = true
	sc.Path = "/"
	sc.MaxAge = 600 // 10 分
	sc.SameSite = http.SameSiteLaxMode
	c.SetCookie(sc)

	url := p.OAuth2.AuthCodeURL(state, oauth2.AccessTypeOnline)
	return c.Redirect(http.StatusFound, url)
}

// Callback GET /api/auth/:provider/callback — コード交換 → JIT → JWT 発行
func (h *AuthHandler) Callback(c *echo.Context) error {
	providerName := c.Param("provider")
	p, ok := h.providers[providerName]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "unknown provider: " + providerName})
	}

	// state 検証
	stateCookieVal, err := c.Cookie(stateCookie)
	if err != nil || stateCookieVal.Value != c.QueryParam("state") {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid state"})
	}
	// state Cookie を削除
	sc := new(http.Cookie)
	sc.Name = stateCookie
	sc.Value = ""
	sc.MaxAge = -1
	sc.Path = "/"
	c.SetCookie(sc)

	// Authorization Code → Access Token
	token, err := p.OAuth2.Exchange(c.Request().Context(), c.QueryParam("code"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "code exchange failed: " + err.Error()})
	}

	// ユーザー情報取得
	info, err := fetchUserInfo(c.Request().Context(), token.AccessToken, p.UserinfoURL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "fetch userinfo: " + err.Error()})
	}

	// JIT プロビジョニング
	user, err := h.userRepo.UpsertWithIdentity(
		c.Request().Context(),
		info.Email, info.Name, providerName, info.Sub,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "upsert user: " + err.Error()})
	}

	// アプリ独自 JWT を発行して HttpOnly Cookie にセット
	appToken, err := h.issueToken(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "issue token: " + err.Error()})
	}
	setSessionCookie(c, appToken, cookieMaxAge)

	return c.Redirect(http.StatusFound, h.frontendURL)
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

// userFromContext は echo.Context からユーザーを取り出すヘルパー
// (middleware パッケージへの依存を避けるため handler 内で定義)
func userFromContext(c *echo.Context) *domain.User {
	u, _ := (*c).Get("auth_user").(*domain.User)
	return u
}

// --- ユーザー情報取得 ---

type oauthUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func fetchUserInfo(ctx context.Context, accessToken, userinfoURL string) (*oauthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userinfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info oauthUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	if info.Email == "" || info.Sub == "" {
		return nil, fmt.Errorf("missing email or sub in userinfo response")
	}
	return &info, nil
}
