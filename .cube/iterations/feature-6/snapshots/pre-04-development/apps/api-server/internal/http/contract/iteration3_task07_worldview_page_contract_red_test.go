package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestTask07WorldviewPageSupportsViewEditSaveAndFeedbackStates(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/projects/[projectId]/novel/worldview/page.tsx")
	if err != nil {
		t.Fatalf("read worldview page: %v", err)
	}
	page := string(raw)
	for _, required := range []string{
		"ProjectWorkspaceNav",
		"fetchWorldview",
		"updateWorldview",
		"世界观 JSON",
		"Forbidden Rules",
		"保存世界观",
		"保存成功",
		"世界观 JSON 格式错误",
		"role=\"alert\"",
		"role=\"status\"",
		"request_id",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected worldview page to contain %s", required)
		}
	}
}
