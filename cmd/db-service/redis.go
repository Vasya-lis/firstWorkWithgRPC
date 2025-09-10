package main

import (
	"context"
	"log"
	"sync"

	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
)

var (
	Rdb *redis.Client
	ctx = context.Background()
	MU  sync.RWMutex
)

func InitRedis() {

	var config Config
	err := envconfig.Process("", &config)

	if err != nil {
		log.Fatalf("failed to process envconfig variables: %v", err)
	}

	Rdb = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: "",
		DB:       0,
	})

	_, err = Rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("connection error to redis: %v", err)
	}

	// очищаем кэш
	ClearTaskCache(ctx)

	log.Println("Redis is connected")
}
