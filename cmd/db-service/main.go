package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"
)

func main() {
	dbFile := os.Getenv("DB_FILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	// инициализация базы
	if err := Init(dbFile); err != nil {
		log.Fatalf("DB init failed: %v", err)
	}

	// запуск gRPC
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSchedulerServiceServer(grpcServer, &TaskServer{})

	log.Println("db-service running on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
