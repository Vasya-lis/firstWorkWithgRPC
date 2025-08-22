package main

import (
	"log"
	"net/http"
	"time"

	pb "github.com/Vasya-lis/firstWorkWithgRPC/proto"
)

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	if dateStr == "" || repeat == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "date and repeat parameters are required"})
		return
	}

	now := time.Now().Format(formDate)
	if nowStr != "" {
		now = nowStr
	}
	client := pb.NewSchedulerServiceClient(conn)

	resp, err := client.NextDate(ctx, &pb.NextDateRequest{
		CurrentDate: now,
		TaskDate:    dateStr,
		RepeatRule:  repeat,
	})
	if err != nil {
		log.Println(err)
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": "failed to calculate next date"})
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(resp.NextDate))
}
