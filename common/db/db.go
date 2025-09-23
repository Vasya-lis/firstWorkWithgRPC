package common

import (
	"fmt"
	"log"

	cfg "github.com/Vasya-lis/firstWorkWithgRPC/config"
	md "github.com/Vasya-lis/firstWorkWithgRPC/services/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

// Init открывает базу данных и при необходимости создает таблицу и индекс
func InitDB() error {
	var err error

	config, err := cfg.NewConfig()
	if err != nil {
		log.Printf("configuration error: %v", err)
		return fmt.Errorf("configuration failed: %w", err)
	}

	// строка подключения
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort, config.DBSSLMode)

	// открываю postgres через GORM
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// создаю таблицу и индекс, если их нет
	if err := db.AutoMigrate(&md.Task{}); err != nil {
		return fmt.Errorf("failed to migrate schema: %w", err)
	}
	return nil
}
func GetDB() *gorm.DB {
	return db
}
