package db

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"

	cm "github.com/Vasya-lis/firstWorkWithgRPC/common"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type AppDB struct {
	conf   *cm.Config      // конфигурация evn
	db     *gorm.DB        // подключение к бд
	redis  *redis.Client   // подключение к кэшу
	mu     sync.RWMutex    // для потокобезопасности
	server *grpc.Server    // сервер
	ctx    context.Context //
}

func NewAppDB() (*AppDB, error) {

	config := cm.NewConfig()

	// строка подключения
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort, config.DBSSLMode)

	// инициализация базы
	if err := InitDB(dsn); err != nil {
		log.Fatalf("DB init failed: %v", err)
	}

	// иниц redis
	InitRedis()

	grpcServer := grpc.NewServer()
	pb.RegisterSchedulerServiceServer(grpcServer, &TaskServer{})

	return &AppDB{
		conf:   config,
		server: grpcServer,
		mu:     sync.RWMutex{},
		ctx:    context.Background(),
		db:     GetDB(),
		redis:  Rdb,
	}, nil
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
	app.redis.Close()
	app.server.GracefulStop()
}
