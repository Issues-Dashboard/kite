package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kite_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kite_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Issue metrics
	IssuesTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kite_issues_total",
			Help: "Total number of issues by namespace, severity, state, and type",
		},
		[]string{"namespace", "severity", "state", "type"},
	)

	IssuesByNamespace = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kite_issues_by_namespace",
			Help: "Number of issues per namespace",
		},
		[]string{"namespace", "state"},
	)

	IssuesBySeverity = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kite_issues_by_severity",
			Help: "Number of issues by severity level",
		},
		[]string{"severity", "state"},
	)

	IssuesCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kite_issues_created_total",
			Help: "Total number of issues created",
		},
		[]string{"namespace", "severity", "type"},
	)

	IssuesResolvedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kite_issues_resolved_total",
			Help: "Total number of issues resolved",
		},
		[]string{"namespace", "severity", "type"},
	)

	IssueResolutionTimeSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kite_issue_resolution_time_seconds",
			Help:    "Time taken to resolve issues in seconds",
			Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 28800, 86400}, // 1m to 1day
		},
		[]string{"namespace", "severity", "type"},
	)

	// Database metrics
	DatabaseConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "kite_database_connections_active",
			Help: "Number of active database connections",
		},
	)

	DatabaseConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "kite_database_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kite_database_query_duration_seconds",
			Help:    "Database query latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

