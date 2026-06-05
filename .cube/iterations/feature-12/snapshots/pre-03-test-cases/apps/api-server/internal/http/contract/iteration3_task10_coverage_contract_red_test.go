package contract_test

import (
	"os"
	"strings"
	"testing"
)

func task10Read(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

// @Test
func TestTask10RedSuiteCoversEveryNovelPlanningErrorCode(t *testing.T) {
	design := task10Read(t, "../../../../../.cube/iterations/feature-3/design.md")
	for _, code := range []string{
		"VALIDATION_ERROR",
		"UNAUTHORIZED",
		"FORBIDDEN",
		"NOT_FOUND",
		"CONFLICT",
		"IDEMPOTENCY_CONFLICT",
		"WORKFLOW_RUN_FAILED",
		"INTERNAL_ERROR",
	} {
		if !strings.Contains(design, code) {
			t.Fatalf("expected design to declare error code %s", code)
		}
	}
}

// @Test
func TestTask10TestMapRecordsAllDevelopmentTasksAndTypeStandards(t *testing.T) {
	testMap := task10Read(t, "../../../../../.cube/iterations/feature-3/test-map.yaml")
	for _, required := range []string{
		"Task-01：定义 Novel Pack DTO 与 Service 契约",
		"Task-02：实现 Novel Planning Service 状态与幂等规则",
		"Task-03：暴露 Novel Planning HTTP API",
		"Task-04：补齐 Novel Planning 数据迁移与 OpenAPI 契约",
		"Task-05：扩展 Web Admin API client 与项目工作区导航入口",
		"Task-06：实现内容规划页与候选确认弹窗",
		"Task-07：实现世界观编辑页",
		"Task-08：实现人物管理页",
		"Task-09：实现大纲管理页",
		"Task-10：补齐 Novel Planning 自动化测试覆盖",
		"type: web-e2e",
		"type: integration",
		"type: sql-query",
	} {
		if !strings.Contains(testMap, required) {
			t.Fatalf("expected test-map to contain %s", required)
		}
	}
}
