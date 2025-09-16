package db

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"

	cmDB "github.com/Vasya-lis/firstWorkWithgRPC/common/db"
	cmR "github.com/Vasya-lis/firstWorkWithgRPC/common/redis"
	cfg "github.com/Vasya-lis/firstWorkWithgRPC/config"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type AppDB struct {
	conf   *cfg.Config     // конфигурация evn
	db     *gorm.DB        // подключение к бд
	redis  *redis.Client   // подключение к кэшу
	mu     sync.RWMutex    // для потокобезопасности
	server *grpc.Server    // сервер
	ctx    context.Context //
}

func NewAppDB() (*AppDB, error) {

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

	// иниц redis
	cmR.InitRedis(config.RedisAddr)

	grpcServer := grpc.NewServer()

	app := &AppDB{
		conf:   config,
		server: grpcServer,
		mu:     sync.RWMutex{},
		ctx:    context.Background(),
		db:     cmDB.GetDB(),
		redis:  cmR.GetRedis(),
	}

	pb.RegisterSchedulerServiceServer(grpcServer, &TaskServer{app: app})

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
	app.redis.Close()
	app.server.GracefulStop()
}
