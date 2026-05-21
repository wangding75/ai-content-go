package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readPublishMigration(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean("../../migrations/00009_create_publish_tables.sql"))
	if err != nil {
		t.Fatalf("read publish migration: %v", err)
	}
	return string(content)
}

func readPublishServiceSource(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean("../modules/publish/service.go"))
	if err != nil {
		t.Fatalf("read publish service: %v", err)
	}
	return string(content)
}

// @Test
func TestTask01MigrationCreatesPublishTablesAndStateConstraints(t *testing.T) {
	migration := readPublishMigration(t)
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS publish_target",
		"CREATE TABLE IF NOT EXISTS publish_job",
		"CREATE TABLE IF NOT EXISTS publish_log",
		"CHECK (status IN ('queued', 'copied', 'published', 'failed', 'canceled'))",
		"CHECK (event_type IN ('job_created', 'payload_copied', 'marked_published', 'marked_failed', 'requeued'))",
		"retry_count INTEGER NOT NULL DEFAULT 0 CHECK (retry_count >= 0)",
	} {
		if !strings.Contains(migration, want) {
			t.Fatalf("publish migration missing contract %q", want)
		}
	}
}

// @Test
func TestTask01MigrationDeclaresIndexesUniquenessAndNonSensitiveConfigColumns(t *testing.T) {
	migration := readPublishMigration(t)
	for _, want := range []string{
		"config JSONB NOT NULL DEFAULT '{}'::jsonb",
		"config_summary TEXT NOT NULL DEFAULT ''",
		"UNIQUE(project_id, platform, account_name, display_name)",
		"idx_publish_target_project_enabled",
		"idx_publish_job_project_status_created",
		"idx_publish_job_target_status",
		"idx_publish_job_scheduled",
		"uq_publish_job_version_target_active",
		"payload_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb",
	} {
		if !strings.Contains(migration, want) {
			t.Fatalf("publish migration missing index/column contract %q", want)
		}
	}
}

// @Test
func TestTask01QueueListSQLUsesWhitelistedJoinedProjectScopedQuery(t *testing.T) {
	service := readPublishServiceSource(t)
	for _, want := range []string{
		"FROM publish_job j",
		"JOIN publish_target t ON t.id = j.target_id",
		"WHERE j.project_id =",
		"j.target_id =",
		"j.status =",
		"j.scheduled_at >=",
		"ORDER BY",
		"LIMIT",
		"OFFSET",
		"created_at",
		"scheduled_at",
		"status",
	} {
		if !strings.Contains(service, want) {
			t.Fatalf("publish queue SQL contract missing %q", want)
		}
	}
}
