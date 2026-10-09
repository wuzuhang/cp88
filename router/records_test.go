package router

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"shuangseqiu/config"
)

func TestRecordsParametersAndUnavailableDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previous := config.DB
	config.DB = nil
	defer func() { config.DB = previous }()
	r := InitRouter()
	for _, tc := range []struct {
		query  string
		status int
	}{
		{"?limit=0", 400}, {"?limit=501", 400}, {"?limit=x", 400}, {"?offset=-1", 400}, {"?offset=100001", 400}, {"?offset=x", 400}, {"", 503},
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/records"+tc.query, nil))
		if w.Code != tc.status {
			t.Fatalf("%s: status=%d, body=%s", tc.query, w.Code, w.Body.String())
		}
	}
}
