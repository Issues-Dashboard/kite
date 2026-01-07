package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal tracks the total number of HTTP requests by method, endpoint, and status.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kite_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTPRequestDuration tracks HTTP request latency in seconds.
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kite_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// IssuesTotal tracks the total number of issues by namespace, severity, state, and type.
	// Updated on each scrape from the database.
	IssuesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kite_issues_total",
			Help: "Total number of issues by namespace, severity, state, and type",
		},
		[]string{"namespace", "severity", "state", "type"},
	)

	// IssuesByNamespace tracks the number of issues per namespace and state.
	// Updated on each scrape from the database.
	IssuesByNamespace = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kite_issues_by_namespace",
			Help: "Number of issues per namespace",
		},
		[]string{"namespace", "state"},
	)

	// IssuesBySeverity tracks the number of issues by severity level and state.
	// Updated on each scrape from the database.
	IssuesBySeverity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kite_issues_by_severity",
			Help: "Number of issues by severity level",
		},
		[]string{"severity", "state"},
	)

	// IssuesCreatedTotal tracks the total number of issues created.
	IssuesCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kite_issues_created_total",
			Help: "Total number of issues created",
		},
		[]string{"namespace", "severity", "type"},
	)

	// IssuesResolvedTotal tracks the total number of issues resolved.
	IssuesResolvedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kite_issues_resolved_total",
			Help: "Total number of issues resolved",
		},
		[]string{"namespace", "severity", "type"},
	)

	// IssueResolutionTimeSeconds tracks the time taken to resolve issues in seconds.
	IssueResolutionTimeSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kite_issue_resolution_time_seconds",
			Help:    "Time taken to resolve issues in seconds",
			Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 28800, 86400},
		},
		[]string{"namespace", "severity", "type"},
	)

	// DatabaseConnectionsActive tracks the number of active database connections.
	// Updated on each scrape from the database.
	DatabaseConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "kite_database_connections_active",
			Help: "Number of active database connections",
		},
	)

	// DatabaseConnectionsIdle tracks the number of idle database connections.
	// Updated on each scrape from the database.
	DatabaseConnectionsIdle = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "kite_database_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	// DatabaseQueryDuration tracks database query latency in seconds.
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kite_database_query_duration_seconds",
			Help:    "Database query latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

