package novel

import (
	"context"
	"errors"
	"testing"
)

// @Test
func TestTask02CreatePlanningRunPersistsRunAndCandidateSnapshot(t *testing.T) {
	svc := NewService()
	created, err := svc.CreatePlanningRun(context.Background(), "project-1", CreatePlanningRunRequest{
		Genre:             "fantasy",
		Audience:          "young-adult",
		Count:             2,
		TemplateVersionID: "wftv-1",
		InputOverride:     map[string]any{"theme": "found family"},
	}, "wfr-1", "idem-1")
	if err != nil {
		t.Fatalf("create planning run: %v", err)
	}
	detail, err := svc.GetPlanningRun(context.Background(), "project-1", created.PlanningRunID)
	if err != nil {
		t.Fatalf("get planning run: %v", err)
	}
	if detail.ID != created.PlanningRunID || detail.WorkflowRunID != "wfr-1" || len(detail.Topics) != 2 {
		t.Fatalf("expected persisted run with 2 candidates, got %#v", detail)
	}
	if detail.Topics[0].Status != "candidate" || detail.Topics[0].SnapshotID == "" {
		t.Fatalf("expected candidate snapshot state, got %#v", detail.Topics[0])
	}
}

// @Test
func TestTask02CreatePlanningRunRejectsSameIdempotencyKeyWithDifferentPayload(t *testing.T) {
	svc := NewService()
	_, err := svc.CreatePlanningRun(context.Background(), "project-1", CreatePlanningRunRequest{Genre: "fantasy", Audience: "ya", Count: 1, TemplateVersionID: "wftv-1"}, "wfr-1", "idem-1")
	if err != nil {
		t.Fatalf("first create should succeed: %v", err)
	}
	_, err = svc.CreatePlanningRun(context.Background(), "project-1", CreatePlanningRunRequest{Genre: "sci-fi", Audience: "ya", Count: 1, TemplateVersionID: "wftv-1"}, "wfr-2", "idem-1")
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("expected idempotency conflict, got %v", err)
	}
}

// @Test
func TestTask02ConfirmTopicTransitionsCandidateAndReturnsOperationLog(t *testing.T) {
	svc := NewService()
	created, err := svc.CreatePlanningRun(context.Background(), "project-1", CreatePlanningRunRequest{Genre: "fantasy", Audience: "ya", Count: 1, TemplateVersionID: "wftv-1"}, "wfr-1", "idem-1")
	if err != nil {
		t.Fatalf("create planning run: %v", err)
	}
	detail, err := svc.GetPlanningRun(context.Background(), "project-1", created.PlanningRunID)
	if err != nil {
		t.Fatalf("get planning run: %v", err)
	}
	confirmed, err := svc.ConfirmTopic(context.Background(), "project-1", detail.Topics[0].CandidateID, ConfirmTopicRequest{Note: "approve"}, "idem-confirm")
	if err != nil {
		t.Fatalf("confirm topic: %v", err)
	}
	if confirmed.PreviousStatus != "candidate" || confirmed.CurrentStatus != "confirmed" || confirmed.ConfirmedTopicID == "" || confirmed.OperationLogID == "" {
		t.Fatalf("unexpected confirm response: %#v", confirmed)
	}
}
