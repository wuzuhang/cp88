package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHistoryUsesConfiguredSourceURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/custom-history.txt" {
			t.Errorf("unexpected source path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(drawFixture))
	}))
	defer server.Close()
	records, err := NewDrawSyncService(nil, server.URL+"/custom-history.txt").FetchHistory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].Issue != "2026115" {
		t.Fatalf("unexpected records: %+v", records)
	}
}

func TestParseHistory(t *testing.T) {
	first := strings.ReplaceAll(strings.ReplaceAll(drawFixture, "2026115", "2026114"), "2026-10-08", "2026-10-06")
	records, err := ParseHistory(strings.NewReader("\uFEFF" + first + "\r\n\r\n" + drawFixture + "\r\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || records[0].Issue != "2026114" || records[1].Issue != "2026115" {
		t.Fatalf("历史解析错误: %+v", records)
	}
}

func TestParseHistoryRejectsWholeInvalidFile(t *testing.T) {
	for _, body := range []string{"", drawFixture + "\n" + drawFixture, drawFixture + "\n2026116 2026-10-09 invalid"} {
		records, err := ParseHistory(strings.NewReader(body))
		if err == nil || records != nil {
			t.Fatalf("非法历史文件应整体失败: %v %v", records, err)
		}
	}
}
