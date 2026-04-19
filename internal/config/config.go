package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port               string
	DatabaseURL        string
	AuthSecret         string
	TokenTTL           time.Duration
	ReminderWindowDays int
	CORSAllowedOrigins []string
	Storage            StorageConfig
}

type StorageConfig struct {
	Endpoint       string
	AccessKey      string
	SecretKey      string
	Bucket         string
	UseSSL         bool
	PublicBaseURL  string
	MaxUploadBytes int64
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Port:               get("PORT", "8080"),
		DatabaseURL:        get("DATABASE_URL", "postgres://ams:ams@localhost:5432/ams?sslmode=disable"),
		AuthSecret:         get("AUTH_SECRET", "local-dev-secret-change-me"),
		TokenTTL:           durationMinutes("TOKEN_TTL_MINUTES", 24*60),
		ReminderWindowDays: intEnv("REMINDER_WINDOW_DAYS", 30),
		CORSAllowedOrigins: csvEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"),
		Storage: StorageConfig{
			Endpoint:       get("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:      get("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey:      get("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:         get("MINIO_BUCKET", "ams-documents"),
			UseSSL:         boolEnv("MINIO_USE_SSL", false),
			PublicBaseURL:  get("MINIO_PUBLIC_BASE_URL", ""),
			MaxUploadBytes: int64Env("MAX_UPLOAD_BYTES", 20*1024*1024),
			ConnectTimeout: 5 * time.Second,
			RequestTimeout: 60 * time.Second,
		},
	}
	if len(cfg.AuthSecret) < 16 {
		return Config{}, fmt.Errorf("AUTH_SECRET must be at least 16 characters")
	}
	if cfg.ReminderWindowDays < 1 {
		return Config{}, fmt.Errorf("REMINDER_WINDOW_DAYS must be greater than zero")
	}
	return cfg, nil
}

func get(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func intEnv(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func int64Env(key string, fallback int64) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return n
}

func boolEnv(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return n
}

func durationMinutes(key string, fallback int) time.Duration {
	return time.Duration(intEnv(key, fallback)) * time.Minute
}

func csvEnv(key, fallback string) []string {
	raw := get(key, fallback)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
