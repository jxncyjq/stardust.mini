package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMetricsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Metrics("test_metrics_svc"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequestResultByStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   string
	}{
		{name: "2xx success", status: http.StatusOK, want: "success"},
		{name: "3xx success", status: http.StatusMovedPermanently, want: "success"},
		{name: "4xx fail", status: http.StatusBadRequest, want: "fail"},
		{name: "5xx fail", status: http.StatusInternalServerError, want: "fail"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, requestResultByStatus(tc.status))
		})
	}
}

func TestRequestStatusClass(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   string
	}{
		{name: "1xx", status: 101, want: "1xx"},
		{name: "2xx", status: http.StatusOK, want: "2xx"},
		{name: "3xx", status: http.StatusFound, want: "3xx"},
		{name: "4xx", status: http.StatusForbidden, want: "4xx"},
		{name: "5xx", status: http.StatusBadGateway, want: "5xx"},
		{name: "unknown", status: 0, want: "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, requestStatusClass(tc.status))
		})
	}
}
