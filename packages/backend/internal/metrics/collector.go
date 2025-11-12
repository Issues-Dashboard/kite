package metrics

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Collector struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewCollector(db *gorm.DB, logger *logrus.Logger) *Collector {
	return &Collector{
		db:     db,
		logger: logger,
	}
}

// Start begins collecting metrics at regular intervals
func (c *Collector) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Collect initial metrics
	c.CollectAll()

	for {
		select {
		case <-ticker.C:
			c.CollectAll()
		case <-ctx.Done():
			c.logger.Info("Metrics collector stopped")
			return
		}
	}
}

// CollectAll collects all metrics
func (c *Collector) CollectAll() {
	c.CollectIssueMetrics()
	c.CollectDatabaseMetrics()
}

// CollectIssueMetrics collects issue-related metrics
func (c *Collector) CollectIssueMetrics() {
	// Reset gauges before collecting new values
	IssuesTotal.Reset()
	IssuesByNamespace.Reset()
	IssuesBySeverity.Reset()

	// Count issues by namespace, severity, state, and type
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

	// Count issues by namespace and state
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

	// Count issues by severity and state
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

	c.logger.Debug("Issue metrics collected successfully")
}

// CollectDatabaseMetrics collects database connection metrics
func (c *Collector) CollectDatabaseMetrics() {
	sqlDB, err := c.db.DB()
	if err != nil {
		c.logger.WithError(err).Error("Failed to get database instance")
		return
	}

	stats := sqlDB.Stats()
	DatabaseConnectionsActive.Set(float64(stats.InUse))
	DatabaseConnectionsIdle.Set(float64(stats.Idle))

	c.logger.Debug("Database metrics collected successfully")
}

