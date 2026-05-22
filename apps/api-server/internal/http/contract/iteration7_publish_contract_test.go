package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	serverhttp "github.com/wangding75/ai-content-go/apps/api-server/internal/http"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/publish"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/system"
)

type iteration7SystemService struct{}

func (iteration7SystemService) Health(ctx context.Context) (system.HealthResponse, error) {
	return system.HealthResponse{}, nil
}

func (iteration7SystemService) Info(ctx context.Context) (system.InfoResponse, error) {
	return system.InfoResponse{}, nil
}

func (iteration7SystemService) ConfigCheck(ctx context.Context) (system.ConfigCheckResponse, error) {
	return system.ConfigCheckResponse{}, nil
}

func (iteration7SystemService) DBCheck(ctx context.Context) (system.DBCheckResponse, error) {
	return system.DBCheckResponse{}, nil
}

func (iteration7SystemService) MigrationStatus(ctx context.Context) (system.MigrationStatusResponse, error) {
	return system.MigrationStatusResponse{}, nil
}

func iteration7Router() http.Handler {
	return serverhttp.NewRouter(iteration7SystemService{}, nil)
}

func iteration7Request(method, path string, body []byte, idempotencyKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("X-Request-Id", "req-iteration-7")
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	iteration7Router().ServeHTTP(rr, req)
	return rr
}

func decodeIteration7Envelope(t *testing.T, body []byte) struct {
	Success   bool
	Data      map[string]any
	Error     *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	RequestID string `json:"request_id"`
} {
	t.Helper()
	var env struct {
		Success   bool
		Data      map[string]any
		Error     *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, string(body))
	}
	return env
}

func readIteration7RepoFile(t *testing.T, parts ...string) string {
	t.Helper()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(append([]string{root}, parts...)...))
	if err != nil {
		t.Fatalf("read %v: %v", parts, err)
	}
	return string(content)
}

func hasIteration7JSONField(t reflect.Type, jsonName string) bool {
	for i := 0; i < t.NumField(); i++ {
		name := strings.Split(t.Field(i).Tag.Get("json"), ",")[0]
		if name == jsonName {
			return true
		}
	}
	return false
}

// @Test
func TestTask02PublishDTOAndErrorConstantsDeclareStableContracts(t *testing.T) {
	for _, status := range []string{publish.JobStatusQueued, publish.JobStatusCopied, publish.JobStatusPublished, publish.JobStatusFailed, publish.JobStatusCanceled} {
		if status == "" {
			t.Fatalf("publish job status constants must be non-empty")
		}
	}
	for _, event := range []string{publish.EventJobCreated, publish.EventPayloadCopied, publish.EventMarkedPublished, publish.EventMarkedFailed, publish.EventRequeued} {
		if event == "" {
			t.Fatalf("publish log event constants must be non-empty")
		}
	}
	for _, errValue := range []error{publish.ErrValidation, publish.ErrNotFound, publish.ErrForbidden, publish.ErrConflict, publish.ErrIdempotencyConflict, publish.ErrInternal} {
		if errValue == nil {
			t.Fatalf("publish domain errors must be declared")
		}
	}
	for _, field := range []string{"id", "project_id", "platform", "account_name", "display_name", "config_summary", "enabled", "updated_at"} {
		if !hasIteration7JSONField(reflect.TypeOf(publish.PublishTargetResponse{}), field) {
			t.Fatalf("PublishTargetResponse missing json field %q", field)
		}
	}
	for _, field := range []string{"content_version_id", "payload_hash", "external_url", "logs"} {
		if !hasIteration7JSONField(reflect.TypeOf(publish.PublishJobDetailResponse{}), field) {
			t.Fatalf("PublishJobDetailResponse missing json field %q", field)
		}
	}
}

// @Test
func TestTask03ServiceRejectsSensitiveConfigAndDetectsIdempotencyConflict(t *testing.T) {
	svc := publish.NewService()
	_, err := svc.CreateTarget(context.Background(), "project-1", publish.CreatePublishTargetRequest{
		Platform: "wechat", AccountName: "official", DisplayName: "Official",
		Config: map[string]any{"display": map[string]any{"api_key": "secret"}},
		Enabled: true,
	}, "target-idem-1")
	if !errors.Is(err, publish.ErrValidation) {
		t.Fatalf("sensitive nested config must return ErrValidation, got %v", err)
	}

	_, err = svc.CreateTarget(context.Background(), "project-1", publish.CreatePublishTargetRequest{
		Platform: "wechat", AccountName: "official", DisplayName: "Official",
		Config: map[string]any{"section": "news"}, Enabled: true,
	}, "target-idem-2")
	if err != nil {
		t.Fatalf("first create target must succeed: %v", err)
	}
	_, err = svc.CreateTarget(context.Background(), "project-1", publish.CreatePublishTargetRequest{
		Platform: "wechat", AccountName: "official", DisplayName: "Different",
		Config: map[string]any{"section": "news"}, Enabled: true,
	}, "target-idem-2")
	if !errors.Is(err, publish.ErrIdempotencyConflict) {
		t.Fatalf("same idempotency key with different body must return ErrIdempotencyConflict, got %v", err)
	}
}

// @Test
func TestTask03ServiceRequiresApprovedContentAndEnforcesStateMachine(t *testing.T) {
	svc := publish.NewService()
	_, err := svc.CreateJob(context.Background(), "project-1", publish.CreatePublishJobRequest{
		ContentItemID: "content-item-draft", ContentVersionID: "version-draft", TargetID: "publish-target-1",
	}, "job-idem-1")
	if !errors.Is(err, publish.ErrConflict) {
		t.Fatalf("unapproved content version must return ErrConflict, got %v", err)
	}
	_, err = svc.MarkPublished(context.Background(), "project-1", "queued-job-1", publish.MarkPublishedRequest{
		ExternalURL: "https://example.com/post/1",
	}, "published-idem-1")
	if !errors.Is(err, publish.ErrConflict) {
		t.Fatalf("mark published must reject non-copied jobs with ErrConflict, got %v", err)
	}
	_, err = svc.Requeue(context.Background(), "project-1", "published-job-1", publish.RequeuePublishJobRequest{
		Reason: "republish",
	}, "requeue-idem-1")
	if !errors.Is(err, publish.ErrConflict) {
		t.Fatalf("published jobs must not be requeued, got %v", err)
	}
}

// @Test
func TestTask04PublishTargetHTTPCoversSuccessValidationAndIdempotencyConflict(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/publish-targets", nil)
	rr := httptest.NewRecorder()
	iteration7Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("publish targets without bearer token = %d, want 401", rr.Code)
	}

	body := []byte(`{"platform":"wechat","account_name":"official","display_name":"Official","config":{"section":"news"},"enabled":true}`)
	rr = iteration7Request(http.MethodPost, "/api/v1/projects/project-1/publish-targets", body, "target-http-idem")
	if rr.Code != http.StatusCreated {
		t.Fatalf("first create target = %d body=%s", rr.Code, rr.Body.String())
	}
	conflictingBody := []byte(`{"platform":"wechat","account_name":"official","display_name":"Changed","config":{"section":"news"},"enabled":true}`)
	rr = iteration7Request(http.MethodPost, "/api/v1/projects/project-1/publish-targets", conflictingBody, "target-http-idem")
	if rr.Code != http.StatusConflict {
		t.Fatalf("same Idempotency-Key with changed target body = %d, want 409 body=%s", rr.Code, rr.Body.String())
	}
	env := decodeIteration7Envelope(t, rr.Body.Bytes())
	if env.Error == nil || env.Error.Code != "IDEMPOTENCY_CONFLICT" {
		t.Fatalf("idempotency conflict must use IDEMPOTENCY_CONFLICT envelope: %s", rr.Body.String())
	}
}

// @Test
func TestTask05PublishJobHTTPReturnsQueueDetailAndNotFoundContracts(t *testing.T) {
	rr := iteration7Request(http.MethodGet, "/api/v1/projects/project-1/publish-jobs?status=queued&page=1&page_size=20&sort=created_at&order=desc", nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("list publish jobs = %d body=%s", rr.Code, rr.Body.String())
	}
	env := decodeIteration7Envelope(t, rr.Body.Bytes())
	items, ok := env.Data["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("publish queue response must contain items: %s", rr.Body.String())
	}
	item, ok := items[0].(map[string]any)
	if !ok || item["status"] == nil || item["actions"] == nil || item["retry_count"] == nil {
		t.Fatalf("publish queue item missing status/actions/retry_count: %s", rr.Body.String())
	}

	rr = iteration7Request(http.MethodGet, "/api/v1/projects/project-1/publish-jobs/unknown-job", nil, "")
	if rr.Code != http.StatusNotFound {
		t.Fatalf("unknown publish job detail = %d, want 404 body=%s", rr.Code, rr.Body.String())
	}
}

// @Test
func TestTask06CopyPayloadHTTPSeparatesPreviewFromCopyMutation(t *testing.T) {
	rr := iteration7Request(http.MethodGet, "/api/v1/projects/project-1/publish-jobs/publish-job-1/copy-payload", nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("get copy payload = %d body=%s", rr.Code, rr.Body.String())
	}
	env := decodeIteration7Envelope(t, rr.Body.Bytes())
	if env.Data["content_version_id"] == "" || env.Data["payload_hash"] == "sha256-placeholder" {
		t.Fatalf("copy payload must bind a real content_version_id and stable payload_hash: %s", rr.Body.String())
	}

	rr = iteration7Request(http.MethodPost, "/api/v1/projects/project-1/publish-jobs/publish-job-1/copy", []byte(`{"copy_scope":"full","note":"manual copy"}`), "copy-idem")
	if rr.Code != http.StatusOK {
		t.Fatalf("copy publish payload = %d body=%s", rr.Code, rr.Body.String())
	}
	rr = iteration7Request(http.MethodPost, "/api/v1/projects/project-1/publish-jobs/publish-job-1/copy", []byte(`{"copy_scope":"title","note":"changed"}`), "copy-idem")
	if rr.Code != http.StatusConflict {
		t.Fatalf("copy with same Idempotency-Key but different body = %d, want 409 body=%s", rr.Code, rr.Body.String())
	}
}

// @Test
func TestTask07BackfillHTTPEnforcesStatusAndReasonRules(t *testing.T) {
	rr := iteration7Request(http.MethodPost, "/api/v1/projects/project-1/publish-jobs/publish-job-1/mark-published", []byte(`{"external_url":"https://example.com/post/1","note":"manual"}`), "mark-published-idem")
	if rr.Code != http.StatusConflict {
		t.Fatalf("mark published before copied = %d, want 409 body=%s", rr.Code, rr.Body.String())
	}

	rr = iteration7Request(http.MethodPost, "/api/v1/projects/project-1/publish-jobs/publish-job-1/mark-failed", []byte(`{"retryable":true,"note":"missing reason"}`), "mark-failed-idem")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("mark failed without reason = %d, want 400 body=%s", rr.Code, rr.Body.String())
	}
}

// @Test
func TestTask08OpenAPIDeclaresCompletePublishContract(t *testing.T) {
	openapi := readIteration7RepoFile(t, "openapi", "openapi.yaml")
	for _, want := range []string{
		"operationId: listPublishTargets",
		"operationId: createPublishTarget",
		"operationId: updatePublishTarget",
		"operationId: createPublishJob",
		"operationId: listPublishJobs",
		"operationId: getPublishJob",
		"operationId: getPublishCopyPayload",
		"operationId: copyPublishPayload",
		"operationId: markPublishJobPublished",
		"operationId: markPublishJobFailed",
		"operationId: requeuePublishJob",
		"Idempotency-Key",
		"IDEMPOTENCY_CONFLICT",
		"examples:",
	} {
		if !strings.Contains(openapi, want) {
			t.Fatalf("OpenAPI publish contract missing %q", want)
		}
	}
	for _, op := range []string{"operationId: createPublishTarget", "operationId: createPublishJob", "operationId: copyPublishPayload", "operationId: markPublishJobPublished", "operationId: markPublishJobFailed", "operationId: requeuePublishJob"} {
		idx := strings.Index(openapi, op)
		if idx < 0 {
			t.Fatalf("OpenAPI missing %s", op)
		}
		window := openapi[idx:min(len(openapi), idx+1600)]
		if !strings.Contains(window, "IDEMPOTENCY_CONFLICT") {
			t.Fatalf("%s must declare IDEMPOTENCY_CONFLICT response", op)
		}
	}
}

// @Test
func TestTask09WebClientDeclaresPublishFunctionsWithIdempotencyAndFiltering(t *testing.T) {
	apiClient := readIteration7RepoFile(t, "apps", "web-admin", "lib", "api.ts")
	for _, want := range []string{
		"export type PublishTargetResponse",
		"export type PublishJobResponse",
		"export type PublishCopyPayloadResponse",
		"fetchPublishTargets(projectID",
		"createPublishTarget(projectID",
		"updatePublishTarget(projectID",
		"fetchPublishJobs(projectID",
		"createPublishJob(projectID",
		"fetchPublishJob(projectID",
		"fetchPublishCopyPayload(projectID",
		"copyPublishPayload(projectID",
		"markPublishJobPublished(projectID",
		"markPublishJobFailed(projectID",
		"requeuePublishJob(projectID",
		"Idempotency-Key",
		"sort",
		"order",
		"scheduled_from",
	} {
		if !strings.Contains(apiClient, want) {
			t.Fatalf("web publish API client missing %q", want)
		}
	}
	targetFn := apiClient[strings.Index(apiClient, "fetchPublishTargets(projectID"):strings.Index(apiClient, "export async function createPublishTarget")]
	if !strings.Contains(targetFn, "sort") || !strings.Contains(targetFn, "order") {
		t.Fatalf("fetchPublishTargets must expose sort/order query support")
	}
}

// @Test
func TestTask10PublishQueuePageExposesFiltersPaginationTargetManagementAndCreateFlow(t *testing.T) {
	page := readIteration7RepoFile(t, "apps", "web-admin", "app", "projects", "[projectId]", "publish-jobs", "page.tsx")
	nav := readIteration7RepoFile(t, "apps", "web-admin", "app", "projects", "[projectId]", "workspace-nav.tsx")
	for _, want := range []string{"发布队列", "fetchPublishJobs", "fetchPublishTargets", "createPublishJob", "状态筛选", "目标筛选", "上一页", "下一页", "新建发布目标", "request_id"} {
		if !strings.Contains(page, want) {
			t.Fatalf("publish queue page missing %q", want)
		}
	}
	if !strings.Contains(nav, "publish-jobs") || !strings.Contains(nav, "发布队列") {
		t.Fatalf("project workspace nav must expose publish queue entry")
	}
}

// @Test
func TestTask11PublishDetailCopyBackfillPagesExposeLogsPayloadAndErrors(t *testing.T) {
	detail := readIteration7RepoFile(t, "apps", "web-admin", "app", "publish-jobs", "[jobId]", "page.tsx")
	copyPage := readIteration7RepoFile(t, "apps", "web-admin", "app", "publish-jobs", "[jobId]", "copy", "page.tsx")
	backfill := readIteration7RepoFile(t, "apps", "web-admin", "app", "publish-jobs", "[jobId]", "backfill", "page.tsx")
	for _, want := range []string{"fetchPublishJob", "行为摘要", "暂无发布日志", "payload_hash", "external_url", "request_id"} {
		if !strings.Contains(detail, want) {
			t.Fatalf("publish detail page missing %q", want)
		}
	}
	for _, want := range []string{"fetchPublishCopyPayload", "copyPublishPayload", "content_version_id", "payload_hash", "复制标题", "复制正文", "复制完整载荷"} {
		if !strings.Contains(copyPage, want) {
			t.Fatalf("publish copy page missing %q", want)
		}
	}
	for _, want := range []string{"markPublishJobPublished", "markPublishJobFailed", "requeuePublishJob", "published_at", "external_url", "reason", "request_id"} {
		if !strings.Contains(backfill, want) {
			t.Fatalf("publish backfill page missing %q", want)
		}
	}
}
