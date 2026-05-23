package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal *prometheus.CounterVec // HTTP 请求总量，按 method/path/status 统计
	httpRequestsByResult *prometheus.CounterVec // HTTP 请求结果总量，按 method/path/result(success|fail) 统计
	httpRequestsByClass  *prometheus.CounterVec // HTTP 状态码分层总量，按 method/path/status_class(2xx/4xx/5xx...) 统计
	httpRequestDuration  *prometheus.HistogramVec // HTTP 请求耗时直方图（毫秒），按 method/path 统计
	httpRequestsInFlight prometheus.Gauge // 当前正在处理中的 HTTP 请求数
	metricsOnce          sync.Once
)

func initMetrics(serviceName string) {
	metricsOnce.Do(func() {
		httpRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: serviceName,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		)
		httpRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: serviceName,
				Name:      "http_request_duration_ms",
				Help:      "HTTP request duration in milliseconds",
				Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500},
			},
			[]string{"method", "path"},
		)
		httpRequestsByResult = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: serviceName,
				Name:      "http_requests_result_total",
				Help:      "Total number of HTTP requests by result",
			},
			[]string{"method", "path", "result"},
		)
		httpRequestsByClass = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: serviceName,
				Name:      "http_requests_status_class_total",
				Help:      "Total number of HTTP requests by status class",
			},
			[]string{"method", "path", "status_class"},
		)
		httpRequestsInFlight = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: serviceName,
				Name:      "http_requests_in_flight",
				Help:      "Number of HTTP requests currently being processed",
			},
		)
		prometheus.MustRegister(
			httpRequestsTotal,
			httpRequestsByResult,
			httpRequestsByClass,
			httpRequestDuration,
			httpRequestsInFlight,
		)
	})
}

func requestResultByStatus(status int) string {
	if status >= 400 {
		return "fail"
	}
	return "success"
}

func requestStatusClass(status int) string {
	switch {
	case status >= 100 && status < 200:
		return "1xx"
	case status >= 200 && status < 300:
		return "2xx"
	case status >= 300 && status < 400:
		return "3xx"
	case status >= 400 && status < 500:
		return "4xx"
	case status >= 500 && status < 600:
		return "5xx"
	default:
		return "unknown"
	}
}

// Metrics HTTP 指标采集中间件
func Metrics(serviceName string) gin.HandlerFunc {
	initMetrics(serviceName)
	return func(c *gin.Context) {
		start := time.Now()
		httpRequestsInFlight.Inc()

		c.Next()

		httpRequestsInFlight.Dec()
		duration := float64(time.Since(start).Milliseconds())
		statusCode := c.Writer.Status()
		status := strconv.Itoa(statusCode)
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		result := requestResultByStatus(statusCode)
		statusClass := requestStatusClass(statusCode)

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestsByResult.WithLabelValues(c.Request.Method, path, result).Inc()
		httpRequestsByClass.WithLabelValues(c.Request.Method, path, statusClass).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// MetricsHandler 返回 Prometheus /metrics 端点处理器
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
