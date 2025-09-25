package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	cm "github.com/Vasya-lis/firstWorkWithgRPC/common"
	md "github.com/Vasya-lis/firstWorkWithgRPC/services/models"
)

func CheckDate(task *md.Task) error {
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

func validateAdd(t *md.Task) error {
	if t.Title == "" {
		return fmt.Errorf("the issue title is not specified")
	}
	return nil
}

func validate(t *md.Task) error {
	if t.ID == 0 {
		return fmt.Errorf("the issue ID is not specified")
	}
	if t.Title == "" {
		return fmt.Errorf("the issue title is not specified")
	}
	return nil
}

func GetIDFromQuery(w http.ResponseWriter, r *http.Request) (int, error) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {

		return 0, fmt.Errorf("id parameter is required")
	}
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("conversion error: %w", err)
	}
	return idInt, nil
}
