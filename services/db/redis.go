package db

import (
	"log"

	"github.com/redis/go-redis/v9"
)

func InitRedis(redisAddr string) {

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
