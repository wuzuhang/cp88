package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"shuangseqiu/backend/models"
	"shuangseqiu/utils"
)

// 31 列：最后两列是源文件附加数据，不属于 ssq_records。
const drawFixture = "2026115 2026-10-08 04 13 14 18 22 24 03 22 04 18 24 14 13 337573290 909402631 15 6031127 131 147585 1912 3000 80737 200 1374142 10 10490302 5 0 5"

type fakeRecordStore struct {
	records []*models.SSQRecord
	err     error
}

func (s *fakeRecordStore) Upsert(_ context.Context, record *models.SSQRecord) error {
	if s.err != nil {
		return s.err
	}
	s.records = append(s.records, record)
	return nil
}

func TestParseDrawMapping(t *testing.T) {
	r, err := parseDrawLine(strings.Fields(drawFixture))
	if err != nil {
		t.Fatal(err)
	}
	if r.Issue != "2026115" || r.DrawDate != "2026-10-08" {
		t.Fatalf("期号日期错误: %+v", r)
	}
	got := []int64{int64(r.Red1), int64(r.Red2), int64(r.Red3), int64(r.Red4), int64(r.Red5), int64(r.Red6), int64(r.Blue),
		int64(r.SeqRed1), int64(r.SeqRed2), int64(r.SeqRed3), int64(r.SeqRed4), int64(r.SeqRed5), int64(r.SeqRed6),
		r.TotalSales, r.PoolAmount, int64(r.Prize1Count), r.Prize1Amount, int64(r.Prize2Count), r.Prize2Amount,
		int64(r.Prize3Count), r.Prize3Amount, int64(r.Prize4Count), r.Prize4Amount, int64(r.Prize5Count), r.Prize5Amount, int64(r.Prize6Count), r.Prize6Amount}
	want := []int64{4, 13, 14, 18, 22, 24, 3, 22, 4, 18, 24, 14, 13, 337573290, 909402631, 15, 6031127, 131, 147585, 1912, 3000, 80737, 200, 1374142, 10, 10490302, 5}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("第 %d 列=%d，预期 %d", i+3, got[i], want[i])
		}
	}
	if _, err := parseDrawLine(strings.Fields(drawFixture)[:29]); err != nil {
		t.Fatalf("29 列格式: %v", err)
	}
}

func TestParseDrawRejectsInvalidData(t *testing.T) {
	cases := []struct {
		name   string
		column int
		value  string
	}{
		{"issue", 0, "invalid"}, {"date", 1, "2026-02-30"}, {"duplicate red", 3, "04"},
		{"red range", 7, "34"}, {"blue range", 8, "17"}, {"sequence mismatch", 9, "33"},
		{"negative sales", 15, "-1"}, {"invalid amount", 18, "oops"}, {"count overflow", 17, "2147483648"},
		{"amount overflow", 18, "9223372036854775808"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fields := strings.Fields(drawFixture)
			fields[tc.column] = tc.value
			if _, err := parseDrawLine(fields); err == nil {
				t.Fatal("接受了错误数据")
			}
		})
	}
	if _, err := parseDrawLine(strings.Fields(drawFixture)[:28]); err == nil {
		t.Fatal("接受了缺列数据")
	}
}

func TestSyncUsesPreviousChinaDate(t *testing.T) {
	// UTC 10 月 8 日 22 点 = 东八区 10 月 9 日 6 点。
	now := time.Date(2026, 10, 8, 22, 0, 0, 0, time.UTC)
	nextDay := strings.ReplaceAll(drawFixture, "2026-10-08", "2026-10-09")
	store := &fakeRecordStore{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Error("请求应使用 GET")
		}
		// 文件最后一行故意不是目标日期，并使用 BOM、CRLF 和空行。
		_, _ = w.Write([]byte("\uFEFF" + drawFixture + "\r\n\r\n" + nextDay + "\r\n"))
	}))
	defer server.Close()
	s := &DrawSyncService{client: server.Client(), sourceURL: server.URL, store: store}
	issue, err := s.SyncPreviousDay(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}
	if issue != "2026115" || len(store.records) != 1 || store.records[0].DrawDate != "2026-10-08" {
		t.Fatalf("错误入库: %+v", store.records)
	}
}

func TestSyncFailuresDoNotWrite(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"http error", 502, drawFixture},
		{"missing yesterday", 200, strings.ReplaceAll(drawFixture, "2026-10-08", "2026-10-07")},
		{"html", 200, "<html>error</html>"},
		{"malformed target", 200, "2026115 2026-10-08 04"},
		{"ambiguous target", 200, drawFixture + "\n" + drawFixture},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeRecordStore{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			s := &DrawSyncService{client: server.Client(), sourceURL: server.URL, store: store}
			_, err := s.SyncPreviousDay(context.Background(), time.Date(2026, 10, 9, 6, 0, 0, 0, utils.ChinaTimezone))
			if err == nil || len(store.records) != 0 {
				t.Fatalf("失败不应入库: err=%v count=%d", err, len(store.records))
			}
		})
	}
}

func TestSyncStoreError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(drawFixture)) }))
	defer server.Close()
	want := errors.New("database unavailable")
	s := &DrawSyncService{client: server.Client(), sourceURL: server.URL, store: &fakeRecordStore{err: want}}
	if _, err := s.SyncPreviousDay(context.Background(), time.Date(2026, 10, 9, 6, 0, 0, 0, utils.ChinaTimezone)); !errors.Is(err, want) {
		t.Fatalf("入库错误未返回: %v", err)
	}
}

func TestPreviousChinaDateAtYearBoundary(t *testing.T) {
	line := strings.ReplaceAll(drawFixture, "2026-10-08", "2025-12-31")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(line)) }))
	defer server.Close()
	store := &fakeRecordStore{}
	s := &DrawSyncService{client: server.Client(), sourceURL: server.URL, store: store}
	// UTC 12 月 31 日 16:30 已经是东八区 1 月 1 日。
	if _, err := s.SyncPreviousDay(context.Background(), time.Date(2025, 12, 31, 16, 30, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if store.records[0].DrawDate != "2025-12-31" {
		t.Fatal("跨年日期错误")
	}
}
