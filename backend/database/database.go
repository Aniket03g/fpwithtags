package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func InitDB() (*Database, error) {
	db, err := gorm.Open(sqlite.Open("test.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}

func (d *Database) Migrate(models ...interface{}) error {
	return d.DB.AutoMigrate(models...)
}
