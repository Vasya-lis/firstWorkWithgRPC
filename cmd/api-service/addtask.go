package main

import (
	"encoding/json"
	"log"
	"net/http"

	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"
)

func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Println("error: ", err)
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	if task.Title == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Title is required"})
		return
	}

	if err := checkDate(&task); err != nil {
		log.Println("error: ", err)
		writeJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
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
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": "Failed to add task"})

		return
	}

	writeJson(w, http.StatusOK, map[string]string{"id": resp.Id})
}
