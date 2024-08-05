package db

import (
	"gorm.io/gorm"
)

type StaticServer struct {
	gorm.Model
	Host        string `gorm:"primaryKey"`
	Port        uint16 `gorm:"primaryKey;autoincrement:false"`
	Name        string
	Description string
	UserCount   uint16
	ServerOrder uint16
}

type StaticServerStore struct {
	db *gorm.DB
}

func NewStaticServerStore(db *gorm.DB) (*StaticServerStore, error) {
	if err := db.AutoMigrate(&StaticServer{}); err != nil {
		return nil, err
	}

	return &StaticServerStore{db}, nil
}

func (s *StaticServerStore) RegisterStaticServer(host string, port uint16, name string, description string, userCount uint16, order uint16) (StaticServer, error) {
	server := StaticServer{
		Host:        host,
		Port:        port,
		Name:        name,
		Description: description,
		UserCount:   userCount,
		ServerOrder: order,
	}

	return server, s.db.Create(&server).Error
}

func (s *StaticServerStore) GetStaticServers() ([]StaticServer, error) {
	var servers []StaticServer
	if err := s.db.Find(&servers).Order("server_order").Error; err != nil {
		return nil, err
	}

	return servers, nil
}
