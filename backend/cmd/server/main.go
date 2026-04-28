package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
	"golang.org/x/oauth2"

	"daileycoding/backend/internal/handler"
	dockerexec "daileycoding/backend/internal/infrastructure/docker"
	"daileycoding/backend/internal/infrastructure/db"
	"daileycoding/backend/internal/infrastructure/db/repository"
	oidcinfra "daileycoding/backend/internal/infrastructure/oidc"
	snekboxexec "daileycoding/backend/internal/infrastructure/snekbox"
	appmw "daileycoding/backend/internal/middleware"
	"daileycoding/backend/internal/usecase"
)

func main() {
	image          := getEnv("PYTHON_IMAGE",    "python:3.14-slim")
	addr           := getEnv("ADDR",            ":8080")
	frontendOrigin := getEnv("FRONTEND_ORIGIN", "http://localhost:3000")
	dsn            := getEnv("DATABASE_URL",
		"postgresql://daileycoding:daileycoding@localhost:5432/daileycoding?sslmode=disable")
	jwtSecret     := []byte(getEnv("JWT_SECRET", "dev-secret-change-in-production"))
	bypass        := getEnv("AUTH_BYPASS", "") == "true"
	allowDomains  := splitCSV(os.Getenv("ALLOWED_DOMAINS"))
	// FRONTEND_ORIGIN が https:// で始まる場合は本番 HTTPS 環境とみなし Cookie に Secure フラグを付与する
	cookieSecure  := strings.HasPrefix(frontendOrigin, "https://")

	// ── DB マイグレーション ──
	log.Println("running migrations...")
	if err := db.RunMigrations(dsn); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// ── DB 接続プール ──
	pool, err := db.NewPool(context.Background(), dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	// ── リポジトリ ──
	problemRepo    := repository.NewProblemRepository(pool)
	submissionRepo := repository.NewSubmissionRepository(pool)
	userRepo       := repository.NewUserRepository(pool)

	// ALLOWED_DOMAINS が未設定のときは allowedEmailRepo を nil にして全員許可（開発用）
	// ALLOWED_DOMAINS が設定されている場合は allowed_emails テーブルも参照する
	var allowedEmailRepo usecase.AuthAllowedEmailRepository
	if len(allowDomains) > 0 {
		allowedEmailRepo = repository.NewAllowedEmailRepository(pool)
	}

	// ── コード実行エンジン ──
	// EXECUTOR_TYPE=snekbox のとき snekbox HTTP API を使う。
	// 未設定または "docker" のときは従来の Docker コンテナ方式を使う。
	var executor usecase.Executor
	switch getEnv("EXECUTOR_TYPE", "docker") {
	case "snekbox":
		snekboxURL := getEnv("SNEKBOX_URL", "http://snekbox:8060")
		executor = snekboxexec.NewExecutor(snekboxURL)
		log.Printf("executor: snekbox (%s)", snekboxURL)
	default:
		dockerExecutor, err := dockerexec.NewExecutor(image)
		if err != nil {
			log.Fatalf("init docker executor: %v", err)
		}
		pullCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if err := dockerExecutor.EnsureImage(pullCtx); err != nil {
			log.Fatalf("ensure image: %v", err)
		}
		executor = dockerExecutor
		log.Printf("executor: docker (%s)", image)
	}

	// ── OIDC プロバイダー設定 ──
	oidcCtx, oidcCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer oidcCancel()
	providers := buildProviders(oidcCtx)

	// ── ユースケース / ハンドラー ──
	judgeUsecase      := usecase.NewJudgeUsecase(executor, problemRepo, submissionRepo)
	submitHandler     := handler.NewSubmitHandler(judgeUsecase)
	problemHandler    := handler.NewProblemHandler(problemRepo)
	submissionHandler := handler.NewSubmissionHandler(submissionRepo)
	authUsecase       := usecase.NewAuthUsecase(userRepo, allowedEmailRepo, allowDomains, jwtSecret)
	authHandler       := handler.NewAuthHandler(authUsecase, providers, frontendOrigin, bypass, cookieSecure)

	// ── Echo ルーター ──
	e := echo.New()
	e.Use(echomw.Recover())
	e.Use(echomw.RequestLogger())
	e.Use(corsMiddleware(frontendOrigin))

	// ── 認証ルート（認証不要）──
	e.GET("/api/auth/providers",          authHandler.Providers)
	e.GET("/api/auth/:provider",          authHandler.Redirect)
	e.GET("/api/auth/:provider/callback", authHandler.Callback)
	e.POST("/api/auth/logout",            authHandler.Logout)

	// ── 公開ルート ──
	e.GET("/api/problems",                 problemHandler.List)
	e.GET("/api/problems/today",           problemHandler.Today)
	e.GET("/api/problems/:id",             problemHandler.GetByID)
	e.GET("/api/problems/:id/submissions", submissionHandler.List)
	e.GET("/api/submissions/:id",          submissionHandler.GetByID)

	// ── 認証必須ルート ──
	authRequired := appmw.Required(jwtSecret, bypass)

	// /api/auth/me は JWT を検証してユーザー情報を返す（未認証なら 401）
	me := e.Group("")
	me.Use(authRequired)
	me.GET("/api/auth/me", authHandler.Me)

	protected := e.Group("")
	protected.Use(authRequired)
	protected.POST("/api/problems",      problemHandler.Create)
	protected.PATCH("/api/problems/:id", problemHandler.Update)
	protected.POST("/api/submit",        submitHandler.Handle)

	// ── admin 限定ルート ──
	admin := e.Group("")
	admin.Use(authRequired)
	admin.Use(appmw.AdminRequired())
	admin.GET("/api/admin/problems", problemHandler.ListAll)

	log.Printf("server starting on %s (image: %s, bypass=%v, allowDomains=%v)",
		addr, image, bypass, allowDomains)
	if err := e.Start(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// buildProviders は利用可能な OIDC プロバイダーを設定する。
// 環境変数が未設定のプロバイダーはスキップされる。
func buildProviders(ctx context.Context) []*handler.ProviderConfig {
	var providers []*handler.ProviderConfig

	if clientID := os.Getenv("GOOGLE_CLIENT_ID"); clientID != "" {
		p, err := oidc.NewProvider(ctx, "https://accounts.google.com")
		if err != nil {
			log.Fatalf("init google oidc provider: %v", err)
		}
		providers = append(providers, &handler.ProviderConfig{
			Name: "google",
			OAuth2: &oauth2.Config{
				ClientID:     clientID,
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
				Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
				Endpoint:     p.Endpoint(),
			},
			Verifier: oidcinfra.NewVerifier(p.Verifier(&oidc.Config{ClientID: clientID})),
		})
		log.Println("oidc provider registered: google")
	}

	if clientID := os.Getenv("KEYCLOAK_CLIENT_ID"); clientID != "" {
		issuer := os.Getenv("KEYCLOAK_ISSUER")
		p, err := oidc.NewProvider(ctx, issuer)
		if err != nil {
			log.Fatalf("init keycloak oidc provider: %v", err)
		}
		providers = append(providers, &handler.ProviderConfig{
			Name: "keycloak",
			OAuth2: &oauth2.Config{
				ClientID:     clientID,
				ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
				RedirectURL:  os.Getenv("KEYCLOAK_CALLBACK_URL"),
				Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
				Endpoint:     p.Endpoint(),
			},
			Verifier: oidcinfra.NewVerifier(p.Verifier(&oidc.Config{ClientID: clientID})),
		})
		log.Println("oidc provider registered: keycloak")
	}

	return providers
}

// corsMiddleware は credentials を許可した CORS ヘッダーを設定する
func corsMiddleware(allowOrigin string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			(*c).Response().Header().Set("Access-Control-Allow-Origin", allowOrigin)
			(*c).Response().Header().Set("Access-Control-Allow-Credentials", "true")
			(*c).Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			(*c).Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			if (*c).Request().Method == http.MethodOptions {
				return (*c).NoContent(http.StatusNoContent)
			}
			return next(c)
		}
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// splitCSV はカンマ区切り文字列をトリムしてスライスに変換する。空文字は除外。
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}
