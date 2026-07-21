package dto

import (
	"testing"
	"time"

	"github.com/konflux-ci/kite/internal/models"
)

func TestCreateIssueRequest_GetResolvedByID(t *testing.T) {
	req := CreateIssueRequest{
		Title:       "Test Issue",
		Description: "Test description",
		Severity:    models.SeverityMajor,
		IssueType:   models.IssueTypeBuild,
		Namespace:   "team-alpha",
		Scope: ScopeReqBody{
			ResourceType: "component",
			ResourceName: "test-component",
		},
	}

	if got := req.GetResolvedByID(); got != "" {
		t.Errorf("expected empty resolvedById for create requests, got %q", got)
	}
}

func TestCreateIssueRequest_GetResolvedAt(t *testing.T) {
	req := CreateIssueRequest{
		Title:       "Test Issue",
		Description: "Test description",
		Severity:    models.SeverityMajor,
		IssueType:   models.IssueTypeBuild,
		Namespace:   "team-alpha",
		Scope: ScopeReqBody{
			ResourceType: "component",
			ResourceName: "test-component",
		},
	}

	if got := req.GetResolvedAt(); !got.IsZero() {
		t.Errorf("expected zero resolvedAt for create requests, got %v", got)
	}
}

func TestUpdateIssueRequest_GetResolvedByID(t *testing.T) {
	expected := "f67079c2-ce41-4bf9-bfb5-fbd9dbc1cf3c"
	req := UpdateIssueRequest{
		State:        models.IssueStateResolved,
		ResolvedByID: expected,
	}

	if got := req.GetResolvedByID(); got != expected {
		t.Errorf("expected resolvedById %q, got %q", expected, got)
	}
}

func TestUpdateIssueRequest_GetResolvedAt(t *testing.T) {
	expected := time.Date(2025, 4, 30, 13, 40, 15, 0, time.UTC)
	req := UpdateIssueRequest{
		ResolvedAt: expected,
	}

	if got := req.GetResolvedAt(); !got.Equal(expected) {
		t.Errorf("expected resolvedAt %v, got %v", expected, got)
	}
}
