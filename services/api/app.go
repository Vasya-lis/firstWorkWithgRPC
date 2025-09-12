package api

import (
	"context"
	"log"
	"net/http"

	cm "github.com/Vasya-lis/firstWorkWithgRPC/common"
	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AppAPI struct {
	conf    *cm.Config       // env
	conn    *grpc.ClientConn // для соединения с db
	client  pb.SchedulerServiceClient
	server  *http.Server // вебсервер
	context context.Context
}

func NewAppApi() (*AppAPI, error) {
	config := cm.NewConfig()

	// Подключение к gRPC серверу
	conn, err := grpc.NewClient(config.DBServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
		return nil, err
	}

	server := &http.Server{
		Addr: ":" + config.TodoPort,
	}

	app := &AppAPI{
		conf:    config,
		conn:    conn,
		client:  pb.NewSchedulerServiceClient(conn),
		server:  server,
		context: context.Background(),
	}
	app.Init()

	return app, nil

}

func (app *AppAPI) Start() {
	log.Printf("API service running on :%s", app.conf.TodoPort)
	if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}

}

func (app *AppAPI) Stop() {
	app.conn.Close()
}
