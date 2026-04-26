package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort     string
	AppEnv      string
	DatabaseURL string
	RedisURL    string
	RedisPass   string
	RedisDB     string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	dbURL := getEnv("DATABASE_URL", "")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL is required but not set")
	}

	return &Config{
		AppPort:     getEnv("APP_PORT", "8080"),
		AppEnv:      getEnv("APP_ENV", "development"),
		DatabaseURL: dbURL,
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
		RedisPass:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:     getEnv("REDIS_DB", "0"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
