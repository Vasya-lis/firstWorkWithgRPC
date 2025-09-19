package db

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"
	"github.com/Vasya-lis/firstWorkWithgRPC/services/db/cache"
	"github.com/Vasya-lis/firstWorkWithgRPC/services/db/repo"

	cmDB "github.com/Vasya-lis/firstWorkWithgRPC/common/db"
	cmR "github.com/Vasya-lis/firstWorkWithgRPC/common/redis"
	cfg "github.com/Vasya-lis/firstWorkWithgRPC/config"
	"google.golang.org/grpc"
)

type AppDB struct {
	conf   *cfg.Config   // конфигурация evn
	tasks  *TasksService // сервис тасков
	server *grpc.Server  // grpc сервер
	ctx    context.Context
}

func NewAppDB() (*AppDB, error) {
	// читаем конфиг
	config, err := cfg.NewConfig()
	if err != nil {
		log.Printf("configuration error: %v", err)
		return nil, fmt.Errorf("configuration failed: %w", err)
	}

	// строка подключения
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort, config.DBSSLMode)

	// инициализация базы
	if err := cmDB.InitDB(dsn); err != nil {
		log.Printf("DB init failed: %v", err)
		return nil, fmt.Errorf("DB init failed: %w", err)
	}
	db := cmDB.GetDB()

	// иниц redis
	cmR.InitRedis(config.RedisAddr)
	client := cmR.GetRedis()

	// создание слоев приложения
	taskRepo := repo.NewTasksRepo(db)                    // работа с бд
	taskCache := cache.NewTasksCache(client)             // кэш
	tasksService := NewTasksService(taskRepo, taskCache) // сервис

	//создание gRPC сервера
	grpcServer := grpc.NewServer()
	tasksServer := NewTasksServer(tasksService)
	pb.RegisterSchedulerServiceServer(grpcServer, tasksServer)

	app := &AppDB{
		conf:   config,
		tasks:  tasksService,
		server: grpcServer,
		ctx:    context.Background(),
	}

	// чистим кэш при старте
	app.tasks.tc.ClearTaskCache(app.ctx)
	return app, nil
}
func (app *AppDB) Start() {
	// запуск gRPC
	lis, err := net.Listen("tcp", ":"+app.conf.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("db-service running on :%s", app.conf.GRPCPort)
	if err := app.server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
func (app *AppDB) Stop() {
	app.server.GracefulStop()
}
