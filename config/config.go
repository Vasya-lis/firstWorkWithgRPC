package common

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// API сервис
	DBServiceAddress string `envconfig:"DB_SERVICE_ADDRESS" required:"true"`
	TodoPort         string `envconfig:"TODO_PORT" required:"true"`

	// DB сервис
	DBHost     string `envconfig:"DB_HOST" required:"true"`
	DBPort     string `envconfig:"DB_PORT" required:"true"`
	DBUser     string `envconfig:"DB_USER" required:"true"`
	DBPassword string `envconfig:"DB_PASSWORD" required:"true"`
	DBName     string `envconfig:"DB_NAME" required:"true"`
	DBSSLMode  string `envconfig:"DB_SSL_MODE" default:"disable"`

	GRPCPort string `envconfig:"GRPC_PORT" required:"true"`

	RedisAddr string `envconfig:"REDIS_ADDR" required:"true"`
}

func NewConfig() (*Config, error) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return nil, fmt.Errorf("failed to process config: %v", err)
	}
	return &config, nil
}
