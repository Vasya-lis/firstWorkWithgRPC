package main

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

// Init открывает базу данных и при необходимости создает таблицу и индекс
func Init(dbFile string) error {
	var err error

	// открываю sqlite через GORM
	db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
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
