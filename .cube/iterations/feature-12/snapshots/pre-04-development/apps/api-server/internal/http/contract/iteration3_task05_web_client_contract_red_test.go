package contract_test

import (
	"os"
	"strings"
	"testing"
)

func readTask05File(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

// @Test
func TestTask05WebAdminAPIClientDeclaresNovelPlanningTypesAndFunctions(t *testing.T) {
	api := readTask05File(t, "../../../../../apps/web-admin/lib/api.ts")
	for _, required := range []string{
		"PlanningRunResponse",
		"TopicCandidateResponse",
		"PlanningRunDetailResponse",
		"WorldviewResponse",
		"CharacterResponse",
		"ArcResponse",
		"fetchPlanningRuns",
		"createPlanningRun",
		"fetchPlanningRun",
		"confirmTopic",
		"fetchWorldview",
		"updateWorldview",
		"fetchCharacters",
		"createCharacter",
		"fetchArcs",
		"Idempotency-Key",
	} {
		if !strings.Contains(api, required) {
			t.Fatalf("expected api.ts to contain %s", required)
		}
	}
}

// @Test
func TestTask05ProjectWorkspaceUsesDynamicProjectPlanningLinks(t *testing.T) {
	home := readTask05File(t, "../../../../../apps/web-admin/app/page.tsx")
	workspaceNav := readTask05File(t, "../../../../../apps/web-admin/app/projects/[projectId]/workspace-nav.tsx")
	if !strings.Contains(home, "`/projects/${selectedProjectID}/planning`") {
		t.Fatalf("expected home page to link selected project to planning workspace")
	}
	for _, required := range []string{
		"`/projects/${projectId}/${item.href}`",
		"planning",
		"planning/topics",
		"novel/worldview",
		"novel/characters",
		"novel/arcs",
		"aria-current",
	} {
		if !strings.Contains(workspaceNav, required) {
			t.Fatalf("expected workspace nav to contain %s", required)
		}
	}
}
