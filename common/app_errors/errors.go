package apperrors

import (
	"errors"
)

// базовые ошибки
var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrInvalidTaskID     = errors.New("invalid task id")
	ErrDateRequired      = errors.New("date is required")
	ErrTitleRequired     = errors.New("title is required")
	ErrTaskRequired      = errors.New("task is required")
	ErrInvalidDateFormat = errors.New("invalid date format")

	// кэш ошибки
	ErrGetTaskCache    = errors.New("get task cache failed")
	ErrSetTaskCache    = errors.New("set task cache failed")
	ErrGetTasksCache   = errors.New("get tasks cache failed")
	ErrSetTasksCache   = errors.New("set tasks cache failed")
	ErrDeleteTaskCache = errors.New("delete task cache failed")

	// репо ошибки
	ErrAddTask        = errors.New("add task failed")
	ErrGetTasks       = errors.New("get tasks failed")
	ErrGetTask        = errors.New("get task failed")
	ErrUpdateTask     = errors.New("update task failed")
	ErrDeleteTask     = errors.New("delete task failed")
	ErrUpdateTaskDate = errors.New("update task date failed")
)
