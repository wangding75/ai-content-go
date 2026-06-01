package app_test

import (
	"os"
	"strings"
	"testing"
)

func TestCIIncludesFrontendBuildJob(t *testing.T) {
	content, err := os.ReadFile("../../../../.github/workflows/ci.yml")
	if err != nil {
		t.Fatalf("read ci workflow: %v", err)
	}
	workflow := string(content)
	for _, required := range []string{"apps/web-admin", "npm", "npm run build"} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("expected CI workflow to include %s", required)
		}
	}
}
