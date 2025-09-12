package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	cm "github.com/Vasya-lis/firstWorkWithgRPC/common"
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

func (t *Task) ValidateAdd() error {
	if t.Title == "" {
		return fmt.Errorf("не указан заголовок задачи")
	}
	return nil
}

func (t *Task) Validate() error {
	if t.ID == 0 {
		return fmt.Errorf("не указан идентификатор задачи")
	}
	if t.Title == "" {
		return fmt.Errorf("не указан заголовок задачи")
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
		return 0, fmt.Errorf("ошибка конвертации: %w", err)
	}
	return idInt, nil
}
