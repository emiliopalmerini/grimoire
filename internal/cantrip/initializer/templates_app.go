package initializer

import "fmt"

func appTemplate(opts ProjectOptions) string {
	fields := `	Addr            string        ` + "`envconfig:\"ADDR\" default:\":8080\"`" + `
	ShutdownTimeout time.Duration ` + "`envconfig:\"SHUTDOWN_TIMEOUT\" default:\"5s\"`"

	if hasTransport(opts.Transports, "grpc") {
		fields += `
	GRPCAddr string ` + "`envconfig:\"GRPC_ADDR\" default:\":9090\"`"
	}

	if hasTransport(opts.Transports, "amqp") {
		fields += `
	AMQPURL string ` + "`envconfig:\"AMQP_URL\" default:\"amqp://guest:guest@localhost:5672/\"`"
	}

	return fmt.Sprintf(`package app

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
%s
}

func New() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
`, fields)
}
