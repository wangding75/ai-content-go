package app_test

import (
	"os"
	"strings"
	"testing"
)

func TestWebAdminPackageUsesBuildablePinnedDependencies(t *testing.T) {
	content, err := os.ReadFile("../../../web-admin/package.json")
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "next lint") {
		t.Fatalf("expected lint script not to use removed next lint command")
	}
	if strings.Contains(text, "latest") {
		t.Fatalf("expected dependencies to be pinned")
	}
}
