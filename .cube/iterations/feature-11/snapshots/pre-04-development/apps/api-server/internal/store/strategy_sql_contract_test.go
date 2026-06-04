package store

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func readStrategyMigration(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean("../../migrations/00011_create_strategy_tables.sql"))
	if err != nil {
		t.Fatalf("read strategy migration: %v", err)
	}
	return string(content)
}

func strategyGooseUpSQL(t *testing.T) string {
	t.Helper()
	migration := readStrategyMigration(t)
	upIdx := strings.Index(strings.ToLower(migration), "-- +goose up")
	downIdx := strings.Index(strings.ToLower(migration), "-- +goose down")
	if upIdx < 0 || downIdx < 0 || downIdx <= upIdx {
		t.Fatalf("strategy migration must contain ordered goose Up and Down markers")
	}
	return migration[upIdx:downIdx]
}

// @Test
func TestTask01MigrationCreatesStrategyTablesAndValueConstraints(t *testing.T) {
	migration := readStrategyMigration(t)
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS strategy_suggestion_run",
		"CREATE TABLE IF NOT EXISTS strategy_suggestion",
		"CREATE TABLE IF NOT EXISTS strategy_execution_log",
		"CHECK (status IN ('generating', 'completed', 'failed'))",
		"CHECK (suggestion_type IN ('keep', 'optimize', 'suspend', 'promote', 'cost_control'))",
		"CHECK (risk_level IN ('low', 'medium', 'high'))",
		"CHECK (confidence IN ('low', 'medium', 'high'))",
		"CHECK (status IN ('pending', 'confirmed', 'ignored', 'executed', 'execution_failed'))",
		"CHECK (result IN ('success', 'failed'))",
		"CHECK (suggestion_count >= 0)",
	} {
		if !strings.Contains(migration, want) {
			t.Fatalf("strategy migration missing contract %q", want)
		}
	}
}

// @Test
func TestTask01MigrationDeclaresForeignKeysAndIndexes(t *testing.T) {
	migration := readStrategyMigration(t)
	for _, want := range []string{
		"REFERENCES strategy_suggestion(id)",
		"idx_strategy_suggestion_run_project",
		"idx_strategy_suggestion_project",
		"idx_strategy_suggestion_run",
		"idx_strategy_suggestion_type_status",
		"idx_strategy_execution_log_suggestion",
	} {
		if !strings.Contains(migration, want) {
			t.Fatalf("strategy migration missing FK/index contract %q", want)
		}
	}
}

// @Test
func TestTask01PostgresMigrationAndStrategyDMLWriteThenReadContract(t *testing.T) {
	dbURL := strings.TrimSpace(os.Getenv("METRICS_TEST_DATABASE_URL"))
	if dbURL == "" {
		t.Fatalf("METRICS_TEST_DATABASE_URL is required to execute PostgreSQL DDL/DML write-then-read contract")
	}
	if _, err := exec.LookPath("psql"); err != nil {
		t.Fatalf("psql is required to execute PostgreSQL contract: %v", err)
	}
	sql := `
DROP SCHEMA IF EXISTS strategy_contract CASCADE;
CREATE SCHEMA strategy_contract;
SET search_path TO strategy_contract;
` + strategyGooseUpSQL(t) + `
INSERT INTO strategy_suggestion_run (
    id, project_id, date_from, date_to, status, suggestion_count
) VALUES (
    'run-001', 'seed-project', DATE '2026-05-01', DATE '2026-05-25', 'completed', 1
);
INSERT INTO strategy_suggestion (
    id, project_id, suggestion_run_id, suggestion_type, title, trigger_reason,
    risk_level, confidence, status, date_from, date_to
) VALUES (
    'sug-001', 'seed-project', 'run-001', 'optimize', '优化发布频率',
    '近7日发布量下降30%', 'medium', 'high', 'pending', DATE '2026-05-01', DATE '2026-05-25'
);
INSERT INTO strategy_execution_log (
    id, suggestion_id, action_type, previous_status, current_status, result
) VALUES (
    'log-001', 'sug-001', 'confirm', 'pending', 'confirmed', 'success'
);
DO $$
DECLARE run_count INTEGER; sug_count INTEGER; log_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO run_count FROM strategy_suggestion_run WHERE id = 'run-001' AND project_id = 'seed-project';
    IF run_count <> 1 THEN RAISE EXCEPTION 'strategy run DML expected 1 row, got %', run_count; END IF;

    SELECT COUNT(*) INTO sug_count FROM strategy_suggestion WHERE id = 'sug-001' AND suggestion_type = 'optimize' AND status = 'pending';
    IF sug_count <> 1 THEN RAISE EXCEPTION 'strategy suggestion DML expected 1 row, got %', sug_count; END IF;

    SELECT COUNT(*) INTO log_count FROM strategy_execution_log WHERE suggestion_id = 'sug-001' AND result = 'success';
    IF log_count <> 1 THEN RAISE EXCEPTION 'strategy execution log DML expected 1 row, got %', log_count; END IF;
END $$;
DROP SCHEMA strategy_contract CASCADE;
`
	cmd := exec.Command("psql", dbURL, "-v", "ON_ERROR_STOP=1", "-c", sql)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("PostgreSQL strategy DDL/DML contract failed: %v\n%s", err, string(output))
	}
}
