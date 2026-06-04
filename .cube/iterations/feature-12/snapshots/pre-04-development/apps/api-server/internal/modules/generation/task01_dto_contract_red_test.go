package generation

import (
	"errors"
	"reflect"
	"testing"
)

// @Test
func TestTask01DTOContractDeclaresGenerationRunAndContentItemFields(t *testing.T) {
	cases := []struct {
		model any
		field string
		json  string
	}{
		{CreateGenerationRunRequest{}, "ConfirmedTopicID", "confirmed_topic_id"},
		{CreateGenerationRunRequest{}, "StartSequenceNo", "start_sequence_no"},
		{CreateGenerationRunResponse{}, "GenerationRunID", "generation_run_id"},
		{GenerationRunResponse{}, "WorkflowRunID", "workflow_run_id"},
		{ContentItemResponse{}, "ContentTypeCode", "content_type_code"},
		{ContentItemDetailResponse{}, "Extension", "extension"},
	}
	for _, tc := range cases {
		field, ok := reflect.TypeOf(tc.model).FieldByName(tc.field)
		if !ok {
			t.Fatalf("expected %T to declare %s", tc.model, tc.field)
		}
		if got := field.Tag.Get("json"); got != tc.json {
			t.Fatalf("expected %T.%s json tag %q, got %q", tc.model, tc.field, tc.json, got)
		}
	}
}

// @Test
func TestTask01DTOContractExposesDeclaredStatusSetsAndErrors(t *testing.T) {
	generationStatuses := []string{GenerationRunPending, GenerationRunRunning, GenerationRunSucceeded, GenerationRunFailed, GenerationRunRetrying}
	contentStatuses := []string{ContentItemPlanned, ContentItemGenerating, ContentItemGenerated, ContentItemGenerationFailed, ContentItemPendingReview}
	for _, want := range []string{"pending", "running", "succeeded", "failed", "retrying"} {
		if !containsString(generationStatuses, want) {
			t.Fatalf("generation status %q missing from %#v", want, generationStatuses)
		}
	}
	for _, want := range []string{"planned", "generating", "generated", "generation_failed", "pending_review"} {
		if !containsString(contentStatuses, want) {
			t.Fatalf("content item status %q missing from %#v", want, contentStatuses)
		}
	}
	for name, err := range map[string]error{
		"validation":           ErrValidation,
		"not_found":            ErrNotFound,
		"conflict":             ErrConflict,
		"idempotency_conflict": ErrIdempotencyConflict,
		"workflow_run_failed":  ErrWorkflowRunFailed,
		"llm_provider_error":   ErrLLMProviderError,
	} {
		if err == nil || !errors.Is(err, err) {
			t.Fatalf("expected %s sentinel error to be usable", name)
		}
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
