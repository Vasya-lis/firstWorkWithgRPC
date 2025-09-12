package api

import (
	"context"
	"log"
	"net/http"

	cfg "github.com/Vasya-lis/firstWorkWithgRPC/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AppAPI struct {
	conf    *cfg.Config      // env
	conn    *grpc.ClientConn // для соединения с db
	server  *http.Server     // вебсервер
	context context.Context
}

func NewAppApi() (*AppAPI, error) {
	config, err := cfg.NewConfig()
	if err != nil {
		log.Println("configuration error: %w", err)
	}

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
		server:  server,
		context: context.Background(),
	}
	app.init()

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
