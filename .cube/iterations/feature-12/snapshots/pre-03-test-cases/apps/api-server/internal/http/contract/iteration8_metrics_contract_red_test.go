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
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	serverhttp "github.com/wangding75/ai-content-go/apps/api-server/internal/http"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/handlers"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/metrics"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/system"
)

type iteration8SystemService struct{}

func (iteration8SystemService) Health(ctx context.Context) (system.HealthResponse, error) {
	return system.HealthResponse{}, nil
}

func (iteration8SystemService) Info(ctx context.Context) (system.InfoResponse, error) {
	return system.InfoResponse{}, nil
}

func (iteration8SystemService) ConfigCheck(ctx context.Context) (system.ConfigCheckResponse, error) {
	return system.ConfigCheckResponse{}, nil
}

func (iteration8SystemService) DBCheck(ctx context.Context) (system.DBCheckResponse, error) {
	return system.DBCheckResponse{}, nil
}

func (iteration8SystemService) MigrationStatus(ctx context.Context) (system.MigrationStatusResponse, error) {
	return system.MigrationStatusResponse{}, nil
}

func iteration8Router() http.Handler {
	return serverhttp.NewRouter(iteration8SystemService{}, nil)
}

func iteration8Request(method, path string, body []byte, idempotencyKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("X-Request-Id", "req-iteration-8")
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	iteration8Router().ServeHTTP(rr, req)
	return rr
}

func decodeIteration8Envelope(t *testing.T, body []byte) struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string           `json:"code"`
		Message string           `json:"message"`
		Details []map[string]any `json:"details"`
	} `json:"error"`
	RequestID string `json:"request_id"`
} {
	t.Helper()
	var env struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
		Error   *struct {
			Code    string           `json:"code"`
			Message string           `json:"message"`
			Details []map[string]any `json:"details"`
		} `json:"error"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, string(body))
	}
	return env
}

func decodeIteration8Data[T any](t *testing.T, env struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string           `json:"code"`
		Message string           `json:"message"`
		Details []map[string]any `json:"details"`
	} `json:"error"`
	RequestID string `json:"request_id"`
}) T {
	t.Helper()
	var data T
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("decode data: %v data=%s", err, string(env.Data))
	}
	return data
}

func readIteration8RepoFile(t *testing.T, parts ...string) string {
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

type fakeMetricsService struct {
	lastTemplateRequest metrics.CreateMetricTemplateRequest
	lastListTemplates   metrics.ListMetricTemplatesRequest
	lastRecordRequest   metrics.CreateMetricRecordRequest
	lastBatchRequest    metrics.BatchCreateMetricRecordsRequest
	lastListRecords     metrics.ListMetricRecordsRequest
	lastSummaryProject  string
	lastSummaryRequest  metrics.MetricSummaryRequest
	lastTrendProject    string
	lastTrendRequest    metrics.MetricTrendRequest
	lastMissingProject  string
	lastMissingRequest  metrics.MissingMetricDatesRequest
	lastIdempotencyKey  string
	errByMethod         map[string]error
}

func (s *fakeMetricsService) CreateTemplate(ctx context.Context, req metrics.CreateMetricTemplateRequest) (metrics.CreateMetricTemplateResponse, error) {
	s.lastTemplateRequest = req
	if err := s.errByMethod["CreateTemplate"]; err != nil {
		return metrics.CreateMetricTemplateResponse{}, err
	}
	return metrics.CreateMetricTemplateResponse{MetricTemplateID: "metric-template-views"}, nil
}

func (s *fakeMetricsService) ListTemplates(ctx context.Context, req metrics.ListMetricTemplatesRequest) (metrics.PagedMetricTemplatesResponse, error) {
	s.lastListTemplates = req
	if err := s.errByMethod["ListTemplates"]; err != nil {
		return metrics.PagedMetricTemplatesResponse{}, err
	}
	return metrics.PagedMetricTemplatesResponse{
		Items: []metrics.MetricTemplateResponse{{ID: "metric-template-views", ContentType: "article", Platform: "wechat", MetricCode: "views", Unit: "次", AggregationMethod: metrics.AggregationSum, Period: metrics.PeriodDay, Required: true, Enabled: true}},
		Pagination: content.PaginationResponse{
			Page: 1, PageSize: 20, Total: 1, HasNext: false,
		},
	}, nil
}

func (s *fakeMetricsService) CreateRecord(ctx context.Context, req metrics.CreateMetricRecordRequest, idempotencyKey string) (metrics.CreateMetricRecordResponse, error) {
	s.lastRecordRequest = req
	s.lastIdempotencyKey = idempotencyKey
	if err := s.errByMethod["CreateRecord"]; err != nil {
		return metrics.CreateMetricRecordResponse{}, err
	}
	return metrics.CreateMetricRecordResponse{MetricRecordID: "metric-record-views-20260525", NormalizedValue: 1200, OperationLogID: "operation-log-metric-record"}, nil
}

func (s *fakeMetricsService) BatchCreateRecords(ctx context.Context, req metrics.BatchCreateMetricRecordsRequest, idempotencyKey string) (metrics.BatchCreateMetricRecordsResponse, error) {
	s.lastBatchRequest = req
	s.lastIdempotencyKey = idempotencyKey
	if err := s.errByMethod["BatchCreateRecords"]; err != nil {
		return metrics.BatchCreateMetricRecordsResponse{}, err
	}
	return metrics.BatchCreateMetricRecordsResponse{
		CreatedCount: 1,
		FailedCount:  1,
		Errors: []metrics.BatchMetricRecordError{{
			Index: 1, MetricCode: "likes", Field: "metric_date", Code: "VALIDATION_ERROR", Message: "invalid metric_date", SourceRef: "row-2",
		}},
		OperationLogID: "operation-log-metric-batch",
	}, nil
}

func (s *fakeMetricsService) ListRecords(ctx context.Context, req metrics.ListMetricRecordsRequest) (metrics.PagedMetricRecordsResponse, error) {
	s.lastListRecords = req
	if err := s.errByMethod["ListRecords"]; err != nil {
		return metrics.PagedMetricRecordsResponse{}, err
	}
	return metrics.PagedMetricRecordsResponse{
		Items: []metrics.MetricRecordResponse{{ID: "metric-record-views-20260525", ProjectID: req.ProjectID, ContentItemID: "content-item-1", ContentVersionID: "content-version-approved-1", PublishJobID: "publish-job-1", TargetID: "publish-target-1", ContentType: "article", MetricTemplateID: "metric-template-views", Platform: "wechat", MetricCode: "views", MetricDate: "2026-05-25", Period: metrics.PeriodDay, RawValue: "1200", NormalizedValue: 1200, SourceType: metrics.SourceManual, SourceRef: "manual-entry"}},
		Pagination: content.PaginationResponse{
			Page: 1, PageSize: 20, Total: 1, HasNext: false,
		},
	}, nil
}

func (s *fakeMetricsService) GetSummary(ctx context.Context, projectID string, req metrics.MetricSummaryRequest) (metrics.MetricSummaryResponse, error) {
	s.lastSummaryProject = projectID
	s.lastSummaryRequest = req
	if err := s.errByMethod["GetSummary"]; err != nil {
		return metrics.MetricSummaryResponse{}, err
	}
	return metrics.MetricSummaryResponse{
		ProjectID: projectID, DateFrom: req.DateFrom, DateTo: req.DateTo, Platform: req.Platform, TargetID: req.TargetID,
		Items:             []metrics.MetricSummaryItem{{MetricCode: "views", Value: 1200, Unit: "次", AggregationMethod: metrics.AggregationSum, SourceRecordCount: 1}},
		SummarySnapshotID: "metric-summary-snapshot-1", SourceRecordCount: 1,
	}, nil
}

func (s *fakeMetricsService) GetTrends(ctx context.Context, projectID string, req metrics.MetricTrendRequest) (metrics.MetricTrendResponse, error) {
	s.lastTrendProject = projectID
	s.lastTrendRequest = req
	if err := s.errByMethod["GetTrends"]; err != nil {
		return metrics.MetricTrendResponse{}, err
	}
	return metrics.MetricTrendResponse{
		ProjectID: projectID, MetricCode: req.MetricCode, Bucket: req.Bucket, AggregationMethod: metrics.AggregationSum, QuerySignature: "metric-trend-seed-project-views-wechat-publish-target-1-day", SourceRecordCount: 1,
		Series:        []metrics.MetricTrendPoint{{BucketStart: "2026-05-25", Value: 1200, SourceRecordCount: 1, Missing: false}, {BucketStart: "2026-05-24", Missing: true}},
		MissingPoints: []metrics.MetricMissingPoint{{MetricDate: "2026-05-24", Reason: "required_metric_missing"}},
	}, nil
}

func (s *fakeMetricsService) GetMissingDates(ctx context.Context, projectID string, req metrics.MissingMetricDatesRequest) (metrics.MissingMetricDatesResponse, error) {
	s.lastMissingProject = projectID
	s.lastMissingRequest = req
	if err := s.errByMethod["GetMissingDates"]; err != nil {
		return metrics.MissingMetricDatesResponse{}, err
	}
	return metrics.MissingMetricDatesResponse{ProjectID: projectID, Items: []metrics.MissingMetricDateItem{{ContentItemID: "content-item-1", ContentVersionID: "content-version-approved-1", PublishJobID: "publish-job-1", TargetID: "publish-target-1", Platform: "wechat", MetricCode: "views", Period: metrics.PeriodDay, MetricDate: "2026-05-24", MissingReason: "required_metric_missing", BackfillHint: "补录 views 2026-05-24"}}}, nil
}

func (s *fakeMetricsService) SubmitPlatformCollectLog(ctx context.Context, req metrics.SubmitPlatformCollectLogRequest, auth metrics.PlatformCollectLogAuth, idempotencyKey string) (metrics.SubmitPlatformCollectLogResponse, error) {
	return metrics.SubmitPlatformCollectLogResponse{}, nil
}

func (s *fakeMetricsService) ListPlatformCollectLogs(ctx context.Context, req metrics.ListPlatformCollectLogsRequest) (metrics.PagedPlatformCollectLogsResponse, error) {
	return metrics.PagedPlatformCollectLogsResponse{}, nil
}

func (s *fakeMetricsService) GetPlatformCollectLog(ctx context.Context, collectLogID string) (metrics.PlatformCollectLogDetailResponse, error) {
	return metrics.PlatformCollectLogDetailResponse{}, nil
}

func (s *fakeMetricsService) ConfirmPlatformCollectLogMetrics(ctx context.Context, collectLogID string, req metrics.ConfirmPlatformCollectLogMetricsRequest, idempotencyKey string) (metrics.ConfirmPlatformCollectLogMetricsResponse, error) {
	return metrics.ConfirmPlatformCollectLogMetricsResponse{}, nil
}

func iteration8MetricsHandlerRequest(svc *fakeMetricsService, method string, path string, body []byte, idempotencyKey string) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	handler := handlers.NewMetricsHandler(svc, svc, nil)
	r.Post("/metric-templates", handler.CreateTemplate)
	r.Get("/metric-templates", handler.ListTemplates)
	r.Post("/metric-records", handler.CreateRecord)
	r.Post("/metric-records/batch", handler.BatchCreateRecords)
	r.Get("/metric-records", handler.ListRecords)
	r.Get("/projects/{projectId}/metrics/summary", handler.GetSummary)
	r.Get("/projects/{projectId}/metrics/trends", handler.GetTrends)
	r.Get("/projects/{projectId}/metrics/missing-dates", handler.GetMissingDates)

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("X-Request-Id", "req-iteration-8")
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func assertIteration8SuccessEnvelope(t *testing.T, rr *httptest.ResponseRecorder, status int) struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string           `json:"code"`
		Message string           `json:"message"`
		Details []map[string]any `json:"details"`
	} `json:"error"`
	RequestID string `json:"request_id"`
} {
	t.Helper()
	if rr.Code != status {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, status, rr.Body.String())
	}
	env := decodeIteration8Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || len(env.Data) == 0 || string(env.Data) == "null" || env.RequestID == "" {
		t.Fatalf("success envelope must set data, null error, and request_id: %s", rr.Body.String())
	}
	return env
}

func assertIteration8ErrorEnvelope(t *testing.T, rr *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if rr.Code != status {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, status, rr.Body.String())
	}
	env := decodeIteration8Envelope(t, rr.Body.Bytes())
	if env.Success || string(env.Data) != "null" || env.Error == nil || env.Error.Code != code || env.RequestID == "" {
		t.Fatalf("error envelope must set data=null, success=false, code=%s, request_id: %s", code, rr.Body.String())
	}
}

// @Test
func TestTask03MetricTemplateHTTPCoversAuthCreateListAndConflictEnvelope(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/metric-templates", nil)
	rr := httptest.NewRecorder()
	iteration8Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("metric templates without bearer token = %d, want 401", rr.Code)
	}

	svc := &fakeMetricsService{errByMethod: map[string]error{}}
	body := []byte(`{"content_type":"article","platform":"wechat","metric_code":"views","metric_name":"阅读量","unit":"次","value_type":"integer","aggregation_method":"sum","period":"day","required":true,"enabled":true}`)
	rr = iteration8MetricsHandlerRequest(svc, http.MethodPost, "/metric-templates", body, "")
	env := assertIteration8SuccessEnvelope(t, rr, http.StatusCreated)
	created := decodeIteration8Data[metrics.CreateMetricTemplateResponse](t, env)
	if created.MetricTemplateID != "metric-template-views" || svc.lastTemplateRequest.ContentType != "article" || svc.lastTemplateRequest.AggregationMethod != metrics.AggregationSum {
		t.Fatalf("create template must pass body to service and return metric_template_id, got %#v request=%#v", created, svc.lastTemplateRequest)
	}

	rr = iteration8MetricsHandlerRequest(svc, http.MethodGet, "/metric-templates?content_type=article&platform=wechat&enabled=true&page=1&page_size=20&sort=metric_code&order=asc", nil, "")
	env = assertIteration8SuccessEnvelope(t, rr, http.StatusOK)
	list := decodeIteration8Data[metrics.PagedMetricTemplatesResponse](t, env)
	if len(list.Items) != 1 || list.Items[0].MetricCode != "views" || list.Pagination.Page != 1 || svc.lastListTemplates.Enabled == nil || !*svc.lastListTemplates.Enabled {
		t.Fatalf("list templates must preserve filters and pagination, got data=%#v request=%#v", list, svc.lastListTemplates)
	}

	for _, tc := range []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "validation", err: metrics.ErrValidation, status: http.StatusBadRequest, code: "VALIDATION_ERROR"},
		{name: "forbidden", err: metrics.ErrForbidden, status: http.StatusForbidden, code: "FORBIDDEN"},
		{name: "conflict", err: metrics.ErrConflict, status: http.StatusConflict, code: "CONFLICT"},
		{name: "internal", err: metrics.ErrInternal, status: http.StatusInternalServerError, code: "INTERNAL_ERROR"},
	} {
		svc.errByMethod = map[string]error{"CreateTemplate": tc.err}
		rr = iteration8MetricsHandlerRequest(svc, http.MethodPost, "/metric-templates", body, "")
		assertIteration8ErrorEnvelope(t, rr, tc.status, tc.code)
	}
}

// @Test
func TestTask04MetricRecordHTTPRequiresIdempotencyAndPreservesBatchRowErrors(t *testing.T) {
	svc := &fakeMetricsService{errByMethod: map[string]error{}}
	body := []byte(`{"project_id":"seed-project","content_item_id":"content-item-1","content_version_id":"content-version-approved-1","publish_job_id":"publish-job-1","target_id":"publish-target-1","platform":"wechat","external_url":"https://example.com/post/1","metric_code":"views","metric_date":"2026-05-25","period":"day","raw_value":"1200","source_type":"manual","source_ref":"manual-entry"}`)
	svc.errByMethod = map[string]error{"CreateRecord": metrics.ErrValidation}
	rr := iteration8MetricsHandlerRequest(svc, http.MethodPost, "/metric-records", body, "")
	assertIteration8ErrorEnvelope(t, rr, http.StatusBadRequest, "VALIDATION_ERROR")

	svc.errByMethod = map[string]error{}
	rr = iteration8MetricsHandlerRequest(svc, http.MethodPost, "/metric-records", body, "metrics-record-idem-1")
	env := assertIteration8SuccessEnvelope(t, rr, http.StatusCreated)
	created := decodeIteration8Data[metrics.CreateMetricRecordResponse](t, env)
	if created.MetricRecordID == "" || created.NormalizedValue != 1200 || created.OperationLogID == "" || svc.lastIdempotencyKey != "metrics-record-idem-1" || svc.lastRecordRequest.SourceType != metrics.SourceManual {
		t.Fatalf("create record must return id, normalized value, operation log and preserve idempotency/body, got data=%#v request=%#v key=%q", created, svc.lastRecordRequest, svc.lastIdempotencyKey)
	}

	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{err: metrics.ErrNotFound, status: http.StatusNotFound, code: "NOT_FOUND"},
		{err: metrics.ErrConflict, status: http.StatusConflict, code: "CONFLICT"},
		{err: metrics.ErrIdempotencyConflict, status: http.StatusConflict, code: "IDEMPOTENCY_CONFLICT"},
		{err: metrics.ErrInternal, status: http.StatusInternalServerError, code: "INTERNAL_ERROR"},
	} {
		svc.errByMethod = map[string]error{"CreateRecord": tc.err}
		rr = iteration8MetricsHandlerRequest(svc, http.MethodPost, "/metric-records", body, "metrics-record-idem-2")
		assertIteration8ErrorEnvelope(t, rr, tc.status, tc.code)
	}

	batchBody := []byte(`{"import_source":"csv-upload","records":[{"project_id":"seed-project","content_item_id":"content-item-1","content_version_id":"content-version-approved-1","publish_job_id":"publish-job-1","target_id":"publish-target-1","platform":"wechat","external_url":"https://example.com/post/1","metric_code":"likes","metric_date":"2026-05-25","period":"day","raw_value":"33","source_type":"import","source_ref":"row-1"},{"project_id":"seed-project","content_item_id":"","content_version_id":"content-version-approved-1","publish_job_id":"publish-job-1","target_id":"publish-target-1","platform":"wechat","metric_code":"likes","metric_date":"bad-date","period":"day","raw_value":"oops","source_type":"import","source_ref":"row-2"}]}`)
	svc.errByMethod = map[string]error{}
	rr = iteration8MetricsHandlerRequest(svc, http.MethodPost, "/metric-records/batch", batchBody, "metrics-batch-idem-1")
	env = assertIteration8SuccessEnvelope(t, rr, http.StatusOK)
	batch := decodeIteration8Data[metrics.BatchCreateMetricRecordsResponse](t, env)
	if batch.CreatedCount != 1 || batch.FailedCount != 1 || len(batch.Errors) != 1 || batch.Errors[0].Index != 1 || batch.Errors[0].Code != "VALIDATION_ERROR" || batch.Errors[0].SourceRef != "row-2" || svc.lastBatchRequest.ImportSource != "csv-upload" {
		t.Fatalf("batch response must preserve counts and row error details, got data=%#v request=%#v", batch, svc.lastBatchRequest)
	}
}

// @Test
func TestTask05MetricReadHTTPCoversProjectScopedListSummaryTrendAndMissingDates(t *testing.T) {
	svc := &fakeMetricsService{errByMethod: map[string]error{"ListRecords": metrics.ErrValidation}}
	rr := iteration8MetricsHandlerRequest(svc, http.MethodGet, "/metric-records?platform=wechat", nil, "")
	assertIteration8ErrorEnvelope(t, rr, http.StatusBadRequest, "VALIDATION_ERROR")

	svc.errByMethod = map[string]error{}
	rr = iteration8MetricsHandlerRequest(svc, http.MethodGet, "/metric-records?project_id=seed-project&metric_code=views&date_from=2026-05-01&date_to=2026-05-25&page=1&page_size=20&sort=metric_date&order=desc", nil, "")
	env := assertIteration8SuccessEnvelope(t, rr, http.StatusOK)
	records := decodeIteration8Data[metrics.PagedMetricRecordsResponse](t, env)
	if len(records.Items) != 1 || records.Items[0].MetricTemplateID == "" || records.Pagination.Page != 1 || svc.lastListRecords.ProjectID != "seed-project" || svc.lastListRecords.DateTo != "2026-05-25" {
		t.Fatalf("list records must preserve project/date filters and pagination, got data=%#v request=%#v", records, svc.lastListRecords)
	}

	rr = iteration8MetricsHandlerRequest(svc, http.MethodGet, "/projects/seed-project/metrics/summary?date_from=2026-05-01&date_to=2026-05-25&metric_codes=views,likes&platform=wechat&target_id=publish-target-1", nil, "")
	env = assertIteration8SuccessEnvelope(t, rr, http.StatusOK)
	summary := decodeIteration8Data[metrics.MetricSummaryResponse](t, env)
	if summary.SummarySnapshotID == "" || summary.SourceRecordCount != 1 || len(summary.Items) != 1 || summary.Items[0].AggregationMethod != metrics.AggregationSum || svc.lastSummaryProject != "seed-project" || len(svc.lastSummaryRequest.MetricCodes) != 2 {
		t.Fatalf("summary must include stable snapshot, aggregation method and source count, got data=%#v request=%#v project=%q", summary, svc.lastSummaryRequest, svc.lastSummaryProject)
	}

	rr = iteration8MetricsHandlerRequest(svc, http.MethodGet, "/projects/seed-project/metrics/trends?metric_code=views&date_from=2026-05-01&date_to=2026-05-25&bucket=day&platform=wechat&target_id=publish-target-1", nil, "")
	env = assertIteration8SuccessEnvelope(t, rr, http.StatusOK)
	trend := decodeIteration8Data[metrics.MetricTrendResponse](t, env)
	if trend.QuerySignature == "" || len(trend.Series) != 2 || len(trend.MissingPoints) != 1 || !trend.Series[1].Missing || svc.lastTrendRequest.TargetID != "publish-target-1" {
		t.Fatalf("trends must preserve target filter and express missing point structurally without nonzero value, got data=%#v request=%#v", trend, svc.lastTrendRequest)
	}
	var trendJSON struct {
		Series []map[string]any `json:"series"`
	}
	if err := json.Unmarshal(env.Data, &trendJSON); err != nil {
		t.Fatalf("decode trend raw json: %v", err)
	}
	for _, point := range trendJSON.Series {
		if point["missing"] == true {
			if _, exists := point["value"]; exists {
				t.Fatalf("missing trend points must omit value instead of encoding missing as 0: %s", string(env.Data))
			}
		}
	}

	rr = iteration8MetricsHandlerRequest(svc, http.MethodGet, "/projects/seed-project/metrics/missing-dates?date_from=2026-05-01&date_to=2026-05-25&platform=wechat&target_id=publish-target-1", nil, "")
	env = assertIteration8SuccessEnvelope(t, rr, http.StatusOK)
	missing := decodeIteration8Data[metrics.MissingMetricDatesResponse](t, env)
	if len(missing.Items) != 1 || missing.Items[0].Period != metrics.PeriodDay || missing.Items[0].MissingReason != "required_metric_missing" || svc.lastMissingRequest.DateFrom != "2026-05-01" {
		t.Fatalf("missing dates must include period, reason and query filters, got data=%#v request=%#v", missing, svc.lastMissingRequest)
	}

	svc.errByMethod = map[string]error{"GetSummary": metrics.ErrNotFound, "GetTrends": metrics.ErrForbidden, "GetMissingDates": metrics.ErrInternal}
	assertIteration8ErrorEnvelope(t, iteration8MetricsHandlerRequest(svc, http.MethodGet, "/projects/seed-project/metrics/summary?date_from=2026-05-01&date_to=2026-05-25&metric_codes=views", nil, ""), http.StatusNotFound, "NOT_FOUND")
	assertIteration8ErrorEnvelope(t, iteration8MetricsHandlerRequest(svc, http.MethodGet, "/projects/seed-project/metrics/trends?metric_code=views&date_from=2026-05-01&date_to=2026-05-25&bucket=bad", nil, ""), http.StatusForbidden, "FORBIDDEN")
	assertIteration8ErrorEnvelope(t, iteration8MetricsHandlerRequest(svc, http.MethodGet, "/projects/seed-project/metrics/missing-dates?date_from=2026-05-01&date_to=2026-05-25", nil, ""), http.StatusInternalServerError, "INTERNAL_ERROR")
}

// @Test
func TestTask05MetricsHandlerMapsUnknownErrorsToInternalEnvelope(t *testing.T) {
	svc := &fakeMetricsService{errByMethod: map[string]error{"ListRecords": errors.New("database unavailable")}}
	rr := iteration8MetricsHandlerRequest(svc, http.MethodGet, "/metric-records?project_id=seed-project", nil, "")
	assertIteration8ErrorEnvelope(t, rr, http.StatusInternalServerError, "INTERNAL_ERROR")
}

// @Test
func TestTask06OpenAPIContainsMetricsPathsSchemasHeadersAndErrors(t *testing.T) {
	openapi := readIteration8RepoFile(t, "openapi", "openapi.yaml")
	for _, want := range []string{
		"/api/v1/metric-templates:",
		"/api/v1/metric-records:",
		"/api/v1/metric-records/batch:",
		"/api/v1/projects/{projectId}/metrics/summary:",
		"/api/v1/projects/{projectId}/metrics/trends:",
		"/api/v1/projects/{projectId}/metrics/missing-dates:",
		"IdempotencyKeyHeader",
		"CreateMetricTemplateRequest",
		"CreateMetricRecordRequest",
		"BatchMetricRecordError",
		"MetricSummaryResponse",
		"MetricTrendResponse",
		"MissingMetricDatesResponse",
		"IDEMPOTENCY_CONFLICT",
	} {
		if !strings.Contains(openapi, want) {
			t.Fatalf("OpenAPI metrics contract missing %q", want)
		}
	}
}

// @Test
func TestTask07WebClientAndPagesBindAllMetricsAPIs(t *testing.T) {
	apiClient := readIteration8RepoFile(t, "apps", "web-admin", "lib", "api.ts")
	for _, want := range []string{
		"fetchMetricTemplates",
		"createMetricTemplate",
		"createMetricRecord",
		"batchCreateMetricRecords",
		"fetchMetricRecords",
		"fetchMetricSummary",
		"fetchMetricTrends",
		"fetchMissingMetricDates",
		"Idempotency-Key",
		"pathSegment(projectID)",
	} {
		if !strings.Contains(apiClient, want) {
			t.Fatalf("web metrics API client missing %q", want)
		}
	}
	nav := readIteration8RepoFile(t, "apps", "web-admin", "app", "projects", "[projectId]", "workspace-nav.tsx")
	for _, want := range []string{"metrics", "metrics/input", "metrics/trends", "metrics/missing"} {
		if !strings.Contains(nav, want) {
			t.Fatalf("workspace nav missing metrics link %q", want)
		}
	}
}

// @Test
func TestTask08MetricsDashboardPageBindsSummaryRecordsMissingAndNavigation(t *testing.T) {
	page := readIteration8RepoFile(t, "apps", "web-admin", "app", "projects", "[projectId]", "metrics", "page.tsx")
	for _, want := range []string{
		"fetchMetricSummary",
		"fetchMetricRecords",
		"fetchMissingMetricDates",
		"summary_snapshot_id",
		"source_record_count",
		"指标表现",
		"趋势图",
		"缺失提醒",
	} {
		if !strings.Contains(page, want) {
			t.Fatalf("metrics dashboard page missing contract %q", want)
		}
	}
}

// @Test
func TestTask09MetricInputPageBindsTemplateRecordBatchAndErrorStates(t *testing.T) {
	page := readIteration8RepoFile(t, "apps", "web-admin", "app", "projects", "[projectId]", "metrics", "input", "page.tsx")
	for _, want := range []string{
		"fetchMetricTemplates",
		"createMetricTemplate",
		"createMetricRecord",
		"batchCreateMetricRecords",
		"指标录入",
		"创建模板",
		"保存指标",
		"批量导入",
		"errors",
	} {
		if !strings.Contains(page, want) {
			t.Fatalf("metric input page missing contract %q", want)
		}
	}
}

// @Test
func TestTask10TrendAndMissingPagesBindSignatureMissingReasonsAndBackfillLinks(t *testing.T) {
	trends := readIteration8RepoFile(t, "apps", "web-admin", "app", "projects", "[projectId]", "metrics", "trends", "page.tsx")
	for _, want := range []string{
		"fetchMetricTrends",
		"aggregation_method",
		"query_signature",
		"missing_points",
		"趋势图",
	} {
		if !strings.Contains(trends, want) {
			t.Fatalf("metric trends page missing contract %q", want)
		}
	}
	missing := readIteration8RepoFile(t, "apps", "web-admin", "app", "projects", "[projectId]", "metrics", "missing", "page.tsx")
	for _, want := range []string{
		"fetchMissingMetricDates",
		"missing_reason",
		"backfill_hint",
		"metric_date",
		"period",
		"metrics/input",
		"缺失提醒",
	} {
		if !strings.Contains(missing, want) {
			t.Fatalf("missing metrics page missing contract %q", want)
		}
	}
}
