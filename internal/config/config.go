// internal/config/config.go
package config

import (
	"os"
)

type Config struct {
	Port      string
	DBDsn     string
	DBDriver  string
	JWTSecret string
}

func Load() *Config {
	driver := getEnv("DB_DRIVER", "postgres")
	dsn := getEnv("DB_DSN", "")
	if dsn == "" {
		if driver == "sqlite" {
			dsn = "data/clinic.db"
		} else {
			dsn = "host=localhost user=clinic password=clinic123 dbname=clinic port=5432 sslmode=disable"
		}
	}
	return &Config{
		Port:      getEnv("PORT", "8080"),
		DBDsn:     dsn,
		DBDriver:  driver,
		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production-please"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
