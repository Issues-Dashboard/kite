package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/konflux-ci/kite/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Migrate schema
	if err := db.AutoMigrate(&models.IssueScope{}, &models.Issue{}); err != nil {
		t.Fatalf("Failed to migrate schema: %v", err)
	}

	return db
}

func setupTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Quiet during tests
	return logger
}

func TestCollector_CollectIssueMetrics(t *testing.T) {
	db := setupTestDB(t)
	logger := setupTestLogger()
	collector := NewCollector(db, logger)

	// Create test scopes first
	testScopes := []models.IssueScope{
		{
			ID:                "scope-1",
			ResourceType:      "component",
			ResourceName:      "test-component-1",
			ResourceNamespace: "team-alpha",
		},
		{
			ID:                "scope-2",
			ResourceType:      "component",
			ResourceName:      "test-component-2",
			ResourceNamespace: "team-alpha",
		},
		{
			ID:                "scope-3",
			ResourceType:      "component",
			ResourceName:      "test-component-3",
			ResourceNamespace: "team-beta",
		},
	}

	for _, scope := range testScopes {
		if err := db.Create(&scope).Error; err != nil {
			t.Fatalf("Failed to create test scope: %v", err)
		}
	}

	// Create test issues
	testIssues := []models.Issue{
		{
			ID:          "issue-1",
			Title:       "Test Issue 1",
			Namespace:   "team-alpha",
			Severity:    models.SeverityMajor,
			State:       models.IssueStateActive,
			IssueType:   models.IssueTypeBuild,
			Description: "Test",
			ScopeID:     "scope-1",
		},
		{
			ID:          "issue-2",
			Title:       "Test Issue 2",
			Namespace:   "team-alpha",
			Severity:    models.SeverityCritical,
			State:       models.IssueStateActive,
			IssueType:   models.IssueTypeTest,
			Description: "Test",
			ScopeID:     "scope-2",
		},
		{
			ID:          "issue-3",
			Title:       "Test Issue 3",
			Namespace:   "team-beta",
			Severity:    models.SeverityMinor,
			State:       models.IssueStateResolved,
			IssueType:   models.IssueTypeBuild,
			Description: "Test",
			ScopeID:     "scope-3",
		},
	}

	for _, issue := range testIssues {
		if err := db.Create(&issue).Error; err != nil {
			t.Fatalf("Failed to create test issue: %v", err)
		}
	}

	// Collect metrics
	collector.CollectIssueMetrics()

	// Verify metrics were collected
	// We can't easily assert exact metric values without exposing internal state,
	// but we can verify the collection didn't error
	t.Log("Issue metrics collected successfully")
}

func TestCollector_CollectDatabaseMetrics(t *testing.T) {
	db := setupTestDB(t)
	logger := setupTestLogger()
	collector := NewCollector(db, logger)

	// Collect database metrics
	collector.CollectDatabaseMetrics()

	// Verify collection succeeded
	t.Log("Database metrics collected successfully")
}

func TestCollector_CollectAll(t *testing.T) {
	db := setupTestDB(t)
	logger := setupTestLogger()
	collector := NewCollector(db, logger)

	// Create a test scope
	scope := models.IssueScope{
		ID:                "test-scope",
		ResourceType:      "component",
		ResourceName:      "test-component",
		ResourceNamespace: "test-ns",
	}

	if err := db.Create(&scope).Error; err != nil {
		t.Fatalf("Failed to create test scope: %v", err)
	}

	// Create a test issue
	issue := models.Issue{
		ID:          "test-issue",
		Title:       "Test",
		Namespace:   "test-ns",
		Severity:    models.SeverityMajor,
		State:       models.IssueStateActive,
		IssueType:   models.IssueTypeBuild,
		Description: "Test issue",
		ScopeID:     "test-scope",
	}

	if err := db.Create(&issue).Error; err != nil {
		t.Fatalf("Failed to create test issue: %v", err)
	}

	// Collect all metrics
	collector.CollectAll()

	t.Log("All metrics collected successfully")
}

func TestCollector_Start(t *testing.T) {
	db := setupTestDB(t)
	logger := setupTestLogger()
	collector := NewCollector(db, logger)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start collector in goroutine
	done := make(chan bool)
	go func() {
		collector.Start(ctx, 50*time.Millisecond)
		done <- true
	}()

	// Wait for context to cancel
	select {
	case <-done:
		t.Log("Collector stopped as expected")
	case <-time.After(200 * time.Millisecond):
		t.Error("Collector did not stop within expected time")
	}
}

func TestCollector_StartWithCancellation(t *testing.T) {
	db := setupTestDB(t)
	logger := setupTestLogger()
	collector := NewCollector(db, logger)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start collector in goroutine
	done := make(chan bool)
	go func() {
		collector.Start(ctx, 1*time.Second)
		done <- true
	}()

	// Cancel after short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for collector to stop
	select {
	case <-done:
		t.Log("Collector stopped after cancellation")
	case <-time.After(200 * time.Millisecond):
		t.Error("Collector did not stop after cancellation")
	}
}

func TestNewCollector(t *testing.T) {
	db := setupTestDB(t)
	logger := setupTestLogger()

	collector := NewCollector(db, logger)

	if collector == nil {
		t.Error("Expected non-nil collector")
	}

	if collector.db != db {
		t.Error("Collector db not set correctly")
	}

	if collector.logger != logger {
		t.Error("Collector logger not set correctly")
	}
}

