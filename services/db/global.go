package db

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB      // основное соединение с Postgres
	Rdb *redis.Client // соединение с Redis
	MU  sync.RWMutex  // глобальная блокировка для потокобезопасности
	ctx = context.Background()
)
