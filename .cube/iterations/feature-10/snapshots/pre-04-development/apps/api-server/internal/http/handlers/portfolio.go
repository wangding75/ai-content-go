package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/portfolio"
)

type PortfolioHandler struct {
	service portfolio.Service
	logger  *slog.Logger
}

func NewPortfolioHandler(service portfolio.Service, logger *slog.Logger) *PortfolioHandler {
	return &PortfolioHandler{service: service, logger: logger}
}

func (h *PortfolioHandler) CreatePortfolio(w http.ResponseWriter, r *http.Request) {
	var req portfolio.CreatePortfolioRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writePortfolioError(w, r, portfolio.ErrValidation, "invalid portfolio request")
		return
	}
	data, err := h.service.CreatePortfolio(r.Context(), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writePortfolioError(w, r, err, "create portfolio failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, data)
}

func (h *PortfolioHandler) ListPortfolios(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	pagination, err := parsePagination(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid query parameters", nil)
		return
	}
	data, err := h.service.ListPortfolios(r.Context(), portfolio.ListPortfoliosRequest{PaginationRequest: pagination, Q: r.URL.Query().Get("q"), Status: r.URL.Query().Get("status"), ScopeType: r.URL.Query().Get("scope_type"), OwnerID: r.URL.Query().Get("owner_id")})
	if err != nil {
		writePortfolioError(w, r, err, "list portfolios failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PortfolioHandler) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	data, err := h.service.GetPortfolio(r.Context(), chi.URLParam(r, "portfolioId"))
	if err != nil {
		writePortfolioError(w, r, err, "get portfolio failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PortfolioHandler) UpdatePortfolio(w http.ResponseWriter, r *http.Request) {
	var req portfolio.UpdatePortfolioRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writePortfolioError(w, r, portfolio.ErrValidation, "invalid portfolio update request")
		return
	}
	data, err := h.service.UpdatePortfolio(r.Context(), chi.URLParam(r, "portfolioId"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writePortfolioError(w, r, err, "update portfolio failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PortfolioHandler) AddProject(w http.ResponseWriter, r *http.Request) {
	var req portfolio.AddPortfolioProjectRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writePortfolioError(w, r, portfolio.ErrValidation, "invalid portfolio project request")
		return
	}
	data, err := h.service.AddProject(r.Context(), chi.URLParam(r, "portfolioId"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writePortfolioError(w, r, err, "add portfolio project failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, data)
}

func (h *PortfolioHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	pagination, err := parsePagination(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid query parameters", nil)
		return
	}
	data, err := h.service.ListProjects(r.Context(), chi.URLParam(r, "portfolioId"), portfolio.ListPortfolioProjectsRequest{PaginationRequest: pagination, Role: r.URL.Query().Get("role")})
	if err != nil {
		writePortfolioError(w, r, err, "list portfolio projects failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PortfolioHandler) UpdateProjectPriority(w http.ResponseWriter, r *http.Request) {
	var req portfolio.UpdatePortfolioProjectPriorityRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writePortfolioError(w, r, portfolio.ErrValidation, "invalid portfolio project priority request")
		return
	}
	data, err := h.service.UpdateProjectPriority(r.Context(), chi.URLParam(r, "portfolioId"), chi.URLParam(r, "projectId"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writePortfolioError(w, r, err, "update portfolio project priority failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PortfolioHandler) RemoveProject(w http.ResponseWriter, r *http.Request) {
	var req portfolio.RemovePortfolioProjectRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writePortfolioError(w, r, portfolio.ErrValidation, "invalid portfolio project removal request")
		return
	}
	data, err := h.service.RemoveProject(r.Context(), chi.URLParam(r, "portfolioId"), chi.URLParam(r, "projectId"), req)
	if err != nil {
		writePortfolioError(w, r, err, "remove portfolio project failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PortfolioHandler) RecalculateStatusSnapshot(w http.ResponseWriter, r *http.Request) {
	var req portfolio.RecalculatePortfolioStatusSnapshotRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writePortfolioError(w, r, portfolio.ErrValidation, "invalid status snapshot request")
		return
	}
	data, err := h.service.RecalculateStatusSnapshot(r.Context(), chi.URLParam(r, "portfolioId"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writePortfolioError(w, r, err, "recalculate portfolio status snapshot failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusAccepted, data)
}

func (h *PortfolioHandler) ListStatusSnapshots(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	pagination, err := parsePagination(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid query parameters", nil)
		return
	}
	data, err := h.service.ListStatusSnapshots(r.Context(), chi.URLParam(r, "portfolioId"), portfolio.ListPortfolioStatusSnapshotsRequest{PaginationRequest: pagination, DateRange: dateRangeFromQuery(r)})
	if err != nil {
		writePortfolioError(w, r, err, "list portfolio status snapshots failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PortfolioHandler) GetHealthSummary(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	data, err := h.service.GetHealthSummary(r.Context(), chi.URLParam(r, "portfolioId"), portfolio.PortfolioSummaryRequest{DateRange: dateRangeFromQuery(r)})
	if err != nil {
		writePortfolioError(w, r, err, "get portfolio health summary failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PortfolioHandler) GetCostSummary(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	data, err := h.service.GetCostSummary(r.Context(), chi.URLParam(r, "portfolioId"), portfolio.PortfolioSummaryRequest{DateRange: dateRangeFromQuery(r)})
	if err != nil {
		writePortfolioError(w, r, err, "get portfolio cost summary failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PortfolioHandler) GetStrategySummary(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	data, err := h.service.GetStrategySummary(r.Context(), chi.URLParam(r, "portfolioId"), portfolio.PortfolioSummaryRequest{DateRange: dateRangeFromQuery(r)})
	if err != nil {
		writePortfolioError(w, r, err, "get portfolio strategy summary failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func dateRangeFromQuery(r *http.Request) portfolio.DateRange {
	return portfolio.DateRange{Start: r.URL.Query().Get("date_from"), End: r.URL.Query().Get("date_to")}
}

func writePortfolioError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, portfolio.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, portfolio.ErrForbidden):
		api.WriteError(w, r, http.StatusForbidden, api.ErrorForbidden, message, nil)
	case errors.Is(err, portfolio.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, portfolio.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	case errors.Is(err, portfolio.ErrIdempotencyConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorIdempotencyConflict, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}
