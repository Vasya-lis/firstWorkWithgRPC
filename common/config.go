package common

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// API сервис
	DBServiceAddress string `envconfig:"DB_SERVICE_ADDRESS" default:"db-service:50051"`
	TodoPort         string `envconfig:"TODO_PORT" default:"7540"`

	// DB сервис
	DBHost     string `envconfig:"DB_HOST" default:"postgres"`
	DBPort     string `envconfig:"DB_PORT" default:"5432"`
	DBUser     string `envconfig:"DB_USER" default:"postgres"`
	DBPassword string `envconfig:"DB_PASSWORD" default:"12345"`
	DBName     string `envconfig:"DB_NAME" default:"schedulerdb"`
	DBSSLMode  string `envconfig:"DB_SSL_MODE" default:"disable"`

	GRPCPort string `envconfig:"GRPC_PORT" default:"50051"`

	RedisAddr string `envconfig:"REDIS_ADDR" default:"my-redis:6379"`
}

func NewConfig() *Config {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		log.Fatalf("failed to process config: %v", err)
	}
	return &config
}
