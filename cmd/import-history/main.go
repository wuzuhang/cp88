package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"shuangseqiu/backend/models"
	"shuangseqiu/backend/service"
	"shuangseqiu/config"
)

func run() error {
	game := flag.String("game", "ssq", "玩法：ssq 或 fucai3d")
	file := flag.String("file", "", "从本地数据文件导入，默认在线下载")
	validateOnly := flag.Bool("validate-only", false, "只校验历史数据，不连接数据库")
	saveSource := flag.String("save-source", "", "下载校验后保存原始历史文件")
	flag.Parse()
	if *game != "ssq" && *game != "fucai3d" {
		return fmt.Errorf("game 必须为 ssq 或 fucai3d")
	}
	if *file == "" || !*validateOnly {
		if err := config.InitConfig(); err != nil {
			return err
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if *game == "fucai3d" {
		return runFucai3DImport(ctx, *file, *validateOnly, *saveSource)
	}
	var records []models.SSQRecord
	var err error
	if *file != "" {
		f, openErr := os.Open(*file)
		if openErr != nil {
			return openErr
		}
		defer f.Close()
		records, err = service.ParseHistory(f)
	} else {
		var body []byte
		records, body, err = service.NewDrawSyncService(nil, config.AppConfig.SSQSourceURL).FetchHistorySnapshot(ctx)
		if err == nil && *saveSource != "" {
			if err = os.MkdirAll(filepath.Dir(*saveSource), 0755); err == nil {
				err = os.WriteFile(*saveSource, body, 0644)
			}
			if err == nil {
				log.Printf("原始历史文件已保存: %s", *saveSource)
			}
		}
	}
	if err != nil {
		return err
	}
	log.Printf("历史数据校验通过: %d 期", len(records))
	if *validateOnly {
		return nil
	}
	if !config.AppConfig.Database.Enabled {
		return fmt.Errorf("历史导入需要启用数据库")
	}
	if err := config.InitDB(); err != nil {
		return err
	}
	// 批量导入只输出 SQL 错误，避免慢查询日志打印整批历史号码。
	config.DB = config.DB.Session(&gorm.Session{Logger: config.DB.Logger.LogMode(logger.Error)})
	sqlDB, err := config.DB.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	if config.AppConfig.Database.AutoMigrate {
		if err := config.DB.WithContext(ctx).Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='双色球开奖明细表'").AutoMigrate(&models.SSQRecord{}); err != nil {
			return err
		}
	}
	if err := (&models.SSQRecordStore{DB: config.DB}).UpsertHistory(ctx, records); err != nil {
		return err
	}
	// 导入后读取数据库确认数量及最早/最新期号。
	var summary struct {
		Count      int64
		FirstIssue string
		LastIssue  string
		FirstDate  string
		LastDate   string
	}
	if err := config.DB.WithContext(ctx).Model(&models.SSQRecord{}).Select("COUNT(*) AS count, MIN(issue) AS first_issue, MAX(issue) AS last_issue, CAST(MIN(draw_date) AS CHAR) AS first_date, CAST(MAX(draw_date) AS CHAR) AS last_date").Scan(&summary).Error; err != nil {
		return err
	}
	var imported int64
	issues := make([]string, len(records))
	for i := range records {
		issues[i] = records[i].Issue
	}
	if err := config.DB.WithContext(ctx).Model(&models.SSQRecord{}).Where("issue IN ?", issues).Count(&imported).Error; err != nil {
		return err
	}
	if imported != int64(len(records)) {
		return fmt.Errorf("导入后核验失败: 期望 %d 期，数据库找到 %d 期", len(records), imported)
	}
	log.Printf("历史导入成功并核验: 本次 %d 期，表内 %d 期；期号 %s–%s；日期 %s–%s", imported, summary.Count, summary.FirstIssue, summary.LastIssue, summary.FirstDate, summary.LastDate)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
