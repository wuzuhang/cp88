package controller

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestManualSyncResponseAndLock(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name       string
		busy, fail bool
		status     int
	}{
		{"success", false, false, 200}, {"failure", false, true, 502}, {"already updating", true, false, 409},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var guard sync.Mutex
			if tc.busy {
				guard.Lock()
				defer guard.Unlock()
			}
			called := false
			r := gin.New()
			r.POST("/sync", func(c *gin.Context) {
				runLatestSync(c, &guard, func(ctx context.Context) (string, error) {
					called = true
					if _, ok := ctx.Deadline(); !ok {
						t.Error("手动更新必须有限时")
					}
					if tc.fail {
						return "", errors.New("private database detail")
					}
					return "2026267", nil
				})
			})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("POST", "/sync", nil))
			if w.Code != tc.status {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if tc.busy && called {
				t.Fatal("忙碌时不应再次更新")
			}
			if strings.Contains(w.Body.String(), "private database detail") {
				t.Fatal("不应向C端暴露内部错误")
			}
			if !tc.busy && !guard.TryLock() {
				t.Fatal("更新完成后必须释放锁")
			} else if !tc.busy {
				guard.Unlock()
			}
		})
	}
}
