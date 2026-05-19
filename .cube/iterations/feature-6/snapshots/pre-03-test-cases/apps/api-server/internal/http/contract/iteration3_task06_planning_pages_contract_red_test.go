package contract_test

import (
	"os"
	"strings"
	"testing"
)

func readTask06File(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

// @Test
func TestTask06PlanningPageCoversRunCreationListDetailAndTopicDialog(t *testing.T) {
	page := readTask06File(t, "../../../../../apps/web-admin/app/projects/[projectId]/planning/page.tsx")
	for _, required := range []string{
		"ProjectWorkspaceNav",
		"createPlanningRun",
		"fetchPlanningRuns",
		"fetchPlanningRun",
		"confirmTopic",
		"dialog-backdrop",
		"候选选题",
		"加载态",
		"空状态",
		"role=\"alert\"",
		"role=\"status\"",
		"request_id",
		"Idempotency",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected planning page to contain %s", required)
		}
	}
}

// @Test
func TestTask06TopicsPageProvidesDirectRecoveryConfirmationEntry(t *testing.T) {
	page := readTask06File(t, "../../../../../apps/web-admin/app/projects/[projectId]/planning/topics/page.tsx")
	for _, required := range []string{
		"ProjectWorkspaceNav",
		"fetchPlanningRuns",
		"fetchPlanningRun",
		"confirmTopic",
		"候选选题确认",
		"role=\"alert\"",
		"role=\"status\"",
		"request_id",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected topics page to contain %s", required)
		}
	}
}
