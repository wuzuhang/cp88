package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"shuangseqiu/backend/service"
	"shuangseqiu/config"
	"shuangseqiu/pkg/response"
	"sync"
	"time"
)

var ssqManualSync sync.Mutex
var fucai3DManualSync sync.Mutex

func SyncLatestSSQ(c *gin.Context) {
	if !syncDatabaseReady(c) {
		return
	}
	runLatestSync(c, &ssqManualSync, service.NewDrawSyncService(config.DB, config.AppConfig.SSQSourceURL).SyncLatest)
}

func SyncLatestFucai3D(c *gin.Context) {
	if !syncDatabaseReady(c) {
		return
	}
	runLatestSync(c, &fucai3DManualSync, service.NewFucai3DSyncService(config.DB, config.AppConfig.Fucai3DSourceURL).SyncLatest)
}

func syncDatabaseReady(c *gin.Context) bool {
	if config.DB == nil || config.AppConfig == nil {
		response.JSON(c, http.StatusServiceUnavailable, "数据库未启用，无法更新开奖数据", nil)
		return false
	}
	return true
}

func runLatestSync(c *gin.Context, guard *sync.Mutex, update func(context.Context) (string, error)) {
	if !guard.TryLock() {
		response.JSON(c, http.StatusConflict, "正在更新开奖数据，请稍后刷新", nil)
		return
	}
	defer guard.Unlock()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()
	issue, err := update(ctx)
	if err != nil {
		log.Printf("手动更新开奖失败: %v", err)
		response.JSON(c, http.StatusBadGateway, "更新失败，请稍后重试或检查数据源与数据库连接", nil)
		return
	}
	response.JSON(c, http.StatusOK, "最新开奖已更新", gin.H{"issue": issue})
}
