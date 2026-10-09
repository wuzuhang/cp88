package cron

import (
	"testing"
	"time"
)

func TestFucai3DDailySchedule(t *testing.T) {
	manager := NewManager()
	id, err := manager.AddFunc(Fucai3DSyncSpec, func() {})
	if err != nil {
		t.Fatal(err)
	}
	from, _ := time.Parse(time.RFC3339, "2026-10-09T06:00:00+08:00")
	got := manager.Entry(id).Schedule.Next(from.In(manager.Location())).Format(time.RFC3339)
	if got != "2026-10-10T06:00:00+08:00" {
		t.Fatalf("每日东八区 6 点错误: %s", got)
	}
}
