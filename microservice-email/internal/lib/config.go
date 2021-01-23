package lib

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Config is all you need to configure email service.
type Config struct {
	// SMTP is all you need to configure your email server.
	SMTP struct {
		Host     string
		Port     int
		User     string
		Password string
	}
	// RabbitMQ is all you need to configure your rabbit mq server.
	RabbitMQ struct {
		Host         string
		User         string
		Password     string
		QueueName    string `yaml:"queue_name"`
		ExchangeName string `yaml:"exchange_name"`
		ExchangeKind string `yaml:"exchange_kind"`
		Declare      bool
	}
}

// Conf is holder of config load file
var Conf *Config
// ConfigFilePath is path of config file
var ConfigFilePath string

// ReadConfig read config file and unmarshal it to Config model
func ReadConfig() {
	fileData, err := ioutil.ReadFile(ConfigFilePath)
	if err != nil {
		panic(err)
	}

	Conf = new(Config)

	err = yaml.Unmarshal(fileData, Conf)
	if err != nil {
		panic(err)
	}
}
