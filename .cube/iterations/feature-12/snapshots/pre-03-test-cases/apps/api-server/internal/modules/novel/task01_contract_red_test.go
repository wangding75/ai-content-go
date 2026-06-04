package novel

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

// @Test
func TestTask01ServiceContractCreatesPlanningRunDTO(t *testing.T) {
	svc := NewService()
	resp, err := svc.CreatePlanningRun(context.Background(), "project-1", CreatePlanningRunRequest{
		Genre:             "fantasy",
		Audience:          "young-adult",
		Count:             3,
		TemplateVersionID: "wftv-1",
		InputOverride:     map[string]any{"tone": "hopeful"},
	}, "wfr-1", "idem-1")
	if err != nil {
		t.Fatalf("expected create planning run to satisfy contract, got %v", err)
	}
	if resp.PlanningRunID == "" || resp.WorkflowRunID != "wfr-1" || resp.Status != "pending" {
		t.Fatalf("unexpected planning run response: %#v", resp)
	}
}

// @Test
func TestTask01ServiceContractExposesAllDeclaredErrorSentinels(t *testing.T) {
	for name, err := range map[string]error{
		"validation":           ErrValidation,
		"not_found":            ErrNotFound,
		"forbidden":            ErrForbidden,
		"conflict":             ErrConflict,
		"idempotency_conflict": ErrIdempotencyConflict,
		"workflow_run_failed":  ErrWorkflowRunFailed,
	} {
		if err == nil || !errors.Is(err, err) {
			t.Fatalf("expected %s sentinel error to be usable", name)
		}
	}
}

// @Test
func TestTask01DTOsDeclareJSONFieldsForNovelPackResources(t *testing.T) {
	cases := []struct {
		model any
		field string
		json  string
	}{
		{CreatePlanningRunRequest{}, "TemplateVersionID", "template_version_id"},
		{CreatePlanningRunResponse{}, "PlanningRunID", "planning_run_id"},
		{PlanningRunResponse{}, "CandidateCount", "candidate_count"},
		{TopicCandidateResponse{}, "ConfirmedTopicID", "confirmed_topic_id,omitempty"},
		{WorldviewResponse{}, "ForbiddenRules", "forbidden_rules"},
		{CreateCharacterResponse{}, "OperationLogID", "operation_log_id"},
		{ArcResponse{}, "OrderIndex", "order_index"},
	}
	for _, tc := range cases {
		field, ok := reflect.TypeOf(tc.model).FieldByName(tc.field)
		if !ok {
			t.Fatalf("expected %T to declare field %s", tc.model, tc.field)
		}
		if got := field.Tag.Get("json"); got != tc.json {
			t.Fatalf("expected %T.%s json tag %q, got %q", tc.model, tc.field, tc.json, got)
		}
	}
}
