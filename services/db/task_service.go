package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	apperrors "github.com/Vasya-lis/firstWorkWithgRPC/common/app_errors"
	"github.com/Vasya-lis/firstWorkWithgRPC/services/db/cache"
	"github.com/Vasya-lis/firstWorkWithgRPC/services/db/repo"
	md "github.com/Vasya-lis/firstWorkWithgRPC/services/models"
)

type TasksService struct {
	tr *repo.TasksRepo
	tc *cache.TasksCache // подключение к кэшу
	mu sync.RWMutex
}

func NewTasksService(tr *repo.TasksRepo, tc *cache.TasksCache) *TasksService {
	return &TasksService{
		tr: tr,
		tc: tc,
		mu: sync.RWMutex{},
	}
}

func (s *TasksService) GetTasks(ctx context.Context, limit int, search string) ([]*md.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 1. пробуем из кеша
	tasks, err := s.tc.GetTasksCache(ctx, limit, search)
	if err != nil {
		log.Printf("failed get cache: %v", err)

		// 2. получаем из бд без фильтра
		tasks, err = s.tr.Tasks(-1, "") // логика репозитория
		if err != nil {
			return nil, fmt.Errorf("failed to get tasks from db: %w", err)
		}
		// 3. сохраняем список в кэш если список из бд

		err = s.tc.SetTasksCache(ctx, tasks)
		if err != nil {
			log.Printf("failed set tasks: %v", err)
		}
		// фильтруем
		tasks, err = s.tr.Tasks(limit, search)
		if err != nil {
			return nil, fmt.Errorf("failed to filter tasks with limit= %d search= %s : %w", limit, search, err)
		}

	}
	return tasks, nil
}
func (s *TasksService) GetTask(ctx context.Context, id int) (*md.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// 1. проверяем кэш

	task, err := s.tc.GetTaskCache(ctx, id)
	if err != nil {
		log.Println("failed get cache: ", err)
		// достаем из бд
		task, err = s.GetTask(ctx, id)
		if err != nil {
			if errors.Is(err, apperrors.ErrTaskNotFound) {
				return nil, err
			}
			return nil, fmt.Errorf("failed to get task ")
		}
		// созраняем задачу в кэш
		if err = s.tc.SetTaskCache(ctx, id, task); err != nil {
			log.Printf("failed set task: %v", err)
		}

	}
	return task, nil
}

func (s *TasksService) AddTask(ctx context.Context, task *md.Task) (int, error) {
	id, err := s.tr.AddTask(task)
	if err != nil {
		return 0, err
	}
	// обновляем кэш

	if err := s.tc.SetTaskCache(ctx, id, task); err != nil {
		log.Printf("failed update task in cahe: %v", err)
	}
	return id, nil
}

func (s *TasksService) UpdateTask(ctx context.Context, task *md.Task) error {
	// обновляем в бд
	err := s.tr.UpdateTask(task)
	if err != nil {
		return err
	}
	// обновляем кэш

	if err := s.tc.SetTaskCache(ctx, task.ID, task); err != nil {
		log.Printf("failed update task in cahe: %v", err)
	}
	return nil
}

func (s *TasksService) DeleteTask(ctx context.Context, id int) error {
	// удаляем из базы
	err := s.tr.DeleteTask(id)
	if err != nil {
		log.Println("error: ", err)
		return err
	}
	// удаляем из кэша
	s.tc.DeleteTaskCache(ctx, id)
	return nil
}

func (s *TasksService) UpdateDateTask(ctx context.Context, next string, id int) error {
	err := s.tr.UpdateDate(next, id)
	if err != nil {
		log.Println("error: ", err)
		return err
	}

	// получаем обновленную задачу из бд
	task, err := s.tr.GetTask(id)
	if err != nil {
		log.Println("failed get updated task:", err)
		return err
	}

	// обновляем кэш

	if err := s.tc.SetTaskCache(ctx, task.ID, task); err != nil {
		log.Printf("failed update task in cache: %s", err)
	}

	return nil
}
