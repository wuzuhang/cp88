package router

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := InitRouter()
	cases := []struct {
		name, method, path, body string
		status                   int
	}{
		{"health", "GET", "/health", "", 200},
		{"analysis", "POST", "/api/v1/analysis", `{"red_balls":[1,8,12,19,25,33],"blue_ball":6}`, 200},
		{"duplicate", "POST", "/api/v1/analysis", `{"red_balls":[1,1,12,19,25,33],"blue_ball":6}`, 400},
		{"out of range", "POST", "/api/v1/analysis", `{"red_balls":[1,8,12,19,25,34],"blue_ball":6}`, 400},
		{"missing", "POST", "/api/v1/analysis", `{}`, 400},
		{"malformed", "POST", "/api/v1/analysis", `{`, 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			var result struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Code != tc.status {
				t.Fatalf("response code=%d", result.Code)
			}
		})
	}
}
