package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Collector implements prometheus.Collector to collect metrics from the database on each scrape.
type Collector struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewCollector creates a new Collector instance.
func NewCollector(db *gorm.DB, logger *logrus.Logger) *Collector {
	return &Collector{
		db:     db,
		logger: logger,
	}
}

// Describe implements prometheus.Collector.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	IssuesTotal.Describe(ch)
	IssuesByNamespace.Describe(ch)
	IssuesBySeverity.Describe(ch)
	DatabaseConnectionsActive.Describe(ch)
	DatabaseConnectionsIdle.Describe(ch)
}

// Collect implements prometheus.Collector.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.CollectIssueMetrics()
	c.CollectDatabaseMetrics()

	IssuesTotal.Collect(ch)
	IssuesByNamespace.Collect(ch)
	IssuesBySeverity.Collect(ch)
	DatabaseConnectionsActive.Collect(ch)
	DatabaseConnectionsIdle.Collect(ch)
}

// CollectIssueMetrics queries the database and updates issue-related gauges.
func (c *Collector) CollectIssueMetrics() {
	IssuesTotal.Reset()
	IssuesByNamespace.Reset()
	IssuesBySeverity.Reset()

	var results []struct {
		Namespace string
		Severity  string
		State     string
		IssueType string
		Count     int64
	}

	err := c.db.Table("issues").
		Select("namespace, severity, state, issue_type, COUNT(*) as count").
		Group("namespace, severity, state, issue_type").
		Scan(&results).Error
	if err != nil {
		c.logger.WithError(err).Error("Failed to collect issue metrics")
		return
	}

	for _, r := range results {
		IssuesTotal.WithLabelValues(r.Namespace, r.Severity, r.State, r.IssueType).Set(float64(r.Count))
	}

	var namespaceResults []struct {
		Namespace string
		State     string
		Count     int64
	}

	err = c.db.Table("issues").
		Select("namespace, state, COUNT(*) as count").
		Group("namespace, state").
		Scan(&namespaceResults).Error
	if err != nil {
		c.logger.WithError(err).Error("Failed to collect namespace metrics")
		return
	}

	for _, r := range namespaceResults {
		IssuesByNamespace.WithLabelValues(r.Namespace, r.State).Set(float64(r.Count))
	}

	var severityResults []struct {
		Severity string
		State    string
		Count    int64
	}

	err = c.db.Table("issues").
		Select("severity, state, COUNT(*) as count").
		Group("severity, state").
		Scan(&severityResults).Error
	if err != nil {
		c.logger.WithError(err).Error("Failed to collect severity metrics")
		return
	}

	for _, r := range severityResults {
		IssuesBySeverity.WithLabelValues(r.Severity, r.State).Set(float64(r.Count))
	}
}

// CollectDatabaseMetrics queries the database and updates connection-related gauges.
func (c *Collector) CollectDatabaseMetrics() {
	sqlDB, err := c.db.DB()
	if err != nil {
		c.logger.WithError(err).Error("Failed to get database instance")
		return
	}

	stats := sqlDB.Stats()
	DatabaseConnectionsActive.Set(float64(stats.InUse))
	DatabaseConnectionsIdle.Set(float64(stats.Idle))
}

