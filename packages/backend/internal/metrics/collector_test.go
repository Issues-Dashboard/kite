package metrics

import (
	"testing"

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
	logger.SetLevel(logrus.ErrorLevel)
	return logger
}

func TestCollector_CollectIssueMetrics(t *testing.T) {
	db := setupTestDB(t)
	logger := setupTestLogger()
	collector := NewCollector(db, logger)

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

	collector.CollectIssueMetrics()
}

func TestCollector_CollectDatabaseMetrics(t *testing.T) {
	db := setupTestDB(t)
	logger := setupTestLogger()
	collector := NewCollector(db, logger)

	collector.CollectDatabaseMetrics()
}

func TestCollector_Collect(t *testing.T) {
	db := setupTestDB(t)
	logger := setupTestLogger()
	collector := NewCollector(db, logger)

	scope := models.IssueScope{
		ID:                "test-scope",
		ResourceType:      "component",
		ResourceName:      "test-component",
		ResourceNamespace: "test-ns",
	}

	if err := db.Create(&scope).Error; err != nil {
		t.Fatalf("Failed to create test scope: %v", err)
	}

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

	collector.CollectIssueMetrics()
	collector.CollectDatabaseMetrics()
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

