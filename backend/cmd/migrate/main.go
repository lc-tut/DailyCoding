package main

import (
	"log"
	"os"

	"daileycoding/backend/internal/infrastructure/db"
)

func main() {
	dsn := getEnv("DATABASE_URL",
		"postgresql://daileycoding:daileycoding@localhost:5432/daileycoding?sslmode=disable")

	log.Println("running migrations...")
	if err := db.RunMigrations(dsn); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migrations done")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
