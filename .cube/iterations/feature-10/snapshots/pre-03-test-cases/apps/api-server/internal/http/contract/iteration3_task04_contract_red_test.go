package contract_test

import (
	"os"
	"strings"
	"testing"
)

func readTask04File(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

// @Test
func TestTask04MigrationDefinesNovelPlanningTablesAndProjectScopedForeignKeys(t *testing.T) {
	ddl := readTask04File(t, "../../../migrations/00005_create_novel_planning_tables.sql")
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS planning_run",
		"CREATE TABLE IF NOT EXISTS planning_snapshot",
		"CREATE TABLE IF NOT EXISTS novel_topic_candidate",
		"CREATE TABLE IF NOT EXISTS novel_worldview",
		"CREATE TABLE IF NOT EXISTS novel_character",
		"CREATE TABLE IF NOT EXISTS novel_arc",
		"workflow_run_id TEXT NOT NULL UNIQUE",
		"UNIQUE (project_id, id)",
		"FOREIGN KEY (project_id, planning_run_id) REFERENCES planning_run(project_id, id)",
		"FOREIGN KEY (project_id, snapshot_id) REFERENCES planning_snapshot(project_id, id)",
	} {
		if !strings.Contains(ddl, required) {
			t.Fatalf("expected migration to contain %q", required)
		}
	}
}

// @Test
func TestTask04MigrationRejectsInvalidPlanningAndTopicStatuses(t *testing.T) {
	ddl := readTask04File(t, "../../../migrations/00005_create_novel_planning_tables.sql")
	for _, required := range []string{
		"CHECK (status IN ('pending', 'running', 'success', 'failed', 'cancelled'))",
		"CHECK (status IN ('candidate', 'confirmed', 'rejected'))",
	} {
		if !strings.Contains(ddl, required) {
			t.Fatalf("expected migration status constraint %q", required)
		}
	}
}

// @Test
func TestTask04OpenAPIDeclaresNovelPathsIdempotencyAndSchemas(t *testing.T) {
	openapi := readTask04File(t, "../../../../../openapi/openapi.yaml")
	for _, required := range []string{
		"/api/v1/projects/{projectId}/novel/planning-runs:",
		"/api/v1/projects/{projectId}/novel/planning-runs/{runId}:",
		"/api/v1/projects/{projectId}/novel/topics/{topicId}/confirm:",
		"/api/v1/projects/{projectId}/novel/worldview:",
		"/api/v1/projects/{projectId}/novel/characters:",
		"/api/v1/projects/{projectId}/novel/arcs:",
		"IdempotencyKeyHeader",
		"PlanningRunDetailResponse",
		"TopicCandidate",
		"WorldviewResponse",
		"CharactersPage",
		"ArcsPage",
		"WORKFLOW_RUN_FAILED",
	} {
		if !strings.Contains(openapi, required) {
			t.Fatalf("expected openapi contract to contain %q", required)
		}
	}
}
