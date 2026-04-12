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
	"daileycoding/backend/internal/usecase"
)

func main() {
	image := getEnv("PYTHON_IMAGE", "python:3.14-slim")
	addr := getEnv("ADDR", ":8080")
	frontendOrigin := getEnv("FRONTEND_ORIGIN", "http://localhost:3000")

	executor, err := dockerexec.NewExecutor(image)
	if err != nil {
		log.Fatalf("init docker executor: %v", err)
	}

	// サーバー起動時にイメージを pull しておく
	pullCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := executor.EnsureImage(pullCtx); err != nil {
		log.Fatalf("ensure image: %v", err)
	}

	judgeUsecase := usecase.NewJudgeUsecase(executor)
	submitHandler := handler.NewSubmitHandler(judgeUsecase)

	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.CORS(frontendOrigin))

	e.POST("/api/submit", submitHandler.Handle)

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
