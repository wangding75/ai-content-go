package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestIteration21HomePageSupportsExplicitEmptyProjectFixture(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/page.tsx")
	if err != nil {
		t.Fatalf("read home page: %v", err)
	}
	page := string(raw)
	for _, required := range []string{
		"searchParams.get('fixture')",
		"fixture === 'empty'",
		"fetchProjects(fixture === 'empty' ? '&status=__empty_fixture__' : '')",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("home page must support explicit empty fixture guard %s", required)
		}
	}
}

// @Test
func TestIteration21HomePageUsesScopedProjectSelectionWithoutSeedFallback(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/page.tsx")
	if err != nil {
		t.Fatalf("read home page: %v", err)
	}
	page := string(raw)
	for _, forbidden := range []string{
		"openProject('seed-project')",
		"openProject(\"seed-project\")",
		"useState('seed-project')",
		"useState(\"seed-project\")",
	} {
		if strings.Contains(page, forbidden) {
			t.Fatalf("home page must not use seed project fallback %s", forbidden)
		}
	}
	for _, required := range []string{
		"openProject(project.id)",
		"projects[0]?.id",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("home page must use scoped project selection %s", required)
		}
	}
}
