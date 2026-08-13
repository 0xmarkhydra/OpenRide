package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv                string
	HTTPAddr              string
	DatabaseURL           string
	DatabaseName          string
	RedisAddr             string
	RedisPassword         string
	RedisDB               int
	Persistence           string
	JWTSecret             string
	AllowDevIdentity      bool
	AdminPhone            string
	AdminEmail            string
	AdminDisplayName      string
	MapsProvider          string
	GoogleMapsAPIKey      string
	SMSProvider           string
	SMSWebhookURL         string
	SMSAPIKey             string
	ObjectStorageProvider string
	S3Endpoint            string
	S3Region              string
	S3Bucket              string
	S3AccessKeyID         string
	S3SecretAccessKey     string
	S3ForcePathStyle      bool
	S3PresignTTLSeconds   int
}

func Load() Config {
	return Config{
		AppEnv:                env("APP_ENV", "development"),
		HTTPAddr:              httpAddr(),
		DatabaseURL:           env("DATABASE_URL", "postgres://flashx:flashx@localhost:55432/flashx?sslmode=disable"),
		DatabaseName:          env("DATABASE_NAME", ""),
		RedisAddr:             env("REDIS_ADDR", "localhost:6379"),
		RedisPassword:         env("REDIS_PASSWORD", ""),
		RedisDB:               envNonNegativeInt("REDIS_DB", 0),
		Persistence:           env("PERSISTENCE", "postgres"),
		JWTSecret:             env("JWT_SECRET", "dev-change-me-please"),
		AllowDevIdentity:      envBool("ALLOW_DEV_IDENTITY", true),
		AdminPhone:            env("ADMIN_PHONE", ""),
		AdminEmail:            env("ADMIN_EMAIL", ""),
		AdminDisplayName:      env("ADMIN_DISPLAY_NAME", "FlashX Admin"),
		MapsProvider:          env("MAPS_PROVIDER", "mock"),
		GoogleMapsAPIKey:      env("GOOGLE_MAPS_API_KEY", ""),
		SMSProvider:           env("SMS_PROVIDER", "development"),
		SMSWebhookURL:         env("SMS_WEBHOOK_URL", ""),
		SMSAPIKey:             env("SMS_API_KEY", ""),
		ObjectStorageProvider: env("OBJECT_STORAGE_PROVIDER", "disabled"),
		S3Endpoint:            env("S3_ENDPOINT", ""),
		S3Region:              env("S3_REGION", "us-east-1"),
		S3Bucket:              env("S3_BUCKET", ""),
		S3AccessKeyID:         env("S3_ACCESS_KEY_ID", ""),
		S3SecretAccessKey:     env("S3_SECRET_ACCESS_KEY", ""),
		S3ForcePathStyle:      envBool("S3_FORCE_PATH_STYLE", true),
		S3PresignTTLSeconds:   envInt("S3_PRESIGN_TTL_SECONDS", 600),
	}
}

func httpAddr() string {
	if value := os.Getenv("HTTP_ADDR"); value != "" {
		return value
	}
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		return ":" + port
	}
	return ":8080"
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

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envNonNegativeInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}
