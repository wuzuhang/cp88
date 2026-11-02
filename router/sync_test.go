package router

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"shuangseqiu/config"
	"testing"
)

func TestManualSyncRequiresDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previous := config.DB
	config.DB = nil
	defer func() { config.DB = previous }()
	r := InitRouter()
	for _, path := range []string{"/api/v1/records/sync", "/api/v1/fucai3d/records/sync"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("POST", path, nil))
		if w.Code != 503 {
			t.Fatalf("%s: status=%d body=%s", path, w.Code, w.Body.String())
		}
	}
}
