package config

import (
	"os"

	"github.com/pelletier/go-toml/v2"

	"github.com/shurco/gosign/internal/trust"
	"github.com/shurco/gosign/pkg/storage/postgres"
	"github.com/shurco/gosign/pkg/storage/redis"
	"github.com/shurco/gosign/pkg/utils/fsutil"
)

const (
	ConfigFile = "./gosign.toml"
)

var cfg *Config

// Config is ...
type Config struct {
	HTTPAddr string          `toml:"http-addr" comment:"Ports <= 1024 are privileged ports. You can't use them unless you're root or have the explicit\npermission to use them. See this answer for an explanation or wikipedia or something you trust more.\nsudo setcap 'cap_net_bind_service=+ep' /opt/yourGoBinary"`
	DevMode  bool            `toml:"dev-mode" comment:"Active develop mode"`
	Postgres postgres.Config `toml:"postgres" comment:"Postgres section"`
	Redis    redis.Config    `toml:"redis" comment:"Redis section"`
	Trust    trust.Config    `toml:"trust-certs" comment:"Trust certs section"`
}

// DefaultConfig is ...
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
		Trust: trust.Config{
			List:   []string{"eutl12", "tl12"},
			Update: 1,
		},
	}
}

// Load is ...
func Load() error {
	config := Default()

	if fsutil.IsFile(ConfigFile) {
		file, err := os.ReadFile(ConfigFile)
		if err != nil {
			return err
		}

		if err := toml.Unmarshal(file, &config); err != nil {
			return err
		}
	}

	cfg = config
	return nil
}

// Save is ...
func Save(config *Config) error {
	byteConfig, err := toml.Marshal(config)
	if err != nil {
		return err
	}
	if err := os.WriteFile(ConfigFile, byteConfig, 0o666); err != nil {
		return err
	}

	return nil
}

// Data is ...
func Data() *Config {
	if cfg == nil {
		cfg = Default()
	}
	return cfg
}
