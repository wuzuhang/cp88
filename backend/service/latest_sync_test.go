package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestManualSSQSyncSelectsLatestAndRejectsInvalidSource(t *testing.T) {
	older := strings.NewReplacer("2026115", "2026114", "2026-10-08", "2026-10-06").Replace(drawFixture)
	for _, tc := range []struct {
		name, body string
		status     int
		storeErr   error
		success    bool
	}{
		{"unordered", older + "\n" + drawFixture + "\n" + strings.NewReplacer("2026115", "2026113", "2026-10-08", "2026-10-04").Replace(drawFixture), 200, nil, true},
		{"invalid", drawFixture + "\ninvalid", 200, nil, false},
		{"http failure", drawFixture, 502, nil, false},
		{"db failure", drawFixture, 200, errors.New("write failure"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			store := &fakeRecordStore{err: tc.storeErr}
			s := NewDrawSyncService(nil, server.URL)
			s.store = store
			issue, err := s.SyncLatest(context.Background())
			if tc.success {
				if err != nil || issue != "2026115" || len(store.records) != 1 || store.records[0].DrawDate != "2026-10-08" {
					t.Fatalf("最新期错误: %q %v %+v", issue, err, store.records)
				}
			} else if err == nil || len(store.records) != 0 {
				t.Fatalf("失败不应入库: %v %+v", err, store.records)
			}
		})
	}
}

func TestManualFucai3DSyncSelectsLatestAndRejectsInvalidSource(t *testing.T) {
	older := strings.NewReplacer("2026267", "2026266", "2026-10-08", "2026-10-07").Replace(fucai3DFixture)
	for _, tc := range []struct {
		name, body string
		status     int
		storeErr   error
		success    bool
	}{
		{"unordered", fucai3DFixture + "\n" + older, 200, nil, true},
		{"invalid", fucai3DFixture + "\ninvalid", 200, nil, false},
		{"http failure", fucai3DFixture, 502, nil, false},
		{"db failure", fucai3DFixture, 200, errors.New("write failure"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			store := &fakeFucai3DStore{err: tc.storeErr}
			s := NewFucai3DSyncService(nil, server.URL)
			s.store = store
			issue, err := s.SyncLatest(context.Background())
			if tc.success {
				if err != nil || issue != "2026267" || len(store.records) != 1 || store.records[0].Num1 != 0 {
					t.Fatalf("最新期错误: %q %v %+v", issue, err, store.records)
				}
			} else if err == nil || len(store.records) != 0 {
				t.Fatalf("失败不应入库: %v %+v", err, store.records)
			}
		})
	}
}
