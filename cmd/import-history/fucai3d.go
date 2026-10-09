package main

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"path/filepath"
	"shuangseqiu/backend/models"
	"shuangseqiu/backend/service"
	"shuangseqiu/config"
)

func runFucai3DImport(ctx context.Context, file string, validateOnly bool, saveSource string) error {
	var records []models.Fucai3DRecord
	var err error
	if file != "" {
		f, openErr := os.Open(file)
		if openErr != nil {
			return openErr
		}
		defer f.Close()
		records, err = service.ParseFucai3DHistory(f)
	} else {
		var body []byte
		records, body, err = service.NewFucai3DSyncService(nil, config.AppConfig.Fucai3DSourceURL).FetchHistorySnapshot(ctx)
		if err == nil && saveSource != "" {
			if err = os.MkdirAll(filepath.Dir(saveSource), 0755); err == nil {
				err = os.WriteFile(saveSource, body, 0644)
			}
		}
	}
	if err != nil {
		return err
	}
	log.Printf("福彩3D历史数据校验通过: %d 期", len(records))
	if validateOnly {
		return nil
	}
	if !config.AppConfig.Database.Enabled {
		return fmt.Errorf("历史导入需要启用数据库")
	}
	if err := config.InitDB(); err != nil {
		return err
	}
	config.DB = config.DB.Session(&gorm.Session{Logger: config.DB.Logger.LogMode(logger.Error)})
	sqlDB, err := config.DB.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	if config.AppConfig.Database.AutoMigrate {
		if err := config.DB.WithContext(ctx).Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='福彩3D开奖明细表'").AutoMigrate(&models.Fucai3DRecord{}); err != nil {
			return err
		}
	}
	if err := (&models.Fucai3DRecordStore{DB: config.DB}).UpsertHistory(ctx, records); err != nil {
		return err
	}
	issues := make([]string, len(records))
	for i := range records {
		issues[i] = records[i].Issue
	}
	var imported, total int64
	if err := config.DB.WithContext(ctx).Model(&models.Fucai3DRecord{}).Where("issue IN ?", issues).Count(&imported).Error; err != nil {
		return err
	}
	if imported != int64(len(records)) {
		return fmt.Errorf("福彩3D导入核验失败: 期望 %d 期，数据库找到 %d 期", len(records), imported)
	}
	if err := config.DB.WithContext(ctx).Model(&models.Fucai3DRecord{}).Count(&total).Error; err != nil {
		return err
	}
	log.Printf("福彩3D历史导入成功并核验: 本次 %d 期，表内 %d 期", imported, total)
	return nil
}
