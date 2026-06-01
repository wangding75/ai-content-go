package store_test

import (
	"os"
	"strings"
	"testing"
)

func TestOperationLogMigrationDeclaresRequiredIndexesAndSecretGuard(t *testing.T) {
	content, err := os.ReadFile("../../migrations/00001_create_operation_log.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(content)
	for _, required := range []string{"idx_operation_log_resource", "idx_operation_log_created_at", "idx_operation_log_request_id", "CHECK"} {
		if !strings.Contains(sql, required) {
			t.Fatalf("expected migration to include %s", required)
		}
	}
}
