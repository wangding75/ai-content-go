package contract

import (
	"strings"
	"testing"
)

// @Test
func TestTask09ContextPreviewPageSupportsPreviewWithoutPersistenceAndSnapshotGeneration(t *testing.T) {
	page := readIteration6WebPage(t, "app", "projects", "[projectId]", "memory", "context-preview", "page.tsx")
	for _, want := range []string{"预览上下文", "生成上下文快照", "用途", "Token 预算", "内容单元 ID", "未落库", "已生成上下文快照", "快照 ID", "role=\"status\""} {
		if !strings.Contains(page, want) {
			t.Fatalf("context preview page missing %q", want)
		}
	}
}

// @Test
func TestTask09ContextPreviewPageDisplaysBudgetSourcesTruncationPreviewAndErrors(t *testing.T) {
	page := readIteration6WebPage(t, "app", "projects", "[projectId]", "memory", "context-preview", "page.tsx")
	for _, want := range []string{"上下文来源", "Token 预算", "预估 Token", "截断策略", "预览内容", "role=\"alert\"", "错误码", "错误信息", "request_id"} {
		if !strings.Contains(page, want) {
			t.Fatalf("context preview page missing output/error contract %q", want)
		}
	}
}
