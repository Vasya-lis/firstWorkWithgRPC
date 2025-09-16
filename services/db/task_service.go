package db

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// AddTask — добавление задачи
func (app *AppDB) AddTask(task *Task) (int, error) {
	app.mu.Lock()
	defer app.mu.Unlock()

	if task == nil {
		return 0, fmt.Errorf("task is nil")
	}

	if task.Date == "" {
		task.Date = time.Now().Format("20060102")
	}

	if task.Title == "" {
		return 0, fmt.Errorf("title is required")
	}

	result := app.db.Create(task)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to insert task: %w", result.Error)
	}

	return task.ID, nil
}

// список задач с поиском и лимитом
func (app *AppDB) Tasks(limit int, search string) ([]*Task, error) {
	app.mu.RLock()
	defer app.mu.RUnlock()

	var tasks []*Task
	query := app.db.Session(&gorm.Session{}).Model(&Task{})

	search = strings.TrimSpace(search)

	switch {
	case search == "":
		// ничего не фильтруем
	case isDateSearch(search):
		date, err := parseSearchDate(search)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %v", err)
		}
		query = query.Where("date = ?", date)
	default:
		like := "%" + search + "%"
		query = query.Where("title LIKE ? OR comment LIKE ?", like, like)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("date ASC").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if tasks == nil {
		tasks = []*Task{}
	}

	return tasks, nil
}

func isDateSearch(s string) bool {
	_, err := time.Parse("02.01.2006", s)
	return err == nil
}

func parseSearchDate(s string) (string, error) {
	t, err := time.Parse("02.01.2006", s)
	if err != nil {
		return "", err
	}
	return t.Format("20060102"), nil
}

// одна задача по id
func (app *AppDB) GetTask(id int) (*Task, error) {
	app.mu.RLock()
	defer app.mu.RUnlock()

	var task Task
	result := app.db.First(&task, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("task was not found")
		}
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	return &task, nil
}
func (app *AppDB) UpdateTask(task *Task) error {
	app.mu.Lock()
	defer app.mu.Unlock()

	if task.ID == 0 {
		return fmt.Errorf("task ID is required")
	}

	result := app.db.Save(task)
	if result.Error != nil {
		return fmt.Errorf("failed to update task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("task was not found")
	}

	return nil
}

func (app *AppDB) DeleteTask(id int) error {
	app.mu.Lock()
	defer app.mu.Unlock()

	if id <= 0 {
		return fmt.Errorf("invalid task ID")
	}

	result := app.db.Delete(&Task{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("task was not found")
	}

	return nil
}

func (app *AppDB) UpdateDate(next string, id int) error {
	app.mu.Lock()
	defer app.mu.Unlock()

	if id <= 0 {
		return fmt.Errorf("invalid task ID")
	}
	if next == "" {
		return fmt.Errorf("date is required")
	}

	result := app.db.Model(&Task{}).Where("id = ?", id).Update("date", next)
	if result.Error != nil {
		return fmt.Errorf("failed to update date: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("task was not found")
	}

	return nil
}
