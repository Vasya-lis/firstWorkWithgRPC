package api

import "net/http"

func (s *TaskServer) init() {
	http.HandleFunc("/api/nextdate", s.nextDateHandler)
	http.HandleFunc("/api/task", s.taskHandler)
	http.HandleFunc("/api/tasks", s.tasksHandler)
	http.HandleFunc("/api/task/done", s.doneTaskHandler)
	http.Handle("/", http.FileServer(http.Dir("./web")))

}
