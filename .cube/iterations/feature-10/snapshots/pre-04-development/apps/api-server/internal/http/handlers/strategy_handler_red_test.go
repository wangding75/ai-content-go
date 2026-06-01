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
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/strategy"
)

type stubStrategyService struct {
	generateFn  func(ctx context.Context, projectID string, req strategy.GenerateSuggestionsRequest, idempotencyKey string) (strategy.GenerateSuggestionsResponse, error)
	listFn      func(ctx context.Context, projectID string, req strategy.ListStrategySuggestionsRequest) (strategy.PagedStrategySuggestionsResponse, error)
	getFn       func(ctx context.Context, suggestionID string) (strategy.StrategySuggestionDetailResponse, error)
	confirmFn   func(ctx context.Context, suggestionID string, req strategy.ConfirmSuggestionRequest, idempotencyKey string) (strategy.SuggestionStatusChangeResponse, error)
	ignoreFn    func(ctx context.Context, suggestionID string, req strategy.IgnoreSuggestionRequest, idempotencyKey string) (strategy.SuggestionStatusChangeResponse, error)
	executeFn   func(ctx context.Context, suggestionID string, req strategy.ExecuteSuggestionRequest, idempotencyKey string) (strategy.ExecuteSuggestionResponse, error)
	retryFn     func(ctx context.Context, suggestionID string, req strategy.RetrySuggestionRequest, idempotencyKey string) (strategy.ExecuteSuggestionResponse, error)
	listLogsFn  func(ctx context.Context, suggestionID string, req strategy.ListExecutionLogsRequest) (strategy.PagedExecutionLogsResponse, error)
}

func (s *stubStrategyService) GenerateSuggestions(ctx context.Context, projectID string, req strategy.GenerateSuggestionsRequest, idempotencyKey string) (strategy.GenerateSuggestionsResponse, error) {
	if s.generateFn != nil {
		return s.generateFn(ctx, projectID, req, idempotencyKey)
	}
	return strategy.GenerateSuggestionsResponse{SuggestionRunID: "run-1", Status: "generating"}, nil
}
func (s *stubStrategyService) ListSuggestions(ctx context.Context, projectID string, req strategy.ListStrategySuggestionsRequest) (strategy.PagedStrategySuggestionsResponse, error) {
	if s.listFn != nil {
		return s.listFn(ctx, projectID, req)
	}
	return strategy.PagedStrategySuggestionsResponse{Items: []strategy.StrategySuggestionItem{{ID: "sug-1", ProjectID: projectID, Status: "pending"}}, Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: 1}}, nil
}
func (s *stubStrategyService) GetSuggestion(ctx context.Context, suggestionID string) (strategy.StrategySuggestionDetailResponse, error) {
	if s.getFn != nil {
		return s.getFn(ctx, suggestionID)
	}
	return strategy.StrategySuggestionDetailResponse{ID: suggestionID, Status: "pending"}, nil
}
func (s *stubStrategyService) ConfirmSuggestion(ctx context.Context, suggestionID string, req strategy.ConfirmSuggestionRequest, idempotencyKey string) (strategy.SuggestionStatusChangeResponse, error) {
	if s.confirmFn != nil {
		return s.confirmFn(ctx, suggestionID, req, idempotencyKey)
	}
	return strategy.SuggestionStatusChangeResponse{SuggestionID: suggestionID, PreviousStatus: "pending", CurrentStatus: "confirmed", OperationLogID: "oplog-1"}, nil
}
func (s *stubStrategyService) IgnoreSuggestion(ctx context.Context, suggestionID string, req strategy.IgnoreSuggestionRequest, idempotencyKey string) (strategy.SuggestionStatusChangeResponse, error) {
	if s.ignoreFn != nil {
		return s.ignoreFn(ctx, suggestionID, req, idempotencyKey)
	}
	return strategy.SuggestionStatusChangeResponse{SuggestionID: suggestionID, PreviousStatus: "pending", CurrentStatus: "ignored", OperationLogID: "oplog-2"}, nil
}
func (s *stubStrategyService) ExecuteSuggestion(ctx context.Context, suggestionID string, req strategy.ExecuteSuggestionRequest, idempotencyKey string) (strategy.ExecuteSuggestionResponse, error) {
	if s.executeFn != nil {
		return s.executeFn(ctx, suggestionID, req, idempotencyKey)
	}
	return strategy.ExecuteSuggestionResponse{ExecutionLogID: "elog-1", SuggestionID: suggestionID, PreviousStatus: "confirmed", CurrentStatus: "executed", OperationLogID: "oplog-3"}, nil
}
func (s *stubStrategyService) RetrySuggestion(ctx context.Context, suggestionID string, req strategy.RetrySuggestionRequest, idempotencyKey string) (strategy.ExecuteSuggestionResponse, error) {
	if s.retryFn != nil {
		return s.retryFn(ctx, suggestionID, req, idempotencyKey)
	}
	return strategy.ExecuteSuggestionResponse{ExecutionLogID: "elog-2", SuggestionID: suggestionID, PreviousStatus: "execution_failed", CurrentStatus: "executed", OperationLogID: "oplog-4"}, nil
}
func (s *stubStrategyService) ListExecutionLogs(ctx context.Context, suggestionID string, req strategy.ListExecutionLogsRequest) (strategy.PagedExecutionLogsResponse, error) {
	if s.listLogsFn != nil {
		return s.listLogsFn(ctx, suggestionID, req)
	}
	return strategy.PagedExecutionLogsResponse{Items: []strategy.ExecutionLogResponse{{ID: "elog-1", SuggestionID: suggestionID}}, Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: 1}}, nil
}

// @Test
func TestStrategyGenerateSuggestionsReturns202(t *testing.T) {
	h := NewStrategyHandler(&stubStrategyService{}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"date_from":"2026-05-01","date_to":"2026-05-25"}`))
	req.Header.Set("Idempotency-Key", "idem-gen-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GenerateSuggestions(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body %s", rec.Code, rec.Body.String())
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if !envelope.Success || envelope.Data == nil {
		t.Fatalf("expected success envelope with data")
	}
}

// @Test
func TestStrategyListSuggestionsReturns200(t *testing.T) {
	h := NewStrategyHandler(&stubStrategyService{}, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/?page=1&page_size=20", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.ListSuggestions(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// @Test
func TestStrategyGetSuggestionReturns200(t *testing.T) {
	h := NewStrategyHandler(&stubStrategyService{}, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("suggestionId", "sug-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetSuggestion(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// @Test
func TestStrategyConfirmSuggestionReturns200(t *testing.T) {
	h := NewStrategyHandler(&stubStrategyService{}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"note":"ok"}`))
	req.Header.Set("Idempotency-Key", "idem-conf-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("suggestionId", "sug-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.ConfirmSuggestion(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
}

// @Test
func TestStrategyIgnoreSuggestionRequiresReason(t *testing.T) {
	h := NewStrategyHandler(&stubStrategyService{ignoreFn: func(_ context.Context, _ string, req strategy.IgnoreSuggestionRequest, _ string) (strategy.SuggestionStatusChangeResponse, error) {
		if req.Reason == "" {
			return strategy.SuggestionStatusChangeResponse{}, strategy.ErrValidation
		}
		return strategy.SuggestionStatusChangeResponse{}, nil
	}}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"reason":"","note":"x"}`))
	req.Header.Set("Idempotency-Key", "idem-ign-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("suggestionId", "sug-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.IgnoreSuggestion(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing reason, got %d", rec.Code)
	}
}

// @Test
func TestStrategyExecuteSuggestionReturns200(t *testing.T) {
	h := NewStrategyHandler(&stubStrategyService{}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"action_type":"adjust_schedule","target_type":"workflow_schedule","target_id":"sched-1"}`))
	req.Header.Set("Idempotency-Key", "idem-exec-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("suggestionId", "sug-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.ExecuteSuggestion(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
}

// @Test
func TestStrategyRetrySuggestionReturns200(t *testing.T) {
	h := NewStrategyHandler(&stubStrategyService{}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"operator_note":"retry"}`))
	req.Header.Set("Idempotency-Key", "idem-retry-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("suggestionId", "sug-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.RetrySuggestion(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
}

// @Test
func TestStrategyListExecutionLogsReturns200(t *testing.T) {
	h := NewStrategyHandler(&stubStrategyService{}, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/?page=1&page_size=20", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("suggestionId", "sug-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.ListExecutionLogs(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// @Test
func TestStrategyHandlerMapsConflictTo409(t *testing.T) {
	h := NewStrategyHandler(&stubStrategyService{confirmFn: func(_ context.Context, _ string, _ strategy.ConfirmSuggestionRequest, _ string) (strategy.SuggestionStatusChangeResponse, error) {
		return strategy.SuggestionStatusChangeResponse{}, strategy.ErrConflict
	}}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"note":"x"}`))
	req.Header.Set("Idempotency-Key", "idem-conflict")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("suggestionId", "sug-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.ConfirmSuggestion(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for conflict, got %d", rec.Code)
	}
}

// @Test
func TestStrategyHandlerMapsNotFoundTo404(t *testing.T) {
	h := NewStrategyHandler(&stubStrategyService{getFn: func(_ context.Context, _ string) (strategy.StrategySuggestionDetailResponse, error) {
		return strategy.StrategySuggestionDetailResponse{}, strategy.ErrNotFound
	}}, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("suggestionId", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetSuggestion(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for not found, got %d", rec.Code)
	}
}

// @Test
func TestStrategyCompileGuardUsesErrorsPackage(t *testing.T) {
	if !errors.Is(strategy.ErrConflict, strategy.ErrConflict) {
		t.Fatal("unreachable")
	}
	if !errors.Is(strategy.ErrIdempotencyConflict, strategy.ErrIdempotencyConflict) {
		t.Fatal("unreachable")
	}
}
