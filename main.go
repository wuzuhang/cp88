package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"shuangseqiu/backend/models"
	"shuangseqiu/backend/service"
	"shuangseqiu/config"
	"shuangseqiu/cron"
	"shuangseqiu/router"
)

func run() error {
	if err := config.InitConfig(); err != nil {
		return err
	}
	gin.SetMode(config.AppConfig.Server.Mode)
	if err := config.InitDB(); err != nil {
		return err
	}
	if config.DB != nil {
		sqlDB, err := config.DB.DB()
		if err != nil {
			return err
		}
		defer sqlDB.Close()
		if config.AppConfig.Database.AutoMigrate {
			if err := config.DB.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='双色球开奖明细表'").AutoMigrate(&models.SSQRecord{}); err != nil {
				return fmt.Errorf("迁移开奖表: %w", err)
			}
			if err := config.DB.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='福彩3D开奖明细表'").AutoMigrate(&models.Fucai3DRecord{}); err != nil {
				return fmt.Errorf("迁移福彩3D开奖表: %w", err)
			}
		}
	}
	manager := cron.NewManager()
	if config.DB != nil {
		if err := cron.RegisterSSQSync(manager, service.NewDrawSyncService(config.DB, config.AppConfig.SSQSourceURL)); err != nil {
			return fmt.Errorf("注册开奖同步任务: %w", err)
		}
		if err := cron.RegisterFucai3DSync(manager, service.NewFucai3DSyncService(config.DB, config.AppConfig.Fucai3DSourceURL)); err != nil {
			return fmt.Errorf("注册福彩3D开奖同步任务: %w", err)
		}
	} else {
		log.Println("数据库未启用，跳过双色球开奖同步任务")
	}
	manager.Start()
	defer func() { <-manager.Stop().Done() }()
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", config.AppConfig.Server.Port), Handler: router.InitRouter(),
		ReadTimeout: 15 * time.Second, ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	log.Printf("双色球分析后端启动，端口 %d", config.AppConfig.Server.Port)
	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			_ = srv.Close()
			return err
		}
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
