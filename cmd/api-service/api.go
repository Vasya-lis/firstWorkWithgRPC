package main

import (
	"net/http"
)

// Обработчик для /api/task
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		AddTaskHandler(w, r)
	case http.MethodGet:
		GetTaskHandler(w, r)
	case http.MethodPut:
		UpdateTaskHandler(w, r)
	case http.MethodDelete:
		DeleteTaskHandler(w, r)
	default:
		WriteJson(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "Method not allowed",
		})
	}
}

func Init() {
	http.HandleFunc("/api/nextdate", (NextDateHandler))
	http.HandleFunc("/api/task", (taskHandler))
	http.HandleFunc("/api/tasks", (tasksHandler))
	http.HandleFunc("/api/task/done", (DoneTaskHandler))

}
