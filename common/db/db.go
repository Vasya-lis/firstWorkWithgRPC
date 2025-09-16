package common

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

type Task struct {
	ID      int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Date    string `gorm:"size:8;not null;default:''" json:"date"`
	Title   string `gorm:"size:255;not null;default:''" json:"title"`
	Comment string `gorm:"not null;default:''" json:"comment"`
	Repeat  string `gorm:"size:128;not null;default:''" json:"repeat"`
}

// Init открывает базу данных и при необходимости создает таблицу и индекс
func InitDB(dsn string) error {
	var err error

	// открываю postgres через GORM
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// создаю таблицу и индекс, если их нет
	if err := db.AutoMigrate(&Task{}); err != nil {
		return fmt.Errorf("failed to migrate schema: %w", err)
	}
	return nil
}
func GetDB() *gorm.DB {
	return db
}
