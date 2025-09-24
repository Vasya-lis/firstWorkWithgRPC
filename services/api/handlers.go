package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	cm "github.com/Vasya-lis/firstWorkWithgRPC/common"
	md "github.com/Vasya-lis/firstWorkWithgRPC/services/models"
)

func (app *AppAPI) taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		app.AddTaskHandler(w, r)
	case http.MethodGet:
		app.GetTaskHandler(w, r)
	case http.MethodPut:
		app.UpdateTaskHandler(w, r)
	case http.MethodDelete:
		app.DeleteTaskHandler(w, r)
	default:
		WriteJson(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "Method not allowed",
		})
	}
}

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

	id, err := app.taskService.AddTask(r.Context(), &task)
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"error": "Failed to add task"})

		return
	}
	WriteJson(w, http.StatusOK, map[string]int{"id": id})
}

// GetTaskHandler обработчик GET /api/task
func (app *AppAPI) GetTaskHandler(w http.ResponseWriter, r *http.Request) {

	id, err := GetIDFromQuery(w, r)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	task, err := app.taskService.GetTask(app.context, id)
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	WriteJson(w, http.StatusOK, task)
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

	err = app.taskService.UpdateTask(app.context, &task)
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

	err = app.taskService.DeleteTask(app.context, id)
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

	err = app.taskService.DoneTask(app.context, id)
	if err != nil {
		log.Println("error: ", err)
		WriteJson(w, http.StatusNotFound, map[string]string{"error": err.Error()})
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

	nextDate, err := app.taskService.NextDateCalc(app.context, now, dateStr, repeat)
	if err != nil {
		log.Println(err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"error": "failed to calculate next date"})
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(nextDate))
}

func (app *AppAPI) tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	search := r.URL.Query().Get("search")
	limit := 50

	tasks, err := app.taskService.ListTasks(r.Context(), limit, search)
	if err != nil {
		log.Println("error:", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch tasks"})
		return
	}

	if tasks == nil {
		tasks = []*md.Task{}
	}

	WriteJson(w, http.StatusOK, TasksResponse{Tasks: tasks})
}
