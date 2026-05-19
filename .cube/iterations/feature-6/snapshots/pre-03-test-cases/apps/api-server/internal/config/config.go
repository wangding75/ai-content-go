package config

import "os"

type Config struct {
	AppName     string
	Environment string
	BuildCommit string
	HTTPAddr    string
	DatabaseURL string
}

func Load() (Config, error) {
	return Config{
		AppName:     valueOrDefault("APP_NAME", "ai-content-go"),
		Environment: valueOrDefault("APP_ENV", "development"),
		BuildCommit: os.Getenv("BUILD_COMMIT"),
		HTTPAddr:    valueOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}, nil
}

func valueOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
