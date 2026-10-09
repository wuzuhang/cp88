package cron

import (
	"context"
	cronlib "github.com/robfig/cron/v3"
	"log"
	"shuangseqiu/backend/service"
	"time"
)

const Fucai3DSyncSpec = "0 0 6 * * *"

func RegisterFucai3DSync(manager *cronlib.Cron, syncService *service.Fucai3DSyncService) error {
	_, err := manager.AddFunc(Fucai3DSyncSpec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		issue, err := syncService.SyncPreviousDay(ctx, time.Now())
		if err != nil {
			log.Printf("[福彩3D Sync] 同步失败: %v", err)
			return
		}
		log.Printf("[福彩3D Sync] 同步成功，期号 %s", issue)
	})
	if err == nil {
		log.Println("[福彩3D Sync] 已注册：每天东八区 06:00 抓取前一日开奖")
	}
	return err
}
