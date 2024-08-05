package db

import (
	"gorm.io/gorm"
	"time"
)

type FederatedTrackerId struct {
	Host string `gorm:"primaryKey"`
	Port uint16 `gorm:"primaryKey;autoincrement:false"`
}

type FederatedTracker struct {
	gorm.Model
	FederatedTrackerId
	Name         string `gorm:"not null"`
	Description  string
	UserCount    uint16 `gorm:"not null"`
	FirstSeen    time.Time
	LastSeen     time.Time
	TrackerOrder uint16 `gorm:"not null"`
}

type FederatedTrackerStore struct {
	db *gorm.DB
}

func NewFederatedTrackerStore(db *gorm.DB) (*FederatedTrackerStore, error) {

	if err := db.AutoMigrate(&FederatedTracker{}); err != nil {
		return nil, err
	}

	return &FederatedTrackerStore{db}, nil
}

func (s *FederatedTrackerStore) RegisterFederatedTracker(host string, port uint16, name string, description string, userCount uint16, order uint16) (FederatedTracker, error) {

	tracker := FederatedTracker{
		FederatedTrackerId: FederatedTrackerId{
			Host: host,
			Port: port,
		},
		Name:         name,
		Description:  description,
		UserCount:    userCount,
		TrackerOrder: order,
	}

	return tracker, s.db.Create(&tracker).Error
}

func (s *FederatedTrackerStore) UpdateFirstSeen(host string, port uint16) error {

	return s.db.Model(&FederatedTracker{}).Where("host = ? AND port = ?", host, port).Update("first_seen", time.Now()).Error
}

func (s *FederatedTrackerStore) UpdateLastSeen(host string, port uint16) error {

	return s.db.Model(&FederatedTracker{}).Where("host = ? AND port = ?", host, port).Update("last_seen", time.Now()).Error
}

func (s *FederatedTrackerStore) GetFederatedTracker(host string, port uint16) (FederatedTracker, error) {

	var tracker FederatedTracker

	err := s.db.Where("host = ? AND port = ?", host, port).First(&tracker).Error

	return tracker, err
}

func (s *FederatedTrackerStore) GetFederatedTrackers() ([]FederatedTracker, error) {

	var trackers []FederatedTracker

	err := s.db.Find(&trackers).Order("tracker_order").Error

	return trackers, err
}
