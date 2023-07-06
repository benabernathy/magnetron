package db

import (
	"gorm.io/gorm"
	"time"
)

type RegisteredServer struct {
	PassID      uint32 `gorm:"primaryKey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Host        string
	Port        uint16
	Name        string
	Description string
	UserCount   uint16
	LastSeen    time.Time
	FirstSeen   time.Time
}
