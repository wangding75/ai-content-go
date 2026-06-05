package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestWebAdminAPIClientDeclaresIteration2WorkflowTypes(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/lib/api.ts")
	if err != nil {
		t.Fatalf("read api.ts: %v", err)
	}
	api := string(raw)

	for _, required := range []string{
		"WorkflowTemplateResponse",
		"WorkflowTemplateVersionResponse",
		"WorkflowStepTemplateResponse",
		"WorkflowRunResponse",
		"WorkflowStepRunResponse",
		"fetchWorkflowTemplates",
		"createWorkflowTemplate",
		"createWorkflowVersion",
		"publishWorkflowVersion",
		"fetchWorkflowRuns",
		"createWorkflowRun",
		"cancelWorkflowRun",
		"retryWorkflowRun",
	} {
		if !strings.Contains(api, required) {
			t.Fatalf("expected api.ts to declare %s", required)
		}
	}
}

// @Test
func TestWebAdminAPIClientDeclaresIteration2AgentAndLLMTypes(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/lib/api.ts")
	if err != nil {
		t.Fatalf("read api.ts: %v", err)
	}
	api := string(raw)

	for _, required := range []string{
		"AgentTaskResponse",
		"AgentTaskDetailResponse",
		"LLMCallLogResponse",
		"fetchAgentTasks",
		"fetchAgentTask",
		"fetchLLMCallLogs",
		"fetchLLMCallLog",
		"llm_call_log_ids",
	} {
		if !strings.Contains(api, required) {
			t.Fatalf("expected api.ts to declare %s", required)
		}
	}
}
