package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/shurco/gosign/pkg/storage/postgres"
	"github.com/shurco/gosign/pkg/storage/redis"
)

const envPrefix = "GOSIGN_"

var (
	cfg   *Config
	cfgMu sync.Mutex
)

// Config is the application configuration (infrastructure only; app settings are in DB).
// All values are read from environment variables with GOSIGN_ prefix.
type Config struct {
	HTTPAddr string
	// WorkerGRPCAddr is the internal gRPC endpoint of the worker: the worker
	// listens on it, the server dials it. It must never be exposed publicly.
	// The server side also accepts a comma-separated list of addresses or a
	// dns:/// target for load balancing across worker replicas.
	WorkerGRPCAddr string
	// WorkerGRPCToken authenticates server->worker gRPC calls (bearer token).
	// Defaults to JWTSecret so auth is always on.
	WorkerGRPCToken    string
	DevMode            bool
	JWTSecret          string
	CORSAllowedOrigins []string
	// AppURL is the public URL of the web app (e.g. https://app.example.com).
	// Used for links that leave the API origin: signing links in emails,
	// OAuth redirects, the /embed iframe. Empty means same-origin relative links.
	AppURL   string
	Postgres postgres.Config
	Redis    redis.Config
}

// AppLink joins the configured app URL with a path.
// With empty AppURL the path is returned as-is (same-origin deployment).
func (c *Config) AppLink(path string) string {
	return c.AppURL + path
}

// Default returns config with default values (used when env vars are not set).
func Default() *Config {
	return &Config{
		DevMode:        false,
		HTTPAddr:       "0.0.0.0:8088",
		WorkerGRPCAddr: "127.0.0.1:8089",
		JWTSecret:      "",
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
	config.WorkerGRPCAddr = getenv("WORKER_GRPC_ADDR", config.WorkerGRPCAddr)
	config.DevMode = getenvBool("DEV_MODE", config.DevMode)
	config.Postgres.URL = getenv("POSTGRES_URL", config.Postgres.URL)
	config.Redis.Address = getenv("REDIS_ADDRESS", config.Redis.Address)
	config.Redis.Password = getenv("REDIS_PASSWORD", config.Redis.Password)
	config.JWTSecret = getenv("JWT_SECRET", config.JWTSecret)
	if config.JWTSecret == "" {
		return fmt.Errorf("GOSIGN_JWT_SECRET environment variable is required")
	}
	config.WorkerGRPCToken = getenv("WORKER_GRPC_TOKEN", config.JWTSecret)
	config.AppURL = strings.TrimRight(getenv("APP_URL", ""), "/")
	if raw := getenv("CORS_ALLOWED_ORIGINS", ""); raw != "" {
		config.CORSAllowedOrigins = splitCommaNonEmpty(raw)
	} else if config.DevMode {
		config.CORSAllowedOrigins = append([]string(nil), defaultDevCORSOrigins...)
	}

	cfgMu.Lock()
	cfg = config
	cfgMu.Unlock()
	return nil
}

// Data returns the loaded config (or default if Load was not called).
func Data() *Config {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	if cfg == nil {
		cfg = Default()
	}
	return cfg
}
