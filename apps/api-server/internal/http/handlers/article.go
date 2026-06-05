package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/article"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/metrics"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type ArticleHandler struct {
	svc        article.Service
	contentSvc content.Service
	workflow   workflow.Service
	metricsSvc metrics.Service
	submitter  engine.Submitter
	logger     *slog.Logger
}

func NewArticleHandler(svc article.Service, contentSvc content.Service, wfSvc workflow.Service, metricsSvc metrics.Service, submitter engine.Submitter, logger *slog.Logger) *ArticleHandler {
	return &ArticleHandler{svc: svc, contentSvc: contentSvc, workflow: wfSvc, metricsSvc: metricsSvc, submitter: submitter, logger: logger}
}

// mapArticleError maps domain errors to HTTP status codes.
func mapArticleError(err error) int {
	if errors.Is(err, article.ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, article.ErrForbidden) {
		return http.StatusForbidden
	}
	if errors.Is(err, article.ErrConflict) || errors.Is(err, article.ErrIdempotencyConflict) {
		return http.StatusConflict
	}
	if errors.Is(err, article.ErrValidation) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func (h *ArticleHandler) RegisterPack(w http.ResponseWriter, r *http.Request) {
	var req article.RegisterPackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid JSON body", []api.ErrorDetail{{Reason: err.Error()}})
		return
	}
	idempotencyKey := r.Header.Get("Idempotency-Key")

	resp, err := h.svc.RegisterPack(r.Context(), req, idempotencyKey)
	if err != nil {
		code := mapArticleError(err)
		errCode := api.ErrorInternal
		if errors.Is(err, article.ErrIdempotencyConflict) {
			errCode = api.ErrorIdempotencyConflict
		} else if errors.Is(err, article.ErrValidation) {
			errCode = api.ErrorValidation
		}
		api.WriteError(w, r, code, errCode, err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, resp)
}

func (h *ArticleHandler) GetPackStatus(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetPackStatus(r.Context())
	if err != nil {
		api.WriteError(w, r, mapArticleError(err), api.ErrorInternal, err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *ArticleHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId", nil)
		return
	}

	resp, err := h.svc.GetConfig(r.Context(), projectID)
	if err != nil {
		api.WriteError(w, r, mapArticleError(err), api.ErrorInternal, err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *ArticleHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId", nil)
		return
	}

	var req article.UpdateArticleConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid JSON body", []api.ErrorDetail{{Reason: err.Error()}})
		return
	}
	idempotencyKey := r.Header.Get("Idempotency-Key")

	resp, err := h.svc.UpdateConfig(r.Context(), projectID, req, idempotencyKey)
	if err != nil {
		code := mapArticleError(err)
		errCode := api.ErrorInternal
		if errors.Is(err, article.ErrValidation) {
			errCode = api.ErrorValidation
		} else if errors.Is(err, article.ErrIdempotencyConflict) {
			errCode = api.ErrorIdempotencyConflict
		}
		api.WriteError(w, r, code, errCode, err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *ArticleHandler) CreateGenerationRun(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId", nil)
		return
	}

	var req article.CreateArticleGenerationRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid JSON body", []api.ErrorDetail{{Reason: err.Error()}})
		return
	}
	idempotencyKey := r.Header.Get("Idempotency-Key")

	workflowRunID := "" // Let the service generate one if needed
	resp, err := h.svc.CreateGenerationRun(r.Context(), projectID, req, workflowRunID, idempotencyKey)
	if err != nil {
		code := mapArticleError(err)
		errCode := api.ErrorInternal
		if errors.Is(err, article.ErrValidation) {
			errCode = api.ErrorValidation
		}
		api.WriteError(w, r, code, errCode, err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusAccepted, resp)
}

func (h *ArticleHandler) ListGenerationRuns(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId", nil)
		return
	}

	page, pageSize := 1, 20
	if p := r.URL.Query().Get("page"); p != "" {
		parsed, err := parsePage(p)
		if err != nil {
			api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid page parameter", nil)
			return
		}
		page = parsed
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		parsed, err := parsePage(ps)
		if err != nil {
			api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid page_size parameter", nil)
			return
		}
		if parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	req := article.ListGenerationRunsRequest{
		PaginationRequest: content.PaginationRequest{
			Page:     page,
			PageSize: pageSize,
		},
		Status:         r.URL.Query().Get("status"),
		Topic:          r.URL.Query().Get("topic"),
		TargetPlatform: r.URL.Query().Get("target_platform"),
	}

	resp, err := h.svc.ListGenerationRuns(r.Context(), projectID, req)
	if err != nil {
		api.WriteError(w, r, mapArticleError(err), api.ErrorInternal, err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *ArticleHandler) GetGenerationRun(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	id := chi.URLParam(r, "id")
	if projectID == "" || id == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId or id", nil)
		return
	}

	resp, err := h.svc.GetGenerationRun(r.Context(), projectID, id)
	if err != nil {
		api.WriteError(w, r, mapArticleError(err), api.ErrorInternal, err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *ArticleHandler) RetryGenerationRun(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	id := chi.URLParam(r, "id")
	if projectID == "" || id == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId or id", nil)
		return
	}

	var req article.RetryGenerationRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid JSON body", []api.ErrorDetail{{Reason: err.Error()}})
		return
	}
	idempotencyKey := r.Header.Get("Idempotency-Key")

	resp, err := h.svc.RetryGenerationRun(r.Context(), projectID, id, req, "", idempotencyKey)
	if err != nil {
		code := mapArticleError(err)
		errCode := api.ErrorInternal
		if errors.Is(err, article.ErrConflict) || errors.Is(err, article.ErrIdempotencyConflict) {
			errCode = api.ErrorConflict
		}
		if errors.Is(err, article.ErrValidation) {
			errCode = api.ErrorValidation
		}
		api.WriteError(w, r, code, errCode, err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *ArticleHandler) GetContentSnapshot(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	if itemID == "" {
		// Try "id" as fallback
		itemID = chi.URLParam(r, "id")
	}
	if itemID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing itemId", nil)
		return
	}

	resp, err := h.svc.GetContentSnapshot(r.Context(), itemID)
	if err != nil {
		api.WriteError(w, r, mapArticleError(err), api.ErrorInternal, err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *ArticleHandler) GetMetricsConfig(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId", nil)
		return
	}

	resp, err := h.svc.GetProjectArticleMetrics(r.Context(), projectID)
	if err != nil {
		api.WriteError(w, r, mapArticleError(err), api.ErrorInternal, err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *ArticleHandler) UpdateMetricsConfig(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId", nil)
		return
	}

	var req article.UpdateProjectArticleMetricsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid JSON body", []api.ErrorDetail{{Reason: err.Error()}})
		return
	}
	idempotencyKey := r.Header.Get("Idempotency-Key")

	resp, err := h.svc.UpdateProjectArticleMetrics(r.Context(), projectID, req, idempotencyKey)
	if err != nil {
		code := mapArticleError(err)
		errCode := api.ErrorInternal
		if errors.Is(err, article.ErrValidation) {
			errCode = api.ErrorValidation
		}
		api.WriteError(w, r, code, errCode, err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

// parsePage is a helper to validate integer page numbers.
func parsePage(s string) (int, error) {
	var v int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("invalid page")
		}
		v = v*10 + int(c-'0')
	}
	return v, nil
}
