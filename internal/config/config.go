package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/shurco/gosign/pkg/storage/postgres"
	"github.com/shurco/gosign/pkg/storage/redis"
)

const envPrefix = "GOSIGN_"

var cfg *Config

// Config is the application configuration (infrastructure only; app settings are in DB).
// All values are read from environment variables with GOSIGN_ prefix.
type Config struct {
	HTTPAddr           string
	DevMode            bool
	JWTSecret          string
	CORSAllowedOrigins []string
	Postgres           postgres.Config
	Redis              redis.Config
}

// Default returns config with default values (used when env vars are not set).
func Default() *Config {
	return &Config{
		DevMode:   false,
		HTTPAddr:  "0.0.0.0:8088",
		JWTSecret: "",
		Postgres: postgres.Config{
			URL: "postgres://goSign:postgresPassword@localhost:5432/goSign?pool_max_conns=10",
		},
		Redis: redis.Config{
			Address:  "localhost:6379",
			Password: "redisPassword",
		},
	}
}

func getenv(key, defaultVal string) string {
	if v := os.Getenv(envPrefix + key); v != "" {
		return v
	}
	return defaultVal
}

func getenvBool(key string, defaultVal bool) bool {
	v := os.Getenv(envPrefix + key)
	if v == "" {
		return defaultVal
	}
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "1" || v == "true" || v == "yes"
}

func splitCommaNonEmpty(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// defaultDevCORSOrigins covers typical local frontends (Vite, Next, etc.).
var defaultDevCORSOrigins = []string{
	"http://localhost:3000",
	"http://127.0.0.1:3000",
	"http://localhost:5173",
	"http://127.0.0.1:5173",
	"http://localhost:4173",
	"http://127.0.0.1:4173",
	"http://localhost:8080",
	"http://127.0.0.1:8080",
}

// Load reads configuration from environment variables and sets global config.
func Load() error {
	config := Default()
	config.HTTPAddr = getenv("HTTP_ADDR", config.HTTPAddr)
	config.DevMode = getenvBool("DEV_MODE", config.DevMode)
	config.Postgres.URL = getenv("POSTGRES_URL", config.Postgres.URL)
	config.Redis.Address = getenv("REDIS_ADDRESS", config.Redis.Address)
	config.Redis.Password = getenv("REDIS_PASSWORD", config.Redis.Password)
	config.JWTSecret = getenv("JWT_SECRET", config.JWTSecret)
	if config.JWTSecret == "" {
		return fmt.Errorf("GOSIGN_JWT_SECRET environment variable is required")
	}
	if raw := getenv("CORS_ALLOWED_ORIGINS", ""); raw != "" {
		config.CORSAllowedOrigins = splitCommaNonEmpty(raw)
	} else if config.DevMode {
		config.CORSAllowedOrigins = append([]string(nil), defaultDevCORSOrigins...)
	}
	cfg = config
	return nil
}

// Data returns the loaded config (or default if Load was not called).
func Data() *Config {
	if cfg == nil {
		cfg = Default()
	}
	return cfg
}
