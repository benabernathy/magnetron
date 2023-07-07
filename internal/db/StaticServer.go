package db

import (
	"gorm.io/gorm"
)

type StaticServer struct {
	gorm.Model
	Host        string
	Port        uint16
	Name        string
	Description string
	UserCount   uint16
}
