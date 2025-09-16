package api

import (
	"context"
	"fmt"
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
		log.Printf("configuration error: %v", err)
		return nil, fmt.Errorf("configuration failed: %w", err)
	}

	// Подключение к gRPC серверу
	conn, err := grpc.NewClient(config.DBServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to gRPC server: %v", err)
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
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
	if err := app.server.Shutdown(app.context); err != nil {
		log.Printf("error when stopping the HTTP server: %v", err)
	}
	app.conn.Close()
}
