package models

type Task struct {
	ID      int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Date    string `gorm:"size:8;not null;default:''" json:"date"`
	Title   string `gorm:"size:255;not null;default:''" json:"title"`
	Comment string `gorm:"not null;default:''" json:"comment"`
	Repeat  string `gorm:"size:128;not null;default:''" json:"repeat"`
}
