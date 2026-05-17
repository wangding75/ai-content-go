package cube_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestSkeletonMapCoversEveryDevelopmentTask(t *testing.T) {
	content, err := os.ReadFile("testdata/skeleton-map.yaml")
	if err != nil {
		t.Fatalf("read skeleton map: %v", err)
	}
	skeletonMap := string(content)
	for _, task := range []string{
		"Task-01：补充 API 错误码与分页契约",
		"Task-02：新增数据库迁移契约",
		"Task-03：实现系统大盘摘要 API",
		"Task-04：实现内容类型管理 API",
		"Task-05：实现内容项目管理 API",
		"Task-06：实现 Prompt 模板管理 API",
		"Task-07：实现 LLM Provider 管理 API 与 API Key 脱敏",
		"Task-08：补充 OpenAPI 契约",
		"Task-09：实现 Web Admin 首页、项目管理与项目模板交互入口",
		"Task-10：实现 Prompt 与 Provider 管理页面交互入口",
		"Task-11：生成骨架映射与阶段追踪文件",
	} {
		if !strings.Contains(skeletonMap, task) {
			t.Fatalf("expected skeleton map to cover %s", task)
		}
	}
}

// @Test
func TestIterationStateTracksTestCasesBeforeDevelopment(t *testing.T) {
	content, err := os.ReadFile("testdata/state.yaml")
	if err != nil {
		t.Fatalf("read iteration state: %v", err)
	}
	state := string(content)
	for _, required := range []string{"current_stage: 03-test-cases", "01-prd: PASS", "02-design: PASS", "03-test-cases: IN_PROGRESS", "04-development: PENDING"} {
		if !strings.Contains(state, required) {
			t.Fatalf("expected state to contain %s", required)
		}
	}
}
