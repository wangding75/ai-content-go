package store_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func readSocialPostMigration(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile("../../migrations/00014_create_social_post_tables.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	return string(content)
}

// @Test
func TestTask02SocialPostMigrationDeclaresGooseMarkersAndThreeTables(t *testing.T) {
	sql := readSocialPostMigration(t)

	if !strings.Contains(sql, "-- +goose Up") {
		t.Fatalf("migration must contain goose Up marker")
	}
	if !strings.Contains(sql, "-- +goose Down") {
		t.Fatalf("migration must contain goose Down marker")
	}

	for _, table := range []string{
		"CREATE TABLE IF NOT EXISTS social_post_extension",
		"CREATE TABLE IF NOT EXISTS social_post_variant",
		"CREATE TABLE IF NOT EXISTS social_post_asset",
	} {
		if !strings.Contains(sql, table) {
			t.Fatalf("migration must contain: %s", table)
		}
	}
}

// @Test
func TestTask02SocialPostMigrationDeclaresRequiredColumnsAndConstraints(t *testing.T) {
	sql := readSocialPostMigration(t)

	for _, required := range []string{
		"target_platforms",
		"default_variant_count",
		"caption_length_policy",
		"hashtag_policy",
		"cover_copy_policy",
		"tone_style",
		"forbidden_terms",
		"config_version",
		"UNIQUE(project_id)",
		"variant_index",
		"content_version_id",
		"selected_at",
		"UNIQUE(content_item_id, generation_run_id, variant_index)",
		"WHERE status = 'selected'",
		"asset_type",
		"source_variant_id",
		"asset_suggestions",
		"result JSONB",
		"operation_log_id",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration must contain: %s", required)
		}
	}
}

// @Test
func TestTask02SocialPostMigrationDownDropsExactlyThreeTables(t *testing.T) {
	sql := strings.ToLower(readSocialPostMigration(t))

	downIdx := strings.Index(sql, "-- +goose down")
	if downIdx < 0 {
		t.Fatalf("missing goose down marker")
	}
	downSQL := sql[downIdx:]

	dropRe := regexp.MustCompile(`drop\s+table\s+if\s+exists\s+(\w+)`)
	matches := dropRe.FindAllStringSubmatch(downSQL, -1)

	owned := map[string]bool{
		"social_post_asset":     true,
		"social_post_variant":   true,
		"social_post_extension": true,
	}

	found := make(map[string]bool)
	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		if owned[name] {
			found[name] = true
		}
	}

	for name := range owned {
		if !found[name] {
			t.Fatalf("down must drop %s", name)
		}
	}
	if len(found) != 3 {
		t.Fatalf("down must drop exactly 3 tables, got %d: %v", len(found), found)
	}
}

// @Test
func TestTask02SocialPostMigrationVariantStatusCheckConstraint(t *testing.T) {
	sql := readSocialPostMigration(t)

	for _, status := range []string{"generated", "selected", "rejected", "archived"} {
		if !strings.Contains(sql, "'"+status+"'") {
			t.Fatalf("variant status check must include: %s", status)
		}
	}
}

// @Test
func TestTask02SocialPostMigrationAssetTypeCheckConstraint(t *testing.T) {
	sql := readSocialPostMigration(t)

	for _, at := range []string{"tags", "cover_copy"} {
		if !strings.Contains(sql, "'"+at+"'") {
			t.Fatalf("asset asset_type check must include: %s", at)
		}
	}
}

// @Test
func TestTask02SocialPostMigrationSelectedUniqueIndexIsPartial(t *testing.T) {
	sql := readSocialPostMigration(t)

	selectedIdx := strings.Index(sql, "idx_social_post_variant_selected_unique")
	if selectedIdx < 0 {
		t.Fatalf("migration must include selected unique index")
	}
	start := selectedIdx - 50
	if start < 0 {
		start = 0
	}
	near := sql[start : selectedIdx+200]
	if !strings.Contains(near, "UNIQUE INDEX") {
		t.Fatalf("selected index must be a UNIQUE INDEX")
	}
	if !strings.Contains(near, "WHERE status = 'selected'") {
		t.Fatalf("selected index must be partial index WHERE status = 'selected'")
	}
}