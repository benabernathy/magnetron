package db

import (
	"fmt"
	"gorm.io/gorm"
	"log"
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

type RegisteredServerStore struct {
	db *gorm.DB
}

func NewRegisteredServerStore(db *gorm.DB) (*RegisteredServerStore, error) {
	if err := db.AutoMigrate(&RegisteredServer{}); err != nil {
		return nil, err
	}

	return &RegisteredServerStore{db}, nil
}

func (r *RegisteredServerStore) RegisterNewServer(passID uint32, host string, port uint16, name string, description string, userCount uint16) error {

	server := RegisteredServer{
		PassID:      passID,
		Host:        host,
		Port:        port,
		Name:        name,
		Description: description,
		UserCount:   userCount,
		LastSeen:    time.Now(),
		FirstSeen:   time.Now(),
	}

	var count int64
	existingServer := RegisteredServer{}
	r.db.First(&existingServer, passID).Count(&count)

	if createError := r.db.Save(&server).Error; createError != nil {
		return fmt.Errorf("could not register server because of an internal error: %s", createError)
	} else if count == 0 {
		log.Printf("Registered new server: %s (%s:%d)", name, host, port)
	}

	return nil
}

func (r *RegisteredServerStore) GetAllRegisteredServers() ([]RegisteredServer, error) {
	var servers []RegisteredServer
	if err := r.db.Find(&servers).Error; err != nil {
		return nil, err
	}

	return servers, nil
}

func (r *RegisteredServerStore) RemoveExpiredServers(expirationTime time.Duration) {
	var registeredServers []RegisteredServer

	r.db.Find(&registeredServers)

	for _, server := range registeredServers {
		if time.Since(server.LastSeen).Minutes() > expirationTime.Minutes() {
			r.db.Delete(&server).Commit()
			log.Printf("Removed expired server %s", server.Name)
		}
	}
}
