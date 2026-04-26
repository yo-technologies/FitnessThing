package main

import (
	"context"
	"database/sql"
	"fitness-trainer/internal/logger"
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func init() {
	logger.Init()
	godotenv.Load()
}

func loadPostgresURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.PathEscape(os.Getenv("POSTGRES_USER")),
		url.PathEscape(os.Getenv("POSTGRES_PASSWORD")),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_SSL_MODE"),
	)
}

func main() {
	postgresURL := loadPostgresURL()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sqlDB, err := sql.Open("postgres", postgresURL)
	if err != nil {
		logger.Fatal(err.Error())
	}

	if err := goose.RunContext(ctx, "up", sqlDB, "migrations"); err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info("migrations applied")
}
