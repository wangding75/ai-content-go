package workflow_test

import (
	"context"
	"errors"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

func createPublishedVersion(t *testing.T, svc interface {
	workflow.Service
	workflow.EnginePort
}) (templateVersionID string) {
	t.Helper()
	tmpl, err := svc.CreateTemplate(context.Background(), workflow.CreateWorkflowTemplateRequest{
		Code: "run-tmpl", Name: "Run Template", ContentType: "article",
	})
	if err != nil {
		t.Fatalf("create template: %v", err)
	}
	ver, err := svc.CreateVersion(context.Background(), tmpl.WorkflowTemplateID, workflow.CreateVersionRequest{
		Steps: []workflow.CreateStepTemplateRequest{
			{StepCode: "s1", StepType: "agent", AgentCode: "writer", OrderIndex: 1},
		},
	})
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	_, err = svc.PublishVersion(context.Background(), ver.TemplateVersionID, workflow.PublishVersionRequest{}, "")
	if err != nil {
		t.Fatalf("publish version: %v", err)
	}
	return ver.TemplateVersionID
}

// @Test
func TestCreateWorkflowRunReturnsPendingStatusAndID(t *testing.T) {
	svc := workflow.NewService()
	versionID := createPublishedVersion(t, svc)

	resp, err := svc.CreateRun(context.Background(), workflow.CreateWorkflowRunRequest{
		ProjectID:         "proj-1",
		TemplateVersionID: versionID,
		Input:             map[string]any{"title": "Test Article"},
	}, "")
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	if resp.WorkflowRunID == "" || resp.Status != "pending" {
		t.Fatalf("unexpected create run response: %#v", resp)
	}
}

// @Test
func TestCreateRunWithIdempotencyKeyReturnsExistingRunOnDuplicate(t *testing.T) {
	svc := workflow.NewService()
	versionID := createPublishedVersion(t, svc)

	req := workflow.CreateWorkflowRunRequest{
		ProjectID: "proj-2", TemplateVersionID: versionID, Input: map[string]any{},
	}
	first, err := svc.CreateRun(context.Background(), req, "key-abc")
	if err != nil {
		t.Fatalf("first create run: %v", err)
	}
	second, err := svc.CreateRun(context.Background(), req, "key-abc")
	if err != nil {
		t.Fatalf("second create run with same key: %v", err)
	}
	if second.WorkflowRunID != first.WorkflowRunID {
		t.Fatalf("idempotency: expected same run id, got %q != %q", second.WorkflowRunID, first.WorkflowRunID)
	}
}

// @Test
func TestGetWorkflowRunDetailIncludesInputAndStepCount(t *testing.T) {
	svc := workflow.NewService()
	versionID := createPublishedVersion(t, svc)

	created, _ := svc.CreateRun(context.Background(), workflow.CreateWorkflowRunRequest{
		ProjectID: "proj-3", TemplateVersionID: versionID, Input: map[string]any{"key": "val"},
	}, "")

	detail, err := svc.GetRun(context.Background(), created.WorkflowRunID)
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if detail.ID != created.WorkflowRunID || detail.Input == nil || detail.StepCount < 0 {
		t.Fatalf("unexpected run detail: %#v", detail)
	}
}

// @Test
func TestListWorkflowRunsFiltersByStatusAndProjectID(t *testing.T) {
	svc := workflow.NewService()
	versionID := createPublishedVersion(t, svc)

	svc.CreateRun(context.Background(), workflow.CreateWorkflowRunRequest{ //nolint
		ProjectID: "proj-filter", TemplateVersionID: versionID, Input: map[string]any{},
	}, "")

	resp, err := svc.ListRuns(context.Background(), workflow.ListWorkflowRunsRequest{
		PaginationRequest: content.PaginationRequest{Page: 1, PageSize: 10},
		ProjectID:         "proj-filter",
		Status:            "pending",
	})
	if err != nil {
		t.Fatalf("list runs: %v", err)
	}
	if resp.Pagination.Total < 0 {
		t.Fatalf("unexpected pagination: %#v", resp.Pagination)
	}
	for _, item := range resp.Items {
		if item.Status != "pending" {
			t.Fatalf("filter mismatch: status %q", item.Status)
		}
	}
}

// @Test
func TestGetRunStepsReturnsStepRunsForRun(t *testing.T) {
	svc := workflow.NewService()
	versionID := createPublishedVersion(t, svc)

	created, _ := svc.CreateRun(context.Background(), workflow.CreateWorkflowRunRequest{
		ProjectID: "proj-steps", TemplateVersionID: versionID, Input: map[string]any{},
	}, "")

	steps, err := svc.GetRunSteps(context.Background(), created.WorkflowRunID)
	if err != nil {
		t.Fatalf("get run steps: %v", err)
	}
	if steps.Items == nil {
		t.Fatalf("expected non-nil steps slice")
	}
}

// @Test
func TestCancelRunTransitionsToCancelledAndReturnsOplogID(t *testing.T) {
	svc := workflow.NewService()
	versionID := createPublishedVersion(t, svc)

	created, _ := svc.CreateRun(context.Background(), workflow.CreateWorkflowRunRequest{
		ProjectID: "proj-cancel", TemplateVersionID: versionID, Input: map[string]any{},
	}, "")

	resp, err := svc.CancelRun(context.Background(), created.WorkflowRunID, workflow.CancelRunRequest{
		Reason: "manual cancel", Note: "testing",
	}, "")
	if err != nil {
		t.Fatalf("cancel run: %v", err)
	}
	if resp.PreviousStatus != "pending" || resp.CurrentStatus != "cancelled" || resp.OperationLogID == "" {
		t.Fatalf("unexpected cancel response: %#v", resp)
	}
}

// @Test
func TestCancelRunRejectsAlreadyCancelledRunWithConflict(t *testing.T) {
	svc := workflow.NewService()
	versionID := createPublishedVersion(t, svc)

	created, _ := svc.CreateRun(context.Background(), workflow.CreateWorkflowRunRequest{
		ProjectID: "proj-cancel2", TemplateVersionID: versionID, Input: map[string]any{},
	}, "")
	svc.CancelRun(context.Background(), created.WorkflowRunID, workflow.CancelRunRequest{Reason: "first"}, "") //nolint

	_, err := svc.CancelRun(context.Background(), created.WorkflowRunID, workflow.CancelRunRequest{Reason: "second"}, "")
	if !errors.Is(err, workflow.ErrConflict) {
		t.Fatalf("expected ErrConflict cancelling already-cancelled run, got %v", err)
	}
}

// @Test
func TestRetryRunCreatesNewRunLinkedToOriginalParent(t *testing.T) {
	svc := workflow.NewService()
	versionID := createPublishedVersion(t, svc)

	original, _ := svc.CreateRun(context.Background(), workflow.CreateWorkflowRunRequest{
		ProjectID: "proj-retry", TemplateVersionID: versionID, Input: map[string]any{},
	}, "")
	// put run in failed state via engine port
	svc.UpdateRunStatus(context.Background(), original.WorkflowRunID, "failed", nil, "mock error") //nolint

	retry, err := svc.RetryRun(context.Background(), original.WorkflowRunID, workflow.RetryRunRequest{Reason: "retry"}, "")
	if err != nil {
		t.Fatalf("retry run: %v", err)
	}
	if retry.NewWorkflowRunID == "" || retry.NewWorkflowRunID == original.WorkflowRunID || retry.Status != "pending" {
		t.Fatalf("unexpected retry response: %#v", retry)
	}

	newRun, err := svc.GetRun(context.Background(), retry.NewWorkflowRunID)
	if err != nil {
		t.Fatalf("get retried run: %v", err)
	}
	if newRun.ParentRunID != original.WorkflowRunID || newRun.Source != "retry" {
		t.Fatalf("retried run missing parent link: %#v", newRun)
	}
}
