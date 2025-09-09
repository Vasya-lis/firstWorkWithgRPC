package main

import (
	"context"
	"log"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ctx = context.Background()
var conn *grpc.ClientConn

type Config struct {
	DBserviseAddress string `envconfig:"DB_SERVICE_ADDRESS" default:"db-service:50051"`
	TodoPort         string `envconfig:"TODO_PORT" default:"7540"`
}

func main() {
	var err error

	var config Config
	err = envconfig.Process("", &config)
	if err != nil {
		log.Fatalf("failed to process enviroment variables: %v", err)
	}

	// Подключение к gRPC серверу
	conn, err = grpc.NewClient(config.DBserviseAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	// инициализация API обработчика
	Init()

	http.Handle("/", http.FileServer(http.Dir("./web")))

	log.Printf("API service running on :%s", config.TodoPort)
	err = http.ListenAndServe(":"+config.TodoPort, nil)
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
