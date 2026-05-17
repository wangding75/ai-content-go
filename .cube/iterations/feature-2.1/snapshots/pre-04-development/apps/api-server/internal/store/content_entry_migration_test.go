package store_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func readIteration1Migration(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile("../../migrations/00002_create_content_entry_tables.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	return string(content)
}

// @Test
func TestContentEntryMigrationDeclaresRequiredPostgreSQLTablesIndexesAndConstraints(t *testing.T) {
	sql := readIteration1Migration(t)
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS content_type",
		"CREATE TABLE IF NOT EXISTS content_project",
		"CREATE TABLE IF NOT EXISTS prompt_template",
		"CREATE TABLE IF NOT EXISTS llm_provider_config",
		"project_schema JSONB NOT NULL DEFAULT '{}'::jsonb",
		"variables JSONB NOT NULL DEFAULT '[]'::jsonb",
		"CONSTRAINT uq_llm_provider_type_base_url UNIQUE(provider_type, base_url)",
		"idx_content_type_enabled",
		"idx_content_project_status",
		"idx_content_project_content_type",
		"idx_prompt_template_agent_code",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("expected migration to contain %s", required)
		}
	}
}

// @Test
func TestContentEntryMigrationDoesNotSeedPlaintextSecretsOrDestructiveDML(t *testing.T) {
	sql := strings.ToLower(readIteration1Migration(t))
	for _, forbidden := range []string{"truncate ", "insert into llm_provider_config", "password", "token"} {
		if strings.Contains(sql, forbidden) {
			t.Fatalf("migration must not contain forbidden pattern %q", forbidden)
		}
	}
	dropRe := regexp.MustCompile(`drop\s+table\s+if\s+exists\s+(\w+)`)
	matches := dropRe.FindAllStringSubmatch(sql, -1)
	allowed := map[string]bool{"llm_provider_config": true, "prompt_template": true, "content_project": true, "content_type": true}
	for _, match := range matches {
		if !allowed[match[1]] {
			t.Fatalf("down migration must not drop table outside this migration: %s", match[1])
		}
	}
	if len(matches) != 4 {
		t.Fatalf("expected down migration to drop exactly 4 owned tables, got %d", len(matches))
	}
}
