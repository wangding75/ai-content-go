package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestIteration21HistoricalPagesUseStyledAdminRenderingInsteadOfSkeletonOrBareHTML(t *testing.T) {
	pages := []string{
		"../../../../../apps/web-admin/app/page.tsx",
		"../../../../../apps/web-admin/app/swagger-openapi/page.tsx",
		"../../../../../apps/web-admin/app/workflow/templates/page.tsx",
		"../../../../../apps/web-admin/app/workflow/runs/page.tsx",
		"../../../../../apps/web-admin/app/agent/tasks/page.tsx",
		"../../../../../apps/web-admin/app/llm/logs/page.tsx",
	}
	for _, path := range pages {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		page := string(raw)
		for _, forbidden := range []string{"裸 HTML", "skeleton", "骨架"} {
			if strings.Contains(page, forbidden) {
				t.Fatalf("%s must not contain placeholder %q", path, forbidden)
			}
		}
		if !strings.Contains(page, "request_id") && !strings.Contains(page, "requestId") {
			t.Fatalf("%s must render request_id/requestId in error states", path)
		}
	}
}

// @Test
func TestIteration21HomePageUsesExplicitEmptyFixtureInsteadOfDefaultEmptyProjectList(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/page.tsx")
	if err != nil {
		t.Fatalf("read home page: %v", err)
	}
	page := string(raw)
	if !strings.Contains(page, "fixture=empty") {
		t.Fatalf("home page must expose explicit fixture=empty handling")
	}
	if strings.Contains(page, "fetchProjects('&status=__empty_fixture__')") {
		t.Fatalf("default project list must not always request the empty fixture")
	}
	if strings.Contains(page, "openProject('seed-project')") {
		t.Fatalf("project flow must not use seed-project fallback")
	}
}
