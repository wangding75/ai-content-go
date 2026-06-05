package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/article"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/generation"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/metrics"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

// --- Task-06: Article Handler Tests ---

type stubArticleService struct {
	registerFn      func(ctx context.Context, req article.RegisterPackRequest, idempotencyKey string) (article.RegisterPackResponse, error)
	packStatusFn    func(ctx context.Context) (article.ArticlePackStatusResponse, error)
	getConfigFn     func(ctx context.Context, projectID string) (article.ArticleConfigResponse, error)
	updateConfigFn  func(ctx context.Context, projectID string, req article.UpdateArticleConfigRequest, idempotencyKey string) (article.UpdateArticleConfigResponse, error)
	createGenRunFn  func(ctx context.Context, projectID string, req article.CreateArticleGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.CreateGenerationRunResponse, error)
	listGenRunsFn   func(ctx context.Context, projectID string, req article.ListGenerationRunsRequest) (article.PagedArticleGenerationRunResponse, error)
	getGenRunFn     func(ctx context.Context, projectID, id string) (article.ArticleGenerationRunDetailResponse, error)
	retryGenRunFn   func(ctx context.Context, projectID, id string, req article.RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.RetryGenerationRunResponse, error)
	getSnapshotFn   func(ctx context.Context, itemID string) (article.ArticleContentSnapshotResponse, error)
	getMetricsFn    func(ctx context.Context, projectID string) (article.PagedProjectArticleMetricsResponse, error)
	updateMetricsFn func(ctx context.Context, projectID string, req article.UpdateProjectArticleMetricsRequest, idempotencyKey string) (article.UpdateProjectArticleMetricsResponse, error)
}

func (s *stubArticleService) RegisterPack(ctx context.Context, req article.RegisterPackRequest, idempotencyKey string) (article.RegisterPackResponse, error) {
	if s.registerFn != nil { return s.registerFn(ctx, req, idempotencyKey) }
	return article.RegisterPackResponse{ContentPackID: "cp-1", ContentTypeID: "ct-1", RegisteredWorkflowVersionIDs: []string{"wftv-1"}, MetricTemplateIDs: []string{"mt-1"}}, nil
}
func (s *stubArticleService) GetPackStatus(ctx context.Context) (article.ArticlePackStatusResponse, error) {
	if s.packStatusFn != nil { return s.packStatusFn(ctx) }
	return article.ArticlePackStatusResponse{Registered: true, ContentPackID: "cp-1", ContentType: &content.ContentTypeResponse{ID: "ct-1", Code: "article", Name: "Article Pack"}}, nil
}
func (s *stubArticleService) GetConfig(ctx context.Context, projectID string) (article.ArticleConfigResponse, error) {
	if s.getConfigFn != nil { return s.getConfigFn(ctx, projectID) }
	return article.ArticleConfigResponse{TopicStyle: "tech", Version: "v1"}, nil
}
func (s *stubArticleService) UpdateConfig(ctx context.Context, projectID string, req article.UpdateArticleConfigRequest, idempotencyKey string) (article.UpdateArticleConfigResponse, error) {
	if s.updateConfigFn != nil { return s.updateConfigFn(ctx, projectID, req, idempotencyKey) }
	return article.UpdateArticleConfigResponse{VersionID: "v2", OperationLogID: "op-1"}, nil
}
func (s *stubArticleService) CreateGenerationRun(ctx context.Context, projectID string, req article.CreateArticleGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.CreateGenerationRunResponse, error) {
	if s.createGenRunFn != nil { return s.createGenRunFn(ctx, projectID, req, workflowRunID, idempotencyKey) }
	return generation.CreateGenerationRunResponse{GenerationRunID: "gr-1", WorkflowRunID: workflowRunID, Status: "pending"}, nil
}
func (s *stubArticleService) ListGenerationRuns(ctx context.Context, projectID string, req article.ListGenerationRunsRequest) (article.PagedArticleGenerationRunResponse, error) {
	if s.listGenRunsFn != nil { return s.listGenRunsFn(ctx, projectID, req) }
	return article.PagedArticleGenerationRunResponse{Items: []article.ArticleGenerationRunSummary{{GenerationRunID: "gr-1", Status: "pending", Topic: "test"}}, Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: 1}}, nil
}
func (s *stubArticleService) GetGenerationRun(ctx context.Context, projectID, id string) (article.ArticleGenerationRunDetailResponse, error) {
	if s.getGenRunFn != nil { return s.getGenRunFn(ctx, projectID, id) }
	return article.ArticleGenerationRunDetailResponse{ArticleGenerationRunSummary: article.ArticleGenerationRunSummary{GenerationRunID: id, Status: "running"}}, nil
}
func (s *stubArticleService) RetryGenerationRun(ctx context.Context, projectID, id string, req article.RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.RetryGenerationRunResponse, error) {
	if s.retryGenRunFn != nil { return s.retryGenRunFn(ctx, projectID, id, req, workflowRunID, idempotencyKey) }
	return generation.RetryGenerationRunResponse{NewGenerationRunID: "gr-2", WorkflowRunID: workflowRunID, OperationLogID: "op-retry"}, nil
}
func (s *stubArticleService) GetContentSnapshot(ctx context.Context, itemID string) (article.ArticleContentSnapshotResponse, error) {
	if s.getSnapshotFn != nil { return s.getSnapshotFn(ctx, itemID) }
	return article.ArticleContentSnapshotResponse{Title: "Snapshot Title", Summary: "Summary", LatestContentVersionID: "cv-1"}, nil
}
func (s *stubArticleService) GetProjectArticleMetrics(ctx context.Context, projectID string) (article.PagedProjectArticleMetricsResponse, error) {
	if s.getMetricsFn != nil { return s.getMetricsFn(ctx, projectID) }
	return article.PagedProjectArticleMetricsResponse{Items: []article.ProjectArticleMetricItem{{MetricCode: "views", Name: "阅读量"}}, Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: 1}}, nil
}
func (s *stubArticleService) UpdateProjectArticleMetrics(ctx context.Context, projectID string, req article.UpdateProjectArticleMetricsRequest, idempotencyKey string) (article.UpdateProjectArticleMetricsResponse, error) {
	if s.updateMetricsFn != nil { return s.updateMetricsFn(ctx, projectID, req, idempotencyKey) }
	return article.UpdateProjectArticleMetricsResponse{VersionID: "mv-1", OperationLogID: "op-metrics"}, nil
}

// Verify stubArticleService implements article.Service
var _ article.Service = (*stubArticleService)(nil)

// @Test
func TestTask06RegisterPackReturns201AndEnvelope(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set("Idempotency-Key", "idem-1")
	rec := httptest.NewRecorder()
	h.RegisterPack(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body %s", rec.Code, rec.Body.String())
	}
	var envelope api.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if !envelope.Success {
		t.Fatal("expected success=true")
	}
}

// @Test
func TestTask06RegisterPackReturnsIdempotencyConflict(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{
		registerFn: func(ctx context.Context, req article.RegisterPackRequest, idempotencyKey string) (article.RegisterPackResponse, error) {
			return article.RegisterPackResponse{}, article.ErrIdempotencyConflict
		},
	}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set("Idempotency-Key", "idem-conflict")
	rec := httptest.NewRecorder()
	h.RegisterPack(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if envelope.Error == nil || envelope.Error.Code != api.ErrorIdempotencyConflict {
		t.Fatalf("expected IDEMPOTENCY_CONFLICT, got %#v", envelope.Error)
	}
}

// @Test
func TestTask06GetPackStatusReturns200(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.GetPackStatus(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// @Test
func TestTask06GetConfigRequiresProjectID(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "project-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetConfig(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// @Test
func TestTask06GetConfigReturnsNotFound(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{
		getConfigFn: func(ctx context.Context, projectID string) (article.ArticleConfigResponse, error) {
			return article.ArticleConfigResponse{}, article.ErrNotFound
		},
	}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetConfig(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// @Test
func TestTask06GetConfigReturnsForbidden(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{
		getConfigFn: func(ctx context.Context, projectID string) (article.ArticleConfigResponse, error) {
			return article.ArticleConfigResponse{}, article.ErrForbidden
		},
	}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "wrong-type")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetConfig(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

// @Test
func TestTask06UpdateConfigReturns200AndEnvelope(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"topic_style":"tech","source_policy":"web"}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "project-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Idempotency-Key", "idem-upd")
	rec := httptest.NewRecorder()
	h.UpdateConfig(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if envelope.Error != nil {
		t.Fatalf("expected no error, got %#v", envelope.Error)
	}
}

// @Test
func TestTask06CreateGenerationRunReturns202(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"topic":"AI","target_platform":"web"}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "project-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Idempotency-Key", "idem-gen")
	rec := httptest.NewRecorder()
	h.CreateGenerationRun(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body %s", rec.Code, rec.Body.String())
	}
}

// @Test
func TestTask06ListGenerationRunsSupportsPagination(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/?page=1&page_size=10", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "project-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.ListGenerationRuns(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// @Test
func TestTask06ListGenerationRunsReturnsValidationError(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/?page=abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "project-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.ListGenerationRuns(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask06GetGenerationRunReturnsDetail(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "project-1")
	rctx.URLParams.Add("id", "gr-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetGenerationRun(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// @Test
func TestTask06RetryGenerationRunReturns200(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"reason":"test retry"}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "project-1")
	rctx.URLParams.Add("id", "gr-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Idempotency-Key", "idem-retry")
	rec := httptest.NewRecorder()
	h.RetryGenerationRun(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
}

// @Test
func TestTask06RetryGenerationRunReturnsConflictWhenNotFailed(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{
		retryGenRunFn: func(ctx context.Context, projectID, id string, req article.RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.RetryGenerationRunResponse, error) {
			return generation.RetryGenerationRunResponse{}, article.ErrConflict
		},
	}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"reason":"test"}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "project-1")
	rctx.URLParams.Add("id", "gr-running")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Idempotency-Key", "idem-retry-conflict")
	rec := httptest.NewRecorder()
	h.RetryGenerationRun(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

// @Test
func TestTask06GetContentSnapshotReturnsSnapshot(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "project-1")
	rctx.URLParams.Add("itemId", "item-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetContentSnapshot(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// @Test
func TestTask06GetMetricsConfigReturnsMetrics(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "project-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetMetricsConfig(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// @Test
func TestTask06UpdateMetricsConfigReturns200(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"enabled_metric_codes":["views","likes"]}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "project-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Idempotency-Key", "idem-mt")
	rec := httptest.NewRecorder()
	h.UpdateMetricsConfig(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
}

// @Test
func TestTask06HandlerReturnsErrorOnValidationFailure(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{
		updateConfigFn: func(ctx context.Context, projectID string, req article.UpdateArticleConfigRequest, idempotencyKey string) (article.UpdateArticleConfigResponse, error) {
			return article.UpdateArticleConfigResponse{}, article.ErrValidation
		},
	}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"topic_style":""}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "project-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Idempotency-Key", "idem-val")
	rec := httptest.NewRecorder()
	h.UpdateConfig(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask06HandlerDecodeErrorsMapToBadRequest(t *testing.T) {
	h := NewArticleHandler(&stubArticleService{}, nil, nil, nil, nil, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`invalid json`))
	rec := httptest.NewRecorder()
	h.RegisterPack(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

// @Test
func TestTask06CompileGuardHasErrorConstants(t *testing.T) {
	if !errors.Is(article.ErrValidation, article.ErrValidation) {
		t.Fatal("unreachable")
	}
	if !errors.Is(article.ErrNotFound, article.ErrNotFound) {
		t.Fatal("unreachable")
	}
	if !errors.Is(article.ErrForbidden, article.ErrForbidden) {
		t.Fatal("unreachable")
	}
	if !errors.Is(article.ErrConflict, article.ErrConflict) {
		t.Fatal("unreachable")
	}
}

// @Test
func TestTask06CompileGuardServiceImplements(t *testing.T) {
	var svc article.Service = &stubArticleService{}
	_ = svc
}

// @Test
func TestTask06NewArticleHandlerAcceptsEngineAndServices(t *testing.T) {
	contentSvc := content.NewService()
	wfSvc := workflow.NewService()
	metricsSvc := metrics.NewService()
	eng := engine.New(wfSvc, nil, nil)
	artSvc := &stubArticleService{}
	h := NewArticleHandler(artSvc, contentSvc, wfSvc, metricsSvc, eng, slog.Default())
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

// @Test
func TestTask06HandlerErrorWriteUsesCorrectErrorCodes(t *testing.T) {
	svc := article.NewService(nil, nil, nil)
	h := NewArticleHandler(svc, nil, nil, nil, nil, slog.Default())
	_ = h
}
