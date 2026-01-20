package models

import (
	"time"
)

type NonWorkingDay struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Date      time.Time `gorm:"uniqueIndex" json:"date"`
	Year      int       `gorm:"index" json:"year"`
	Month     int       `gorm:"index" json:"month"`
	Day       int       `json:"day"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}