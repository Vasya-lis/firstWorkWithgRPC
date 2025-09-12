package db

import (
	"log"

	cm "github.com/Vasya-lis/firstWorkWithgRPC/common"
	"github.com/redis/go-redis/v9"
)

func InitRedis() {

	config := cm.NewConfig()

	Rdb = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
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
