package cron

import (
	"testing"
	"time"

	"shuangseqiu/utils"
)

func TestSSQScheduleUsesChinaTimezone(t *testing.T) {
	manager := NewManager()
	if manager.Location() != utils.ChinaTimezone {
		t.Fatal("调度器必须使用固定东八区")
	}
	id, err := manager.AddFunc(SSQSyncSpec, func() {})
	if err != nil {
		t.Fatal(err)
	}
	schedule := manager.Entry(id).Schedule
	cases := []struct{ from, want string }{
		{"2026-10-08T21:59:59Z", "2026-10-09T06:00:00+08:00"},
		{"2026-10-09T22:00:00Z", "2026-10-12T06:00:00+08:00"},
		{"2026-10-12T06:00:00+08:00", "2026-10-14T06:00:00+08:00"},
		{"2026-10-14T06:00:00+08:00", "2026-10-16T06:00:00+08:00"},
	}
	for _, tc := range cases {
		from, _ := time.Parse(time.RFC3339, tc.from)
		// Cron 将当前时间转换到其 Location 后计算下一次执行时间。
		got := schedule.Next(from.In(manager.Location())).In(utils.ChinaTimezone).Format(time.RFC3339)
		if got != tc.want {
			t.Errorf("Next(%s)=%s，预期 %s", tc.from, got, tc.want)
		}
	}
}
