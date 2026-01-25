package config

import (
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
	HTTPAddr string
	DevMode  bool
	Postgres postgres.Config
	Redis    redis.Config
}

// Default returns config with default values (used when env vars are not set).
func Default() *Config {
	return &Config{
		DevMode:  false,
		HTTPAddr: "0.0.0.0:8088",
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

// Load reads configuration from environment variables and sets global config.
func Load() error {
	config := Default()
	config.HTTPAddr = getenv("HTTP_ADDR", config.HTTPAddr)
	config.DevMode = getenvBool("DEV_MODE", config.DevMode)
	config.Postgres.URL = getenv("POSTGRES_URL", config.Postgres.URL)
	config.Redis.Address = getenv("REDIS_ADDRESS", config.Redis.Address)
	config.Redis.Password = getenv("REDIS_PASSWORD", config.Redis.Password)
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
