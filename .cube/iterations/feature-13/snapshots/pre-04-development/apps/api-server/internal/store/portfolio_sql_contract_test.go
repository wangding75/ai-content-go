package store

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func readPortfolioMigration(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean("../../migrations/00012_create_portfolio_tables.sql"))
	if err != nil {
		t.Fatalf("read portfolio migration: %v", err)
	}
	return string(content)
}

func portfolioGooseUpSQL(t *testing.T) string {
	t.Helper()
	migration := readPortfolioMigration(t)
	upIdx := strings.Index(strings.ToLower(migration), "-- +goose up")
	downIdx := strings.Index(strings.ToLower(migration), "-- +goose down")
	if upIdx < 0 || downIdx < 0 || downIdx <= upIdx {
		t.Fatalf("portfolio migration must contain ordered goose Up and Down markers")
	}
	return migration[upIdx:downIdx]
}

// @Test
func TestTask02MigrationCreatesPortfolioTablesAndValueConstraints(t *testing.T) {
	migration := readPortfolioMigration(t)
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS project_portfolio",
		"CREATE TABLE IF NOT EXISTS portfolio_project",
		"CREATE TABLE IF NOT EXISTS portfolio_status_snapshot",
		"scope_type text NOT NULL",
		"owner_id text NOT NULL DEFAULT ''",
		"health_policy jsonb NOT NULL DEFAULT '{}'",
		"role text NOT NULL DEFAULT 'member'",
		"priority integer NOT NULL CHECK (priority >= 1)",
		"weight numeric(10, 2) NOT NULL DEFAULT 1 CHECK (weight >= 0)",
		"date_range_start date NOT NULL",
		"date_range_end date NOT NULL",
		"health_score numeric(5, 2) NOT NULL",
		"source_refs jsonb NOT NULL DEFAULT '[]'",
		"calculated_at timestamptz",
		"calculation_status text NOT NULL",
	} {
		if !strings.Contains(migration, want) {
			t.Fatalf("portfolio migration missing contract %q", want)
		}
	}
}

// @Test
func TestTask02MigrationDeclaresForeignKeysAndIndexes(t *testing.T) {
	migration := readPortfolioMigration(t)
	for _, want := range []string{
		"REFERENCES project_portfolio(id)",
		"idx_project_portfolio_status_created_at",
		"idx_project_portfolio_scope_type",
		"idx_project_portfolio_owner_id",
		"idx_project_portfolio_name",
		"idx_portfolio_project_project_id",
		"idx_portfolio_project_priority",
		"idx_portfolio_project_role",
		"idx_portfolio_status_snapshot_portfolio_created",
		"idx_portfolio_status_snapshot_status",
		"idx_portfolio_status_snapshot_date_range",
	} {
		if !strings.Contains(migration, want) {
			t.Fatalf("portfolio migration missing FK/index contract %q", want)
		}
	}
}

// @Test
func TestTask02MigrationDownDropsTablesInCorrectOrder(t *testing.T) {
	migration := readPortfolioMigration(t)
	downIdx := strings.Index(strings.ToLower(migration), "-- +goose down")
	if downIdx < 0 {
		t.Fatalf("portfolio migration must contain goose Down marker")
	}
	down := migration[downIdx:]
	for _, want := range []string{
		"DROP TABLE IF EXISTS portfolio_status_snapshot",
		"DROP TABLE IF EXISTS portfolio_project",
		"DROP TABLE IF EXISTS project_portfolio",
	} {
		if !strings.Contains(down, want) {
			t.Fatalf("portfolio migration down section missing %q", want)
		}
	}
}

// @Test
func TestTask02PostgresMigrationAndPortfolioDMLWriteThenReadContract(t *testing.T) {
	dbURL := strings.TrimSpace(os.Getenv("METRICS_TEST_DATABASE_URL"))
	if dbURL == "" {
		t.Fatalf("METRICS_TEST_DATABASE_URL is required to execute PostgreSQL DDL/DML write-then-read contract")
	}
	if _, err := exec.LookPath("psql"); err != nil {
		t.Fatalf("psql is required to execute PostgreSQL contract: %v", err)
	}
	sql := `
DROP SCHEMA IF EXISTS portfolio_contract CASCADE;
CREATE SCHEMA portfolio_contract;
SET search_path TO portfolio_contract;
` + portfolioGooseUpSQL(t) + `
INSERT INTO project_portfolio (
    id, name, description, scope_type, owner_id, health_policy, status
) VALUES (
    'pf-001', '增长组合', '跨项目增长', 'manual', 'growth-team', '{"warning_threshold": 60}', 'active'
);
INSERT INTO portfolio_project (
    portfolio_id, project_id, role, priority, weight, note, added_by
) VALUES (
    'pf-001', 'proj-001', 'member', 1, 1.50, '核心项目', 'admin'
);
INSERT INTO portfolio_status_snapshot (
    id, portfolio_id, date_range_start, date_range_end,
    health_score, health_status, total_projects, active_projects, warning_projects,
    estimated_monthly_cost, currency, risk_summary, cost_summary, strategy_summary,
    source_refs, calculation_status, calculated_at
) VALUES (
    'pss-001', 'pf-001', DATE '2026-05-01', DATE '2026-05-31',
    85.50, 'healthy', 1, 1, 0,
    386.00, 'CNY', '{}', '{}', '{}',
    '[{"source":"llm_call_logs","source_id":"project:proj-001","updated_at":"2026-05-29T10:00:00Z"}]',
    'completed', NOW()
);
DO $$
DECLARE pf_count INTEGER; pp_count INTEGER; pss_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO pf_count FROM project_portfolio WHERE id = 'pf-001' AND scope_type = 'manual' AND owner_id = 'growth-team';
    IF pf_count <> 1 THEN RAISE EXCEPTION 'portfolio DML expected 1 row, got %', pf_count; END IF;

    SELECT COUNT(*) INTO pp_count FROM portfolio_project WHERE portfolio_id = 'pf-001' AND project_id = 'proj-001' AND role = 'member';
    IF pp_count <> 1 THEN RAISE EXCEPTION 'portfolio project DML expected 1 row, got %', pp_count; END IF;

    SELECT COUNT(*) INTO pss_count FROM portfolio_status_snapshot WHERE id = 'pss-001' AND health_status = 'healthy' AND calculation_status = 'completed';
    IF pss_count <> 1 THEN RAISE EXCEPTION 'portfolio snapshot DML expected 1 row, got %', pss_count; END IF;
END $$;
DROP SCHEMA portfolio_contract CASCADE;
`
	cmd := exec.Command("psql", dbURL, "-v", "ON_ERROR_STOP=1", "-c", sql)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("PostgreSQL portfolio DDL/DML contract failed: %v\n%s", err, string(output))
	}
}
