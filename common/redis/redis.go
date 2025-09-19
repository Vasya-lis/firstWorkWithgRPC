package common

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func InitRedis(redisAddr string) {

	Rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	_, err := Rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("connection error to redis: %v", err)
	}
	log.Println("Redis is connected")
}
func GetRedis() *redis.Client {
	return Rdb
}
