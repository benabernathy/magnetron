package db

import (
	"gorm.io/gorm"
	"time"
)

type FederatedServerId struct {
	TrackerHost string `gorm:"primaryKey"`
	TrackerPort uint16 `gorm:"primaryKey;autoincrement:false"`
	Host        string `gorm:"primaryKey"`
	Port        uint16 `gorm:"primaryKey;autoincrement:false"`
}

type FederatedServer struct {
	gorm.Model
	FederatedServerId `gorm:"embedded"`
	Name              string
	Description       string
	UserCount         uint16
	ServerOrder       uint16
	LastSeen          time.Time
	FirstSeen         time.Time
}

type FederatedServerStore struct {
	db *gorm.DB
}

func NewFederatedServerStore(db *gorm.DB) (*FederatedServerStore, error) {
	if err := db.AutoMigrate(&FederatedServer{}); err != nil {
		return nil, err
	}

	return &FederatedServerStore{db}, nil
}

func (s *FederatedServerStore) RegisterFederatedServer(trackerHost string, trackerPort uint16, host string, port uint16, name string, description string, userCount uint16, order uint16) (FederatedServer, error) {
	server := FederatedServer{
		FederatedServerId: FederatedServerId{
			TrackerHost: trackerHost,
			TrackerPort: trackerPort,
			Host:        host,
			Port:        port,
		},
		Name:        name,
		Description: description,
		UserCount:   userCount,
		ServerOrder: order,
		FirstSeen:   time.Now(),
		LastSeen:    time.Now(),
	}

	return server, s.db.Save(&server).Error
}

func (s *FederatedServerStore) UpdateFederatedServer(trackerHost string, trackerPort uint16, host string, port uint16, name string, description string, userCount uint16, order uint16) error {
	updatedServer := FederatedServer{
		Name:        name,
		Description: description,
		UserCount:   userCount,
		ServerOrder: order,
		LastSeen:    time.Now(),
	}
	return s.db.Model(&FederatedServer{}).Where("tracker_host = ? AND tracker_port = ? AND host = ? AND port = ?", trackerHost, trackerPort, host, port).Updates(updatedServer).Error
}

func (s *FederatedServerStore) GetFederatedServer(trackerHost string, trackerPort uint16, host string, port uint16) (FederatedServer, error) {
	var server FederatedServer
	err := s.db.Where("tracker_host = ? AND tracker_port = ? AND host = ? AND port = ?", trackerHost, trackerPort, host, port).First(&server).Error
	return server, err
}

func (s *FederatedServerStore) GetFederatedServers(trackerHost string, trackerPort uint16) ([]FederatedServer, error) {
	var servers []FederatedServer
	err := s.db.Where("tracker_host = ? AND tracker_port = ?", trackerHost, trackerPort).Find(&servers).Order("server_order").Error
	return servers, err
}

func (s *FederatedServerStore) ExpireFederatedServers(expiration time.Duration) error {
	return s.db.Where("last_seen < ?", time.Now().Add(-expiration)).Delete(&FederatedServer{}).Error
}
