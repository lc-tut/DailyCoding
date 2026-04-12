package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"daileycoding/backend/internal/handler"
	dockerexec "daileycoding/backend/internal/infrastructure/docker"
	"daileycoding/backend/internal/usecase"
)

func main() {
	image := getEnv("PYTHON_IMAGE", "python:3.14-slim")
	addr := getEnv("ADDR", ":8080")

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

	mux := http.NewServeMux()
	mux.Handle("/api/submit", withCORS(submitHandler))

	log.Printf("server starting on %s (image: %s)", addr, image)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// withCORS はフロントエンド(localhost:3000)からのリクエストを許可する
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
