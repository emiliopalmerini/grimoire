package app

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr            string        `envconfig:"ADDR" default:":8080"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"5s"`
	GRPCAddr string `envconfig:"GRPC_ADDR" default:":9090"`
	AMQPURL string `envconfig:"AMQP_URL" default:"amqp://guest:guest@localhost:5672/"`
}

func New() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
