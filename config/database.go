package config

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	if !AppConfig.Database.Enabled {
		return nil
	}
	if AppConfig.Database.Driver != "mysql" {
		return fmt.Errorf("不支持的数据库驱动: %s", AppConfig.Database.Driver)
	}
	if AppConfig.Database.DSN == "" {
		return fmt.Errorf("启用数据库后必须设置 database.dsn 或 SSQ_DATABASE_DSN")
	}
	db, err := gorm.Open(mysql.Open(AppConfig.Database.DSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接数据库: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)
	DB = db
	return nil
}
