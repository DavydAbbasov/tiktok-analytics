package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ListenAddr  string         `yaml:"listen_addr" env:"HTTP_PORT" env-required:"true"`
	Server      ServerOpts     `yaml:"server_opts"`
	SQLDataBase SQLDataBase    `yaml:"sql_database"`
	Provider    ProviderConfig `yaml:"provider"`
	Earnings    EarningsConfig `yaml:"earnings"`
}

type ServerOpts struct {
	ReadTimeoutSeconds  int `yaml:"read_timeout"  env:"HTTP_READ_TIMEOUT"  env-default:"10"`
	WriteTimeoutSeconds int `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-default:"10"`
	IdleTimeoutSeconds  int `yaml:"idle_timeout"  env:"HTTP_IDLE_TIMEOUT"  env-default:"60"`
}

type SQLDataBase struct {
	Server             string `yaml:"server"      env:"DB_HOST"      env-default:"localhost"`
	Database           string `yaml:"database"      env:"DB_PORT"      env-default:"5432"`
	Username           string `yaml:"username"      env:"DB_NAME"      env-required:"true"`
	Password           string `yaml:"password"  env:"DB_PASSWORD"  env-default:"postgres"`
	Port               int    `yaml:"port" env:"DB_PORT"`
	MaxIdleConns       int32  `yaml:"max_idle_conns"  env:"DB_MAX_IDLE_CONNS"  env-default:"5"`
	MaxOpenConns       int32  `yaml:"max_open_conns"  env:"DB_MAX_OPEN_CONNS"  env-default:"10"`
	ConnMaxLifetimeMin int    `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME" env-default:"5"` // minutes
	QueryTimeoutSec    int    `yaml:"query_timeout"    env:"DB_QUERY_TIMEOUT"     env-default:"2"`  // seconds
}

type ProviderConfig struct {
	Type  string `yaml:"type"  env:"PROVIDER_TYPE"  env-default:"ensemble"`
	URL   string `yaml:"url"   env:"PROVIDER_URL"   env-default:"https://ensembledata.com/apis"`
	Token string `yaml:"token" env:"PROVIDER_TOKEN" env-required:"true"`
}

type EarningsConfig struct {
	Rate float64 `yaml:"rate" env:"EARNINGS_RATE" env-default:"0.10"`
	Per  int64   `yaml:"per"  env:"EARNINGS_PER"  env-default:"1000"`
}

func ParseConfig() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig("internal/config/config.yaml", &cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return &cfg, nil
}
