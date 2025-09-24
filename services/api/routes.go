package api

import "net/http"

func (app *AppAPI) init() {
	http.HandleFunc("/api/nextdate", app.nextDateHandler)
	http.HandleFunc("/api/task", app.taskHandler)
	http.HandleFunc("/api/tasks", app.tasksHandler)
	http.HandleFunc("/api/task/done", app.doneTaskHandler)
	http.Handle("/", http.FileServer(http.Dir("./web")))

}
