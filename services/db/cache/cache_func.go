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

	key := fmt.Sprintf("task:%d", id)

	val, err := s.redis.Get(ctx, key).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%w: task id=%d not found in cache", apperrors.ErrTaskNotFound, id)
		}
		return nil, fmt.Errorf("%w: %w: task id=%d, key=%s", apperrors.ErrGetTaskCache, err, id, key)
	}

	var task models.Task

	if err := json.Unmarshal([]byte(val), &task); err != nil {
		s.redis.Del(ctx, key)
		return nil, fmt.Errorf("%w: %w: task id=%d, key=%s", apperrors.ErrGetTaskCache, err, id, key)
	}

	return &task, nil
}

func (s *TasksCache) SetTaskCache(ctx context.Context, id int, task *models.Task) error {

	key := fmt.Sprintf("task:%d", id)

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("%w: %w: task id=%d, key=%s", apperrors.ErrSetTaskCache, err, id, key)
	}
	if err := s.redis.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("%w: %w: task id=%d, key=%s", apperrors.ErrSetTaskCache, err, id, key)
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
				log.Printf("invalid task date: task id=%d: %v", task.ID, err)
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
		return nil, fmt.Errorf("%w: %w: limit=%d, search=%s", apperrors.ErrGetTasksCache, err, limit, search)
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("%w: no tasks found with limit=%d, search=%s", apperrors.ErrTaskNotFound, limit, search)
	}
	return tasks, nil
}

func isDateSearch(s string) bool {
	_, err := time.Parse("02.01.2006", s)
	return err == nil
}

func (s *TasksCache) SetTasksCache(ctx context.Context, tasks []*models.Task) error {

	for _, task := range tasks {
		key := fmt.Sprintf("task:%d", task.ID)
		tasksData, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("%w: %w: task id=%d, key=%s", apperrors.ErrSetTasksCache, err, task.ID, key)
		}
		if err := s.redis.Set(ctx, key, tasksData, 0).Err(); err != nil {
			return fmt.Errorf("%w: %w: task id=%d, key=%s", apperrors.ErrSetTasksCache, err, task.ID, key)
		}

	}
	return nil
}

func (s *TasksCache) DeleteTaskCache(ctx context.Context, id int) {

	key := fmt.Sprintf("task:%d", id)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		log.Printf("%v: task id=%d, key=%s: %v", apperrors.ErrDeleteTaskCache, id, key, err)
	}
}
