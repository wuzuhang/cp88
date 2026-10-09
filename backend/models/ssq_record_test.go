package models

import (
	"context"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestSSQUpsertSQL(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: "user:pass@tcp(localhost:3306)/test", SkipInitializeWithVersion: true}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatal(err)
	}
	var query string
	if err := db.Callback().Create().After("gorm:create").Register("test:capture", func(tx *gorm.DB) { query = tx.Statement.SQL.String() }); err != nil {
		t.Fatal(err)
	}
	if err := (&SSQRecordStore{DB: db}).Upsert(context.Background(), &SSQRecord{Issue: "2026115"}); err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(query, "ON DUPLICATE KEY UPDATE")
	if len(parts) != 2 {
		t.Fatalf("SQL 缺少重复期号更新: %s", query)
	}
	for _, field := range []string{"draw_date", "red_1", "blue", "seq_red_6", "total_sales", "pool_amount", "prize1_count", "prize6_amount"} {
		if !strings.Contains(parts[1], "`"+field+"`") {
			t.Errorf("更新缺少列 %s: %s", field, query)
		}
	}
	if !strings.Contains(parts[0], "`ssq_records`") || strings.Contains(parts[1], "`issue`=") || strings.Contains(query, "created_at") || strings.Contains(query, "updated_at") {
		t.Fatalf("表或列不符合要求: %s", query)
	}
}

func TestDatePreservesCalendarDay(t *testing.T) {
	for _, input := range []any{"2026-10-08", []byte("2026-10-08"), time.Date(2026, 10, 8, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))} {
		var date Date
		if err := date.Scan(input); err != nil {
			t.Fatal(err)
		}
		value, err := date.Value()
		if err != nil || value != "2026-10-08" {
			t.Fatalf("日期偏移: %v %v", value, err)
		}
	}
	var invalid Date
	if err := invalid.Scan("2026-02-30"); err == nil {
		t.Fatal("接受了非法日期")
	}
}
