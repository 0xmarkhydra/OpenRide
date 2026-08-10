package config

import (
	"os"
	"strings"
)

type Config struct {
	AppEnv           string
	HTTPAddr         string
	DatabaseURL      string
	RedisAddr        string
	RedisPassword    string
	Persistence      string
	JWTSecret        string
	AllowDevIdentity bool
	AdminPhone       string
	AdminEmail       string
	AdminDisplayName string
	MapsProvider     string
	GoogleMapsAPIKey string
	SMSProvider      string
	SMSWebhookURL    string
	SMSAPIKey        string
}

func Load() Config {
	return Config{
		AppEnv:           env("APP_ENV", "development"),
		HTTPAddr:         env("HTTP_ADDR", ":8080"),
		DatabaseURL:      env("DATABASE_URL", "postgres://flashx:flashx@localhost:55432/flashx?sslmode=disable"),
		RedisAddr:        env("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    env("REDIS_PASSWORD", ""),
		Persistence:      env("PERSISTENCE", "postgres"),
		JWTSecret:        env("JWT_SECRET", "dev-change-me-please"),
		AllowDevIdentity: envBool("ALLOW_DEV_IDENTITY", true),
		AdminPhone:       env("ADMIN_PHONE", ""),
		AdminEmail:       env("ADMIN_EMAIL", ""),
		AdminDisplayName: env("ADMIN_DISPLAY_NAME", "FlashX Admin"),
		MapsProvider:     env("MAPS_PROVIDER", "mock"),
		GoogleMapsAPIKey: env("GOOGLE_MAPS_API_KEY", ""),
		SMSProvider:      env("SMS_PROVIDER", "development"),
		SMSWebhookURL:    env("SMS_WEBHOOK_URL", ""),
		SMSAPIKey:        env("SMS_API_KEY", ""),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
