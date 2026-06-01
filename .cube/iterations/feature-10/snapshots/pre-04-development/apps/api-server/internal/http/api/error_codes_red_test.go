package api_test

import (
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
)

// @Test
func TestNewIterationErrorCodesAreDefinedWithCorrectValues(t *testing.T) {
	codes := map[api.ErrorCode]string{
		api.ErrorIdempotencyConflict: "IDEMPOTENCY_CONFLICT",
		api.ErrorWorkflowRunFailed:   "WORKFLOW_RUN_FAILED",
		api.ErrorAgentOutputInvalid:  "AGENT_OUTPUT_INVALID",
		api.ErrorLLMProviderError:    "LLM_PROVIDER_ERROR",
		api.ErrorExternalAutomation:  "EXTERNAL_AUTOMATION_ERROR",
	}
	for code, expected := range codes {
		if string(code) != expected {
			t.Fatalf("error code %q has wrong value, want %q", string(code), expected)
		}
	}
}
