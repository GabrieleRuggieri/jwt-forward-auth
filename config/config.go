package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.comcom/joho/godotenv"
)

type Config struct {
	ServerPort          string
	JWKSUrl             string
	AllowedIssuers      []string
	AllowedAudiences    []string
	AllowedAlg          string
	JWKSCacheTTL        time.Duration
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	ttlMinutes, _ := strconv.Atoi(getEnv("JWKS_CACHE_TTL_MINUTES", "60"))

	return &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		JWKSUrl:          getEnvOrFatal("JWKS_URL"),
		AllowedIssuers:   splitAndTrim(getEnvOrFatal("ALLOWED_ISSUERS")),
		AllowedAudiences: splitAndTrim(getEnvOrFatal("ALLOWED_AUDIENCES")),
		AllowedAlg:       getEnv("ALLOWED_ALG", "RS256"),
		JWKSCacheTTL:     time.Duration(ttlMinutes) * time.Minute,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvOrFatal(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		log.Fatalf("FATAL: Environment variable %s is not set.", key)
	}
	return value
}

func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}