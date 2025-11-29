package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ListenAddr  string         `yaml:"listen_addr" env:"HTTP_LISTEN_ADDR" env-required:"true"`
	ServerOpts  ServerOpts     `yaml:"server_opts"`
	SQLDataBase SQLDataBase    `yaml:"sql_database"`
	Provider    ProviderConfig `yaml:"provider"`
	Earnings    EarningsConfig `yaml:"earnings"`
	Updater     UpdaterConfig  `yaml:"updater"`
}

type ServerOpts struct {
	ReadTimeout  int `yaml:"read_timeout"  env:"HTTP_READ_TIMEOUT"  env-default:"10"`
	WriteTimeout int `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-default:"10"`
	IdleTimeout  int `yaml:"idle_timeout"  env:"HTTP_IDLE_TIMEOUT"  env-default:"60"`
}

type SQLDataBase struct {
	Server             string `yaml:"server"            env:"DB_HOST"               env-required:"true"`
	Database           string `yaml:"database"          env:"DB_NAME"               env-required:"true"`
	Username           string `yaml:"username"          env:"DB_USER"`
	Password           string `yaml:"password"          env:"DB_PASSWORD"`
	Port               int    `yaml:"port"              env:"DB_PORT"`
	MaxIdleConns       int32  `yaml:"max_idle_conns"    env:"DB_MAX_IDLE_CONNS"     env-default:"5"`
	MaxOpenConns       int32  `yaml:"max_open_conns"    env:"DB_MAX_OPEN_CONNS"     env-default:"10"`
	ConnMaxLifetimeMin int    `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME"  env-default:"5"` // minutes
	QueryTimeoutSec    int    `yaml:"query_timeout"     env:"DB_QUERY_TIMEOUT"      env-default:"2"` // seconds
}

type ProviderConfig struct {
	Type  string `yaml:"type"  env:"PROVIDER_TYPE"`
	URL   string `yaml:"url"   env:"PROVIDER_URL"`
	Token string `yaml:"token" env:"PROVIDER_TOKEN"`
}

type EarningsConfig struct {
	Rate float64 `yaml:"rate"`
	Per  int64   `yaml:"per"`
}
type UpdaterConfig struct {
	Interval     int `yaml:"interval"`
	BatchSize    int `yaml:"batch_size"`
	MinUpdateAge int `yaml:"min_update_age"`
}

const (
	envConfigPath     = "CONFIG_PATH"
	defaultConfigPath = "internal/config/config.yaml"
)

func ParseConfig() (*Config, error) {
	var cfg Config

	path := os.Getenv(envConfigPath) //prod
	if path == "" {
		path = defaultConfigPath // local
	}

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return &cfg, nil
}
