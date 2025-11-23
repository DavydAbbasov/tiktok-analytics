package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Server   ServerOpts     `yaml:"server_opts"`
	DB       DBConfig       `yaml:"db"`
	Provider ProviderConfig `yaml:"provider"`
	Earnings EarningsConfig `yaml:"earnings"`
}
type HTTPConfig struct {
	ListenAddr string `yaml:"listen_addr" env:"HTTP_LISTEN_ADDR" env-default:":8080"`
}

type ServerOpts struct {
	ReadTimeoutSeconds  int `yaml:"read_timeout"  env:"HTTP_READ_TIMEOUT"  env-default:"10"`
	WriteTimeoutSeconds int `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-default:"10"`
	IdleTimeoutSeconds  int `yaml:"idle_timeout"  env:"HTTP_IDLE_TIMEOUT"  env-default:"60"`
}

type DBConfig struct {
	Host               string `yaml:"host"      env:"DB_HOST"      env-default:"localhost"`
	Port               int    `yaml:"port"      env:"DB_PORT"      env-default:"5432"`
	Name               string `yaml:"name"      env:"DB_NAME"      env-required:"true"`
	User               string `yaml:"user"      env:"DB_USER"      env-default:"postgres"`
	Password           string `yaml:"password"  env:"DB_PASSWORD"  env-default:"postgres"`
	MaxIdleConns       int    `yaml:"max_idle_conns"  env:"DB_MAX_IDLE_CONNS"  env-default:"5"`
	MaxOpenConns       int    `yaml:"max_open_conns"  env:"DB_MAX_OPEN_CONNS"  env-default:"10"`
	ConnMaxLifetimeMin int    `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME" env-default:"5"` // minutes
	QueryTimeoutSec    int    `yaml:"query_timeout"    env:"DB_QUERY_TIMEOUT"     env-default:"2"`  // seconds
}

type ProviderConfig struct {
	Type  string `yaml:"type"  env:"PROVIDER_TYPE"  env-default:"ensemble"` // ensemble | mock | tiktok
	URL   string `yaml:"url"   env:"PROVIDER_URL"   env-default:"https://ensembledata.com/apis/tiktok"`
	Token string `yaml:"token" env:"PROVIDER_TOKEN" env-required:"true"`
}

type EarningsConfig struct {
	Rate float64 `yaml:"rate" env:"EARNINGS_RATE" env-default:"0.10"` // $0.10
	Per  int64   `yaml:"per"  env:"EARNINGS_PER"  env-default:"1000"` // за 1000 просмотров
}


func ParseConfig() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig("internal/config/config.yaml", &cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return &cfg, nil
}
