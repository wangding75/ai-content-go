package contract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func iteration5RepoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}

func readIteration5File(t *testing.T, parts ...string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(append([]string{iteration5RepoRoot(t)}, parts...)...))
	if err != nil {
		t.Fatalf("read %v: %v", parts, err)
	}
	return string(content)
}

// @Test
func TestTask02ReviewMigrationDeclaresTablesConstraintsAndIndexes(t *testing.T) {
	migration := readIteration5File(t, "apps", "api-server", "migrations", "00007_create_content_review_tables.sql")
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS content_review",
		"CREATE TABLE IF NOT EXISTS content_version",
		"CREATE TABLE IF NOT EXISTS review_report",
		"CHECK (status IN ('pending', 'in_review', 'approved', 'rejected', 'approved_with_edit'))",
		"CHECK (status IN ('pending', 'generating', 'succeeded', 'failed'))",
		"UNIQUE(content_item_id, version_no)",
		"idx_content_review_project_status",
		"idx_content_version_item_version",
	} {
		if !strings.Contains(migration, want) {
			t.Fatalf("migration missing %q", want)
		}
	}
	for _, forbidden := range []string{"book_", "chapter_", "novel_"} {
		if strings.Contains(migration, forbidden) {
			t.Fatalf("core migration must not introduce content-type-specific resource %q", forbidden)
		}
	}
}

// @Test
func TestTask05OpenAPIDeclaresReviewPathsSchemasAndErrorCodes(t *testing.T) {
	openapi := readIteration5File(t, "openapi", "openapi.yaml")
	for _, want := range []string{
		"/api/v1/content-reviews:",
		"/api/v1/content-items/{id}/reviews:",
		"/api/v1/content-reviews/{id}/ai-report:",
		"/api/v1/content-reviews/{id}/approve-with-edit:",
		"/api/v1/content-items/{id}/versions:",
		"CreateReviewRequest",
		"ContentReviewDetailResponse",
		"TriggerAIReportResponse",
		"RejectReviewRequest",
		"ApproveWithEditResponse",
		"WORKFLOW_RUN_FAILED",
		"LLM_PROVIDER_ERROR",
	} {
		if !strings.Contains(openapi, want) {
			t.Fatalf("openapi missing %q", want)
		}
	}
}

// @Test
func TestTask06WebAPIClientDeclaresReviewTypesFunctionsAndIdempotency(t *testing.T) {
	api := readIteration5File(t, "apps", "web-admin", "lib", "api.ts")
	for _, want := range []string{
		"export type ContentReviewResponse",
		"export type ContentReviewDetailResponse",
		"export type ReviewReportResponse",
		"fetchContentReviews(projectID",
		"createContentReview(contentItemID",
		"fetchContentReview(reviewID",
		"triggerAIReport(reviewID",
		"approveReview(reviewID",
		"rejectReview(reviewID",
		"approveWithEdit(reviewID",
		"fetchContentVersions(contentItemID",
		"Idempotency-Key",
	} {
		if !strings.Contains(api, want) {
			t.Fatalf("web API client missing %q", want)
		}
	}
}

// @Test
func TestTask07ReviewCenterPageAndNavigationExposeCreateReviewFlow(t *testing.T) {
	page := readIteration5File(t, "apps", "web-admin", "app", "projects", "[projectId]", "reviews", "page.tsx")
	nav := readIteration5File(t, "apps", "web-admin", "app", "projects", "[projectId]", "workspace-nav.tsx")
	for _, want := range []string{"审稿中心", "fetchContentReviews", "createContentReview", "pending_review", "request_id", "/content-reviews/"} {
		if !strings.Contains(page, want) {
			t.Fatalf("review center page missing %q", want)
		}
	}
	for _, want := range []string{"reviews", "审稿中心"} {
		if !strings.Contains(nav, want) {
			t.Fatalf("workspace nav missing %q", want)
		}
	}
}

// @Test
func TestTask08ReviewDetailPageExposesApproveRejectReportAndVersionContracts(t *testing.T) {
	page := readIteration5File(t, "apps", "web-admin", "app", "content-reviews", "[reviewId]", "page.tsx")
	for _, want := range []string{"fetchContentReview", "approveReview", "rejectReview", "AI 质检报告", "编辑后通过", "版本", "request_id", "打回并重生成"} {
		if !strings.Contains(page, want) {
			t.Fatalf("review detail page missing %q", want)
		}
	}
}

// @Test
func TestTask09AIReportPageExposesAsyncTriggerAndReportContracts(t *testing.T) {
	page := readIteration5File(t, "apps", "web-admin", "app", "content-reviews", "[reviewId]", "ai-report", "page.tsx")
	for _, want := range []string{"fetchAIReport", "triggerAIReport", "生成质检报告", "workflow_run_id", "request_id", "issues", "suggestions"} {
		if !strings.Contains(page, want) {
			t.Fatalf("AI report page missing %q", want)
		}
	}
}

// @Test
func TestTask10EditApprovePageExposesEditableFieldsVersionAndOperationLog(t *testing.T) {
	page := readIteration5File(t, "apps", "web-admin", "app", "content-reviews", "[reviewId]", "edit-approve", "page.tsx")
	for _, want := range []string{"approveWithEdit", "editable_fields", "content_version_id", "operation_log_id", "request_id", "保存并通过"} {
		if !strings.Contains(page, want) {
			t.Fatalf("edit approve page missing %q", want)
		}
	}
}
