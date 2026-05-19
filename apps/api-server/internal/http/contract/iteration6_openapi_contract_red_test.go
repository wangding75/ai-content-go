package contract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readIteration6OpenAPI(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean("../../../../../openapi/openapi.yaml"))
	if err != nil {
		t.Fatalf("read openapi.yaml: %v", err)
	}
	return string(content)
}

// @Test
func TestTask06OpenAPIDeclaresAllKnowledgeMemoryPaths(t *testing.T) {
	openapi := readIteration6OpenAPI(t)
	for _, want := range []string{"/api/v1/projects/{projectId}/knowledge-memory:", "/api/v1/projects/{projectId}/knowledge-memory/static-context:", "/api/v1/projects/{projectId}/knowledge-memory/style-guide:", "/api/v1/projects/{projectId}/knowledge-memory/dynamic-state-correction:", "/api/v1/projects/{projectId}/knowledge-memory/recent-window-policy:", "/api/v1/projects/{projectId}/knowledge-memory/snapshots:", "/api/v1/projects/{projectId}/knowledge-memory/context-preview:", "/api/v1/projects/{projectId}/knowledge-memory/assemble-context:", "/api/v1/content-items/{id}/update-dynamic-state:", "/api/v1/projects/{projectId}/consistency-reports:", "/api/v1/projects/{projectId}/consistency-reports/{reportId}:"} {
		if !strings.Contains(openapi, want) {
			t.Fatalf("openapi missing path %q", want)
		}
	}
}

// @Test
func TestTask06OpenAPIDeclaresSchemasErrorsBearerAndIdempotency(t *testing.T) {
	openapi := readIteration6OpenAPI(t)
	for _, want := range []string{"KnowledgeMemoryResponse", "recent_snapshot_summary", "CorrectDynamicStateRequest", "DynamicStateCorrectionResponse", "ContextPreviewResponse", "AssembleContextResponse", "ConsistencyIssue", "ConsistencyReportDetailResponse", "bearerAuth", "Idempotency-Key", "IDEMPOTENCY_CONFLICT"} {
		if !strings.Contains(openapi, want) {
			t.Fatalf("openapi missing schema/security/error contract %q", want)
		}
	}
}

// @Test
func TestTask06OpenAPIUsesSynchronousSnapshotAndAcceptedReportStatusCodes(t *testing.T) {
	openapi := readIteration6OpenAPI(t)
	for _, section := range []string{"operationId: assembleContext", "operationId: updateDynamicState"} {
		idx := strings.Index(openapi, section)
		if idx < 0 {
			t.Fatalf("openapi missing %s", section)
		}
		window := openapi[idx:min(len(openapi), idx+900)]
		if !strings.Contains(window, "'200':") || strings.Contains(window, "'202':") {
			t.Fatalf("%s must be synchronous 200 and not 202", section)
		}
	}
	idx := strings.Index(openapi, "operationId: createConsistencyReport")
	if idx < 0 || !strings.Contains(openapi[idx:min(len(openapi), idx+900)], "'202':") {
		t.Fatalf("createConsistencyReport must declare 202 accepted")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
