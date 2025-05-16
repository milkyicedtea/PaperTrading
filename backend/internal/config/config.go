package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	AppPort  string
	AppEnv   string
	LogLevel string

	JWTSecret              string
	JWTExpiration          time.Duration
	RefreshTokenExpiration time.Duration

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSslMode  string
}

// Load loads configuration from environment variables.
// It will attempt to load a .env file if present.
func Load() (*Config, error) {
	// attempt to load .env file.
	// in production, variables are set directly or through docker secrets (needs implementation).
	// the error is ignored because we don't strictly require a .env file.
	_ = godotenv.Load(".env")

	jwtExpMinutes, err := strconv.Atoi(getEnv("JWT_EXPIRATION_MINUTES", "15"))
	if err != nil {
		log.Printf("Warning: Invalid JWT_EXPIRATION_MINUTES, using default 15: %v", err)
		jwtExpMinutes = 15
	}

	refreshExpDays, err := strconv.Atoi(getEnv("REFRESH_TOKEN_EXPIRATION_DAYS", "7"))
	if err != nil {
		log.Printf("Warning: Invalid REFRESH_TOKEN_EXPIRATION_DAYS, using default 7: %v", err)
		refreshExpDays = 7
	}

	cfg := &Config{
		AppPort:                getEnv("APP_PORT", "8080"),
		AppEnv:                 getEnv("APP_ENV", "development"),
		LogLevel:               getEnv("LOG_LEVEL", "info"),
		JWTSecret:              getEnv("JWT_SECRET", "default"), // fallback for error handling
		JWTExpiration:          time.Duration(jwtExpMinutes) * time.Minute,
		RefreshTokenExpiration: time.Duration(refreshExpDays) * 24 * time.Hour,
		DBHost:                 getEnv("DB_HOST", "localhost"),
		DBPort:                 getEnv("DB_PORT", "5432"),
		DBUser:                 getEnv("DB_USER", "postgres"),
		DBPassword:             getEnv("DB_PASSWORD", ""), // on linux the default is empty, on others is postgres
		DBName:                 getEnv("DB_NAME", "papertrading"),
		DBSslMode:              getEnv("DB_SSLMODE", "disable"),
	}

	if cfg.JWTSecret == "default" || cfg.JWTSecret == "" {
		log.Fatalf("WARNING: JWT_SECRET is not set or is using the default. This is insecure for production.")
	}

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
