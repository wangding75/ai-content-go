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
func TestTask03ServiceReturnsNotFoundForUnknownProject(t *testing.T) {
	svc := NewService()
	projectMethods := []struct {
		name string
		fn   func() error
	}{
		{"GetKnowledgeMemory", func() error { _, err := svc.GetKnowledgeMemory(context.Background(), "unknown-project"); return err }},
		{"UpdateStaticContext", func() error { _, err := svc.UpdateStaticContext(context.Background(), "unknown-project", UpdateStaticContextRequest{StaticContext: map[string]any{"summary": "v2"}, Note: "manual"}); return err }},
		{"UpdateStyleGuide", func() error { _, err := svc.UpdateStyleGuide(context.Background(), "unknown-project", UpdateStyleGuideRequest{StyleGuide: map[string]any{"tone": "formal"}, Note: "manual"}); return err }},
		{"CorrectDynamicState", func() error { _, err := svc.CorrectDynamicState(context.Background(), "unknown-project", CorrectDynamicStateRequest{Reason: "人工纠偏", Changes: map[string]any{"status": "corrected"}, SourceRefs: []string{"item_001"}}, "idem-key"); return err }},
		{"UpdateRecentWindowPolicy", func() error { _, err := svc.UpdateRecentWindowPolicy(context.Background(), "unknown-project", UpdateRecentWindowPolicyRequest{ItemCount: 5, TokenLimit: 2000, TruncationPolicy: "time"}); return err }},
		{"ListSnapshots", func() error { _, err := svc.ListSnapshots(context.Background(), "unknown-project", ListSnapshotsRequest{Page: 1, PageSize: 20}); return err }},
		{"PreviewContext", func() error { _, err := svc.PreviewContext(context.Background(), "unknown-project", ContextPreviewRequest{Purpose: "draft_generation", Budget: 2000}); return err }},
		{"AssembleContext", func() error { _, err := svc.AssembleContext(context.Background(), "unknown-project", AssembleContextRequest{Purpose: "draft_generation", Budget: 2000}, "idem-key"); return err }},
		{"CreateConsistencyReport", func() error { _, err := svc.CreateConsistencyReport(context.Background(), "unknown-project", CreateConsistencyReportRequest{Range: map[string]any{"latest": true}, Scope: "project", SeverityThreshold: "low"}, "idem-key"); return err }},
		{"ListConsistencyReports", func() error { _, err := svc.ListConsistencyReports(context.Background(), "unknown-project", ListConsistencyReportsRequest{Page: 1, PageSize: 20}); return err }},
	}
	for _, tc := range projectMethods {
		if err := tc.fn(); !errors.Is(err, ErrNotFound) {
			t.Fatalf("%s for unknown project must return ErrNotFound, got %v", tc.name, err)
		}
	}
}

// @Test
func TestTask03ServiceRejectsForbiddenAccessWithErrForbidden(t *testing.T) {
	_, err := NewService().GetKnowledgeMemory(context.Background(), "project-forbidden")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("forbidden project must return ErrForbidden, got %v", err)
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
func TestTask03ServiceRejectsVersionConflict(t *testing.T) {
	_, err := NewService().UpdateStaticContext(context.Background(), "project-conflict", UpdateStaticContextRequest{StaticContext: map[string]any{"summary": "v2"}, Note: "manual"})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("version conflict must return ErrConflict, got %v", err)
	}
	_, err = NewService().UpdateStyleGuide(context.Background(), "project-conflict", UpdateStyleGuideRequest{StyleGuide: map[string]any{"tone": "formal"}, Note: "manual"})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("version conflict on style guide must return ErrConflict, got %v", err)
	}
}

// @Test
func TestTask03ServiceUpdateDynamicStateReturnsNotFoundForUnknownContentItem(t *testing.T) {
	_, err := NewService().UpdateDynamicState(context.Background(), "unknown-content-item", UpdateDynamicStateRequest{Summary: "summary", Changes: map[string]any{"state": "changed"}, SourceVersionID: "version-1"}, "idem-key")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdateDynamicState for unknown content item must return ErrNotFound, got %v", err)
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

// @Test
func TestTask03ServiceReportLifecycleIsQueryableAfterCreation(t *testing.T) {
	svc := NewService()
	report, err := svc.CreateConsistencyReport(context.Background(), "project-1", CreateConsistencyReportRequest{Range: map[string]any{"latest": true}, Scope: "project", SeverityThreshold: "low"}, "idem-lifecycle-1")
	if err != nil {
		t.Fatalf("CreateConsistencyReport returned error: %v", err)
	}
	got, err := svc.GetConsistencyReport(context.Background(), "project-1", report.ReportID)
	if err != nil {
		t.Fatalf("GetConsistencyReport returned error: %v", err)
	}
	if got.ID != report.ReportID || got.Status != string(ReportStatusPending) {
		t.Fatalf("get report must reflect created pending report: got=%+v", got)
	}
	list, err := svc.ListConsistencyReports(context.Background(), "project-1", ListConsistencyReportsRequest{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListConsistencyReports returned error: %v", err)
	}
	found := false
	for _, r := range list.Items {
		if r.ID == report.ReportID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListConsistencyReports must include the created report %s", report.ReportID)
	}
}

// @Test
func TestTask03ServiceGetConsistencyReportReturnsNotFoundForUnknownOrMismatchedProject(t *testing.T) {
	svc := NewService()
	_, err := svc.GetConsistencyReport(context.Background(), "project-1", "unknown-report")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetConsistencyReport for unknown report must return ErrNotFound, got %v", err)
	}
	report, _ := svc.CreateConsistencyReport(context.Background(), "project-1", CreateConsistencyReportRequest{Range: map[string]any{"latest": true}, Scope: "project", SeverityThreshold: "low"}, "idem-mismatch-1")
	_, err = svc.GetConsistencyReport(context.Background(), "seed-project", report.ReportID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetConsistencyReport for report owned by different project must return ErrNotFound, got %v", err)
	}
}

// @Test
func TestTask03ServiceLatestSnapshotSummaryReflectsCreatedSnapshots(t *testing.T) {
	svc := NewService()
	assembled, err := svc.AssembleContext(context.Background(), "project-1", AssembleContextRequest{Purpose: "draft_generation", Budget: 2000}, "idem-snapshot-summary-1")
	if err != nil {
		t.Fatalf("AssembleContext returned error: %v", err)
	}
	got, err := svc.GetKnowledgeMemory(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("GetKnowledgeMemory returned error: %v", err)
	}
	if got.RecentSnapshotSummary.ID != assembled.ContextSnapshotID {
		t.Fatalf("recent_snapshot_summary must reflect the latest created snapshot: got=%s want=%s", got.RecentSnapshotSummary.ID, assembled.ContextSnapshotID)
	}
}
