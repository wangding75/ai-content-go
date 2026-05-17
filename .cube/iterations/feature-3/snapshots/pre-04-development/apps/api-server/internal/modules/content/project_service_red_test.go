package content_test

import (
	"context"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

// @Test
func TestListProjectsSupportsStatusAndContentTypeFilters(t *testing.T) {
	service := content.NewService()

	resp, err := service.ListProjects(context.Background(), content.ListProjectsRequest{
		PaginationRequest: content.PaginationRequest{Page: 1, PageSize: 10, Sort: "created_at", Order: "desc"},
		Status:            "active",
		ContentType:       "article",
	})
	if err != nil {
		t.Fatalf("list projects should not fail: %v", err)
	}
	if resp.Pagination.Page != 1 || resp.Pagination.PageSize != 10 || resp.Pagination.Total < 0 {
		t.Fatalf("unexpected project page: %#v", resp)
	}
	for _, item := range resp.Items {
		if item.ID == "" || item.Name == "" || item.ContentTypeID == "" || item.ContentTypeCode != "article" || item.Status != "active" || item.ProjectConfig == nil {
			t.Fatalf("project item does not satisfy filters or response contract: %#v", item)
		}
	}
}

// @Test
func TestCreateProjectRequiresExistingContentTypeAndValidSchemaConfig(t *testing.T) {
	service := content.NewService()

	resp, err := service.CreateProject(context.Background(), content.CreateProjectRequest{
		Name:          "Daily Articles",
		ContentTypeID: "1",
		ProjectConfig: map[string]any{"title": "daily briefing"},
	})
	if err != nil {
		t.Fatalf("create project should not fail: %v", err)
	}
	if resp.ProjectID == "" || resp.Status == "" {
		t.Fatalf("unexpected create project response: %#v", resp)
	}
}

// @Test
func TestCreateProjectRejectsMissingFieldsAndUnknownContentType(t *testing.T) {
	service := content.NewService()

	if _, err := service.CreateProject(context.Background(), content.CreateProjectRequest{ProjectConfig: map[string]any{}}); err == nil {
		t.Fatalf("expected validation error for missing project fields")
	}
	if _, err := service.CreateProject(context.Background(), content.CreateProjectRequest{Name: "Daily Articles", ContentTypeID: "missing", ProjectConfig: map[string]any{}}); err == nil {
		t.Fatalf("expected not found error for missing content type")
	}
}

// @Test
func TestProjectOverviewReturnsProgressPendingActionsAndCost(t *testing.T) {
	service := content.NewService()

	resp, err := service.ProjectOverview(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("project overview should not fail: %v", err)
	}
	if resp.ProjectID != "project-1" || resp.Progress < 0 || resp.Progress > 1 || resp.PendingActions < 0 || resp.Cost < 0 {
		t.Fatalf("unexpected project overview: %#v", resp)
	}
}

// @Test
func TestProjectOverviewRejectsInvalidOrMissingProjectID(t *testing.T) {
	service := content.NewService()

	if _, err := service.ProjectOverview(context.Background(), ""); err == nil {
		t.Fatalf("expected validation error for missing project id")
	}
	if _, err := service.ProjectOverview(context.Background(), "missing"); err == nil {
		t.Fatalf("expected not found error for missing project")
	}
}

// @Test
func TestPauseProjectReturnsPausedStatusAndOperationLogID(t *testing.T) {
	service := content.NewService()

	resp, err := service.PauseProject(context.Background(), "project-1", content.PauseProjectRequest{Reason: "manual review", Note: "waiting for editor"})
	if err != nil {
		t.Fatalf("pause project should not fail: %v", err)
	}
	if resp.ProjectID != "project-1" || resp.Status != "paused" || resp.OperationLogID == "" {
		t.Fatalf("unexpected pause response: %#v", resp)
	}
}

// @Test
func TestPauseProjectRejectsMissingReasonMissingProjectAndStateConflict(t *testing.T) {
	service := content.NewService()

	if _, err := service.PauseProject(context.Background(), "project-1", content.PauseProjectRequest{}); err == nil {
		t.Fatalf("expected validation error for missing pause reason")
	}
	if _, err := service.PauseProject(context.Background(), "missing", content.PauseProjectRequest{Reason: "manual review"}); err == nil {
		t.Fatalf("expected not found error for missing project")
	}
	if _, err := service.PauseProject(context.Background(), "paused-project", content.PauseProjectRequest{Reason: "manual review"}); err == nil {
		t.Fatalf("expected conflict error for project state")
	}
}
