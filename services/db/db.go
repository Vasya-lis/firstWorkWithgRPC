package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Init открывает базу данных и при необходимости создает таблицу и индекс
func initDB(dsn string) error {
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
