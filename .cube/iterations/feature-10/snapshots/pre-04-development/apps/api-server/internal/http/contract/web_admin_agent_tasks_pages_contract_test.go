package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestAgentTasksListPageDeclaresFilterFields(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/agent/tasks/page.tsx")
	if err != nil {
		t.Fatalf("read agent tasks page: %v", err)
	}
	page := string(raw)

	for _, required := range []string{
		"fetchAgentTasks",
		"AgentTaskResponse",
		"workflow_run_id",
		"step_run_id",
		"agent_code",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected agent tasks page to reference %s", required)
		}
	}
}

// @Test
func TestAgentTaskDetailPageDeclaresLLMCallLogLinks(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/agent/tasks/[id]/page.tsx")
	if err != nil {
		t.Fatalf("read agent task detail page: %v", err)
	}
	page := string(raw)

	for _, required := range []string{
		"fetchAgentTask",
		"AgentTaskDetailResponse",
		"llm_call_log_ids",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected agent task detail page to reference %s", required)
		}
	}
}
