package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"shuangseqiu/backend/models"
	"strings"
	"testing"
	"time"
)

const fucai3DFixture = "2026267 2026-10-08 0 5 0 9 8 4 1 2 99125960 11 1040 22 346 33 173"

func TestFucai3DParseMappingAndValidation(t *testing.T) {
	r, err := parseFucai3DLine(strings.Fields(fucai3DFixture))
	if err != nil {
		t.Fatal(err)
	}
	want := &models.Fucai3DRecord{Issue: "2026267", DrawDate: "2026-10-08", Num1: 0, Num2: 5, Num3: 0, TestNum1: 9, TestNum2: 8, TestNum3: 4, Machine: 1, BallSet: 2, TotalSales: 99125960, DirectCount: 11, DirectAmount: 1040, Group3Count: 22, Group3Amount: 346, Group6Count: 33, Group6Amount: 173}
	if !reflect.DeepEqual(r, want) {
		t.Fatalf("字段映射错误: %+v", r)
	}
	for _, tc := range []struct {
		column int
		value  string
	}{{0, "abc1234"}, {1, "2026-02-30"}, {2, "10"}, {3, "-1"}, {5, "10"}, {8, "128"}, {10, "-1"}, {11, "2147483648"}, {12, "9223372036854775808"}} {
		fields := strings.Fields(fucai3DFixture)
		fields[tc.column] = tc.value
		if _, err := parseFucai3DLine(fields); err == nil {
			t.Errorf("接受非法第 %d 列: %s", tc.column, tc.value)
		}
	}
	for _, body := range []string{"", fucai3DFixture + "\n" + fucai3DFixture, fucai3DFixture + "\ninvalid"} {
		if records, err := ParseFucai3DHistory(strings.NewReader(body)); err == nil || records != nil {
			t.Fatalf("非法文件应整体失败: %v %v", records, err)
		}
	}
	if records, err := ParseFucai3DHistory(strings.NewReader("\uFEFF" + fucai3DFixture + "\r\n\r\n")); err != nil || len(records) != 1 {
		t.Fatalf("BOM/CRLF解析失败: %v", err)
	}
}

type fakeFucai3DStore struct {
	records []*models.Fucai3DRecord
	err     error
}

func (s *fakeFucai3DStore) Upsert(_ context.Context, r *models.Fucai3DRecord) error {
	if s.err != nil {
		return s.err
	}
	s.records = append(s.records, r)
	return nil
}

func TestFucai3DSyncPreviousChinaDayAndFailures(t *testing.T) {
	now := time.Date(2026, 10, 8, 22, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name, body string
		status     int
		storeError error
		success    bool
	}{
		{"previous day", fucai3DFixture + "\n" + strings.NewReplacer("2026267", "2026268", "2026-10-08", "2026-10-09").Replace(fucai3DFixture), 200, nil, true},
		{"not available", strings.ReplaceAll(fucai3DFixture, "2026-10-08", "2026-10-07"), 200, nil, false},
		{"duplicate day", fucai3DFixture + "\n" + strings.ReplaceAll(fucai3DFixture, "2026267", "2026266"), 200, nil, false},
		{"invalid number", strings.ReplaceAll(fucai3DFixture, "0 5 0", "10 5 0"), 200, nil, false},
		{"http error", fucai3DFixture, 502, nil, false},
		{"database error", fucai3DFixture, 200, errors.New("db unavailable"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			s := NewFucai3DSyncService(nil, server.URL)
			store := &fakeFucai3DStore{err: tc.storeError}
			s.store = store
			issue, err := s.SyncPreviousDay(context.Background(), now)
			if tc.success {
				if err != nil || issue != "2026267" || len(store.records) != 1 || store.records[0].DrawDate != "2026-10-08" {
					t.Fatalf("目标日期错误: %s %v %+v", issue, err, store.records)
				}
			} else if err == nil || len(store.records) != 0 {
				t.Fatalf("失败不应入库: %v %+v", err, store.records)
			}
		})
	}
}

func TestFucai3DAnalysis(t *testing.T) {
	for _, tc := range []struct {
		numbers   []int
		pattern   string
		sum, span int
	}{
		{[]int{0, 0, 0}, "豹子", 0, 0}, {[]int{0, 5, 0}, "组三", 5, 5}, {[]int{9, 0, 5}, "组六", 14, 9},
	} {
		r, err := AnalyzeFucai3D(tc.numbers)
		if err != nil || r.Pattern != tc.pattern || r.Sum != tc.sum || r.Span != tc.span || !reflect.DeepEqual(r.Numbers, tc.numbers) || r.OddCount+r.EvenCount != 3 || r.BigCount+r.SmallCount != 3 {
			t.Fatalf("分析错误: %+v %v", r, err)
		}
	}
	for _, numbers := range [][]int{{0, 1}, {0, 1, 10}, {-1, 0, 0}, nil} {
		if _, err := AnalyzeFucai3D(numbers); err == nil {
			t.Fatalf("接受非法号码: %v", numbers)
		}
	}
}
