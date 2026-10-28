package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var Requests = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "http_requests_total", Help: "HTTP request count"}, []string{"method", "path", "status"})
var Latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "HTTP request latency"}, []string{"method", "path"})
var Errors = prometheus.NewCounter(prometheus.CounterOpts{Name: "http_errors_total", Help: "HTTP error count"})

func InitMetrics() {
	prometheus.MustRegister(Requests, Latency, Errors)
}
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		st := time.Now()
		c.Next()
		Requests.WithLabelValues(c.Request.Method, c.FullPath(), strconv.Itoa(c.Writer.Status())).Inc()
		Latency.WithLabelValues(c.Request.Method, c.FullPath()).Observe(time.Since(st).Seconds())
		if c.Writer.Status() >= 500 {
			Errors.Inc()
		}
	}
}
