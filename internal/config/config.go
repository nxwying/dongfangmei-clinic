// internal/config/config.go
package config

import (
	"os"
)

type Config struct {
	Port      string
	DBDsn     string
	JWTSecret string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		DBDsn:     getEnv("DB_DSN", "host=localhost user=clinic password=clinic123 dbname=clinic port=5432 sslmode=disable"),
		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production-please"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
