package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Task struct {
	ID      string `db:"id" json:"id"`
	Date    string `db:"date" json:"date"`
	Title   string `db:"title" json:"title"`
	Comment string `db:"comment" json:"comment"`
	Repeat  string `db:"repeat" json:"repeat"`
}

const formDate = "20060102"

func writeJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func checkDate(task *Task) error {
	now := time.Now()
	// если пустая дата — ставим сегодня
	if task.Date == "" {
		task.Date = now.Format(formDate)
		return nil
	}

	// парсим указанную дату
	t, err := time.ParseInLocation(formDate, task.Date, now.Location())
	if err != nil {
		log.Println("error: ", err)
		return fmt.Errorf("invalid date format: %w", err)
	}

	// если есть правило повторения — проверяем его через NextDate
	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			log.Println("error: ", err)
			return fmt.Errorf("invalid repeat: %w", err)
		}
	}

	// Если указанная дата раньше или равна сегодня (now), корректируем:
	if afterNow(now, t) {
		if task.Repeat == "" {
			// без повтора — ставим сегодняшнюю дату
			task.Date = now.Format(formDate)
		} else {
			// с повтором — ставим вычисленную следующую дату
			task.Date = next
		}
	}
	return nil
}
