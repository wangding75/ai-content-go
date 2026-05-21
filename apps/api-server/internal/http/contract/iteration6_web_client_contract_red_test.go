package contract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readIteration6WebAPI(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean("../../../../../apps/web-admin/lib/api.ts"))
	if err != nil {
		t.Fatalf("read web api client: %v", err)
	}
	return string(content)
}

// @Test
func TestTask07WebClientDeclaresKnowledgeMemoryTypes(t *testing.T) {
	api := readIteration6WebAPI(t)
	for _, want := range []string{"export type RecentWindowPolicy", "export type SnapshotSummaryResponse", "export type KnowledgeMemoryResponse", "recent_snapshot_summary", "export type MemorySnapshotResponse", "export type ContextPreviewResponse", "export type ConsistencyIssue", "export type ConsistencyReportDetailResponse"} {
		if !strings.Contains(api, want) {
			t.Fatalf("web API client missing type contract %q", want)
		}
	}
}

// @Test
func TestTask07WebClientDeclaresAllKnowledgeMemoryFunctionsAndPaths(t *testing.T) {
	api := readIteration6WebAPI(t)
	for _, want := range []string{"fetchKnowledgeMemory(projectID", "updateStaticContext(projectID", "updateStyleGuide(projectID", "correctDynamicState(projectID", "updateRecentWindowPolicy(projectID", "fetchMemorySnapshots(projectID", "previewContext(projectID", "assembleContext(projectID", "updateDynamicState(contentItemID", "createConsistencyReport(projectID", "fetchConsistencyReports(projectID", "fetchConsistencyReport(projectID", "/knowledge-memory/context-preview", "/knowledge-memory/assemble-context", "/update-dynamic-state", "/consistency-reports"} {
		if !strings.Contains(api, want) {
			t.Fatalf("web API client missing function/path %q", want)
		}
	}
}

// @Test
func TestTask07WebClientPassesQueryParametersAndIdempotencyHeaders(t *testing.T) {
	api := readIteration6WebAPI(t)
	for _, want := range []string{"new URLSearchParams", "content_item_id", "page_size", "budget", "purpose", "Idempotency-Key"} {
		if !strings.Contains(api, want) {
			t.Fatalf("web API client missing query/header contract %q", want)
		}
	}
}
