package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	apperrors "github.com/Vasya-lis/firstWorkWithgRPC/common/app_errors"
	"github.com/Vasya-lis/firstWorkWithgRPC/services/models"
	"github.com/redis/go-redis/v9"
)

type TasksCache struct {
	redis *redis.Client
}

func NewTasksCache(redis *redis.Client) *TasksCache {
	return &TasksCache{
		redis: redis,
	}
}

// функция чистки ключей задач
func (s *TasksCache) ClearTaskCache(ctx context.Context) {

	iter := s.redis.Scan(ctx, 0, "task:*", 0).Iterator()
	for iter.Next(ctx) {
		if err := s.redis.Del(ctx, iter.Val()).Err(); err != nil {
			log.Printf("failed to delete cache key %s: %v", iter.Val(), err)
		}
	}
	if err := iter.Err(); err != nil {
		log.Printf("redis scan error: %v", err)
	}
}
func (s *TasksCache) GetTaskCache(ctx context.Context, id int) (*models.Task, error) {

	key := "task:" + fmt.Sprint(id)

	val, err := s.redis.Get(ctx, key).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, apperrors.ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get task from cache: %w", err)
	}

	var task models.Task

	if err := json.Unmarshal([]byte(val), &task); err != nil {
		s.redis.Del(ctx, key)
		log.Printf("deserialization error: %v", err)
		return nil, fmt.Errorf("failed to unmarshl task: %w", err)
	}

	return &task, nil
}

func (s *TasksCache) SetTaskCache(ctx context.Context, id int, task *models.Task) error {

	key := "task:" + fmt.Sprint(id)

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task %s: %v", key, err)
	}
	if err := s.redis.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to set cache for task %s: %w", key, err)
	}
	return nil
}

func (s *TasksCache) GetTasksCache(ctx context.Context, limit int, search string) ([]*models.Task, error) {

	var tasks []*models.Task
	iter := s.redis.Scan(ctx, 0, "task:*", 0).Iterator()

	for iter.Next(ctx) {
		if len(tasks) >= limit && limit > 0 {
			break
		}

		data, err := s.redis.Get(ctx, iter.Val()).Result()
		if err != nil {
			log.Printf("failed get task %s: %v", iter.Val(), err)
			continue
		}

		task := &models.Task{}

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
		return nil, fmt.Errorf("redis scan error: %v", err)
	}
	if len(tasks) == 0 {
		return nil, apperrors.ErrTaskNotFound
	}
	return tasks, nil
}

func isDateSearch(s string) bool {
	_, err := time.Parse("02.01.2006", s)
	return err == nil
}

func (s *TasksCache) SetTasksCache(ctx context.Context, tasks []*models.Task) error {

	for _, task := range tasks {
		tasksData, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("failed to cache task %d: %v", task.ID, err)
		}
		if err := s.redis.Set(ctx, fmt.Sprintf("task:%d", task.ID), tasksData, 0).Err(); err != nil {
			return fmt.Errorf("failed set task: %w", err)
		}

	}
	return nil
}

func (s *TasksCache) DeleteTaskCache(ctx context.Context, id int) {

	key := "task:" + fmt.Sprint(id)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		log.Printf("failed delete task %d from cache: %v", id, err)
	}
}
