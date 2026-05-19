package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readKnowledgeMemoryMigration(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean("../../migrations/00008_create_knowledge_memory_tables.sql"))
	if err != nil {
		t.Fatalf("read knowledge memory migration: %v", err)
	}
	return string(content)
}

// @Test
func TestTask02MigrationCreatesAllKnowledgeMemoryTables(t *testing.T) {
	migration := readKnowledgeMemoryMigration(t)
	for _, want := range []string{"CREATE TABLE IF NOT EXISTS knowledge_memory", "CREATE TABLE IF NOT EXISTS memory_snapshot", "CREATE TABLE IF NOT EXISTS consistency_report", "CREATE TABLE IF NOT EXISTS idempotency_record"} {
		if !strings.Contains(migration, want) {
			t.Fatalf("migration missing %q", want)
		}
	}
}

// @Test
func TestTask02MigrationDeclaresJSONBSourceRefsAndStructuredIssues(t *testing.T) {
	migration := readKnowledgeMemoryMigration(t)
	for _, want := range []string{"static_context JSONB", "dynamic_state JSONB", "recent_window_policy JSONB", "style_guide JSONB", "source_refs JSONB", "assembled_context JSONB", "severity_summary JSONB", "issues JSONB"} {
		if !strings.Contains(migration, want) {
			t.Fatalf("migration missing JSONB contract %q", want)
		}
	}
	if strings.Contains(migration, "issues TEXT") {
		t.Fatalf("consistency issues must not be stored as unstructured TEXT")
	}
}

// @Test
func TestTask02MigrationDeclaresStateChecksIndexesAndPersistentIdempotency(t *testing.T) {
	migration := readKnowledgeMemoryMigration(t)
	for _, want := range []string{"CHECK (source_type IN ('assemble_context', 'dynamic_state_update', 'dynamic_state_correction'))", "CHECK (status IN ('pending', 'running', 'completed', 'failed'))", "idx_memory_snapshot_project_created", "idx_memory_snapshot_content_item", "idx_consistency_report_project_status_created", "UNIQUE(scope, endpoint, idempotency_key)", "request_hash", "response_ref_type", "response_ref_id"} {
		if !strings.Contains(migration, want) {
			t.Fatalf("migration missing constraint/index/idempotency contract %q", want)
		}
	}
}
