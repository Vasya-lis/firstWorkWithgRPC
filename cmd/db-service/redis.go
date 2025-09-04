package main

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	Rdb *redis.Client
	ctx = context.Background()
)

func InitRedis() {

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "my-redis:6379"
	}

	Rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	_, err := Rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("connection error to redis: %v", err)
	}

	// очищаем кэш
	ClearTaskCache(ctx)

	log.Println("Redis is connected")
}
