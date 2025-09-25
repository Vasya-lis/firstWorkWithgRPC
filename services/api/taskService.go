package api

import (
	"context"
	"time"

	cm "github.com/Vasya-lis/firstWorkWithgRPC/common"
	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"
	md "github.com/Vasya-lis/firstWorkWithgRPC/services/models"
)

type TaskService struct {
	client *SchedulerClient
}

func NewTaskService(client *SchedulerClient) *TaskService {
	return &TaskService{
		client: client}
}

func (s *TaskService) AddTask(ctx context.Context, task *md.Task) (int, error) {
	resp, err := s.client.AddTask(ctx, &pb.Task{
		Title:   task.Title,
		Date:    task.Date,
		Comment: task.Comment,
		Repeat:  task.Repeat,
	})
	if err != nil {
		return 0, err
	}
	return int(resp.Id), nil
}

func (s *TaskService) GetTask(ctx context.Context, id int) (*md.Task, error) {
	resp, err := s.client.GetTask(ctx, &pb.IDRequest{
		Id: int32(id),
	})
	if err != nil {
		return nil, err
	}
	return &md.Task{
		ID:      int(resp.Task.Id),
		Date:    resp.Task.Date,
		Title:   resp.Task.Title,
		Comment: resp.Task.Comment,
		Repeat:  resp.Task.Repeat,
	}, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, task *md.Task) error {
	_, err := s.client.UpdateTask(ctx, &pb.UpdateTaskRequest{
		Task: &pb.Task{
			Id:      int32(task.ID),
			Date:    task.Date,
			Title:   task.Title,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		},
	})
	return err
}

func (s *TaskService) DeleteTask(ctx context.Context, id int) error {
	_, err := s.client.DeleteTask(ctx, &pb.IDRequest{
		Id: int32(id),
	})
	return err
}

func (s *TaskService) DoneTask(ctx context.Context, id int) error {
	task, err := s.client.GetTask(ctx, &pb.IDRequest{
		Id: int32(id)})
	if err != nil {
		return err
	}
	if task.Task.Repeat == "" {
		_, err = s.client.DeleteTask(ctx, &pb.IDRequest{
			Id: int32(id),
		})
		return err
	}
	nextDate, err := cm.NextDate(time.Now(), task.Task.Date, task.Task.Repeat)
	if err != nil {
		return err
	}

	_, err = s.client.UpdateDate(ctx, &pb.UpdateDateRequest{
		Id:       int32(id),
		NextDate: nextDate,
	})
	return err

}

func (s *TaskService) NextDateCalc(ctx context.Context, now, date, repeat string) (string, error) {
	resp, err := s.client.NextDate(ctx, &pb.NextDateRequest{
		CurrentDate: now,
		TaskDate:    date,
		RepeatRule:  repeat,
	})
	if err != nil {
		return "", err
	}
	return resp.NextDate, nil
}

func (s *TaskService) ListTasks(ctx context.Context, limit int, search string) ([]*md.Task, error) {
	resp, err := s.client.ListTasks(ctx, &pb.ListTasksRequest{
		Limit:  int32(limit),
		Search: search,
	})
	if err != nil {
		return nil, err
	}

	var tasks []*md.Task
	for _, protoTask := range resp.Tasks {
		tasks = append(tasks, &md.Task{
			ID:      int(protoTask.Id),
			Date:    protoTask.Date,
			Title:   protoTask.Title,
			Comment: protoTask.Comment,
			Repeat:  protoTask.Repeat,
		})
	}
	return tasks, nil
}
