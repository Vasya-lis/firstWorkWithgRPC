package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DBHost     string `envconfig:"DB_HOST" default:"postgres"`
	DBPort     string `envconfig:"DB_PORT" default:"5432"`
	DBUser     string `envconfig:"DB_USER" default:"postgres"`
	DBPassword string `envconfig:"DB_PASSWORD" default:"12345"`
	DBName     string `envconfig:"DB_NAME" default:"schedulerdb"`
	DBSSLMode  string `envconfig:"DB_SSL_MODE" default:"disable"`

	GRPCPort string `envconfig:"GRPC_PORT" default:"50051"`
	GRPCHost string `envconfig:"GRPC_HOST" default:""`

	RedisAddr string `envconfig:"REDIS_ADDR" default:"my-redis:6379"`
}

func main() {

	InitRedis()

	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatalf("failed to process envconfig variables: %v", err)
	}
	// строка подключения
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort, config.DBSSLMode)

	// инициализация базы
	if err := Init(dsn); err != nil {
		log.Fatalf("DB init failed: %v", err)
	}

	// запуск gRPC
	lis, err := net.Listen("tcp", ":"+config.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSchedulerServiceServer(grpcServer, &TaskServer{})

	log.Printf("db-service running on :%s", config.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
