package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

type Task struct {
	ID      string `db:"id" json:"id"`
	Date    string `db:"date" json:"date"`
	Title   string `db:"title" json:"title"`
	Comment string `db:"comment" json:"comment"`
	Repeat  string `db:"repeat" json:"repeat"`
}

// AddTask — добавление задачи
func AddTask(task *Task) (int64, error) {
	if task == nil {
		return 0, fmt.Errorf("task is nil")
	}

	// Устанавливаем текущую дату, если не указана
	if task.Date == "" {
		task.Date = time.Now().Format("20060102")
	}

	// Валидация обязательных полей
	if task.Title == "" {
		return 0, fmt.Errorf("title is required")
	}

	query := `
		INSERT INTO scheduler (date, title, comment, repeat)
		VALUES (?, ?, ?, ?)
	`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("failed to insert task: %w", err)
	}
	return res.LastInsertId()
}

func Tasks(limit int, search string) ([]*Task, error) {
	search = strings.TrimSpace(search)
	var rows *sql.Rows
	var err error

	query := `
		SELECT id, date, title, comment, repeat 
		FROM scheduler 
		%s
		ORDER BY date ASC 
		LIMIT ?
	`

	switch {
	case search == "":
		rows, err = db.Query(fmt.Sprintf(query, ""), limit)
	case isDateSearch(search):
		date, err := parseSearchDate(search)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %v", err)
		}
		rows, err = db.Query(fmt.Sprintf(query, "WHERE date = ?"), date, limit)
		log.Println("error: ", err)
	default:
		like := "%" + search + "%"
		rows, err = db.Query(fmt.Sprintf(query, "WHERE title LIKE ? OR comment LIKE ?"), like, like, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		tasks = append(tasks, &t)
	}

	if err = rows.Err(); err != nil {
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

func GetTask(id string) (*Task, error) {
	var t Task
	row := db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id)
	err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("задача не найдена")
		}
		return nil, err
	}
	return &t, nil
}

func UpdateTask(task *Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}
	return nil
}

func DeleteTask(id string) error {
	res, err := db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}
	return nil
}

func UpdateDate(next string, id string) error {
	res, err := db.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, next, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}
	return nil
}
