package config

import (
	"os"

	"github.com/pelletier/go-toml/v2"

	"github.com/shurco/gosign/internal/trust"
	"github.com/shurco/gosign/pkg/storage/postgres"
	"github.com/shurco/gosign/pkg/storage/redis"
	"github.com/shurco/gosign/pkg/utils"
)

const (
	ConfigFile = "./gosign.toml"
)

var cfg *Config

// Settings represents additional application settings
type Settings struct {
	Email      map[string]string `toml:"email" comment:"Email settings (provider, smtp_host, smtp_port, smtp_user, smtp_pass, from_email, from_name)"`
	Storage    map[string]string `toml:"storage" comment:"Storage settings (provider: local/s3/gcs/azure, bucket, region, endpoint, base_path)"`
	Webhook    map[string]string `toml:"webhook" comment:"Webhook settings (enabled, max_retries, timeout)"`
	Geolocation map[string]string `toml:"geolocation" comment:"Geolocation settings (db_path: path to GeoLite2-City.mmdb file, base_dir: directory for database storage, default ./base)"`
	Features   map[string]bool   `toml:"features" comment:"Feature flags (reminders_enabled, sms_verification, embedded_signing, bulk_operations)"`
}

// Config is ...
type Config struct {
	HTTPAddr string          `toml:"http-addr" comment:"Ports <= 1024 are privileged ports. You can't use them unless you're root or have the explicit\npermission to use them. See this answer for an explanation or wikipedia or something you trust more.\nsudo setcap 'cap_net_bind_service=+ep' /opt/yourGoBinary"`
	DevMode  bool            `toml:"dev-mode" comment:"Active develop mode"`
	Postgres postgres.Config `toml:"postgres" comment:"Postgres section"`
	Redis    redis.Config    `toml:"redis" comment:"Redis section"`
	Trust    trust.Config    `toml:"trust-certs" comment:"Trust certs section"`
	Settings Settings        `toml:"settings" comment:"Additional settings section"`
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
		Settings: Settings{
			Email: map[string]string{
				"provider":   "smtp",
				"smtp_host":  "localhost",
				"smtp_port":  "1025", 
				"from_email": "noreply@gosign.local",
				"from_name":  "goSign",
			},
			Storage: map[string]string{
				"provider":  "local",
				"base_path": "./uploads",
			},
			Webhook: map[string]string{
				"enabled":     "true",
				"max_retries": "3",
				"timeout":     "30", 
			},
			Geolocation: map[string]string{
				"base_dir": "./base",
				"db_path":  "./base/GeoLite2-City.mmdb",
				// Note: maxmind_license_key is now stored in account.settings, not in config file
			},
			Features: map[string]bool{
				"reminders_enabled":   true,
				"sms_verification":    false,
				"embedded_signing":    false,
				"bulk_operations":     false,
			},
		},
	}
}

// Load is ...
func Load() error {
	config := Default()

	if utils.IsFile(ConfigFile) {
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
