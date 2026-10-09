package cron

import (
	"log"

	"shuangseqiu/utils"

	cronlib "github.com/robfig/cron/v3"
)

// NewManager 使用东八区和秒级表达式，恢复异常并防止单个任务重叠执行。
func NewManager() *cronlib.Cron {
	logger := cronlib.PrintfLogger(log.Default())
	return cronlib.New(cronlib.WithLocation(utils.ChinaTimezone), cronlib.WithSeconds(), cronlib.WithChain(
		cronlib.Recover(logger), cronlib.SkipIfStillRunning(logger),
	))
}
