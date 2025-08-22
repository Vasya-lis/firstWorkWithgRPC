package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

const schema = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(255) NOT NULL DEFAULT "",
    comment TEXT NOT NULL DEFAULT "",
    repeat VARCHAR(128) NOT NULL DEFAULT ""
);
CREATE INDEX idx_scheduler_date ON scheduler(date);
`

// Init открывает базу данных и при необходимости создает таблицу и индекс
func Init(dbFile string) error {
	install := false

	// Проверяем наличие файла базы данных
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		install = true
	}

	// Открываем подключение к базе
	var err error
	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("database opening error: %w", err)
	}

	// Если файл не существовал (install == true), создаем таблицу и индекс
	if install {
		if _, err = db.Exec(schema); err != nil {
			return fmt.Errorf("schema creation error: %w", err)
		}

	}

	return nil
}

func GetDB() *sql.DB {
	return db
}

func NextDate(now time.Time, dateStr string, repeat string) (string, error) {
	date, err := time.Parse("2006.01.02", dateStr)
	if err != nil {
		log.Println("error: ", err)
		return "", errors.New("invalid date format")
	}

	if repeat == "" {
		if date.Before(now) {
			return now.Format("2006.01.02"), nil
		}
		return date.Format("2006.01.02"), nil
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", errors.New("invalid repeat format")
	}

	var nextDate time.Time

	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", errors.New("invalid d format")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			log.Println("error: ", err)
			return "", errors.New("invalid day interval")
		}

		nextDate = date
		for {
			nextDate = nextDate.AddDate(0, 0, days)
			if afterNow(nextDate, now) {
				break
			}
		}

	case "y":
		nextDate = date
		// Если дата уже в будущем, добавляем год
		if afterNow(nextDate, now) {
			nextDate = nextDate.AddDate(1, 0, 0)
		} else {
			// Иначе ищем следующий год, когда дата будет в будущем
			for !afterNow(nextDate, now) {
				nextDate = nextDate.AddDate(1, 0, 0)
			}
		}
		// Коррекция для 29 февраля в невисокосный год
		if nextDate.Month() == time.February && nextDate.Day() == 29 && !isLeap(nextDate.Year()) {
			nextDate = time.Date(nextDate.Year(), time.March, 1, 0, 0, 0, 0, time.UTC)
		}

	case "w":
		if len(parts) != 2 {
			return "", errors.New("invalid w format")
		}
		weekdays, err := parseWeekdays(parts[1])
		if err != nil {
			return "", err
		}
		nextDate = findNextWeekday(date, now, weekdays)

	case "m":
		if len(parts) < 2 {
			return "", errors.New("invalid m format")
		}
		days, months, err := parseMonthRules(parts[1:])
		if err != nil {
			return "", err
		}
		nextDate = findNextMonthDay(date, now, days, months)

	default:
		return "", errors.New("unsupported repeat format")
	}

	return nextDate.Format("2006.01.02"), nil
}

func afterNow(date, now time.Time) bool {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return date.After(now)
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func parseWeekdays(s string) (map[int]bool, error) {
	daysStr := strings.Split(s, ",")
	weekdays := make(map[int]bool)
	for _, dayStr := range daysStr {
		day, err := strconv.Atoi(dayStr)
		if err != nil || day < 1 || day > 7 {
			return nil, errors.New("invalid weekday")
		}
		weekdays[day] = true
	}
	return weekdays, nil
}

func parseMonthRules(parts []string) (map[int]bool, map[int]bool, error) {
	daysStr := strings.Split(parts[0], ",")
	days := make(map[int]bool)
	for _, dayStr := range daysStr {
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			return nil, nil, errors.New("invalid day in month")
		}
		if day < -2 || day > 31 || day == 0 {
			return nil, nil, errors.New("invalid day in month")
		}
		days[day] = true
	}

	var months map[int]bool
	if len(parts) > 1 {
		monthsStr := strings.Split(parts[1], ",")
		months = make(map[int]bool)
		for _, monthStr := range monthsStr {
			month, err := strconv.Atoi(monthStr)
			if err != nil || month < 1 || month > 12 {
				return nil, nil, errors.New("invalid month")
			}
			months[month] = true
		}
	}

	return days, months, nil
}

func findNextWeekday(start, now time.Time, weekdays map[int]bool) time.Time {
	date := start
	for {
		date = date.AddDate(0, 0, 1)
		if afterNow(date, now) {
			wd := int(date.Weekday())
			if wd == 0 { // Воскресенье
				wd = 7
			}
			if weekdays[wd] {
				return date
			}
		}
	}
}

func findNextMonthDay(start, now time.Time, days, months map[int]bool) time.Time {
	date := start
	for {
		date = date.AddDate(0, 0, 1)
		if afterNow(date, now) {
			month := int(date.Month())
			day := date.Day()
			lastDay := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()

			// Проверяем ограничения по месяцам
			if len(months) > 0 && !months[month] {
				continue
			}

			// Проверяем специальные дни (-1, -2)
			if days[-1] && day == lastDay {
				return date
			}
			if days[-2] && day == lastDay-1 {
				return date
			}

			// Проверяем обычные дни
			if days[day] {
				return date
			}

			// Обработка дней > lastDay (например, "m 31")
			if day == lastDay {
				for d := range days {
					if d > lastDay {
						nextMonth := time.Date(date.Year(), date.Month()+1, 1, 0, 0, 0, 0, time.UTC)
						lastDayNextMonth := time.Date(nextMonth.Year(), nextMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
						if d <= lastDayNextMonth {
							return time.Date(nextMonth.Year(), nextMonth.Month(), d, 0, 0, 0, 0, time.UTC)
						}
						return time.Date(nextMonth.Year(), nextMonth.Month(), lastDayNextMonth, 0, 0, 0, 0, time.UTC)
					}
				}
			}
		}
	}
}
