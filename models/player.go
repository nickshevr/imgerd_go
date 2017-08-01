package models

import "time"

type Player struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	CurrentBalance uint `json:"currentBalance"`
}
