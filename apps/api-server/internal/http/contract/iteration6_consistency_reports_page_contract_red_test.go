package contract

import (
	"strings"
	"testing"
)

// @Test
func TestTask10ConsistencyReportsPageSupportsCreateListStatusSummaryAndDetailLinks(t *testing.T) {
	page := readIteration6WebPage(t, "app", "projects", "[projectId]", "consistency-reports", "page.tsx")
	for _, want := range []string{"一致性报告", "创建一致性报告", "状态筛选", "排序字段", "排序方向", "报告状态", "问题数量", "严重度", "查看详情", "上一页", "下一页", "页码"} {
		if !strings.Contains(page, want) {
			t.Fatalf("consistency reports page missing %q", want)
		}
	}
}

// @Test
func TestTask10ConsistencyReportsPageShowsEmptyErrorAndSuccessFeedback(t *testing.T) {
	page := readIteration6WebPage(t, "app", "projects", "[projectId]", "consistency-reports", "page.tsx")
	for _, want := range []string{"暂无一致性报告", "role=\"alert\"", "role=\"status\"", "报告已创建", "错误码", "错误信息", "request_id"} {
		if !strings.Contains(page, want) {
			t.Fatalf("consistency reports page missing feedback contract %q", want)
		}
	}
}
