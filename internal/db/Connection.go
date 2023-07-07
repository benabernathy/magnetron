package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func GetDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	return db, err
}

func InitDB(db *gorm.DB) error {
	if err := db.AutoMigrate(&RegisteredServer{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&StaticServer{}); err != nil {
		return err
	}

	return nil

}
