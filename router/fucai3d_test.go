package router

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"shuangseqiu/config"
	"strings"
	"testing"
)

func TestFucai3DEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previous := config.DB
	config.DB = nil
	defer func() { config.DB = previous }()
	r := InitRouter()
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/api/v1/fucai3d/records", "", 503},
		{"GET", "/api/v1/fucai3d/records?limit=501", "", 400},
		{"GET", "/api/v1/fucai3d/records?offset=-1", "", 400},
		{"POST", "/api/v1/fucai3d/analysis", `{"numbers":[0,0,0]}`, 200},
		{"POST", "/api/v1/fucai3d/analysis", `{"numbers":[9,0,5]}`, 200},
		{"POST", "/api/v1/fucai3d/analysis", `{"numbers":[0,0,10]}`, 400},
		{"POST", "/api/v1/fucai3d/analysis", `{"numbers":[0,0]}`, 400},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != tc.status {
			t.Fatalf("%s %s: %d %s", tc.method, tc.path, w.Code, w.Body.String())
		}
	}
}
