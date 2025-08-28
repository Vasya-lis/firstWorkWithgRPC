package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	cm "github.com/Vasya-lis/firstWorkWithgRPC/cmd/common"
	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"
	"github.com/redis/go-redis/v9"
)

type TaskServer struct {
	pb.UnimplementedSchedulerServiceServer
}

// ListTasks возвращает список задач с лимитом и поиском
func (s *TaskServer) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	//ключ для Redis
	cacheKey := fmt.Sprintf("task:limit=%d:search=%s", req.Limit, req.Search)

	// 1. Проверяем в кеше

	val, err := Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var cached pb.ListTasksResponse
		if err := json.Unmarshal([]byte(val), &cached); err == nil {
			log.Printf("cashe hit for %s", cacheKey)
			return &cached, nil
		} else {
			log.Printf("cache hit for %s", cacheKey)
		}
	} else if err != redis.Nil {
		log.Printf("redis get error for %s: %v", cacheKey, err)
	}

	// 2. Если нет в кеше идем в бд

	tasks, err := Tasks(int(req.Limit), req.Search)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks from database: %v", err)
	}

	// конвертируем задачи в пр.буф формат

	var pbTasks []*pb.Task
	for _, t := range tasks {
		pbTasks = append(pbTasks, &pb.Task{
			Id:      t.ID,
			Date:    t.Date,
			Title:   t.Title,
			Comment: t.Comment,
			Repeat:  t.Repeat,
		})
	}

	resp := &pb.ListTasksResponse{Tasks: pbTasks}

	// 3. Сохраняем результат в Redis

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("json marshal error for %s: %v", cacheKey, err)

	} else {
		if err = Rdb.Set(ctx, cacheKey, data, 0).Err(); err != nil {
			log.Printf("failed to set cache for %s: %v", cacheKey, err)
		} else {
			log.Printf("cached response for %s", cacheKey)
		}
	}

	return resp, nil
}

// GetTask возвращает задачу по ID
func (s *TaskServer) GetTask(ctx context.Context, req *pb.IDRequest) (*pb.GetTaskResponse, error) {

	cacheKey := "task:" + req.Id

	// 1. проверяем
	val, err := Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var cached pb.GetTaskResponse
		if err := json.Unmarshal([]byte(val), &cached); err == nil {
			log.Printf("cashe hit for %s", cacheKey)
			return &cached, nil
		} else {
			log.Printf("cache hit for %s", cacheKey)
		}
	} else if err != redis.Nil {
		log.Printf("redis get error for %s: %v", cacheKey, err)
	}

	// 2. берем из бд

	task, err := GetTask(req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get ID: %v", err)
	}

	resp := &pb.GetTaskResponse{
		Task: &pb.Task{
			Id:      task.ID,
			Date:    task.Date,
			Title:   task.Title,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		},
	}

	// 3. сохр в кэш

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("json marshal error for %s: %v", cacheKey, err)

	} else {
		if err = Rdb.Set(ctx, cacheKey, data, 0).Err(); err != nil {
			log.Printf("failed to set cache for %s: %v", cacheKey, err)
		} else {
			log.Printf("cached response for %s", cacheKey)
		}
	}

	return resp, nil
}

// AddTask добавляет новую задачу
func (s *TaskServer) AddTask(ctx context.Context, req *pb.Task) (*pb.AddTaskResponse, error) {
	id, err := AddTask(&Task{
		Date:    req.Date,
		Title:   req.Title,
		Comment: req.Comment,
		Repeat:  req.Repeat,
	})
	if err != nil {
		return nil, err
	}

	// при добавлении задачи стираем кэш

	ClearTaskCache(ctx)

	return &pb.AddTaskResponse{Id: fmt.Sprint(id)}, nil
}

// UpdateTask обновляет задачу
func (s *TaskServer) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.EmptyResponse, error) {
	err := UpdateTask(&Task{
		ID:      req.Task.Id,
		Date:    req.Task.Date,
		Title:   req.Task.Title,
		Comment: req.Task.Comment,
		Repeat:  req.Task.Repeat,
	})
	if err != nil {
		return nil, err
	}

	// задача обновляетя обновляем кэш этой задачи

	cacheKey := "task:" + req.Task.Id
	resp := &pb.GetTaskResponse{
		Task: req.Task,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("json marshal error for %s: %v", cacheKey, err)

	} else {
		if err = Rdb.Set(ctx, cacheKey, data, 0).Err(); err != nil {
			log.Printf("failed to set cache for %s: %v", cacheKey, err)
		} else {
			log.Printf("cached response for %s", cacheKey)
		}
	}

	return &pb.EmptyResponse{}, nil
}

// DeleteTask удаляет задачу по ID
func (s *TaskServer) DeleteTask(ctx context.Context, req *pb.IDRequest) (*pb.EmptyResponse, error) {
	if err := DeleteTask(req.Id); err != nil {
		log.Println("error: ", err)
		return nil, err
	}

	// очищаем кэш

	ClearTaskCache(ctx)

	return &pb.EmptyResponse{}, nil
}

// DoneTask отмечает задачу как выполненную и удаляет
func (s *TaskServer) DoneTask(ctx context.Context, req *pb.IDRequest) (*pb.EmptyResponse, error) {
	if err := DeleteTask(req.Id); err != nil {
		log.Println("error: ", err)
		return nil, err
	}

	// чистим кэш

	ClearTaskCache(ctx)

	return &pb.EmptyResponse{}, nil
}

// UpdateDate обновляет дату задачи
func (s *TaskServer) UpdateDate(ctx context.Context, req *pb.UpdateDateRequest) (*pb.EmptyResponse, error) {
	if err := UpdateDate(req.NextDate, req.Id); err != nil {
		log.Println("error: ", err)
		return nil, err
	}

	// чистим кэш
	ClearTaskCache(ctx)

	return &pb.EmptyResponse{}, nil
}

// NextDate рассчитывает следующую дату по правилу повторения
func (s *TaskServer) NextDate(ctx context.Context, req *pb.NextDateRequest) (*pb.NextDateResponse, error) {
	now := time.Now()
	next, err := cm.NextDate(now, req.TaskDate, req.RepeatRule)
	if err != nil {
		log.Println("error: ", err)
		return nil, err
	}
	return &pb.NextDateResponse{NextDate: next}, nil
}
