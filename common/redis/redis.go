package common

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitRedis(redisAddr string) {

	rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("connection error to redis: %v", err)
	}
	log.Println("Redis is connected")
}
func GetRedis() *redis.Client {
	return rdb
}
