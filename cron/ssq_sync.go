package cron

import (
	"context"
	"log"
	"time"

	cronlib "github.com/robfig/cron/v3"
	"shuangseqiu/backend/service"
)

// SSQSyncSpec 每周一、三、五东八区 06:00:00 执行。
const SSQSyncSpec = "0 0 6 * * 1,3,5"

func RegisterSSQSync(manager *cronlib.Cron, syncService *service.DrawSyncService) error {
	_, err := manager.AddFunc(SSQSyncSpec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		issue, err := syncService.SyncPreviousDay(ctx, time.Now())
		if err != nil {
			log.Printf("[SSQ Sync] 同步失败: %v", err)
			return
		}
		log.Printf("[SSQ Sync] 同步成功，期号 %s", issue)
	})
	if err == nil {
		log.Println("[SSQ Sync] 已注册：每周一、三、五东八区 06:00 抓取前一日开奖")
	}
	return err
}
