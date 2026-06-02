package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestIteration21WebAdminAPIClientDeclaresScheduleExternalAndCostSummaryFunctions(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/lib/api.ts")
	if err != nil {
		t.Fatalf("read api.ts: %v", err)
	}
	api := string(raw)
	for _, required := range []string{
		"WorkflowScheduleResponse",
		"ScheduleTriggerResponse",
		"LLMCostSummaryResponse",
		"ExternalProviderResponse",
		"ExternalBindingResponse",
		"fetchWorkflowSchedules",
		"createWorkflowSchedule",
		"enableWorkflowSchedule",
		"disableWorkflowSchedule",
		"testRunWorkflowSchedule",
		"fetchScheduleTriggers",
		"fetchLLMCostSummary",
		"fetchExternalProviders",
		"createExternalProvider",
		"fetchExternalBindings",
		"createExternalBinding",
	} {
		if !strings.Contains(api, required) {
			t.Fatalf("expected api.ts to declare %s", required)
		}
	}
}

// @Test
func TestIteration21GlobalNavigationDeclaresNewRoutesWithActiveMatches(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/global-nav.tsx")
	if err != nil {
		t.Fatalf("read global-nav.tsx: %v", err)
	}
	nav := string(raw)
	for _, required := range []string{
		"/workflow/schedules",
		"生产计划 / 调度管理",
		"/external-automation/n8n",
		"外部自动化 / n8n",
		"/llm/cost-summary",
		"成本汇总",
		"aria-current",
	} {
		if !strings.Contains(nav, required) {
			t.Fatalf("expected global-nav.tsx to contain %s", required)
		}
	}
}
