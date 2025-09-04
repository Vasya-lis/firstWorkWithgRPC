package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID      int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Date    string `gorm:"size:8;not null;default:''" json:"date"`
	Title   string `gorm:"size:255;not null;default:''" json:"title"`
	Comment string `gorm:"not null;default:''" json:"comment"`
	Repeat  string `gorm:"size:128;not null;default:''" json:"repeat"`
}

// AddTask — добавление задачи
func AddTask(task *Task) (int, error) {
	if task == nil {
		return 0, fmt.Errorf("task is nil")
	}

	if task.Date == "" {
		task.Date = time.Now().Format("20060102")
	}

	if task.Title == "" {
		return 0, fmt.Errorf("title is required")
	}

	result := db.Create(task)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to insert task: %w", result.Error)
	}

	return task.ID, nil
}

// список задач с поиском и лимитом
func Tasks(limit int, search string) ([]*Task, error) {

	var tasks []*Task
	query := db.Model(&Task{})

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
func GetTask(id int) (*Task, error) {

	var task Task
	result := db.First(&task, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("task was not found")
		}
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	return &task, nil
}
func UpdateTask(task *Task) error {
	if task.ID == 0 {
		return fmt.Errorf("task ID is required")
	}

	result := db.Save(task)
	if result.Error != nil {
		return fmt.Errorf("failed to update task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("task was not found")
	}

	return nil
}

func DeleteTask(id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid task ID")
	}

	result := db.Delete(&Task{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("task was not found")
	}

	return nil
}

func UpdateDate(next string, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid task ID")
	}
	if next == "" {
		return fmt.Errorf("date is required")
	}

	result := db.Model(&Task{}).Where("id = ?", id).Update("date", next)
	if result.Error != nil {
		return fmt.Errorf("failed to update date: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("task was not found")
	}

	return nil
}
