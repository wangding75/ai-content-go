package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestIteration21NewWebAdminSchedulePageRendersRealManagementControls(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/workflow/schedules/page.tsx")
	if err != nil {
		t.Fatalf("read schedules page: %v", err)
	}
	page := string(raw)
	for _, required := range []string{"fetchWorkflowSchedules", "createWorkflowSchedule", "enableWorkflowSchedule", "disableWorkflowSchedule", "testRunWorkflowSchedule", "daily_content_count", "request_id"} {
		if !strings.Contains(page, required) {
			t.Fatalf("schedule page must contain %s", required)
		}
	}
	if strings.Contains(page, "workflow-schedules-skeleton") || strings.Contains(page, "骨架") {
		t.Fatalf("schedule page must not remain a skeleton")
	}
}

// @Test
func TestIteration21NewWebAdminExternalAutomationPageRedactsProviderToken(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/external-automation/n8n/page.tsx")
	if err != nil {
		t.Fatalf("read n8n page: %v", err)
	}
	page := string(raw)
	for _, required := range []string{"fetchExternalProviders", "createExternalProvider", "fetchExternalBindings", "createExternalBinding", "token_masked", "request_id"} {
		if !strings.Contains(page, required) {
			t.Fatalf("external automation page must contain %s", required)
		}
	}
	if strings.Contains(page, "external-automation-skeleton") || strings.Contains(page, "骨架") {
		t.Fatalf("external automation page must not remain a skeleton")
	}
}

// @Test
func TestIteration21NewWebAdminCostSummaryPageRendersAggregationStates(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/llm/cost-summary/page.tsx")
	if err != nil {
		t.Fatalf("read cost summary page: %v", err)
	}
	page := string(raw)
	for _, required := range []string{"fetchLLMCostSummary", "calls", "tokens", "cost", "by_model", "request_id"} {
		if !strings.Contains(page, required) {
			t.Fatalf("cost summary page must contain %s", required)
		}
	}
	if strings.Contains(page, "llm-cost-summary-skeleton") || strings.Contains(page, "骨架") {
		t.Fatalf("cost summary page must not remain a skeleton")
	}
}
