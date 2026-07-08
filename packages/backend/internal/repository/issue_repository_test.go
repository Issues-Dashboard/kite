package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/konflux-ci/kite/internal/handlers/dto"
	"github.com/konflux-ci/kite/internal/models"
	"github.com/konflux-ci/kite/internal/testhelpers"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SetupOptions struct {
	UseConcurrentDatabase bool // Use a concurrent database setup
}

// setupTestScenario sets up a context and repository for test scenarios
func setupTestScenario(t *testing.T, options SetupOptions) (context.Context, *gorm.DB, IssueRepository) {
	var db *gorm.DB
	if options.UseConcurrentDatabase {
		db = testhelpers.SetupConcurrentTestDB(t)
	} else {
		db = testhelpers.SetupTestDB(t)
	}
	logger := logrus.New()
	repo := NewIssueRepository(db, logger)
	ctx := context.Background()

	return ctx, db, repo
}

// createTestIssue is a helper function to create test issues
func createTestIssue(title, namespace string) dto.CreateIssueRequest {
	return createTestIssueWithScope(title, namespace, "component", "test-component")
}

func createTestIssueWithScope(title, namespace, resourceType, resourceName string) dto.CreateIssueRequest {
	return dto.CreateIssueRequest{
		Title:       title,
		Description: "Test description",
		Severity:    models.SeverityMajor,
		IssueType:   models.IssueTypeBuild,
		Namespace:   namespace,
		Scope: dto.ScopeReqBody{
			ResourceType:      resourceType,
			ResourceName:      resourceName,
			ResourceNamespace: namespace,
		},
		Links: []dto.CreateLinkRequest{
			{
				URL:   "konflux.test/pipelineruns/failure-xyz",
				Title: "Failed Pipeline Run: xyz",
			},
		},
	}
}

func TestIssueRepository_Create(t *testing.T) {
	// Setup
	ctx, db, repo := setupTestScenario(t, SetupOptions{})

	// Test issue data
	req := createTestIssue("Test Issue", "test-namespace")

	// Create it
	issue, err := repo.Create(ctx, req)

	// Check
	if err != nil {
		t.Fatalf("unexpected error, got %v", err)
	}

	err = testhelpers.CompareIssueToDTO(*issue, req)
	if err != nil {
		t.Errorf("unexpected error, got: %v", err)
	}

	// Confirm that issue was saved to the database
	var currentCount int64
	db.Model(&models.Issue{}).Count(&currentCount)
	if currentCount != 1 {
		t.Errorf("Expected 1 issue in DB, got %d", currentCount)
	}
}

func TestIssueRepository_FindByID(t *testing.T) {
	// Setup
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	// Create a test issue first
	req := createTestIssue("Find Test Issue", "test-namespace")
	createdIssue, err := repo.Create(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error, got %v", err)
	}
	if createdIssue == nil {
		t.Fatalf("Expected issue to be created, got nil")
	}

	// Find the issue
	foundIssue, err := repo.FindByID(ctx, createdIssue.ID)
	if err != nil {
		t.Fatalf("unexpected error, got: %v", err)
	}

	if foundIssue == nil {
		t.Fatal("Expected issue to be found, got nil")
	}

	// Verify
	err = testhelpers.CompareIssues(*createdIssue, *foundIssue)
	if err != nil {
		t.Errorf("unexpected error, got: %v", err)
	}
}

func TestIssueRepository_FindByID_NotFound(t *testing.T) {
	// Setup
	ctx, _, repo := setupTestScenario(t, SetupOptions{})
	// Try to find non-existent issue
	foundIssue, err := repo.FindByID(ctx, "does-not-exist")

	// Verify
	if err != nil {
		t.Fatalf("Expected no error for non-existent issue, got %v", err)
	}

	if foundIssue != nil {
		t.Errorf("Expected nil for non-existent issue, got an issue")
	}
}

func TestIssueRepository_FindAll_WithFilters(t *testing.T) {
	// Setup
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	// Create test issues
	issues := []dto.CreateIssueRequest{
		createTestIssue("Build Issue", "team-test"),
		{
			Title:       "Test Issue",
			Description: "Test Description",
			Severity:    models.SeverityCritical,
			IssueType:   models.IssueTypeTest,
			Namespace:   "team-test",
			Scope: dto.ScopeReqBody{
				ResourceType:      "component",
				ResourceName:      "test-component",
				ResourceNamespace: "team-test",
			},
		},
		createTestIssue("Release Issue", "team-beta"),
	}

	// Write issues to DB
	for _, req := range issues {
		_, err := repo.Create(ctx, req)
		if err != nil {
			t.Fatalf("Failed to create test issue: %v", err)
		}
	}

	// Check: Find all issues in team-alpha
	filters := IssueQueryFilters{
		Namespace: "team-test",
		Limit:     10,
	}

	foundIssues, total, err := repo.FindAll(ctx, filters)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 issues in team-test, got %d", total)
	}

	if len(foundIssues) != 2 {
		t.Errorf("Expected 2 issues returned, got %d", len(foundIssues))
	}

	// Check that all returned issues belong to team-test
	for _, issue := range foundIssues {
		if issue.Namespace != "team-test" {
			t.Errorf("Expected namespace 'team-test', got '%s'", issue.Namespace)
		}
	}
}

func TestIssueRepository_CheckDuplicate(t *testing.T) {
	// Setup
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	// Create an issue
	req := createTestIssue("Duplicate Test", "test-namespace")
	_, err := repo.Create(ctx, req)
	if err != nil {
		t.Fatalf("Unexpected error, got %v", err)
	}

	// Check for duplicates with the same properties
	foundIssue, err := repo.FindDuplicate(ctx, req)

	// Verify
	if err != nil {
		t.Fatalf("Unexpected error, got %v", err)
	}

	if foundIssue == nil {
		t.Fatal("Expected duplicate issue to be returned")
	}
}

func TestIssueRepository_Update(t *testing.T) {
	// Setup
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	// Create an issue
	req := createTestIssue("Some Issue", "test-namespace")
	issue, err := repo.Create(ctx, req)
	if err != nil {
		t.Fatalf("Unexpected error, got %v", err)
	}

	// Get latest issue
	expectedID := issue.ID
	expectedTitle := "Updated Issue"

	updatedIssueReq := dto.UpdateIssueRequest{
		Title: expectedTitle,
	}
	// Update
	updatedIssue, err := repo.Update(ctx, expectedID, updatedIssueReq)

	// Verify
	if err != nil {
		t.Fatalf("Unexpected error, got %v", err)
	}

	if updatedIssue == nil {
		t.Fatal("Expected issue to be returned")
	}

	if updatedIssue.ID != expectedID {
		t.Errorf("Wrong issue returned, got issue with ID %s, expected %s", updatedIssue.ID, expectedID)
	}

	if updatedIssue.Title != expectedTitle {
		t.Errorf("Wrong title, got '%s', expected '%s'", updatedIssue.Title, expectedTitle)
	}
}

func TestIssueRepository_Update_ResolvedByID(t *testing.T) {
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	req := createTestIssue("Resolvable Issue", "test-namespace")
	issue, err := repo.Create(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error creating issue: %v", err)
	}

	resolvedByID := "f67079c2-ce41-4bf9-bfb5-fbd9dbc1cf3c"
	updatedIssue, err := repo.Update(ctx, issue.ID, dto.UpdateIssueRequest{
		State:        models.IssueStateResolved,
		ResolvedByID: resolvedByID,
	})
	if err != nil {
		t.Fatalf("unexpected error updating issue: %v", err)
	}

	if updatedIssue.State != models.IssueStateResolved {
		t.Errorf("expected state RESOLVED, got %q", updatedIssue.State)
	}

	if updatedIssue.ResolvedByID != resolvedByID {
		t.Errorf("expected resolvedById %q, got %q", resolvedByID, updatedIssue.ResolvedByID)
	}

	if updatedIssue.ResolvedAt == nil {
		t.Fatal("expected resolvedAt to be set when resolving an issue")
	}

	reloadedIssue, err := repo.FindByID(ctx, issue.ID)
	if err != nil {
		t.Fatalf("unexpected error reloading issue: %v", err)
	}

	if reloadedIssue.ResolvedByID != resolvedByID {
		t.Errorf("expected persisted resolvedById %q, got %q", resolvedByID, reloadedIssue.ResolvedByID)
	}
}

func TestIssueRepository_Delete(t *testing.T) {
	ctx, db, repo := setupTestScenario(t, SetupOptions{})

	// Create issue with links
	req := createTestIssue("Delete Test", "test-namespace")
	req.Links = append(req.Links,
		dto.CreateLinkRequest{
			Title: "Delete Test Link",
			URL:   "https://konflux.test/some-link",
		},
	)

	createdIssue, err := repo.Create(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create test issue: %v", err)
	}

	// Verify issue and link exists
	var issueCount, linkCount int64
	db.Model(&models.Issue{}).Count(&issueCount)
	db.Model(&models.Link{}).Count(&linkCount)

	if issueCount != 1 {
		t.Errorf("Expected 1 issue before delete, got %d", issueCount)
	}

	if linkCount != 2 {
		t.Errorf("Expected 2 links before delete, got %d", linkCount)
	}

	// Delete the issue
	err = repo.Delete(ctx, createdIssue.ID)

	// Verify
	if err != nil {
		t.Fatalf("Unexpected error, got %v", err)
	}

	// Update variables after deletion
	db.Model(&models.Issue{}).Count(&issueCount)
	db.Model(&models.Link{}).Count(&linkCount)

	if issueCount != 0 {
		t.Errorf("Expected 0 issues after delete, got %d", issueCount)
	}

	if linkCount != 0 {
		t.Errorf("Expected 0 links after delete, got %d", linkCount)
	}
}

func TestIssueRepository_CreateOrUpdate_NoDuplicates(t *testing.T) {
	// Setup
	ctx, _, repo := setupTestScenario(t, SetupOptions{
		UseConcurrentDatabase: true,
	})

	// Create issue
	req := createTestIssue("CreateOrUpdate Test", "test-namespace")

	// Number of concurrent requests
	numGoroutines := 10
	// Lets create a WaitGroup and wait for all
	// goroutines to finish making requests.
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Store all returned issues
	issues := make([]*models.Issue, numGoroutines)
	errors := make([]error, numGoroutines)

	// Launch concurrent CreateOrUpdate operations
	for i := 0; i < numGoroutines; i++ {
		// Launch this in a goroutine
		go func(index int) {
			defer wg.Done()

			// Add a small, random delay to increase chance of race condition.
			//
			// Without the delay: the goroutines would most likely execute sequentially.
			// With Delay: they're more likely to be in different phases of operation
			// at the same time, which is when race conditions occur.
			//
			// Index%3 creates: 0, 1, 2, 0, etc ...
			// So delays are: 0ms, 1ms, 2ms, 0ms, etc ...
			// This creates three waves of goroutines, each in some delay (0ms, 1ms, 2ms)
			time.Sleep(time.Millisecond * time.Duration(index%3))

			issue, err := repo.CreateOrUpdate(ctx, req)
			issues[index] = issue
			errors[index] = err
		}(i)
	}
	// Wait for all goroutines to complete
	wg.Wait()
	// Ensure that all issues returned are the same issue.
	// This means that no duplicates should have been created
	// with the same request payload.
	expectedID := issues[0].ID
	for _, issue := range issues {
		if issue.ID != expectedID {
			t.Fatalf("Expected all issues to have ID %s, got %s", expectedID, issue.ID)
		}
	}

	// There should be no errors reported.
	for _, err := range errors {
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}
}

func TestIssueRepository_FindDuplicate_NoDuplicate(t *testing.T) {
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	foundIssue, err := repo.FindDuplicate(ctx, createTestIssue("Fresh Issue", "test-namespace"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if foundIssue != nil {
		t.Fatalf("expected no duplicate, got issue %s", foundIssue.ID)
	}
}

func TestIssueRepository_Create_UpdatesDuplicate(t *testing.T) {
	ctx, db, repo := setupTestScenario(t, SetupOptions{})

	req := createTestIssue("Duplicate Create", "test-namespace")
	original, err := repo.Create(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error creating issue: %v", err)
	}

	updatedReq := createTestIssue("Duplicate Create Updated", "test-namespace")
	updatedReq.Description = "Updated description"
	updatedReq.Severity = models.SeverityCritical

	updated, err := repo.Create(ctx, updatedReq)
	if err != nil {
		t.Fatalf("unexpected error updating duplicate: %v", err)
	}

	if updated.ID != original.ID {
		t.Errorf("expected same issue ID %s, got %s", original.ID, updated.ID)
	}

	if updated.Title != updatedReq.Title {
		t.Errorf("expected title %q, got %q", updatedReq.Title, updated.Title)
	}

	if updated.Severity != models.SeverityCritical {
		t.Errorf("expected severity %q, got %q", models.SeverityCritical, updated.Severity)
	}

	var issueCount int64
	db.Model(&models.Issue{}).Count(&issueCount)
	if issueCount != 1 {
		t.Errorf("expected 1 issue in DB, got %d", issueCount)
	}
}

func TestIssueRepository_CreateOrUpdate_UpdatesExisting(t *testing.T) {
	ctx, db, repo := setupTestScenario(t, SetupOptions{})

	req := createTestIssue("CreateOrUpdate Existing", "test-namespace")
	original, err := repo.Create(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error creating issue: %v", err)
	}

	req.Title = "CreateOrUpdate Updated"
	updated, err := repo.CreateOrUpdate(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error on create or update: %v", err)
	}

	if updated.ID != original.ID {
		t.Errorf("expected same issue ID %s, got %s", original.ID, updated.ID)
	}

	if updated.Title != "CreateOrUpdate Updated" {
		t.Errorf("expected updated title, got %q", updated.Title)
	}

	var issueCount int64
	db.Model(&models.Issue{}).Count(&issueCount)
	if issueCount != 1 {
		t.Errorf("expected 1 issue in DB, got %d", issueCount)
	}
}

func TestIssueRepository_Create_DefaultScopeNamespaceAndState(t *testing.T) {
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	req := dto.CreateIssueRequest{
		Title:       "Defaulted Issue",
		Description: "Test description",
		Severity:    models.SeverityMajor,
		IssueType:   models.IssueTypeBuild,
		Namespace:   "test-namespace",
		Scope: dto.ScopeReqBody{
			ResourceType: "component",
			ResourceName: "fallback-component",
		},
	}

	issue, err := repo.Create(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error creating issue: %v", err)
	}

	if issue.State != models.IssueStateActive {
		t.Errorf("expected default state ACTIVE, got %q", issue.State)
	}

	if issue.Scope.ResourceNamespace != "test-namespace" {
		t.Errorf("expected scope namespace to fall back to issue namespace, got %q", issue.Scope.ResourceNamespace)
	}
}

func TestIssueRepository_FindAll_AdditionalFilters(t *testing.T) {
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	buildIssue := createTestIssue("Build failure in pipeline", "team-alpha")
	testIssue := dto.CreateIssueRequest{
		Title:       "Unit test regression",
		Description: "Regression in auth package",
		Severity:    models.SeverityCritical,
		IssueType:   models.IssueTypeTest,
		Namespace:   "team-alpha",
		Scope: dto.ScopeReqBody{
			ResourceType:      "pipelinerun",
			ResourceName:      "pipeline-xyz",
			ResourceNamespace: "team-alpha",
		},
	}

	for _, req := range []dto.CreateIssueRequest{buildIssue, testIssue} {
		if _, err := repo.Create(ctx, req); err != nil {
			t.Fatalf("failed to create test issue: %v", err)
		}
	}

	severity := models.SeverityCritical
	issueType := models.IssueTypeTest
	state := models.IssueStateActive

	tests := []struct {
		name          string
		filters       IssueQueryFilters
		expectedTotal int64
		expectedCount int
	}{
		{
			name: "severity filter",
			filters: IssueQueryFilters{
				Namespace: "team-alpha",
				Severity:  &severity,
				Limit:     10,
			},
			expectedTotal: 1,
			expectedCount: 1,
		},
		{
			name: "issue type filter",
			filters: IssueQueryFilters{
				Namespace: "team-alpha",
				IssueType: &issueType,
				Limit:     10,
			},
			expectedTotal: 1,
			expectedCount: 1,
		},
		{
			name: "state filter",
			filters: IssueQueryFilters{
				Namespace: "team-alpha",
				State:     &state,
				Limit:     10,
			},
			expectedTotal: 2,
			expectedCount: 2,
		},
		{
			name: "resource filters",
			filters: IssueQueryFilters{
				Namespace:    "team-alpha",
				ResourceType: "pipelinerun",
				ResourceName: "pipeline-xyz",
				Limit:        10,
			},
			expectedTotal: 1,
			expectedCount: 1,
		},
		{
			name: "search filter",
			filters: IssueQueryFilters{
				Namespace: "team-alpha",
				Search:    "regression",
				Limit:     10,
			},
			expectedTotal: 1,
			expectedCount: 1,
		},
		{
			name: "default limit",
			filters: IssueQueryFilters{
				Namespace: "team-alpha",
			},
			expectedTotal: 2,
			expectedCount: 2,
		},
		{
			name: "pagination",
			filters: IssueQueryFilters{
				Namespace: "team-alpha",
				Limit:     1,
				Offset:    1,
			},
			expectedTotal: 2,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundIssues, total, err := repo.FindAll(ctx, tt.filters)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if total != tt.expectedTotal {
				t.Errorf("expected total %d, got %d", tt.expectedTotal, total)
			}

			if len(foundIssues) != tt.expectedCount {
				t.Errorf("expected %d issues, got %d", tt.expectedCount, len(foundIssues))
			}
		})
	}
}

func TestIssueRepository_Update_NotFound(t *testing.T) {
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	_, err := repo.Update(ctx, "missing-issue-id", dto.UpdateIssueRequest{
		Title: "Should not update",
	})
	if err == nil {
		t.Fatal("expected error for missing issue")
	}
}

func TestIssueRepository_Update_LinksAndScope(t *testing.T) {
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	issue, err := repo.Create(ctx, createTestIssue("Update Fields", "test-namespace"))
	if err != nil {
		t.Fatalf("unexpected error creating issue: %v", err)
	}

	updated, err := repo.Update(ctx, issue.ID, dto.UpdateIssueRequest{
		Description: "Updated description",
		Severity:    models.SeverityMinor,
		IssueType:   models.IssueTypeTest,
		Namespace:   "updated-namespace",
		Links: []dto.CreateLinkRequest{
			{Title: "New Link", URL: "https://example.com/new"},
		},
		Scope: dto.ScopeReqBodyOptional{
			ResourceType:      "pipelinerun",
			ResourceName:      "pipeline-updated",
			ResourceNamespace: "updated-namespace",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error updating issue: %v", err)
	}

	if updated.Description != "Updated description" {
		t.Errorf("expected updated description, got %q", updated.Description)
	}

	if updated.Severity != models.SeverityMinor {
		t.Errorf("expected severity minor, got %q", updated.Severity)
	}

	if updated.IssueType != models.IssueTypeTest {
		t.Errorf("expected issue type test, got %q", updated.IssueType)
	}

	if updated.Namespace != "updated-namespace" {
		t.Errorf("expected namespace updated-namespace, got %q", updated.Namespace)
	}

	if len(updated.Links) != 1 || updated.Links[0].Title != "New Link" {
		t.Fatalf("expected replaced links, got %+v", updated.Links)
	}

	if updated.Scope.ResourceType != "pipelinerun" {
		t.Errorf("expected scope resource type pipelinerun, got %q", updated.Scope.ResourceType)
	}

	if updated.Scope.ResourceName != "pipeline-updated" {
		t.Errorf("expected scope resource name pipeline-updated, got %q", updated.Scope.ResourceName)
	}
}

func TestIssueRepository_Update_ResolvedAt(t *testing.T) {
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	issue, err := repo.Create(ctx, createTestIssue("Resolved At Update", "test-namespace"))
	if err != nil {
		t.Fatalf("unexpected error creating issue: %v", err)
	}

	_, err = repo.Update(ctx, issue.ID, dto.UpdateIssueRequest{
		State: models.IssueStateResolved,
	})
	if err != nil {
		t.Fatalf("unexpected error resolving issue: %v", err)
	}

	resolvedAt := time.Date(2025, 4, 30, 13, 40, 15, 0, time.UTC)
	updated, err := repo.Update(ctx, issue.ID, dto.UpdateIssueRequest{
		State:        models.IssueStateResolved,
		ResolvedAt:   resolvedAt,
		ResolvedByID: "resolver-user-id",
	})
	if err != nil {
		t.Fatalf("unexpected error updating issue: %v", err)
	}

	if updated.ResolvedAt == nil || !updated.ResolvedAt.Equal(resolvedAt) {
		t.Errorf("expected resolvedAt %v, got %v", resolvedAt, updated.ResolvedAt)
	}

	if updated.ResolvedByID != "resolver-user-id" {
		t.Errorf("expected resolvedById resolver-user-id, got %q", updated.ResolvedByID)
	}
}

func TestIssueRepository_Delete_NotFound(t *testing.T) {
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	err := repo.Delete(ctx, "missing-issue-id")
	if err == nil {
		t.Fatal("expected error for missing issue")
	}
}

func TestIssueRepository_Delete_WithRelatedIssues(t *testing.T) {
	ctx, db, repo := setupTestScenario(t, SetupOptions{})

	source, err := repo.Create(ctx, createTestIssue("Source Issue", "test-namespace"))
	if err != nil {
		t.Fatalf("unexpected error creating source issue: %v", err)
	}

	target, err := repo.Create(ctx, createTestIssueWithScope("Target Issue", "test-namespace", "component", "other-component"))
	if err != nil {
		t.Fatalf("unexpected error creating target issue: %v", err)
	}

	if err := repo.AddRelatedIssue(ctx, source.ID, target.ID); err != nil {
		t.Fatalf("unexpected error adding related issue: %v", err)
	}

	if err := repo.Delete(ctx, source.ID); err != nil {
		t.Fatalf("unexpected error deleting issue: %v", err)
	}

	var relatedCount int64
	db.Model(&models.RelatedIssue{}).Count(&relatedCount)
	if relatedCount != 0 {
		t.Errorf("expected related issues to be deleted, got %d", relatedCount)
	}
}

func TestIssueRepository_ResolveByScope(t *testing.T) {
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	req := createTestIssueWithScope("Pipeline failure", "team-alpha", "pipelinerun", "pipeline-xyz")
	issue, err := repo.Create(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error creating issue: %v", err)
	}

	count, err := repo.ResolveByScope(ctx, "pipelinerun", "pipeline-xyz", "team-alpha")
	if err != nil {
		t.Fatalf("unexpected error resolving by scope: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 resolved issue, got %d", count)
	}

	resolved, err := repo.FindByID(ctx, issue.ID)
	if err != nil {
		t.Fatalf("unexpected error reloading issue: %v", err)
	}

	if resolved.State != models.IssueStateResolved {
		t.Errorf("expected state RESOLVED, got %q", resolved.State)
	}

	if resolved.ResolvedAt == nil {
		t.Fatal("expected resolvedAt to be set")
	}

	count, err = repo.ResolveByScope(ctx, "pipelinerun", "missing-pipeline", "team-alpha")
	if err != nil {
		t.Fatalf("unexpected error resolving empty scope: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 resolved issues for missing scope, got %d", count)
	}
}

func TestIssueRepository_AddRelatedIssue(t *testing.T) {
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	source, err := repo.Create(ctx, createTestIssue("Related Source", "test-namespace"))
	if err != nil {
		t.Fatalf("unexpected error creating source issue: %v", err)
	}

	target, err := repo.Create(ctx, createTestIssueWithScope("Related Target", "test-namespace", "component", "related-target"))
	if err != nil {
		t.Fatalf("unexpected error creating target issue: %v", err)
	}

	if err := repo.AddRelatedIssue(ctx, source.ID, target.ID); err != nil {
		t.Fatalf("unexpected error adding related issue: %v", err)
	}

	reloaded, err := repo.FindByID(ctx, source.ID)
	if err != nil {
		t.Fatalf("unexpected error reloading source issue: %v", err)
	}

	if len(reloaded.RelatedFrom) != 1 {
		t.Fatalf("expected 1 related issue, got %d", len(reloaded.RelatedFrom))
	}

	if reloaded.RelatedFrom[0].TargetID != target.ID {
		t.Errorf("expected target ID %s, got %s", target.ID, reloaded.RelatedFrom[0].TargetID)
	}

	if err := repo.AddRelatedIssue(ctx, source.ID, target.ID); err == nil {
		t.Fatal("expected error when adding duplicate relationship")
	}

	if err := repo.AddRelatedIssue(ctx, source.ID, "missing-issue"); err == nil {
		t.Fatal("expected error when target issue is missing")
	}
}

func TestIssueRepository_RemoveRelatedIssue(t *testing.T) {
	ctx, _, repo := setupTestScenario(t, SetupOptions{})

	source, err := repo.Create(ctx, createTestIssue("Remove Source", "test-namespace"))
	if err != nil {
		t.Fatalf("unexpected error creating source issue: %v", err)
	}

	target, err := repo.Create(ctx, createTestIssueWithScope("Remove Target", "test-namespace", "component", "remove-target"))
	if err != nil {
		t.Fatalf("unexpected error creating target issue: %v", err)
	}

	if err := repo.AddRelatedIssue(ctx, source.ID, target.ID); err != nil {
		t.Fatalf("unexpected error adding related issue: %v", err)
	}

	if err := repo.RemoveRelatedIssue(ctx, source.ID, target.ID); err != nil {
		t.Fatalf("unexpected error removing related issue: %v", err)
	}

	reloaded, err := repo.FindByID(ctx, source.ID)
	if err != nil {
		t.Fatalf("unexpected error reloading source issue: %v", err)
	}

	if len(reloaded.RelatedFrom) != 0 {
		t.Errorf("expected related issues to be removed, got %d", len(reloaded.RelatedFrom))
	}

	if err := repo.RemoveRelatedIssue(ctx, source.ID, target.ID); err == nil {
		t.Fatal("expected error when relationship does not exist")
	}
}
