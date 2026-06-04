package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/portfolio"
)

type stubPortfolioService struct {
	createFn      func(ctx context.Context, req portfolio.CreatePortfolioRequest, idempotencyKey string) (portfolio.PortfolioDetailResponse, error)
	listFn        func(ctx context.Context, req portfolio.ListPortfoliosRequest) (portfolio.PagedPortfoliosResponse, error)
	getFn         func(ctx context.Context, portfolioID string) (portfolio.PortfolioDetailResponse, error)
	updateFn      func(ctx context.Context, portfolioID string, req portfolio.UpdatePortfolioRequest, idempotencyKey string) (portfolio.PortfolioDetailResponse, error)
	addProjectFn  func(ctx context.Context, portfolioID string, req portfolio.AddPortfolioProjectRequest, idempotencyKey string) (portfolio.PortfolioProjectResponse, error)
	listProjectsFn func(ctx context.Context, portfolioID string, req portfolio.ListPortfolioProjectsRequest) (portfolio.PagedPortfolioProjectsResponse, error)
	updatePriorityFn func(ctx context.Context, portfolioID string, projectID string, req portfolio.UpdatePortfolioProjectPriorityRequest, idempotencyKey string) (portfolio.PortfolioProjectResponse, error)
	removeProjectFn  func(ctx context.Context, portfolioID string, projectID string, req portfolio.RemovePortfolioProjectRequest) (portfolio.RemovePortfolioProjectResponse, error)
	recalcFn      func(ctx context.Context, portfolioID string, req portfolio.RecalculatePortfolioStatusSnapshotRequest, idempotencyKey string) (portfolio.RecalculatePortfolioStatusSnapshotResponse, error)
	listSnapshotsFn  func(ctx context.Context, portfolioID string, req portfolio.ListPortfolioStatusSnapshotsRequest) (portfolio.PagedPortfolioStatusSnapshotsResponse, error)
	healthFn      func(ctx context.Context, portfolioID string, req portfolio.PortfolioSummaryRequest) (portfolio.PortfolioHealthSummaryResponse, error)
	costFn        func(ctx context.Context, portfolioID string, req portfolio.PortfolioSummaryRequest) (portfolio.PortfolioCostSummaryResponse, error)
	strategyFn    func(ctx context.Context, portfolioID string, req portfolio.PortfolioSummaryRequest) (portfolio.PortfolioStrategySummaryResponse, error)
}

func (s *stubPortfolioService) CreatePortfolio(ctx context.Context, req portfolio.CreatePortfolioRequest, idempotencyKey string) (portfolio.PortfolioDetailResponse, error) {
	if s.createFn != nil { return s.createFn(ctx, req, idempotencyKey) }
	return portfolio.PortfolioDetailResponse{ID: "pf-1", Name: req.Name, Status: req.Status}, nil
}
func (s *stubPortfolioService) ListPortfolios(ctx context.Context, req portfolio.ListPortfoliosRequest) (portfolio.PagedPortfoliosResponse, error) {
	if s.listFn != nil { return s.listFn(ctx, req) }
	return portfolio.PagedPortfoliosResponse{Items: []portfolio.PortfolioListItem{{ID: "pf-1"}}, Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: 1}}, nil
}
func (s *stubPortfolioService) GetPortfolio(ctx context.Context, portfolioID string) (portfolio.PortfolioDetailResponse, error) {
	if s.getFn != nil { return s.getFn(ctx, portfolioID) }
	return portfolio.PortfolioDetailResponse{ID: portfolioID}, nil
}
func (s *stubPortfolioService) UpdatePortfolio(ctx context.Context, portfolioID string, req portfolio.UpdatePortfolioRequest, idempotencyKey string) (portfolio.PortfolioDetailResponse, error) {
	if s.updateFn != nil { return s.updateFn(ctx, portfolioID, req, idempotencyKey) }
	return portfolio.PortfolioDetailResponse{ID: portfolioID}, nil
}
func (s *stubPortfolioService) AddProject(ctx context.Context, portfolioID string, req portfolio.AddPortfolioProjectRequest, idempotencyKey string) (portfolio.PortfolioProjectResponse, error) {
	if s.addProjectFn != nil { return s.addProjectFn(ctx, portfolioID, req, idempotencyKey) }
	return portfolio.PortfolioProjectResponse{PortfolioID: portfolioID, ProjectID: req.ProjectID}, nil
}
func (s *stubPortfolioService) ListProjects(ctx context.Context, portfolioID string, req portfolio.ListPortfolioProjectsRequest) (portfolio.PagedPortfolioProjectsResponse, error) {
	if s.listProjectsFn != nil { return s.listProjectsFn(ctx, portfolioID, req) }
	return portfolio.PagedPortfolioProjectsResponse{Items: []portfolio.PortfolioProjectResponse{}, Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: 0}}, nil
}
func (s *stubPortfolioService) UpdateProjectPriority(ctx context.Context, portfolioID string, projectID string, req portfolio.UpdatePortfolioProjectPriorityRequest, idempotencyKey string) (portfolio.PortfolioProjectResponse, error) {
	if s.updatePriorityFn != nil { return s.updatePriorityFn(ctx, portfolioID, projectID, req, idempotencyKey) }
	return portfolio.PortfolioProjectResponse{PortfolioID: portfolioID, ProjectID: projectID}, nil
}
func (s *stubPortfolioService) RemoveProject(ctx context.Context, portfolioID string, projectID string, req portfolio.RemovePortfolioProjectRequest) (portfolio.RemovePortfolioProjectResponse, error) {
	if s.removeProjectFn != nil { return s.removeProjectFn(ctx, portfolioID, projectID, req) }
	return portfolio.RemovePortfolioProjectResponse{PortfolioID: portfolioID, ProjectID: projectID, Removed: true, OperationID: "op-1"}, nil
}
func (s *stubPortfolioService) RecalculateStatusSnapshot(ctx context.Context, portfolioID string, req portfolio.RecalculatePortfolioStatusSnapshotRequest, idempotencyKey string) (portfolio.RecalculatePortfolioStatusSnapshotResponse, error) {
	if s.recalcFn != nil { return s.recalcFn(ctx, portfolioID, req, idempotencyKey) }
	return portfolio.RecalculatePortfolioStatusSnapshotResponse{PortfolioID: portfolioID, SnapshotID: "pss-1", JobID: "pss-1", CalculationStatus: "queued"}, nil
}
func (s *stubPortfolioService) ListStatusSnapshots(ctx context.Context, portfolioID string, req portfolio.ListPortfolioStatusSnapshotsRequest) (portfolio.PagedPortfolioStatusSnapshotsResponse, error) {
	if s.listSnapshotsFn != nil { return s.listSnapshotsFn(ctx, portfolioID, req) }
	return portfolio.PagedPortfolioStatusSnapshotsResponse{Items: []portfolio.PortfolioStatusSnapshotResponse{}, Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: 0}}, nil
}
func (s *stubPortfolioService) GetHealthSummary(ctx context.Context, portfolioID string, req portfolio.PortfolioSummaryRequest) (portfolio.PortfolioHealthSummaryResponse, error) {
	if s.healthFn != nil { return s.healthFn(ctx, portfolioID, req) }
	return portfolio.PortfolioHealthSummaryResponse{PortfolioID: portfolioID}, nil
}
func (s *stubPortfolioService) GetCostSummary(ctx context.Context, portfolioID string, req portfolio.PortfolioSummaryRequest) (portfolio.PortfolioCostSummaryResponse, error) {
	if s.costFn != nil { return s.costFn(ctx, portfolioID, req) }
	return portfolio.PortfolioCostSummaryResponse{PortfolioID: portfolioID}, nil
}
func (s *stubPortfolioService) GetStrategySummary(ctx context.Context, portfolioID string, req portfolio.PortfolioSummaryRequest) (portfolio.PortfolioStrategySummaryResponse, error) {
	if s.strategyFn != nil { return s.strategyFn(ctx, portfolioID, req) }
	return portfolio.PortfolioStrategySummaryResponse{PortfolioID: portfolioID}, nil
}

// @Test
func TestTask03CreatePortfolioReturns201(t *testing.T) {
	h := NewPortfolioHandler(&stubPortfolioService{}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"test","scope_type":"manual","status":"active"}`))
	req.Header.Set("Idempotency-Key", "idem-pf-1")
	rec := httptest.NewRecorder()
	h.CreatePortfolio(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body %s", rec.Code, rec.Body.String())
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if !envelope.Success || envelope.Data == nil {
		t.Fatalf("expected success envelope with data")
	}
}

// @Test
func TestTask03RecalculateSnapshotReturns202(t *testing.T) {
	h := NewPortfolioHandler(&stubPortfolioService{}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"date_range":{"start":"2026-05-01","end":"2026-05-31"},"force":true}`))
	req.Header.Set("Idempotency-Key", "idem-recalc-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("portfolioId", "pf-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.RecalculateStatusSnapshot(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for recalculate, got %d body %s", rec.Code, rec.Body.String())
	}
}

// @Test
func TestTask03RemoveProjectReturns200(t *testing.T) {
	h := NewPortfolioHandler(&stubPortfolioService{}, slog.Default())
	req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(`{"reason":"scope-update","note":"removed"}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("portfolioId", "pf-1")
	rctx.URLParams.Add("projectId", "proj-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.RemoveProject(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for remove, got %d body %s", rec.Code, rec.Body.String())
	}
}

// @Test
func TestTask03ListPortfoliosReturns200(t *testing.T) {
	h := NewPortfolioHandler(&stubPortfolioService{}, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/?page=1&page_size=20", nil)
	rec := httptest.NewRecorder()
	h.ListPortfolios(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// @Test
func TestTask03GetPortfolioReturns200(t *testing.T) {
	h := NewPortfolioHandler(&stubPortfolioService{}, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("portfolioId", "pf-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetPortfolio(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// @Test
func TestTask03AddProjectReturns201(t *testing.T) {
	h := NewPortfolioHandler(&stubPortfolioService{}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"project_id":"proj-1","priority":1,"weight":1}`))
	req.Header.Set("Idempotency-Key", "idem-add-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("portfolioId", "pf-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.AddProject(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body %s", rec.Code, rec.Body.String())
	}
}

// @Test
func TestTask03HandlerMapsValidationTo400(t *testing.T) {
	h := NewPortfolioHandler(&stubPortfolioService{createFn: func(_ context.Context, _ portfolio.CreatePortfolioRequest, _ string) (portfolio.PortfolioDetailResponse, error) {
		return portfolio.PortfolioDetailResponse{}, portfolio.ErrValidation
	}}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"bad"}`))
	rec := httptest.NewRecorder()
	h.CreatePortfolio(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for validation, got %d", rec.Code)
	}
}

// @Test
func TestTask03HandlerMapsNotFoundTo404(t *testing.T) {
	h := NewPortfolioHandler(&stubPortfolioService{getFn: func(_ context.Context, _ string) (portfolio.PortfolioDetailResponse, error) {
		return portfolio.PortfolioDetailResponse{}, portfolio.ErrNotFound
	}}, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("portfolioId", "missing")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetPortfolio(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for not found, got %d", rec.Code)
	}
}

// @Test
func TestTask03HandlerMapsForbiddenTo403(t *testing.T) {
	h := NewPortfolioHandler(&stubPortfolioService{updateFn: func(_ context.Context, _ string, _ portfolio.UpdatePortfolioRequest, _ string) (portfolio.PortfolioDetailResponse, error) {
		return portfolio.PortfolioDetailResponse{}, portfolio.ErrForbidden
	}}, slog.Default())
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"name":"unauth"}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("portfolioId", "pf-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.UpdatePortfolio(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for forbidden, got %d", rec.Code)
	}
}

// @Test
func TestTask03HandlerMapsIdempotencyConflictTo409(t *testing.T) {
	h := NewPortfolioHandler(&stubPortfolioService{createFn: func(_ context.Context, _ portfolio.CreatePortfolioRequest, _ string) (portfolio.PortfolioDetailResponse, error) {
		return portfolio.PortfolioDetailResponse{}, portfolio.ErrIdempotencyConflict
	}}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"dup","scope_type":"manual","status":"active"}`))
	req.Header.Set("Idempotency-Key", "idem-dup-1")
	rec := httptest.NewRecorder()
	h.CreatePortfolio(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for idempotency conflict, got %d", rec.Code)
	}
}

// @Test
func TestTask03HandlerMapsConflictTo409(t *testing.T) {
	h := NewPortfolioHandler(&stubPortfolioService{addProjectFn: func(_ context.Context, _ string, _ portfolio.AddPortfolioProjectRequest, _ string) (portfolio.PortfolioProjectResponse, error) {
		return portfolio.PortfolioProjectResponse{}, portfolio.ErrConflict
	}}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"project_id":"dup","priority":1,"weight":1}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("portfolioId", "pf-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.AddProject(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for conflict, got %d", rec.Code)
	}
}
