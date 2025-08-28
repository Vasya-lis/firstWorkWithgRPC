package main

import (
	"context"
	"log"
)

// функция чистки ключей задач
func ClearTaskCache(ctx context.Context) {
	iter := Rdb.Scan(ctx, 0, "task:*", 0).Iterator()
	for iter.Next(ctx) {
		if err := Rdb.Del(ctx, iter.Val()).Err(); err != nil {
			log.Printf("failed to delete cache key %s: %v", iter.Val(), err)
		}
	}
	if err := iter.Err(); err != nil {
		log.Printf("redis scan error: %v", err)
	}
}
