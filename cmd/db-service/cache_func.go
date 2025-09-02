package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrNotFound = errors.New("not found")

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
func GetTaskCache(ctx context.Context, id string) (*Task, error) {

	key := "task:" + id

	val, err := Rdb.Get(ctx, key).Result()

	if err != nil {
		if err == redis.Nil {
			return nil, ErrNotFound
		}
		log.Printf("failed get task with key %s: %v", key, err)
		return nil, err
	}

	var task Task

	if err := json.Unmarshal([]byte(val), &task); err != nil {
		Rdb.Del(ctx, key)
		log.Printf("deserialization error: %v", err)
		return nil, err
	}

	return &task, nil
}
func SetTaskCashe(ctx context.Context, id string, task *Task) error {

	key := "task:" + id

	data, err := json.Marshal(task)
	if err != nil {
		log.Printf("failed to marshal task %s: %v", key, err)
		return err
	}
	if err := Rdb.Set(ctx, key, data, 0).Err(); err != nil {
		log.Printf("failed to set cache for task %s: %v", key, err)
		return err
	}
	return nil
}

func GetTasksCache(ctx context.Context, limit int, search string) ([]*Task, error) {

	var tasks []*Task
	iter := Rdb.Scan(ctx, 0, "task:*", 0).Iterator()

	for iter.Next(ctx) {
		if len(tasks) >= limit && limit > 0 {
			break
		}

		data, err := Rdb.Get(ctx, iter.Val()).Result()
		if err != nil {
			log.Printf("failed get task %s: %v", iter.Val(), err)
			continue
		}

		task := &Task{}

		if err := json.Unmarshal([]byte(data), task); err != nil {
			log.Printf("deserialization error: %v", err)
			continue
		}

		switch {
		case search == "":
			tasks = append(tasks, task)
		case isDateSearch(search):
			searchTime, err := time.Parse("02.01.2006", search)
			if err != nil {
				log.Printf("invalid search date: %v", err)
				continue
			}
			taskTime, err := time.Parse("20060102", task.Date)
			if err != nil {
				log.Printf("invalid task date: %v", err)
				continue
			}

			if taskTime.Equal(searchTime) {
				tasks = append(tasks, task)
			}

		default:
			if strings.Contains(task.Title, search) || strings.Contains(task.Comment, search) {
				tasks = append(tasks, task)
			}
		}

	}
	if err := iter.Err(); err != nil {
		log.Printf("redis scan error: %v", err)
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, ErrNotFound
	}
	return tasks, nil
}

func SetTasksCashe(ctx context.Context, tasks []*Task) error {
	for _, task := range tasks {
		tasksData, err := json.Marshal(task)
		if err != nil {

			log.Printf("failed to cache task %s: %v", task.ID, err)
			return err
		}
		if err := Rdb.Set(ctx, fmt.Sprintf("task:%s", task.ID), tasksData, 0).Err(); err != nil {
			log.Println(" failed set task", err)
			return err
		}

	}
	return nil
}

func DeleteTaskCache(ctx context.Context, id string) {
	key := "task:" + id
	if err := Rdb.Del(ctx, key).Err(); err != nil {
		log.Printf("failed delete task %s from cache: %v", id, err)
	}
}
