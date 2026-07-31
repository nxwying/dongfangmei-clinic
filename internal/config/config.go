package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Port            string
	DBDsn           string
	DBDriver        string
	JWTSecret       string
	DataDir         string
	ForcePassChange bool
}

// secretFile is where the auto-generated JWT secret is persisted.
func secretFile(dataDir string) string {
	return filepath.Join(dataDir, ".jwt_secret")
}

// loadOrGenerateSecret reads the persisted JWT secret, or generates and
// saves a random one on first launch. This prevents hard-coded secrets.
func loadOrGenerateSecret(dataDir string) string {
	// If JWT_SECRET env var is explicitly set, honor it
	if envSecret := os.Getenv("JWT_SECRET"); envSecret != "" && envSecret != "change-me-in-production-please" {
		return envSecret
	}

	sf := secretFile(dataDir)
	if data, err := os.ReadFile(sf); err == nil && len(data) >= 32 {
		return string(data)
	}

	// Generate 32 random bytes → 64 hex chars
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		log.Fatalf("failed to generate JWT secret: %v", err)
	}
	secret := hex.EncodeToString(buf)

	os.MkdirAll(dataDir, 0755)
	if err := os.WriteFile(sf, []byte(secret), 0600); err != nil {
		log.Printf("warning: could not persist JWT secret: %v", err)
	}
	log.Println("Generated new JWT secret (persisted)")
	return secret
}

func Load() *Config {
	// Default data dir: user's home ~/.clinic-mgmt/data (consistent across launches)
	homeDir, _ := os.UserHomeDir()
	dataDir := getEnv("DATA_DIR", filepath.Join(homeDir, ".clinic-mgmt", "data"))

	driver := getEnv("DB_DRIVER", "postgres")
	dsn := getEnv("DB_DSN", "")
	if dsn == "" {
		if driver == "sqlite" {
			dsn = filepath.Join(dataDir, "clinic.db")
		} else {
			dsn = "host=localhost user=clinic password=clinic123 dbname=clinic port=5432 sslmode=disable"
		}
	}

	return &Config{
		Port:            getEnv("PORT", "8080"),
		DBDsn:           dsn,
		DBDriver:        driver,
		JWTSecret:       loadOrGenerateSecret(dataDir),
		DataDir:         dataDir,
		ForcePassChange: true,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
