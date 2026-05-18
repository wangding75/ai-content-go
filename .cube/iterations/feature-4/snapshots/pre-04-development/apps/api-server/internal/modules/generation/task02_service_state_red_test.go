package generation

import (
	"context"
	"errors"
	"testing"
)

// @Test
func TestTask02CreateGenerationRunPersistsPendingRunAndPendingReviewContent(t *testing.T) {
	svc := NewService()
	created, err := svc.CreateGenerationRun(context.Background(), "project-1", CreateGenerationRunRequest{
		ConfirmedTopicID:   "topic-1",
		WorldviewVersionID: "worldview-v1",
		ArcID:              "arc-1",
		TargetCount:        1,
		StartSequenceNo:    1,
		TemplateVersionID:  "wftv-generation",
		GenerationConfig:   map[string]any{"tone": "hopeful"},
	}, "wfr-1", "idem-1")
	if err != nil {
		t.Fatalf("create generation run: %v", err)
	}
	if created.GenerationRunID == "" || created.WorkflowRunID != "wfr-1" || created.Status != GenerationRunPending {
		t.Fatalf("unexpected create response: %#v", created)
	}
	detail, err := svc.GetGenerationRun(context.Background(), created.GenerationRunID)
	if err != nil {
		t.Fatalf("get generation run: %v", err)
	}
	if detail.Status != GenerationRunPending || len(detail.ContentItems) != 1 {
		t.Fatalf("expected pending run with one content item, got %#v", detail)
	}
	if detail.ContentItems[0].Status != ContentItemPendingReview || detail.ContentItems[0].GenerationRunID != created.GenerationRunID {
		t.Fatalf("expected generated content to enter pending_review, got %#v", detail.ContentItems[0])
	}
}

// @Test
func TestTask02CreateGenerationRunRejectsSameIdempotencyKeyWithDifferentPayload(t *testing.T) {
	svc := NewService()
	_, err := svc.CreateGenerationRun(context.Background(), "project-1", CreateGenerationRunRequest{ConfirmedTopicID: "topic-1", WorldviewVersionID: "worldview-v1", ArcID: "arc-1", TargetCount: 1, StartSequenceNo: 1, TemplateVersionID: "wftv-generation"}, "wfr-1", "idem-1")
	if err != nil {
		t.Fatalf("first create should succeed: %v", err)
	}
	_, err = svc.CreateGenerationRun(context.Background(), "project-1", CreateGenerationRunRequest{ConfirmedTopicID: "topic-2", WorldviewVersionID: "worldview-v1", ArcID: "arc-1", TargetCount: 1, StartSequenceNo: 1, TemplateVersionID: "wftv-generation"}, "wfr-2", "idem-1")
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("expected idempotency conflict, got %v", err)
	}
}

// @Test
func TestTask02RetryGenerationRunCreatesNewRunWithoutOverwritingOriginal(t *testing.T) {
	svc := NewService()
	original, err := svc.CreateGenerationRun(context.Background(), "project-1", CreateGenerationRunRequest{ConfirmedTopicID: "topic-1", WorldviewVersionID: "worldview-v1", ArcID: "arc-1", TargetCount: 1, StartSequenceNo: 1, TemplateVersionID: "wftv-generation"}, "wfr-1", "idem-1")
	if err != nil {
		t.Fatalf("create original run: %v", err)
	}
	retry, err := svc.RetryGenerationRun(context.Background(), original.GenerationRunID, RetryGenerationRunRequest{Reason: "provider recovered", InputOverride: map[string]any{"temperature": 0.5}}, "wfr-2", "idem-retry")
	if err != nil {
		t.Fatalf("retry generation run: %v", err)
	}
	if retry.NewGenerationRunID == "" || retry.NewGenerationRunID == original.GenerationRunID || retry.WorkflowRunID != "wfr-2" || retry.OperationLogID == "" {
		t.Fatalf("unexpected retry response: %#v", retry)
	}
	originalDetail, err := svc.GetGenerationRun(context.Background(), original.GenerationRunID)
	if err != nil {
		t.Fatalf("get original run: %v", err)
	}
	if originalDetail.ID != original.GenerationRunID || originalDetail.RetryOfGenerationRunID != "" {
		t.Fatalf("original run should remain unchanged, got %#v", originalDetail.GenerationRunResponse)
	}
}
