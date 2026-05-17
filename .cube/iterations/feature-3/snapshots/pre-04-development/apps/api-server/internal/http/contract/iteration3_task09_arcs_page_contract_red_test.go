package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestTask09ArcsPageSupportsSortingListLoadingEmptyAndErrorStates(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/projects/[projectId]/novel/arcs/page.tsx")
	if err != nil {
		t.Fatalf("read arcs page: %v", err)
	}
	page := string(raw)
	for _, required := range []string{
		"ProjectWorkspaceNav",
		"fetchArcs",
		"sort: 'order_index'",
		"排序",
		"升序",
		"降序",
		"加载态",
		"空状态：暂无大纲",
		"planning_run_id",
		"role=\"alert\"",
		"request_id",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected arcs page to contain %s", required)
		}
	}
}
