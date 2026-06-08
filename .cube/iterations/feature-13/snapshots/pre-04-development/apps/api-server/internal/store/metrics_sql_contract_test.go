package store

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func readMetricsMigration(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean("../../migrations/00010_create_metrics_tables.sql"))
	if err != nil {
		t.Fatalf("read metrics migration: %v", err)
	}
	return string(content)
}

func readMetricsServiceSource(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean("../modules/metrics/service.go"))
	if err != nil {
		t.Fatalf("read metrics service: %v", err)
	}
	return string(content)
}

func readMetricsDesign(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean("../../../../.cube/iterations/feature-8/design.md"))
	if err != nil {
		t.Fatalf("read metrics design: %v", err)
	}
	return string(content)
}

func metricsGooseUpSQL(t *testing.T) string {
	t.Helper()
	migration := readMetricsMigration(t)
	upIdx := strings.Index(strings.ToLower(migration), "-- +goose up")
	downIdx := strings.Index(strings.ToLower(migration), "-- +goose down")
	if upIdx < 0 || downIdx < 0 || downIdx <= upIdx {
		t.Fatalf("metrics migration must contain ordered goose Up and Down markers")
	}
	return migration[upIdx:downIdx]
}

// @Test
func TestTask01MigrationCreatesMetricTablesAndValueConstraints(t *testing.T) {
	migration := readMetricsMigration(t)
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS metric_template",
		"CREATE TABLE IF NOT EXISTS metric_record",
		"CREATE TABLE IF NOT EXISTS metric_summary_snapshot",
		"CHECK (value_type IN ('integer', 'decimal', 'percentage', 'currency', 'duration'))",
		"CHECK (aggregation_method IN ('sum', 'avg', 'max', 'min', 'latest'))",
		"CHECK (period IN ('day', 'week', 'month'))",
		"CHECK (source_type IN ('manual', 'import', 'extension', 'external_callback'))",
	} {
		if !strings.Contains(migration, want) {
			t.Fatalf("metrics migration missing contract %q", want)
		}
	}
}

// @Test
func TestTask01MigrationDeclaresTemplateAndRecordUniqueness(t *testing.T) {
	migration := readMetricsMigration(t)
	for _, want := range []string{
		"UNIQUE(content_type, platform, metric_code)",
		"metric_template_id TEXT NOT NULL REFERENCES metric_template(id)",
		"content_type TEXT NOT NULL",
		"UNIQUE(project_id, platform, target_id, content_version_id, metric_code, metric_date, period)",
		"idx_metric_record_project_metric_date",
		"idx_metric_record_template_date",
		"idx_metric_record_target_date",
	} {
		if !strings.Contains(migration, want) {
			t.Fatalf("metrics migration missing uniqueness/index contract %q", want)
		}
	}
}

// @Test
func TestTask01SummaryAndTrendSQLFilterBeforeAggregation(t *testing.T) {
	design := readMetricsDesign(t)
	service := readMetricsServiceSource(t)
	for _, want := range []string{
		"汇总查询模板",
		"DATE_TRUNC($7",
		"AND ($6::text = '' OR r.target_id = $6)",
		"GROUP BY bucket_start, r.metric_code, t.aggregation_method",
	} {
		if !strings.Contains(design, want) {
			t.Fatalf("design SQL contract missing %q", want)
		}
	}
	for _, want := range []string{
		"const metricSummarySQL",
		"const metricTrendSQL",
		"FROM metric_record r",
		"JOIN metric_template t ON t.id = r.metric_template_id",
		"WHERE r.project_id =",
		"r.metric_date >=",
		"r.metric_date <=",
		"r.platform =",
		"r.target_id =",
		"GROUP BY r.metric_code",
		"DATE_TRUNC",
		"GROUP BY bucket_start, r.metric_code, t.aggregation_method",
		"ORDER BY bucket_start ASC",
	} {
		if !strings.Contains(service, want) {
			t.Fatalf("metrics summary/trend SQL contract missing %q", want)
		}
	}
}

// @Test
func TestTask01MissingDatesSQLUsesPublishedJobsTemplatesAndNotExists(t *testing.T) {
	design := readMetricsDesign(t)
	service := readMetricsServiceSource(t)
	for _, want := range []string{
		"缺失提醒 SQL 模板",
		"JOIN LATERAL generate_series",
		"j.published_at::date <= expected.metric_date",
		"r.metric_template_id = t.id",
		"r.period = t.period",
	} {
		if !strings.Contains(design, want) {
			t.Fatalf("design missing-dates SQL contract missing %q", want)
		}
	}
	for _, want := range []string{
		"const metricMissingDatesSQL",
		"FROM publish_job j",
		"JOIN publish_target pt ON pt.id = j.target_id",
		"JOIN content_type ct",
		"JOIN metric_template t",
		"generate_series",
		"j.status = 'published'",
		"t.required = TRUE",
		"t.enabled = TRUE",
		"NOT EXISTS",
		"r.metric_template_id = t.id",
		"r.period = t.period",
	} {
		if !strings.Contains(service, want) {
			t.Fatalf("metrics missing-dates SQL contract missing %q", want)
		}
	}
}

// @Test
func TestTask01SQLContractRequiresParameterizedWhitelistAndNoZeroFilledMissingMetrics(t *testing.T) {
	service := readMetricsServiceSource(t)
	for _, want := range []string{
		"metricSortColumns",
		"metricBucketColumns",
		"metricAggregationMethods",
		"$1",
		"$2",
		"$3",
		"$4",
		"$5",
		"$6",
	} {
		if !strings.Contains(service, want) {
			t.Fatalf("metrics SQL implementation must use parameterized whitelist contract %q", want)
		}
	}
	for _, forbidden := range []string{
		"COALESCE(value, 0)",
		"0 AS value",
		"missing_value_zero",
	} {
		if strings.Contains(service, forbidden) {
			t.Fatalf("metrics SQL must not encode missing metrics as zero via %q", forbidden)
		}
	}
}

// @Test
func TestTask01PostgresMigrationAndMetricDMLWriteThenReadContract(t *testing.T) {
	dbURL := strings.TrimSpace(os.Getenv("METRICS_TEST_DATABASE_URL"))
	if dbURL == "" {
		t.Fatalf("METRICS_TEST_DATABASE_URL is required to execute PostgreSQL DDL/DML write-then-read contract")
	}
	if _, err := exec.LookPath("psql"); err != nil {
		t.Fatalf("psql is required to execute PostgreSQL contract: %v", err)
	}
	sql := `
DROP SCHEMA IF EXISTS metrics_contract CASCADE;
CREATE SCHEMA metrics_contract;
SET search_path TO metrics_contract;
` + metricsGooseUpSQL(t) + `
INSERT INTO metric_template (
    id, content_type, platform, metric_code, metric_name, unit, value_type, aggregation_method, period, required, enabled
) VALUES (
    'metric-template-views', 'article', 'wechat', 'views', '阅读量', '次', 'integer', 'sum', 'day', TRUE, TRUE
);
INSERT INTO metric_record (
    id, project_id, content_item_id, content_version_id, publish_job_id, target_id, content_type, metric_template_id,
    platform, external_url, metric_code, metric_date, period, raw_value, normalized_value, source_type, source_ref
) VALUES (
    'metric-record-views-20260525', 'seed-project', 'content-item-1', 'content-version-approved-1', 'publish-job-1',
    'publish-target-1', 'article', 'metric-template-views', 'wechat', 'https://example.com/post/1', 'views',
    DATE '2026-05-25', 'day', '1200', 1200, 'manual', 'manual-entry'
);
DO $$
DECLARE record_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO record_count
    FROM metric_record r
    JOIN metric_template t ON t.id = r.metric_template_id
    WHERE r.project_id = 'seed-project'
      AND r.content_type = 'article'
      AND t.content_type = 'article'
      AND r.metric_code = 'views'
      AND r.normalized_value = 1200;
    IF record_count <> 1 THEN
        RAISE EXCEPTION 'metrics DML write-then-read expected 1 row, got %', record_count;
    END IF;
END $$;
DROP SCHEMA metrics_contract CASCADE;
`
	cmd := exec.Command("psql", dbURL, "-v", "ON_ERROR_STOP=1", "-c", sql)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("PostgreSQL metrics DDL/DML contract failed: %v\n%s", err, string(output))
	}
}
