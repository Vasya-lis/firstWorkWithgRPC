package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	cm "github.com/Vasya-lis/firstWorkWithgRPC/cmd/common"
	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"
)

type Task struct {
	ID      string `db:"id" json:"id"`
	Date    string `db:"date" json:"date"`
	Title   string `db:"title" json:"title"`
	Comment string `db:"comment" json:"comment"`
	Repeat  string `db:"repeat" json:"repeat"`
}

// Структура для ответа с задачами
type TasksResp struct {
	Tasks []*Task `json:"tasks"`
}

func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	if ok, errResp := task.Validate(); !ok {
		WriteJson(w, http.StatusBadRequest, errResp)
	}

	if err := CheckDate(&task); err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	client := pb.NewSchedulerServiceClient(conn)

	resp, err := client.AddTask(ctx, &pb.Task{
		Title:   task.Title,
		Date:    task.Date,
		Comment: task.Comment,
		Repeat:  task.Repeat,
	})
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"error": "Failed to add task"})

		return
	}
	WriteJson(w, http.StatusOK, map[string]string{"id": resp.Id})
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	search := r.URL.Query().Get("search")
	limit := 50

	client := pb.NewSchedulerServiceClient(conn)

	resp, err := client.ListTasks(ctx, &pb.ListTasksRequest{
		Limit:  int32(limit),
		Search: search,
	})
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch tasks"})
		return
	}

	var tasks []*Task
	for _, protoTask := range resp.Tasks {
		tasks = append(tasks, &Task{
			ID:      protoTask.Id,
			Date:    protoTask.Date,
			Title:   protoTask.Title,
			Comment: protoTask.Comment,
			Repeat:  protoTask.Repeat,
		})
	}

	if tasks == nil {
		tasks = []*Task{}
	}

	WriteJson(w, http.StatusOK, TasksResp{Tasks: tasks})
}

// GetTaskHandler обработчик GET /api/task
func GetTaskHandler(w http.ResponseWriter, r *http.Request) {

	id, err := GetIDFromQuery(w, r)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	client := pb.NewSchedulerServiceClient(conn)

	task, err := client.GetTask(ctx, &pb.IDRequest{
		Id: id,
	})
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	WriteJson(w, http.StatusOK, task.Task)
}

// UpdateTaskHandler обработчик PUT /api/task
func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": "ошибка десериализации JSON"})
		return
	}

	if ok, errResp := task.Validate(); !ok {
		WriteJson(w, http.StatusBadRequest, errResp)
	}

	err = CheckDate(&task)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	client := pb.NewSchedulerServiceClient(conn)

	_, err = client.UpdateTask(ctx, &pb.UpdateTaskRequest{
		Task: &pb.Task{
			Id:      task.ID,
			Title:   task.Title,
			Date:    task.Date,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		},
	})
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Успешный ответ — пустой JSON
	WriteJson(w, http.StatusOK, map[string]interface{}{})
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDFromQuery(w, r)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	client := pb.NewSchedulerServiceClient(conn)

	_, err = client.DeleteTask(ctx, &pb.IDRequest{
		Id: id,
	})
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	WriteJson(w, http.StatusOK, map[string]interface{}{})
}

func DoneTaskHandler(w http.ResponseWriter, r *http.Request) {

	id, err := GetIDFromQuery(w, r)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	client := pb.NewSchedulerServiceClient(conn)

	task, err := client.GetTask(ctx, &pb.IDRequest{
		Id: id,
	})
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	if task.Task.Repeat == "" {
		// Одноразовая задача — удаляем
		_, err := client.DeleteTask(ctx, &pb.IDRequest{
			Id: id,
		})
		if err != nil {
			log.Println("error: ", err)
			WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		WriteJson(w, http.StatusOK, map[string]interface{}{})
		return
	}

	// Периодическая задача — вычисляем следующую дату
	nextDate, err := cm.NextDate(time.Now(), task.Task.Date, task.Task.Repeat)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	_, err = client.UpdateDate(ctx, &pb.UpdateDateRequest{
		Id:       id,
		NextDate: nextDate,
	})
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	WriteJson(w, http.StatusOK, map[string]interface{}{})
}
