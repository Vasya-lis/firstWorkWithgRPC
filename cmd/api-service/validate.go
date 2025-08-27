package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	cm "github.com/Vasya-lis/firstWorkWithgRPC/cmd/common"
)

func CheckDate(task *Task) error {
	now := time.Now()
	// если пустая дата — ставим сегодня
	if task.Date == "" {
		task.Date = now.Format(cm.FormDate)
		return nil
	}

	// парсим указанную дату
	t, err := time.ParseInLocation(cm.FormDate, task.Date, now.Location())
	if err != nil {
		log.Println("error: ", err)
		return fmt.Errorf("invalid date format: %w", err)
	}

	// если есть правило повторения — проверяем его через NextDate
	var next string
	if task.Repeat != "" {
		next, err = cm.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			log.Println("error: ", err)
			return fmt.Errorf("invalid repeat: %w", err)
		}
	}

	// Если указанная дата раньше или равна сегодня (now), корректируем:
	if cm.AfterNow(now, t) {
		if task.Repeat == "" {
			// без повтора — ставим сегодняшнюю дату
			task.Date = now.Format(cm.FormDate)
		} else {
			// с повтором — ставим вычисленную следующую дату
			task.Date = next
		}
	}
	return nil
}

func (t *Task) Validate() (bool, map[string]string) {
	if t.ID == "" {
		return false, map[string]string{"error": "не указан идентификатор задачи"}
	}
	if t.Title == "" {
		return false, map[string]string{"error": "не указан заголовок задачи"}
	}
	return true, nil
}

func GetIDFromQuery(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.URL.Query().Get("id")
	if id == "" {
		WriteJson(w, http.StatusBadRequest, map[string]string{
			"error": "не указан идентификатор",
		})
		return "", false
	}
	return id, true
}
