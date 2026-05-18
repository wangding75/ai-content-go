package contract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repoRootFromContractTest(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}

func readRepoFile(t *testing.T, parts ...string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(append([]string{repoRootFromContractTest(t)}, parts...)...))
	if err != nil {
		t.Fatalf("read %v: %v", parts, err)
	}
	return string(content)
}

// @Test
func TestTask04MigrationAndOpenAPIDeclareGenerationContracts(t *testing.T) {
	migration := readRepoFile(t, "apps", "api-server", "migrations", "00006_create_content_generation_tables.sql")
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS generation_run",
		"CREATE TABLE IF NOT EXISTS content_item",
		"CREATE TABLE IF NOT EXISTS novel_chapter_extension",
		"pending_review",
		"idx_generation_run_project_idempotency_key",
		"uq_content_item_project_type_sequence_version",
	} {
		if !strings.Contains(migration, want) {
			t.Fatalf("migration missing %q", want)
		}
	}
	if strings.Contains(migration, "book_") {
		t.Fatalf("core migration must not introduce book-prefixed core resource names")
	}

	openapi := readRepoFile(t, "openapi", "openapi.yaml")
	for _, want := range []string{
		"/api/v1/projects/{projectId}/generation-runs:",
		"/api/v1/projects/{projectId}/generation-runs/batch:",
		"/api/v1/generation-runs/{id}/retry:",
		"/api/v1/projects/{projectId}/content-items:",
		"CreateGenerationRunRequest",
		"GenerationRunDetailResponse",
		"ContentItemDetailResponse",
		"LLM_PROVIDER_ERROR",
	} {
		if !strings.Contains(openapi, want) {
			t.Fatalf("openapi missing %q", want)
		}
	}
}

// @Test
func TestTask05FrontendAPIClientDeclaresGenerationFunctionsAndIdempotency(t *testing.T) {
	api := readRepoFile(t, "apps", "web-admin", "lib", "api.ts")
	for _, want := range []string{
		"export type GenerationRunResponse",
		"export type ContentItemDetailResponse",
		"createGenerationRun(projectID",
		"createBatchGenerationRuns(projectID",
		"fetchGenerationRuns(projectID",
		"fetchGenerationRun(runID",
		"retryGenerationRun(runID",
		"fetchContentItems(projectID",
		"fetchContentItem(itemID",
		"Idempotency-Key",
	} {
		if !strings.Contains(api, want) {
			t.Fatalf("web API client missing %q", want)
		}
	}
}

// @Test
func TestTask06ProductionPageAndNavigationExposeGenerationControls(t *testing.T) {
	page := readRepoFile(t, "apps", "web-admin", "app", "projects", "[projectId]", "production", "page.tsx")
	nav := readRepoFile(t, "apps", "web-admin", "app", "projects", "[projectId]", "workspace-nav.tsx")
	for _, want := range []string{"内容生产", "手动生成", "批量生成", "fetchGenerationRuns", "createGenerationRun", "createBatchGenerationRuns", "request_id"} {
		if !strings.Contains(page, want) {
			t.Fatalf("production page missing %q", want)
		}
	}
	for _, want := range []string{"production", "content-items", "内容生产", "内容单元"} {
		if !strings.Contains(nav, want) {
			t.Fatalf("workspace nav missing %q", want)
		}
	}
}

// @Test
func TestTask07GenerationDetailAndRetryPagesExposeTraceAndRetryContracts(t *testing.T) {
	detail := readRepoFile(t, "apps", "web-admin", "app", "generation-runs", "[runId]", "page.tsx")
	retry := readRepoFile(t, "apps", "web-admin", "app", "generation-runs", "[runId]", "retry", "page.tsx")
	for _, want := range []string{"fetchGenerationRun", "workflow_run_id", "agent_tasks", "llm_call_logs", "content_items", "retry"} {
		if !strings.Contains(detail, want) {
			t.Fatalf("generation detail page missing %q", want)
		}
	}
	for _, want := range []string{"retryGenerationRun", "reason", "input_override", "operation_log_id", "request_id"} {
		if !strings.Contains(retry, want) {
			t.Fatalf("generation retry page missing %q", want)
		}
	}
}

// @Test
func TestTask08ContentItemPagesExposeListAndDetailContracts(t *testing.T) {
	list := readRepoFile(t, "apps", "web-admin", "app", "projects", "[projectId]", "content-items", "page.tsx")
	detail := readRepoFile(t, "apps", "web-admin", "app", "content-items", "[itemId]", "page.tsx")
	for _, want := range []string{"fetchContentItems", "pending_review", "page_size", "request_id", "/content-items/"} {
		if !strings.Contains(list, want) {
			t.Fatalf("content item list page missing %q", want)
		}
	}
	for _, want := range []string{"fetchContentItem", "body", "metadata", "extension", "generation_run_id"} {
		if !strings.Contains(detail, want) {
			t.Fatalf("content item detail page missing %q", want)
		}
	}
}

// @Test
func TestTask09BackendIntegrationContractCoversAsyncTraceAndErrors(t *testing.T) {
	service := readRepoFile(t, "apps", "api-server", "internal", "modules", "generation", "service.go")
	handler := readRepoFile(t, "apps", "api-server", "internal", "http", "handlers", "generation.go")
	for _, want := range []string{"ValidateGenerationRun", "ReconcileWorkflowResult", "operation_log", "ErrIdempotencyConflict", "ErrLLMProviderError"} {
		if !strings.Contains(service+handler, want) {
			t.Fatalf("backend generation integration contract missing %q", want)
		}
	}
}

// @Test
func TestTask10FrontendE2ECoversNavigationSuccessFailureAndRetryRoute(t *testing.T) {
	spec := readRepoFile(t, "apps", "web-admin", "e2e", "iteration4-content-generation-loop.spec.ts")
	for _, want := range []string{"/projects/seed-project/production", "手动生成", "批量生成", "request_id", "/generation-runs/genrun-1/retry", "pending_review"} {
		if !strings.Contains(spec, want) {
			t.Fatalf("iteration4 e2e spec missing %q", want)
		}
	}
}

// @Test
func TestTask11TestMapCoversAllIteration4Tasks(t *testing.T) {
	testMap := readRepoFile(t, ".cube", "iterations", "feature-4", "test-map.yaml")
	for _, want := range []string{
		"Task-01：定义生成运行与内容单元 DTO 契约",
		"Task-02：实现生成服务接口与状态规则骨架",
		"Task-03：暴露生成运行和 ContentItem HTTP API",
		"Task-04：补充数据库迁移和 OpenAPI 契约",
		"Task-05：补充前端 API client 契约",
		"Task-06：实现内容生产页与项目导航入口",
		"Task-07：实现生成运行详情和失败重试路由",
		"Task-08：实现 ContentItem 列表和详情页面",
		"Task-09：覆盖 Iteration 4 后端契约与异步联调路径",
		"Task-10：覆盖 Iteration 4 前端 e2e 与页面联调路径",
		"Task-11：维护设计骨架映射",
	} {
		if !strings.Contains(testMap, want) {
			t.Fatalf("test-map missing %q", want)
		}
	}
}
