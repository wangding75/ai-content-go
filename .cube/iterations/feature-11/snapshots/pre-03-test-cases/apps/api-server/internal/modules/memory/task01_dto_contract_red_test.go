package memory

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

// @Test
func TestTask01KnowledgeMemoryResponseDeclaresStableJSONContract(t *testing.T) {
	responseType := reflect.TypeOf(KnowledgeMemoryResponse{})
	for _, field := range []string{"id", "project_id", "static_context", "dynamic_state", "recent_window_policy", "style_guide", "version", "updated_at", "recent_snapshot_summary"} {
		if !hasJSONField(responseType, field) {
			t.Fatalf("KnowledgeMemoryResponse missing json field %q", field)
		}
	}

	payload, err := json.Marshal(KnowledgeMemoryResponse{
		ID:                    "memory-project-1",
		ProjectID:             "project-1",
		StaticContext:         map[string]any{"world": "stable"},
		DynamicState:          map[string]any{"status": "active"},
		RecentWindowPolicy:    RecentWindowPolicy{ItemCount: 5, TokenLimit: 2000, TruncationPolicy: "time"},
		StyleGuide:            map[string]any{"tone": "consistent"},
		Version:               7,
		UpdatedAt:             time.Date(2026, 5, 19, 8, 0, 0, 0, time.UTC),
		RecentSnapshotSummary: SnapshotSummaryResponse{ID: "snapshot-1", SourceType: string(SnapshotSourceAssembleContext), EstimatedTokens: 1200, TruncationPolicy: "time", CreatedAt: time.Date(2026, 5, 19, 8, 0, 0, 0, time.UTC)},
	})
	if err != nil {
		t.Fatalf("marshal KnowledgeMemoryResponse: %v", err)
	}
	if !strings.Contains(string(payload), "recent_snapshot_summary") {
		t.Fatalf("serialized response must expose recent_snapshot_summary: %s", payload)
	}
}

// @Test
func TestTask01StatusesErrorsAndIdempotencyDTOAreDeclared(t *testing.T) {
	for _, status := range []ConsistencyReportStatus{ReportStatusPending, ReportStatusRunning, ReportStatusCompleted, ReportStatusFailed} {
		if status == "" {
			t.Fatalf("report status must not be empty")
		}
	}
	for _, sourceType := range []SnapshotSourceType{SnapshotSourceAssembleContext, SnapshotSourceDynamicStateUpdate, SnapshotSourceDynamicStateCorrection} {
		if sourceType == "" {
			t.Fatalf("snapshot source type must not be empty")
		}
	}
	for _, errValue := range []error{ErrValidation, ErrNotFound, ErrConflict, ErrIdempotencyConflict} {
		if errValue == nil {
			t.Fatalf("memory domain errors must be declared")
		}
	}
	content, err := os.ReadFile("dto.go")
	if err != nil {
		t.Fatalf("read dto.go: %v", err)
	}
	if !strings.Contains(string(content), "IdempotencyRecord") {
		t.Fatalf("dto.go must declare IdempotencyRecord contract for persistent idempotency")
	}
}

// @Test
func TestTask01ConsistencyIssueRequiresStructuredFields(t *testing.T) {
	issueType := reflect.TypeOf(ConsistencyIssue{})
	for _, field := range []string{"issue_id", "severity", "type", "title", "description", "affected_content_items", "suggestion"} {
		if !hasJSONField(issueType, field) {
			t.Fatalf("ConsistencyIssue missing json field %q", field)
		}
	}
}

func hasJSONField(t reflect.Type, jsonName string) bool {
	for i := 0; i < t.NumField(); i++ {
		name := strings.Split(t.Field(i).Tag.Get("json"), ",")[0]
		if name == jsonName {
			return true
		}
	}
	return false
}
