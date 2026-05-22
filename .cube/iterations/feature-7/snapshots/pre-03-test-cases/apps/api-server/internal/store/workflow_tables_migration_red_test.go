package store_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func readIteration2Migration(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile("../../migrations/00003_create_workflow_tables.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	return string(content)
}

// @Test
func TestWorkflowTablesMigrationDeclaresGooseMarkersAndAllEightTables(t *testing.T) {
	sql := readIteration2Migration(t)

	if !strings.Contains(sql, "-- +goose Up") {
		t.Fatalf("migration must contain goose Up marker")
	}
	if !strings.Contains(sql, "-- +goose Down") {
		t.Fatalf("migration must contain goose Down marker")
	}

	for _, table := range []string{
		"CREATE TABLE IF NOT EXISTS workflow_template",
		"CREATE TABLE IF NOT EXISTS workflow_template_version",
		"CREATE TABLE IF NOT EXISTS workflow_step_template",
		"CREATE TABLE IF NOT EXISTS workflow_run",
		"CREATE TABLE IF NOT EXISTS workflow_step_run",
		"CREATE TABLE IF NOT EXISTS agent_task",
		"CREATE TABLE IF NOT EXISTS llm_call_log",
		"CREATE TABLE IF NOT EXISTS workflow_schedule",
	} {
		if !strings.Contains(sql, table) {
			t.Fatalf("migration must contain: %s", table)
		}
	}
}

// @Test
func TestWorkflowTablesMigrationDeclaresRequiredColumnsAndConstraints(t *testing.T) {
	sql := readIteration2Migration(t)

	for _, required := range []string{
		"idempotency_key",
		"parent_run_id",
		"REFERENCES workflow_template_version",
		"REFERENCES workflow_run",
		"agent_task_id",
		"input_tokens",
		"output_tokens",
		"latency_ms",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration must contain: %s", required)
		}
	}
}

// @Test
func TestWorkflowTablesMigrationDownDropsExactlyEightOwnedTables(t *testing.T) {
	sql := strings.ToLower(readIteration2Migration(t))

	downIdx := strings.Index(sql, "-- +goose down")
	if downIdx < 0 {
		t.Fatalf("missing goose down marker")
	}
	downSQL := sql[downIdx:]

	dropRe := regexp.MustCompile(`drop\s+table\s+if\s+exists\s+(\w+)`)
	matches := dropRe.FindAllStringSubmatch(downSQL, -1)

	owned := map[string]bool{
		"workflow_schedule": true, "llm_call_log": true, "agent_task": true,
		"workflow_step_run": true, "workflow_run": true, "workflow_step_template": true,
		"workflow_template_version": true, "workflow_template": true,
	}
	for _, m := range matches {
		if !owned[m[1]] {
			t.Fatalf("down migration drops table outside this migration: %s", m[1])
		}
	}
	if len(matches) != 8 {
		t.Fatalf("expected down migration to drop exactly 8 tables, got %d", len(matches))
	}
}
