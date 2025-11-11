package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/konflux-ci/kite/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestMetricsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset metrics before test
	metrics.HTTPRequestsTotal.Reset()
	metrics.HTTPRequestDuration.Reset()

	// Create router with metrics middleware
	router := gin.New()
	router.Use(MetricsMiddleware())

	// Add test endpoints
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fail"})
	})

	// Test successful request
	t.Run("Records successful request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check that counter was incremented
		counter := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", "/test", "200")
		if counter != 1 {
			t.Errorf("Expected counter to be 1, got %f", counter)
		}
	})

	// Test error request
	t.Run("Records error request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		// Check that counter was incremented with correct status
		counter := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", "/error", "500")
		if counter != 1 {
			t.Errorf("Expected counter to be 1, got %f", counter)
		}
	})

	// Test 404 request
	t.Run("Records 404 request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/notfound", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}

		// For 404, endpoint path is used as-is
		counter := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", "/notfound", "404")
		if counter != 1 {
			t.Errorf("Expected counter to be 1, got %f", counter)
		}
	})

	// Test duration histogram
	t.Run("Records request duration", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify histogram metric was created
		_, err := metrics.HTTPRequestDuration.GetMetricWithLabelValues("GET", "/test")
		if err != nil {
			t.Errorf("Expected histogram metric to exist, got error: %v", err)
		}
	})
}

// Helper function to get counter value from a CounterVec
func getCounterValue(t *testing.T, counterVec *prometheus.CounterVec, labels ...string) float64 {
	counter, err := counterVec.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Fatalf("Failed to get counter with labels %v: %v", labels, err)
	}

	var metric dto.Metric
	if err := counter.Write(&metric); err != nil {
		t.Fatalf("Failed to write metric: %v", err)
	}

	return metric.Counter.GetValue()
}

