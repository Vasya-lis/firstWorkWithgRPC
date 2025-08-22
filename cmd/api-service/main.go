package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ctx = context.Background()
var conn *grpc.ClientConn

func main() {
	var err error

	// читаем адрес gRPC из env или ставим дефолт
	grpcAddr := os.Getenv("DB_SERVICE_HOST")
	if grpcAddr == "" {
		grpcAddr = "db-service:50051"
	}

	// Подключение к gRPC серверу
	conn, err = grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	// Запуск HTTP сервера
	// Инициализация порта из переменной окружения или настроек
	Port := os.Getenv("TODO_PORT")
	if Port == "" {
		Port = "7540"
	}

	// инициализация API обработчика
	Init()

	http.Handle("/", http.FileServer(http.Dir("./web")))

	log.Printf("API service running on :%s", Port)
	err = http.ListenAndServe(":"+Port, nil)
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
