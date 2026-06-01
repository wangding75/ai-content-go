package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestWorkflowRunsListPageDeclaresAPICallsAndTriggerButton(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/workflow/runs/page.tsx")
	if err != nil {
		t.Fatalf("read runs list page: %v", err)
	}
	page := string(raw)

	for _, required := range []string{
		"fetchWorkflowRuns",
		"createWorkflowRun",
		"WorkflowRunResponse",
		"project_id",
		"template_version_id",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected runs page to reference %s", required)
		}
	}
}

// @Test
func TestWorkflowRunDetailPageDeclaresStepsCancelRetryAndLLMLinks(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/workflow/runs/[id]/page.tsx")
	if err != nil {
		t.Fatalf("read run detail page: %v", err)
	}
	page := string(raw)

	for _, required := range []string{
		"cancelWorkflowRun",
		"retryWorkflowRun",
		"WorkflowStepRunResponse",
		"step_count",
		"agent_task_count",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected run detail page to reference %s", required)
		}
	}
}
