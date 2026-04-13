package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"daileycoding/backend/internal/handler"
	dockerexec "daileycoding/backend/internal/infrastructure/docker"
	"daileycoding/backend/internal/infrastructure/db"
	"daileycoding/backend/internal/infrastructure/db/repository"
	"daileycoding/backend/internal/usecase"
)

func main() {
	image         := getEnv("PYTHON_IMAGE",    "python:3.14-slim")
	addr          := getEnv("ADDR",            ":8080")
	frontendOrigin := getEnv("FRONTEND_ORIGIN", "http://localhost:3000")
	dsn           := getEnv("DATABASE_URL",
		"postgresql://daileycoding:daileycoding@localhost:5432/daileycoding?sslmode=disable")

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

	// ── コード実行エンジン ──
	executor, err := dockerexec.NewExecutor(image)
	if err != nil {
		log.Fatalf("init docker executor: %v", err)
	}
	pullCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := executor.EnsureImage(pullCtx); err != nil {
		log.Fatalf("ensure image: %v", err)
	}

	// ── ユースケース / ハンドラー ──
	judgeUsecase  := usecase.NewJudgeUsecase(executor, problemRepo, submissionRepo)
	submitHandler := handler.NewSubmitHandler(judgeUsecase)
	problemHandler := handler.NewProblemHandler(problemRepo)

	// ── Echo ルーター ──
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.CORS(frontendOrigin))

	e.GET("/api/problems",          problemHandler.List)
	e.GET("/api/problems/today",    problemHandler.Today)
	e.GET("/api/problems/:id",      problemHandler.GetByID)
	e.POST("/api/problems",         problemHandler.Create)
	e.PATCH("/api/problems/:id",    problemHandler.Update)
	e.GET("/api/admin/problems",    problemHandler.ListAll)
	e.POST("/api/submit",           submitHandler.Handle)

	log.Printf("server starting on %s (image: %s)", addr, image)
	if err := e.Start(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
