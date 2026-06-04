package contract

import (
	"strings"
	"testing"
)

// @Test
func TestTask11ConsistencyReportDetailPageLoadsIndependentRouteAndOverview(t *testing.T) {
	page := readIteration6WebPage(t, "app", "projects", "[projectId]", "consistency-reports", "[reportId]", "page.tsx")
	for _, want := range []string{"一致性报告详情", "来源快照", "报告状态", "问题数量", "严重度摘要"} {
		if !strings.Contains(page, want) {
			t.Fatalf("consistency report detail page missing overview contract %q", want)
		}
	}
}

// @Test
func TestTask11ConsistencyReportDetailPageRendersStructuredIssuesAndStates(t *testing.T) {
	page := readIteration6WebPage(t, "app", "projects", "[projectId]", "consistency-reports", "[reportId]", "page.tsx")
	for _, want := range []string{"加载中", "暂无问题", "role=\"alert\"", "request_id", "错误码", "错误信息", "问题编号", "严重度", "问题类型", "问题标题", "问题描述", "受影响内容", "修复建议", "失败原因"} {
		if !strings.Contains(page, want) {
			t.Fatalf("consistency report detail page missing structured issue/state contract %q", want)
		}
	}
}

// @Test
func TestTask11ConsistencyReportDetailPageShowsErrorCodeAndMessageForFailedReports(t *testing.T) {
	page := readIteration6WebPage(t, "app", "projects", "[projectId]", "consistency-reports", "[reportId]", "page.tsx")
	for _, want := range []string{"error_code", "error_message"} {
		if !strings.Contains(page, want) {
			t.Fatalf("consistency report detail page missing failed report field %q", want)
		}
	}
}
