package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	cm "github.com/Vasya-lis/firstWorkWithgRPC/common"
	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"
	md "github.com/Vasya-lis/firstWorkWithgRPC/services/models"
)

func (app *AppAPI) AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task md.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	errResp := validateAdd(&task)
	if errResp != nil {
		WriteJson(w, http.StatusBadRequest, errResp)
		return
	}

	if err := CheckDate(&task); err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	client := pb.NewSchedulerServiceClient(app.conn)

	resp, err := client.AddTask(app.context, &pb.Task{
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

// GetTaskHandler обработчик GET /api/task
func (app *AppAPI) GetTaskHandler(w http.ResponseWriter, r *http.Request) {

	id, err := GetIDFromQuery(w, r)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	client := pb.NewSchedulerServiceClient(app.conn)

	task, err := client.GetTask(app.context, &pb.IDRequest{
		Id: int32(id),
	})
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	WriteJson(w, http.StatusOK, &md.Task{
		ID:      int(task.Task.Id),
		Date:    task.Task.Date,
		Title:   task.Task.Title,
		Comment: task.Task.Comment,
		Repeat:  task.Task.Repeat,
	})
}

// UpdateTaskHandler обработчик PUT /api/task
func (app *AppAPI) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task md.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": "ошибка десериализации JSON"})
		return
	}

	errResp := validate(&task)
	if errResp != nil {
		WriteJson(w, http.StatusBadRequest, errResp)
		return
	}

	err = CheckDate(&task)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	client := pb.NewSchedulerServiceClient(app.conn)

	_, err = client.UpdateTask(app.context, &pb.UpdateTaskRequest{
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

	client := pb.NewSchedulerServiceClient(app.conn)

	_, err = client.DeleteTask(app.context, &pb.IDRequest{
		Id: int32(id),
	})
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	WriteJson(w, http.StatusOK, map[string]interface{}{})
}

func (app *AppAPI) doneTaskHandler(w http.ResponseWriter, r *http.Request) {

	id, err := GetIDFromQuery(w, r)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	client := pb.NewSchedulerServiceClient(app.conn)

	task, err := client.GetTask(app.context, &pb.IDRequest{
		Id: int32(id),
	})
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	if task.Task.Repeat == "" {
		// Одноразовая задача — удаляем
		_, err := client.DeleteTask(app.context, &pb.IDRequest{
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

	_, err = client.UpdateDate(app.context, &pb.UpdateDateRequest{
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

func (app *AppAPI) nextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	if dateStr == "" || repeat == "" {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": "date and repeat parameters are required"})
		return
	}

	now := time.Now().Format(cm.FormDate)
	if nowStr != "" {
		now = nowStr
	}

	client := pb.NewSchedulerServiceClient(app.conn)

	resp, err := client.NextDate(app.context, &pb.NextDateRequest{
		CurrentDate: now,
		TaskDate:    dateStr,
		RepeatRule:  repeat,
	})
	if err != nil {
		log.Println(err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"error": "failed to calculate next date"})
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(resp.NextDate))
}

func (app *AppAPI) tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	search := r.URL.Query().Get("search")
	limit := 50

	client := pb.NewSchedulerServiceClient(app.conn)

	resp, err := client.ListTasks(app.context, &pb.ListTasksRequest{
		Limit:  int32(limit),
		Search: search,
	})
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch tasks"})
		return
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

	if tasks == nil {
		tasks = []*md.Task{}
	}

	WriteJson(w, http.StatusOK, TasksResponse{Tasks: tasks})
}
