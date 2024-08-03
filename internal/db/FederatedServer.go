package db

import (
	"gorm.io/gorm"
	"strconv"
	"time"
)

type FederatedServer struct {
	gorm.Model
	ID          string `gorm:"primaryKey"`
	Host        string
	Port        uint16
	Name        string
	Description string
	UserCount   uint16
	TrackerHost string
	LastSeen    time.Time
	FirstSeen   time.Time
}

func (*FederatedServer) GenerateId(trackerAddress string, serverHost string, serverPort uint16) string {

	return trackerAddress + ":" + serverHost + ":" + strconv.Itoa(int(serverPort))
}
