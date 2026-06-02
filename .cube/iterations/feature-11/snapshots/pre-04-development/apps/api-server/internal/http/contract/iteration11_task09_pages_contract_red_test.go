package contract

import (
	"testing"
	"strings"
)

func TestTask09PlatformCollectLogsAndN8NPagesDeclareUIContracts(t *testing.T) {
	collectPage := readIteration11RepoFile(t, "apps", "web-admin", "app", "platform-collect-logs", "page.tsx")
	for _, want := range []string{
		"fetchPlatformCollectLogs",
		"fetchPlatformCollectLog",
		"confirmPlatformCollectLogMetrics",
		"PlatformCollectLogResponse",
		"PlatformCollectLogDetailResponse",
		"错误摘要",
		"request_id",
		"page-shell",
		"card",
	} {
		if !strings.Contains(collectPage, want) {
			t.Fatalf("platform collect logs page missing %q", want)
		}
	}

	n8nPage := readIteration11RepoFile(t, "apps", "web-admin", "app", "external-automation", "n8n", "page.tsx")
	for _, want := range []string{
		"rotateCallbackToken",
		"updateCallbackAuth",
		"testExternalCallback",
		"fetchCallbackLogs",
		"callback_token_once",
		"callback_log_id",
		"request_id",
		"边界",
		"role=\"alert\"",
		"role=\"status\"",
	} {
		if !strings.Contains(n8nPage, want) {
			t.Fatalf("external automation n8n page missing %q", want)
		}
	}

	nav := readIteration11RepoFile(t, "apps", "web-admin", "app", "global-nav.tsx")
	for _, want := range []string{"/platform-collect-logs", "采集日志", "/external-automation/n8n", "外部自动化 / n8n"} {
		if !strings.Contains(nav, want) {
			t.Fatalf("global nav missing %q", want)
		}
	}
}