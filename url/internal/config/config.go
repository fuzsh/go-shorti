package config

import (
	"io/ioutil"
	"url/pkg/log"

	"github.com/qiangxue/go-env"
	"gopkg.in/yaml.v2"
)

const (
	defaultServerPort = 8080
)

// Cfg is holder of config load file
var Cfg *Config

// Config represents an application configuration.
type Config struct {
	ServerPort int `yaml:"server_port" env:"SERVER_PORT"`

	Options struct {
		Schema  string `yaml:"schema" env:"SCHEMA"`
		Prefix  string `yaml:"prefix" env:"PREFIX"`
		BaseURL string `yaml:"base_url" env:"BASE_URL"`
	} `yaml:"options"`

	Redis struct {
		Host     string `yaml:"host" env:"REDIS_HOST"`
		Port     string `yaml:"port" env:"REDIS_PORT"`
		Password string `yaml:"password" env:"REDIS_PASSWORD,secret"`
	} `yaml:"redis"`

	Postgres struct {
		Host     string `yaml:"host" env:"POSTGRES_HOST"`
		Port     int    `yaml:"port" env:"POSTGRES_PORT"`
		Password string `yaml:"password" env:"POSTGRES_PASSWORD"`
		User     string `yaml:"user" env:"POSTGRES_USER"`
		DBName   string `yaml:"db_name" env:"POSTGRES_DB_NAME"`
	} `yaml:"postgres"`

	JwtRSAKeys struct{
		Access string `yaml:"access" env:"JWT_RSA_KEYS_ACCESS_KEY"`
		Refresh string `yaml:"refresh" env:"JWT_RSA_KEYS_REFRESH_KEY"`
	} `yaml:"jwt_rsa_keys"`
}

// Load returns an application configuration which is populated from the given configuration file and environment variables.
func Load(file string, logger log.Logger) (*Config, error) {
	// default config
	c := Config{
		ServerPort: defaultServerPort,
	}

	// load from YAML config file
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(bytes, &c); err != nil {
		return nil, err
	}

	// load from environment variables prefixed with "APP_"
	if err = env.New("APP_", logger.Infof).Load(&c); err != nil {
		return nil, err
	}

	return &c, err
}
