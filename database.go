package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	ensureDataDir()

	var err error
	DB, err = gorm.Open(sqlite.Open(getDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	// 自动迁移
	DB.AutoMigrate(
		&User{},
		&Role{},
		&Permission{},
		&RolePermission{},
		&UserPermission{},
		&LandContract{},
		&LandPlot{},
	)
}
