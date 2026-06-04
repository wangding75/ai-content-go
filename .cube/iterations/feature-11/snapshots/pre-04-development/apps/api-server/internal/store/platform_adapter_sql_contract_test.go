package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readPlatformAdapterMigration(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean("../../migrations/00013_create_platform_adapter_extension_tables.sql"))
	if err != nil {
		t.Fatalf("read platform adapter migration: %v", err)
	}
	return string(content)
}

func TestTask01MigrationCreatesPlatformAdapterExtensionTablesAndCoreConstraints(t *testing.T) {
	migration := readPlatformAdapterMigration(t)
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS platform_adapter_config",
		"CREATE TABLE IF NOT EXISTS platform_adapter_revision",
		"CREATE TABLE IF NOT EXISTS plugin_client",
		"CREATE TABLE IF NOT EXISTS plugin_access_token",
		"CREATE TABLE IF NOT EXISTS platform_collect_log",
		"CREATE TABLE IF NOT EXISTS external_callback_log",
		"ALTER TABLE publish_target ADD COLUMN IF NOT EXISTS target_type TEXT NOT NULL DEFAULT 'default'",
		"ALTER TABLE publish_job ADD COLUMN IF NOT EXISTS plugin_lock_id TEXT",
		"ALTER TABLE publish_job ADD COLUMN IF NOT EXISTS plugin_client_id TEXT",
		"ALTER TABLE publish_job ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ",
		"ALTER TABLE publish_job ADD COLUMN IF NOT EXISTS adapter_config_id TEXT",
		"ALTER TABLE publish_job ADD COLUMN IF NOT EXISTS adapter_version INTEGER",
		"CHECK (publish_mode IN ('manual_plugin', 'external_callback', 'manual_only'))",
		"CHECK (client_type IN ('chrome_extension'))",
		"CHECK (status IN ('enabled', 'disabled'))",
		"CHECK (source_type IN ('extension','external_callback'))",
		"CHECK (status IN ('received','parse_failed','ready','confirmed','rejected'))",
		"CHECK (callback_auth_type IN ('token','signature'))",
		"UNIQUE(binding_id, idempotency_key)",
		"api_key_hash TEXT NOT NULL UNIQUE",
		"token_hash TEXT NOT NULL UNIQUE",
	} {
		if !strings.Contains(migration, want) {
			t.Fatalf("platform adapter migration missing contract %q", want)
		}
	}
}

func TestTask01MigrationDeclaresIndexesAndAtomicLockContract(t *testing.T) {
	migration := readPlatformAdapterMigration(t)
	for _, want := range []string{
		"idx_platform_adapter_enabled",
		"idx_platform_adapter_mode",
		"idx_platform_adapter_revision_adapter_version",
		"idx_plugin_client_status",
		"idx_plugin_client_last_active",
		"idx_plugin_access_token_client_expires",
		"idx_plugin_access_token_hash",
		"idx_publish_target_platform_type",
		"idx_publish_job_plugin_lock",
		"idx_publish_job_adapter_status",
		"idx_publish_job_locked_until",
		"idx_platform_collect_log_project_status",
		"idx_platform_collect_log_platform_collected",
		"idx_platform_collect_log_publish_job",
		"idx_external_callback_binding_created",
		"idx_external_callback_event_accepted",
		"UPDATE publish_job SET plugin_lock_id = $2, plugin_client_id = $3, locked_until = $4 WHERE id = $1 AND status IN ('queued','copied','failed') AND (locked_until IS NULL OR locked_until < $5) RETURNING id;",
	} {
		if !strings.Contains(migration, want) {
			t.Fatalf("platform adapter migration missing SQL/index contract %q", want)
		}
	}
	for _, forbid := range []string{
		"api_key TEXT NOT NULL",
		"api_key_plain",
		"password TEXT NOT NULL",
		"secret_plain",
	} {
		if strings.Contains(migration, forbid) {
			t.Fatalf("platform adapter migration must not contain forbidden plaintext pattern %q", forbid)
		}
	}
}
