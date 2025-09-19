package apperrors

import (
	"errors"
	"fmt"
)

// базовые ошибки
var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrInvalidTaskID     = errors.New("invalid task id")
	ErrDateRequired      = errors.New("date is required")
	ErrTitleRequired     = errors.New("title is required")
	ErrTaskRequired      = errors.New("task is required")
	ErrInvalidDateFormat = errors.New("invalid date format")
)

// С обёрткой — если нужно передавать детали
type ErrorWithDetail struct {
	Code    error  // базовая ошибка
	Details string // доп. информация
}

func (e *ErrorWithDetail) Error() string {
	return fmt.Sprintf("%v: %s", e.Code, e.Details)
}

func (e *ErrorWithDetail) Unwrap() error {
	return e.Code
}

// Конструктор
func NewError(code error, details string) error {
	return &ErrorWithDetail{
		Code:    code,
		Details: details,
	}
}
