package contract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readIteration6WebPage(t *testing.T, parts ...string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(append([]string{"../../../../../apps/web-admin"}, parts...)...))
	if err != nil {
		t.Fatalf("read web page %v: %v", parts, err)
	}
	return string(content)
}

// @Test
func TestTask08MemoryPageAndNavigationExposeProjectMemoryWorkspace(t *testing.T) {
	nav := readIteration6WebPage(t, "app", "projects", "[projectId]", "workspace-nav.tsx")
	page := readIteration6WebPage(t, "app", "projects", "[projectId]", "memory", "page.tsx")
	for _, want := range []string{"memory", "记忆上下文", "memory/context-preview", "上下文预览", "consistency-reports", "一致性报告"} {
		if !strings.Contains(nav, want) {
			t.Fatalf("workspace nav missing %q", want)
		}
	}
	for _, want := range []string{"记忆上下文", "StaticContext", "DynamicState", "RecentContentWindow", "StyleGuide", "最近快照摘要", "修正 StaticContext", "修正 StyleGuide", "纠偏 DynamicState", "保存最小策略", "错误码", "错误信息", "request_id"} {
		if !strings.Contains(page, want) {
			t.Fatalf("memory page missing %q", want)
		}
	}
}

// @Test
func TestTask08MemoryPageRendersSnapshotListLoadingEmptyErrorAndSuccessStates(t *testing.T) {
	page := readIteration6WebPage(t, "app", "projects", "[projectId]", "memory", "page.tsx")
	for _, want := range []string{"加载中", "暂无记忆快照", "role=\"alert\"", "role=\"status\"", "来源", "Token 预算", "预估 Token", "截断策略", "创建时间", "内容单元筛选", "排序字段", "排序方向", "上一页", "下一页", "页码"} {
		if !strings.Contains(page, want) {
			t.Fatalf("memory page missing state/snapshot rendering contract %q", want)
		}
	}
}

// @Test
func TestTask08MemoryPageDoesNotExposeDynamicStateFreeTextEditor(t *testing.T) {
	page := readIteration6WebPage(t, "app", "projects", "[projectId]", "memory", "page.tsx")
	if strings.Contains(page, "DynamicState 自由编辑") || strings.Contains(page, "编辑完整 DynamicState") || strings.Contains(page, "name=\"dynamic_state\"") || strings.Contains(page, "aria-label=\"DynamicState JSON\"") {
		t.Fatalf("memory page must not expose free text editing for DynamicState")
	}
	for _, want := range []string{"纠偏原因", "纠偏内容", "来源引用"} {
		if !strings.Contains(page, want) {
			t.Fatalf("DynamicState correction UI must expose %q contract", want)
		}
	}
}
