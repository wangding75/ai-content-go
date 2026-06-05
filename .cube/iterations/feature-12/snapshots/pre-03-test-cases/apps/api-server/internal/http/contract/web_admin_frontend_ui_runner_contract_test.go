package contract_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// @Test
func TestWebAdminPackageDeclaresExecutableFrontendUITestCommand(t *testing.T) {
	content, err := os.ReadFile(filepath.Join(webAdminRoot(t), "package.json"))
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}
	pkg := string(content)
	for _, required := range []string{
		"\"test:ui\"",
		"playwright test",
		"@playwright/test",
	} {
		if !strings.Contains(pkg, required) {
			t.Fatalf("web-admin package must declare executable frontend UI test dependency/script %q", required)
		}
	}
}

// @Test
func TestPlaywrightConfigStartsWebAndAPIServersForFullRoundtrip(t *testing.T) {
	config := readWebAdminFile(t, filepath.Join(webAdminRoot(t), "playwright.config.ts"))
	for _, required := range []string{
		"webServer",
		"npm run dev",
		"go run ./apps/api-server/cmd/api",
		"WEB_BASE_URL",
		"API_BASE_URL",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:18080",
		"Authorization",
		"Bearer dev",
	} {
		if !strings.Contains(config, required) {
			t.Fatalf("playwright config must enable real frontend/backend roundtrip evidence %q", required)
		}
	}
}
