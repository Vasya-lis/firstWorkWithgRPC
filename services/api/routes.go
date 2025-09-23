package api

import (
	"net/http"
)

// Обработчик для /api/task
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
func (app *AppAPI) init() {
	http.HandleFunc("/api/nextdate", func(w http.ResponseWriter, r *http.Request) { app.nextDateHandler(w, r) })
	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) { app.taskHandler(w, r) })
	http.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) { app.tasksHandler(w, r) })
	http.HandleFunc("/api/task/done", func(w http.ResponseWriter, r *http.Request) { app.doneTaskHandler(w, r) })

	http.Handle("/", http.FileServer(http.Dir("./web")))

}
