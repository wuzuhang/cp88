package utils

import "time"

// ChinaTimezone 固定东八区，不依赖主机时区或系统时区数据库。
var ChinaTimezone = time.FixedZone("Asia/Shanghai", 8*60*60)
