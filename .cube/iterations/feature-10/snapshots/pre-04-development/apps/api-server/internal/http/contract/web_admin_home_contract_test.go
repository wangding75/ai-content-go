package contract_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readWebAdminFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func webAdminRoot(t *testing.T) string {
	t.Helper()
	return filepath.Clean("../../../../../apps/web-admin")
}

// @Test
func TestHomePageImplementsPrototypeNavigationAndStatefulDashboardProjectTemplateViews(t *testing.T) {
	page := readWebAdminFile(t, filepath.Join(webAdminRoot(t), "app/page.tsx"))

	for _, required := range []string{
		"首页 / 系统大盘",
		"项目管理",
		"项目工作区",
		"项目概览",
		"项目模板管理",
		"运行中项目",
		"待审稿",
		"待发布",
		"今日模型成本",
		"进入项目",
		"暂停项目",
		"查看 Schema",
		"data-testid=\"dashboard-loading\"",
		"data-testid=\"dashboard-project-count\"",
		"data-testid=\"projects-empty\"",
		"data-testid=\"project-schema\"",
		"role=\"alert\"",
		"role=\"status\"",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("home page must implement prototype UI/state contract %q", required)
		}
	}
}

// @Test
func TestHomePageUsesRealAPIClientForRoundtripInteractionsInsteadOfStaticPlaceholders(t *testing.T) {
	page := readWebAdminFile(t, filepath.Join(webAdminRoot(t), "app/page.tsx"))

	for _, required := range []string{
		"await fetchDashboardSummary(",
		"await fetchProjects(",
		"await createProject(",
		"await fetchProjectOverview(",
		"await pauseProject(",
		"await fetchContentTypes(",
		"await createContentType(",
		"await fetchProjectSchema(",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("home page must execute %s during user-visible UI flows", required)
		}
	}
	for _, forbidden := range []string{
		"将在开发阶段连接 API",
		"摘要接口：fetchDashboardSummary",
	} {
		if strings.Contains(page, forbidden) {
			t.Fatalf("home page must not satisfy UI contract with placeholder text %q", forbidden)
		}
	}
}

// @Test
func TestWebAdminHomeE2ESpecRequiresRealBrowserBackendRoundtripAndFailureStates(t *testing.T) {
	spec := readWebAdminFile(t, filepath.Join(webAdminRoot(t), "e2e/iteration1-ui.spec.ts"))

	for _, required := range []string{
		"page.goto(webBaseURL)",
		"page.waitForResponse",
		"/api/v1/dashboard/summary",
		"/api/v1/projects",
		"/api/v1/projects/seed-project/overview",
		"/api/v1/projects/seed-project/pause",
		"/api/v1/content-types",
		"/api/v1/content-types/seed-content-type/project-schema",
		"dashboard-loading",
		"projects-empty",
		"getByRole('alert')",
		"getByRole('status')",
		"INTERNAL_ERROR",
		"req-dashboard-error",
		"创建成功",
		"已暂停",
	} {
		if !strings.Contains(spec, required) {
			t.Fatalf("home e2e spec must cover frontend-ui requirement %q", required)
		}
	}
	if strings.Contains(spec, "route.fulfill") && !strings.Contains(spec, "page.waitForResponse") {
		t.Fatalf("frontend UI spec must not replace all API evidence with mocked responses")
	}
}
