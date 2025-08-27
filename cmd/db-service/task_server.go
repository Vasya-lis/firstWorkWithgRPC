package main

import (
	"context"
	"fmt"
	"log"
	"time"

	cm "github.com/Vasya-lis/firstWorkWithgRPC/cmd/common"
	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"
)

type TaskServer struct {
	pb.UnimplementedSchedulerServiceServer
}

// ListTasks возвращает список задач с лимитом и поиском
func (s *TaskServer) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	tasks, err := Tasks(int(req.Limit), req.Search)
	if err != nil {
		return nil, err
	}

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

	return &pb.ListTasksResponse{Tasks: pbTasks}, nil
}

// GetTask возвращает задачу по ID
func (s *TaskServer) GetTask(ctx context.Context, req *pb.IDRequest) (*pb.GetTaskResponse, error) {
	task, err := GetTask(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetTaskResponse{
		Task: &pb.Task{
			Id:      task.ID,
			Date:    task.Date,
			Title:   task.Title,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		},
	}, nil
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
	return &pb.EmptyResponse{}, nil
}

// DeleteTask удаляет задачу по ID
func (s *TaskServer) DeleteTask(ctx context.Context, req *pb.IDRequest) (*pb.EmptyResponse, error) {
	if err := DeleteTask(req.Id); err != nil {
		log.Println("error: ", err)
		return nil, err
	}
	return &pb.EmptyResponse{}, nil
}

// DoneTask отмечает задачу как выполненную и удаляет
func (s *TaskServer) DoneTask(ctx context.Context, req *pb.IDRequest) (*pb.EmptyResponse, error) {
	if err := DeleteTask(req.Id); err != nil {
		log.Println("error: ", err)
		return nil, err
	}
	return &pb.EmptyResponse{}, nil
}

// UpdateDate обновляет дату задачи
func (s *TaskServer) UpdateDate(ctx context.Context, req *pb.UpdateDateRequest) (*pb.EmptyResponse, error) {
	if err := UpdateDate(req.NextDate, req.Id); err != nil {
		log.Println("error: ", err)
		return nil, err
	}
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
