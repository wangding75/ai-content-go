package workflow_test

import (
	"context"
	"errors"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

// @Test
func TestListWorkflowTemplatesReturnsPaginatedResults(t *testing.T) {
	svc := workflow.NewService()

	resp, err := svc.ListTemplates(context.Background(), workflow.ListWorkflowTemplatesRequest{
		PaginationRequest: content.PaginationRequest{Page: 1, PageSize: 10, Sort: "created_at", Order: "desc"},
	})
	if err != nil {
		t.Fatalf("list templates should not fail: %v", err)
	}
	if resp.Pagination.Page != 1 || resp.Pagination.PageSize != 10 || resp.Pagination.Total < 0 {
		t.Fatalf("unexpected pagination: %#v", resp.Pagination)
	}
}

// @Test
func TestCreateWorkflowTemplateReturnsDraftStatusAndID(t *testing.T) {
	svc := workflow.NewService()

	resp, err := svc.CreateTemplate(context.Background(), workflow.CreateWorkflowTemplateRequest{
		Code:        "article-workflow",
		Name:        "Article Workflow",
		ContentType: "article",
		Category:    "editorial",
	})
	if err != nil {
		t.Fatalf("create template should not fail: %v", err)
	}
	if resp.WorkflowTemplateID == "" || resp.Status != "draft" {
		t.Fatalf("unexpected create response: %#v", resp)
	}
}

// @Test
func TestCreateWorkflowTemplateRejectsDuplicateCodeWithConflictError(t *testing.T) {
	svc := workflow.NewService()

	req := workflow.CreateWorkflowTemplateRequest{Code: "dup-code", Name: "Dup", ContentType: "article"}
	if _, err := svc.CreateTemplate(context.Background(), req); err != nil {
		t.Fatalf("first create should not fail: %v", err)
	}
	_, err := svc.CreateTemplate(context.Background(), req)
	if !errors.Is(err, workflow.ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate code, got %v", err)
	}
}

// @Test
func TestGetWorkflowTemplateReturnsExistingTemplate(t *testing.T) {
	svc := workflow.NewService()

	created, err := svc.CreateTemplate(context.Background(), workflow.CreateWorkflowTemplateRequest{
		Code: "get-test", Name: "Get Test", ContentType: "video",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	tmpl, err := svc.GetTemplate(context.Background(), created.WorkflowTemplateID)
	if err != nil {
		t.Fatalf("get should not fail: %v", err)
	}
	if tmpl.ID != created.WorkflowTemplateID || tmpl.Code != "get-test" {
		t.Fatalf("unexpected template: %#v", tmpl)
	}
}

// @Test
func TestGetWorkflowTemplateReturnsNotFoundForMissingID(t *testing.T) {
	svc := workflow.NewService()

	if _, err := svc.GetTemplate(context.Background(), "missing-id"); !errors.Is(err, workflow.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// @Test
func TestCreateWorkflowVersionAndListVersionsForTemplate(t *testing.T) {
	svc := workflow.NewService()

	tmpl, err := svc.CreateTemplate(context.Background(), workflow.CreateWorkflowTemplateRequest{
		Code: "v-test", Name: "Version Test", ContentType: "article",
	})
	if err != nil {
		t.Fatalf("create template: %v", err)
	}

	created, err := svc.CreateVersion(context.Background(), tmpl.WorkflowTemplateID, workflow.CreateVersionRequest{
		InputSchema:  map[string]any{"title": "string"},
		OutputSchema: map[string]any{"result": "string"},
		Steps: []workflow.CreateStepTemplateRequest{
			{StepCode: "step-1", StepType: "agent", AgentCode: "writer", OrderIndex: 1},
		},
	})
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	if created.TemplateVersionID == "" || created.StepCount != 1 || created.Status != "draft" {
		t.Fatalf("unexpected version response: %#v", created)
	}

	versions, err := svc.ListVersions(context.Background(), tmpl.WorkflowTemplateID, workflow.PaginationRequest{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(versions.Items) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions.Items))
	}
}

// @Test
func TestGetVersionDetailIncludesStepTemplates(t *testing.T) {
	svc := workflow.NewService()

	tmpl, _ := svc.CreateTemplate(context.Background(), workflow.CreateWorkflowTemplateRequest{
		Code: "detail-test", Name: "Detail", ContentType: "article",
	})
	ver, _ := svc.CreateVersion(context.Background(), tmpl.WorkflowTemplateID, workflow.CreateVersionRequest{
		Steps: []workflow.CreateStepTemplateRequest{
			{StepCode: "step-1", StepType: "agent", AgentCode: "writer", OrderIndex: 1},
			{StepCode: "step-2", StepType: "system_task", OrderIndex: 2},
		},
	})

	detail, err := svc.GetVersion(context.Background(), ver.TemplateVersionID)
	if err != nil {
		t.Fatalf("get version: %v", err)
	}
	if len(detail.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(detail.Steps))
	}
}

// @Test
func TestPublishVersionTransitionsFromDraftToPublished(t *testing.T) {
	svc := workflow.NewService()

	tmpl, _ := svc.CreateTemplate(context.Background(), workflow.CreateWorkflowTemplateRequest{
		Code: "pub-test", Name: "Pub", ContentType: "article",
	})
	ver, _ := svc.CreateVersion(context.Background(), tmpl.WorkflowTemplateID, workflow.CreateVersionRequest{
		Steps: []workflow.CreateStepTemplateRequest{{StepCode: "s1", StepType: "agent", AgentCode: "a", OrderIndex: 1}},
	})

	resp, err := svc.PublishVersion(context.Background(), ver.TemplateVersionID, workflow.PublishVersionRequest{Note: "first release"}, "")
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if resp.PreviousStatus != "draft" || resp.CurrentStatus != "published" || resp.OperationLogID == "" {
		t.Fatalf("unexpected publish response: %#v", resp)
	}
}

// @Test
func TestPublishVersionRejectsAlreadyPublishedVersionWithConflict(t *testing.T) {
	svc := workflow.NewService()

	tmpl, _ := svc.CreateTemplate(context.Background(), workflow.CreateWorkflowTemplateRequest{
		Code: "pub-dup", Name: "Pub Dup", ContentType: "article",
	})
	ver, _ := svc.CreateVersion(context.Background(), tmpl.WorkflowTemplateID, workflow.CreateVersionRequest{
		Steps: []workflow.CreateStepTemplateRequest{{StepCode: "s1", StepType: "agent", AgentCode: "a", OrderIndex: 1}},
	})
	svc.PublishVersion(context.Background(), ver.TemplateVersionID, workflow.PublishVersionRequest{}, "") //nolint

	_, err := svc.PublishVersion(context.Background(), ver.TemplateVersionID, workflow.PublishVersionRequest{}, "")
	if !errors.Is(err, workflow.ErrConflict) {
		t.Fatalf("expected ErrConflict for re-publish, got %v", err)
	}
}
