package db

import (
	"context"
	"errors"
	"log"
	"time"

	cm "github.com/Vasya-lis/firstWorkWithgRPC/common"
	apperrors "github.com/Vasya-lis/firstWorkWithgRPC/common/app_errors"
	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"
	"github.com/Vasya-lis/firstWorkWithgRPC/services/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskServer struct {
	pb.UnimplementedSchedulerServiceServer
	ts *TasksService
}

func NewTasksServer(ts *TasksService) *TaskServer {
	return &TaskServer{
		ts: ts,
	}
}

// ListTasks возвращает список задач с лимитом и поиском
func (s *TaskServer) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {

	tasks, err := s.ts.GetTasks(ctx, int(req.Limit), req.Search)
	if err != nil {
		if errors.Is(err, apperrors.ErrTaskNotFound) {
			return nil, status.Errorf(codes.NotFound, "tasks wish limit= %d and search= %s not found", req.Limit, req.Search)
		}
		return nil, status.Errorf(codes.Internal, "failed to get tasks: %v", err)
	}

	// конвертируем в прото буф
	var pbTasks []*pb.Task
	for _, t := range tasks {
		pbTasks = append(pbTasks, &pb.Task{
			Id:      int32(t.ID),
			Date:    t.Date,
			Title:   t.Title,
			Comment: t.Comment,
			Repeat:  t.Repeat,
		})

	}

	return &pb.ListTasksResponse{
		Tasks: pbTasks,
	}, nil
}

// GetTask возвращает задачу по ID
func (s *TaskServer) GetTask(ctx context.Context, req *pb.IDRequest) (*pb.GetTaskResponse, error) {

	// 1. проверяем кэш

	task, err := s.ts.GetTask(ctx, int(req.Id))
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrTaskNotFound):
			return nil, status.Error(codes.NotFound, "task not found")
		case errors.Is(err, apperrors.ErrInvalidTaskID):
			return nil, status.Error(codes.InvalidArgument, "invalid task id")
		default:
			log.Printf("GetTask error: %v", err)
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &pb.GetTaskResponse{
		Task: &pb.Task{
			Id:      int32(task.ID),
			Date:    task.Date,
			Title:   task.Title,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		},
	}, nil
}

// AddTask добавляет новую задачу
func (s *TaskServer) AddTask(ctx context.Context, req *pb.Task) (*pb.AddTaskResponse, error) {

	task := &models.Task{
		Date:    req.Date,
		Title:   req.Title,
		Comment: req.Comment,
		Repeat:  req.Repeat,
	}

	id, err := s.ts.AddTask(ctx, task)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrTitleRequired):
			return nil, status.Error(codes.InvalidArgument, "title is required")
		default:
			log.Printf("AddTask error: %v", err)
			return nil, status.Error(codes.Internal, "failed to add task")
		}
	}

	return &pb.AddTaskResponse{Id: int32(id)}, nil
}

// UpdateTask обновляет задачу
func (s *TaskServer) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.EmptyResponse, error) {

	task := &models.Task{
		ID:      int(req.Task.Id),
		Date:    req.Task.Date,
		Title:   req.Task.Title,
		Comment: req.Task.Comment,
		Repeat:  req.Task.Repeat,
	}

	err := s.ts.UpdateTask(ctx, task)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInvalidTaskID):
			return nil, status.Error(codes.InvalidArgument, "invalid task id")
		case errors.Is(err, apperrors.ErrTaskNotFound):
			return nil, status.Error(codes.NotFound, "task not found")
		default:
			log.Printf("UpdateTask error: %v", err)
			return nil, status.Error(codes.Internal, "failed to update task")
		}
	}

	return &pb.EmptyResponse{}, nil

}

// DeleteTask удаляет задачу по ID
func (s *TaskServer) DeleteTask(ctx context.Context, req *pb.IDRequest) (*pb.EmptyResponse, error) {

	if err := s.ts.DeleteTask(ctx, int(req.Id)); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInvalidTaskID):
			return nil, status.Error(codes.InvalidArgument, "invalid task id")
		case errors.Is(err, apperrors.ErrTaskNotFound):
			return nil, status.Error(codes.NotFound, "task not found")
		default:
			log.Printf("DeleteTask error: %v", err)
			return nil, status.Error(codes.Internal, "failed to delete task")
		}
	}

	return &pb.EmptyResponse{}, nil
}

// DoneTask отмечает задачу как выполненную и удаляет
func (s *TaskServer) DoneTask(ctx context.Context, req *pb.IDRequest) (*pb.EmptyResponse, error) {

	if err := s.ts.DeleteTask(ctx, int(req.Id)); err != nil {
		log.Println("error: ", err)
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}

// UpdateDate обновляет дату задачи
func (s *TaskServer) UpdateDate(ctx context.Context, req *pb.UpdateDateRequest) (*pb.EmptyResponse, error) {
	err := s.ts.UpdateDateTask(ctx, req.NextDate, int(req.Id))
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInvalidTaskID):
			return nil, status.Error(codes.InvalidArgument, "invalid task id")
		case errors.Is(err, apperrors.ErrDateRequired):
			return nil, status.Error(codes.InvalidArgument, "date is required")
		case errors.Is(err, apperrors.ErrTaskNotFound):
			return nil, status.Error(codes.NotFound, "task not found")
		default:
			log.Printf("UpdateDate error: %v", err)
			return nil, status.Error(codes.Internal, "failed to update date")
		}
	}

	return &pb.EmptyResponse{}, nil
}

// NextDate рассчитывает следующую дату по правилу повторения
func (s *TaskServer) NextDate(ctx context.Context, req *pb.NextDateRequest) (*pb.NextDateResponse, error) {
	now := time.Now()
	next, err := cm.NextDate(now, req.TaskDate, req.RepeatRule)
	if err != nil {
		log.Println("error: ", err)
		return nil, status.Error(codes.InvalidArgument, "invalid repeat rule or date")
	}
	return &pb.NextDateResponse{NextDate: next}, nil
}
