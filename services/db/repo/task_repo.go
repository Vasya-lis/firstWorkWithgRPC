package repo

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	apperrors "github.com/Vasya-lis/firstWorkWithgRPC/common/app_errors"
	md "github.com/Vasya-lis/firstWorkWithgRPC/services/models"
	"gorm.io/gorm"
)

type TasksRepo struct {
	db *gorm.DB
	mu sync.RWMutex
}

func NewTasksRepo(db *gorm.DB) *TasksRepo {
	return &TasksRepo{
		db: db,
	}
}

func (t *TasksRepo) AddTask(task *md.Task) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if task == nil {
		return 0, fmt.Errorf("%w:task is nil", apperrors.ErrTaskNotFound)
	}

	if task.Date == "" {
		task.Date = time.Now().Format("20060102")
	}

	if task.Title == "" {
		return 0, apperrors.ErrTitleRequired
	}

	result := t.db.Create(task)
	if result.Error != nil {
		return 0, fmt.Errorf("%w:%w", apperrors.ErrAddTask, result.Error)
	}

	return task.ID, nil
}

// список задач с поиском и лимитом подлежит репозиторию тасков
func (t *TasksRepo) Tasks(limit int, search string) ([]*md.Task, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var tasks []*md.Task
	query := t.db.Session(&gorm.Session{}).Model(&md.Task{})

	search = strings.TrimSpace(search)

	switch {
	case search == "":
		// ничего не фильтруем
	case isDateSearch(search):
		date, err := parseSearchDate(search)
		if err != nil {
			return nil, fmt.Errorf("%w:%w", apperrors.ErrInvalidDateFormat, err)
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
		return nil, fmt.Errorf("%w:%w", apperrors.ErrGetTasks, err)
	}

	if tasks == nil {
		tasks = []*md.Task{}
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
func (t *TasksRepo) GetTask(id int) (*md.Task, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var task md.Task
	result := t.db.First(&task, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTaskNotFound
		}
		return nil, fmt.Errorf("%w:%w", apperrors.ErrGetTask, result.Error)
	}
	return &task, nil
}
func (t *TasksRepo) Updates(task *md.Task) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if task.ID == 0 {
		return apperrors.ErrInvalidTaskID
	}

	result := t.db.Save(task)
	if result.Error != nil {
		return fmt.Errorf("%w:%w", apperrors.ErrUpdateTask, result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.ErrTaskNotFound
	}

	return nil
}

func (t *TasksRepo) DeleteTask(id int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if id <= 0 {
		return apperrors.ErrInvalidTaskID
	}

	result := t.db.Delete(&md.Task{}, id)
	if result.Error != nil {
		return fmt.Errorf("%w:%w", apperrors.ErrDeleteTask, result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.ErrTaskNotFound
	}

	return nil
}

func (t *TasksRepo) UpdateDate(next string, id int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if id <= 0 {
		return apperrors.ErrInvalidTaskID
	}
	if next == "" {
		return apperrors.ErrDateRequired
	}

	result := t.db.Model(&md.Task{}).Where("id = ?", id).Update("date", next)
	if result.Error != nil {
		return fmt.Errorf("%w:%w", apperrors.ErrUpdateTaskDate, result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.ErrTaskNotFound
	}

	return nil
}
