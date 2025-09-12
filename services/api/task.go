package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	cm "github.com/Vasya-lis/firstWorkWithgRPC/common"
	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"
)

type Task struct {
	ID      int    `db:"id" json:"id"`
	Date    string `db:"date" json:"date"`
	Title   string `db:"title" json:"title"`
	Comment string `db:"comment" json:"comment"`
	Repeat  string `db:"repeat" json:"repeat"`
}

// Структура для ответа с задачами
type TasksResp struct {
	Tasks []*Task `json:"tasks"`
}

func (app *AppAPI) AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	errResp := task.ValidateAdd()
	if errResp != nil {
		WriteJson(w, http.StatusBadRequest, errResp)
		return
	}

	if err := CheckDate(&task); err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	resp, err := app.client.AddTask(app.context, &pb.Task{
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
	WriteJson(w, http.StatusOK, map[string]int{"id": int(resp.Id)})
}

func (app *AppAPI) tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	search := r.URL.Query().Get("search")
	limit := 50

	resp, err := app.client.ListTasks(app.context, &pb.ListTasksRequest{
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
			ID:      int(protoTask.Id),
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
func (app *AppAPI) GetTaskHandler(w http.ResponseWriter, r *http.Request) {

	id, err := GetIDFromQuery(w, r)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	task, err := app.client.GetTask(app.context, &pb.IDRequest{
		Id: int32(id),
	})
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	WriteJson(w, http.StatusOK, &Task{
		ID:      int(task.Task.Id),
		Date:    task.Task.Date,
		Title:   task.Task.Title,
		Comment: task.Task.Comment,
		Repeat:  task.Task.Repeat,
	})
}

// UpdateTaskHandler обработчик PUT /api/task
func (app *AppAPI) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": "ошибка десериализации JSON"})
		return
	}

	errResp := task.Validate()
	if errResp != nil {
		WriteJson(w, http.StatusBadRequest, errResp)
		return
	}

	err = CheckDate(&task)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	_, err = app.client.UpdateTask(app.context, &pb.UpdateTaskRequest{
		Task: &pb.Task{
			Id:      int32(task.ID),
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

func (app *AppAPI) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDFromQuery(w, r)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	_, err = app.client.DeleteTask(app.context, &pb.IDRequest{
		Id: int32(id),
	})
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	WriteJson(w, http.StatusOK, map[string]interface{}{})
}

func (app *AppAPI) DoneTaskHandler(w http.ResponseWriter, r *http.Request) {

	id, err := GetIDFromQuery(w, r)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	task, err := app.client.GetTask(app.context, &pb.IDRequest{
		Id: int32(id),
	})
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	if task.Task.Repeat == "" {
		// Одноразовая задача — удаляем
		_, err := app.client.DeleteTask(app.context, &pb.IDRequest{
			Id: int32(id),
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

	_, err = app.client.UpdateDate(app.context, &pb.UpdateDateRequest{
		Id:       int32(id),
		NextDate: nextDate,
	})
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	WriteJson(w, http.StatusOK, map[string]interface{}{})
}
