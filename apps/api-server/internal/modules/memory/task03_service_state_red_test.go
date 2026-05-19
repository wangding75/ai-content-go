package memory

import (
	"context"
	"errors"
	"testing"
)

// @Test
func TestTask03ServiceReturnsMemoryWithRecentSnapshotSummaryAndConfiguredPolicy(t *testing.T) {
	got, err := NewService().GetKnowledgeMemory(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("GetKnowledgeMemory returned error: %v", err)
	}
	if got.ProjectID != "project-1" || got.RecentSnapshotSummary.ID == "" {
		t.Fatalf("memory response must include project_id and recent_snapshot_summary: %+v", got)
	}
	if got.RecentWindowPolicy.ItemCount <= 0 || got.RecentWindowPolicy.TokenLimit <= 0 || got.RecentWindowPolicy.TruncationPolicy == "" {
		t.Fatalf("service must load a configured recent window policy, got %+v", got.RecentWindowPolicy)
	}
}

// @Test
func TestTask03ServiceValidatesDynamicCorrectionInputsAndRequiresIdempotencyKey(t *testing.T) {
	_, err := NewService().CorrectDynamicState(context.Background(), "project-1", CorrectDynamicStateRequest{Reason: "人工纠偏", Changes: map[string]any{"status": "corrected"}}, "idem-1")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("missing source_refs must return ErrValidation, got %v", err)
	}
	_, err = NewService().CorrectDynamicState(context.Background(), "project-1", CorrectDynamicStateRequest{Reason: "人工纠偏", Changes: map[string]any{"status": "corrected"}, SourceRefs: []string{"item_001"}}, "")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("missing idempotency key must return ErrValidation, got %v", err)
	}
}

// @Test
func TestTask03ServiceRejectsUnsupportedRecentWindowPolicyAndBudget(t *testing.T) {
	_, err := NewService().UpdateRecentWindowPolicy(context.Background(), "project-1", UpdateRecentWindowPolicyRequest{ItemCount: 5, TokenLimit: 2000, TruncationPolicy: "semantic"})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("unsupported truncation policy must return ErrValidation, got %v", err)
	}
	_, err = NewService().PreviewContext(context.Background(), "project-1", ContextPreviewRequest{Purpose: "draft_generation", Budget: 0})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("non-positive preview budget must return ErrValidation, got %v", err)
	}
	_, err = NewService().UpdateRecentWindowPolicy(context.Background(), "project-1", UpdateRecentWindowPolicyRequest{ItemCount: -1, TokenLimit: 2000, TruncationPolicy: "time"})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("negative item_count must return ErrValidation, got %v", err)
	}
	_, err = NewService().UpdateRecentWindowPolicy(context.Background(), "project-1", UpdateRecentWindowPolicyRequest{ItemCount: 5, TokenLimit: 0, TruncationPolicy: "time"})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("zero token_limit must return ErrValidation, got %v", err)
	}
	_, err = NewService().PreviewContext(context.Background(), "project-1", ContextPreviewRequest{Purpose: "draft_generation", Budget: 999999999})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("unreasonably large budget must return ErrValidation, got %v", err)
	}
}

// @Test
func TestTask03ServiceRejectsVersionConflictAndForbiddenAccess(t *testing.T) {
	_, err := NewService().UpdateStaticContext(context.Background(), "project-conflict", UpdateStaticContextRequest{StaticContext: map[string]any{"summary": "v2"}, Note: "manual"})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("version conflict must return ErrConflict, got %v", err)
	}
	_, err = NewService().GetKnowledgeMemory(context.Background(), "project-forbidden")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("forbidden project access must return ErrValidation, got %v", err)
	}
	_, err = NewService().UpdateStyleGuide(context.Background(), "project-conflict", UpdateStyleGuideRequest{StyleGuide: map[string]any{"tone": "formal"}, Note: "manual"})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("version conflict on style guide must return ErrConflict, got %v", err)
	}
}

// @Test
func TestTask03ServiceCreatesTraceableSnapshotsReportsAndIdempotentResults(t *testing.T) {
	svc := NewService()
	assembled, err := svc.AssembleContext(context.Background(), "project-1", AssembleContextRequest{Purpose: "draft_generation", Budget: 2000}, "idem-context-1")
	if err != nil {
		t.Fatalf("AssembleContext returned error: %v", err)
	}
	if assembled.ContextSnapshotID == "" || assembled.TruncationPolicy == "" || assembled.EstimatedTokens <= 0 {
		t.Fatalf("assembled context must return snapshot id, estimated tokens and truncation policy: %+v", assembled)
	}
	corrected, err := svc.CorrectDynamicState(context.Background(), "project-1", CorrectDynamicStateRequest{Reason: "人工纠偏", Changes: map[string]any{"status": "corrected"}, SourceRefs: []string{"item_001"}}, "idem-correction-1")
	if err != nil {
		t.Fatalf("CorrectDynamicState returned error: %v", err)
	}
	if corrected.MemorySnapshotID == "" || corrected.OperationLogID == "" {
		t.Fatalf("dynamic state correction must create traceable snapshot and operation log: %+v", corrected)
	}
	replayedCorrection, err := svc.CorrectDynamicState(context.Background(), "project-1", CorrectDynamicStateRequest{Reason: "人工纠偏", Changes: map[string]any{"status": "corrected"}, SourceRefs: []string{"item_001"}}, "idem-correction-1")
	if err != nil {
		t.Fatalf("CorrectDynamicState replay returned error: %v", err)
	}
	if replayedCorrection.MemorySnapshotID != corrected.MemorySnapshotID || replayedCorrection.OperationLogID != corrected.OperationLogID {
		t.Fatalf("same idempotency key and payload must return same correction result: first=%+v replay=%+v", corrected, replayedCorrection)
	}
	_, err = svc.CorrectDynamicState(context.Background(), "project-1", CorrectDynamicStateRequest{Reason: "人工纠偏", Changes: map[string]any{"status": "different"}, SourceRefs: []string{"item_001"}}, "idem-correction-1")
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("same idempotency key with different correction payload must return ErrIdempotencyConflict, got %v", err)
	}

	updated, err := svc.UpdateDynamicState(context.Background(), "content-item-1", UpdateDynamicStateRequest{Summary: "summary", Changes: map[string]any{"state": "changed"}, SourceVersionID: "version-1"}, "idem-state-1")
	if err != nil {
		t.Fatalf("UpdateDynamicState returned error: %v", err)
	}
	if updated.MemorySnapshotID == "" || updated.DynamicStateVersion <= 1 {
		t.Fatalf("dynamic state update must create snapshot and increment version: %+v", updated)
	}
	replayedState, err := svc.UpdateDynamicState(context.Background(), "content-item-1", UpdateDynamicStateRequest{Summary: "summary", Changes: map[string]any{"state": "changed"}, SourceVersionID: "version-1"}, "idem-state-1")
	if err != nil {
		t.Fatalf("UpdateDynamicState replay returned error: %v", err)
	}
	if replayedState.MemorySnapshotID != updated.MemorySnapshotID || replayedState.DynamicStateVersion != updated.DynamicStateVersion {
		t.Fatalf("same idempotency key and payload must return same dynamic state update result: first=%+v replay=%+v", updated, replayedState)
	}
	_, err = svc.UpdateDynamicState(context.Background(), "content-item-1", UpdateDynamicStateRequest{Summary: "summary", Changes: map[string]any{"state": "different"}, SourceVersionID: "version-1"}, "idem-state-1")
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("same idempotency key with different dynamic state payload must return ErrIdempotencyConflict, got %v", err)
	}

	replayedContext, err := svc.AssembleContext(context.Background(), "project-1", AssembleContextRequest{Purpose: "draft_generation", Budget: 2000}, "idem-context-1")
	if err != nil {
		t.Fatalf("AssembleContext replay returned error: %v", err)
	}
	if replayedContext.ContextSnapshotID != assembled.ContextSnapshotID {
		t.Fatalf("same idempotency key and payload must return same context snapshot: first=%+v replay=%+v", assembled, replayedContext)
	}
	_, err = svc.AssembleContext(context.Background(), "project-1", AssembleContextRequest{Purpose: "draft_generation", Budget: 3000}, "idem-context-1")
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("same idempotency key with different context payload must return ErrIdempotencyConflict, got %v", err)
	}

	report, err := svc.CreateConsistencyReport(context.Background(), "project-1", CreateConsistencyReportRequest{Range: map[string]any{"latest": true}, Scope: "project", SeverityThreshold: "low"}, "idem-report-1")
	if err != nil {
		t.Fatalf("CreateConsistencyReport returned error: %v", err)
	}
	if report.ReportID == "" || report.Status != string(ReportStatusPending) {
		t.Fatalf("report creation must return trackable pending report, got %+v", report)
	}
	replayedReport, err := svc.CreateConsistencyReport(context.Background(), "project-1", CreateConsistencyReportRequest{Range: map[string]any{"latest": true}, Scope: "project", SeverityThreshold: "low"}, "idem-report-1")
	if err != nil {
		t.Fatalf("CreateConsistencyReport replay returned error: %v", err)
	}
	if replayedReport.ReportID != report.ReportID || replayedReport.Status != report.Status {
		t.Fatalf("same idempotency key and payload must return same report result: first=%+v replay=%+v", report, replayedReport)
	}
	_, err = svc.CreateConsistencyReport(context.Background(), "project-1", CreateConsistencyReportRequest{Range: map[string]any{"latest": true}, Scope: "project", SeverityThreshold: "high"}, "idem-report-1")
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("same idempotency key with different report payload must return ErrIdempotencyConflict, got %v", err)
	}
}
