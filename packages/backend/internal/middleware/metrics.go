package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/konflux-ci/kite/internal/metrics"
)

// MetricsMiddleware records HTTP request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		endpoint := c.FullPath()
		method := c.Request.Method

		// If endpoint is empty (404), use the path
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		metrics.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
	}
}

