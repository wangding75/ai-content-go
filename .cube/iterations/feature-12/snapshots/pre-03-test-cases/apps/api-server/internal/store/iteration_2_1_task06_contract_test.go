package store_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func readIteration21Migration(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile("../../migrations/00004_create_iteration_2_1_tables.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	return string(content)
}

func readIteration21OpenAPI(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile("../../../../openapi/openapi.yaml")
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	return string(content)
}

// @Test
func TestIteration21MigrationDeclaresScheduleExternalTablesIndexesAndConstraints(t *testing.T) {
	sql := readIteration21Migration(t)
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS workflow_schedule",
		"CREATE TABLE IF NOT EXISTS production_plan",
		"CREATE TABLE IF NOT EXISTS schedule_trigger_log",
		"CREATE TABLE IF NOT EXISTS external_workflow_provider",
		"CREATE TABLE IF NOT EXISTS external_workflow_binding",
		"CREATE TABLE IF NOT EXISTS external_workflow_call_log",
		"CHECK (daily_content_count > 0)",
		"idx_workflow_schedule_project_id",
		"idx_workflow_schedule_enabled",
		"idx_production_plan_schedule_id",
		"idx_schedule_trigger_log_schedule_id",
		"idx_schedule_trigger_log_run_id",
		"idx_external_binding_provider_id",
		"idx_external_binding_event",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration must contain %s", required)
		}
	}
}

// @Test
func TestIteration21MigrationDoesNotPersistPlainExternalAutomationSecrets(t *testing.T) {
	sql := strings.ToLower(readIteration21Migration(t))
	plainSecretColumn := regexp.MustCompile(`(?m)^\s*(token|secret|password)\s+`)
	if plainSecretColumn.FindString(sql) != "" {
		t.Fatalf("external automation migration must not declare plaintext token/secret/password columns")
	}
	if !strings.Contains(sql, "token_masked") {
		t.Fatalf("provider table must expose token_masked")
	}
}

// @Test
func TestIteration21OpenAPIDeclaresNewScheduleLLMAndExternalAutomationPaths(t *testing.T) {
	openapi := readIteration21OpenAPI(t)
	for _, required := range []string{
		"/api/v1/workflow-schedules",
		"/api/v1/workflow-schedules/{id}/enable",
		"/api/v1/workflow-schedules/{id}/disable",
		"/api/v1/workflow-schedules/{id}/test-run",
		"/api/v1/workflow-schedules/{id}/triggers",
		"/api/v1/llm-call-logs/summary",
		"/api/v1/external-automation/providers",
		"/api/v1/external-automation/bindings",
		"CreateScheduleRequest",
		"LLMCostSummaryResponse",
		"CreateExternalProviderRequest",
	} {
		if !strings.Contains(openapi, required) {
			t.Fatalf("openapi must contain %s", required)
		}
	}
}
