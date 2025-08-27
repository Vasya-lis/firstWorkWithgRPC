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

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	// запуск gRPC
	lis, err := net.Listen("tcp", net.JoinHostPort("", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSchedulerServiceServer(grpcServer, &TaskServer{})

	log.Printf("db-service running on :%s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
