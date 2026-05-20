package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/memory"
)

type MemoryHandler struct {
	service memory.Service
	logger  *slog.Logger
}

func NewMemoryHandler(service memory.Service, logger *slog.Logger) *MemoryHandler {
	return &MemoryHandler{service: service, logger: logger}
}

func (h *MemoryHandler) GetKnowledgeMemory(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetKnowledgeMemory(r.Context(), chi.URLParam(r, "projectId"))
	if err != nil {
		writeMemoryError(w, r, err, "get knowledge memory failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *MemoryHandler) UpdateStaticContext(w http.ResponseWriter, r *http.Request) {
	var req memory.UpdateStaticContextRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeMemoryError(w, r, memory.ErrValidation, "invalid static context request")
		return
	}
	result, err := h.service.UpdateStaticContext(r.Context(), chi.URLParam(r, "projectId"), req)
	if err != nil {
		writeMemoryError(w, r, err, "update static context failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *MemoryHandler) UpdateStyleGuide(w http.ResponseWriter, r *http.Request) {
	var req memory.UpdateStyleGuideRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeMemoryError(w, r, memory.ErrValidation, "invalid style guide request")
		return
	}
	result, err := h.service.UpdateStyleGuide(r.Context(), chi.URLParam(r, "projectId"), req)
	if err != nil {
		writeMemoryError(w, r, err, "update style guide failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *MemoryHandler) CorrectDynamicState(w http.ResponseWriter, r *http.Request) {
	var req memory.CorrectDynamicStateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeMemoryError(w, r, memory.ErrValidation, "invalid dynamic state correction request")
		return
	}
	result, err := h.service.CorrectDynamicState(r.Context(), chi.URLParam(r, "projectId"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeMemoryError(w, r, err, "correct dynamic state failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *MemoryHandler) UpdateRecentWindowPolicy(w http.ResponseWriter, r *http.Request) {
	var req memory.UpdateRecentWindowPolicyRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeMemoryError(w, r, memory.ErrValidation, "invalid recent window policy request")
		return
	}
	result, err := h.service.UpdateRecentWindowPolicy(r.Context(), chi.URLParam(r, "projectId"), req)
	if err != nil {
		writeMemoryError(w, r, err, "update recent window policy failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *MemoryHandler) ListSnapshots(w http.ResponseWriter, r *http.Request) {
	page, ok := queryInt(w, r, "page")
	if !ok {
		return
	}
	pageSize, ok := queryInt(w, r, "page_size")
	if !ok {
		return
	}
	result, err := h.service.ListSnapshots(r.Context(), chi.URLParam(r, "projectId"), memory.ListSnapshotsRequest{ContentItemID: r.URL.Query().Get("content_item_id"), Page: page, PageSize: pageSize, Sort: r.URL.Query().Get("sort"), Order: r.URL.Query().Get("order")})
	if err != nil {
		writeMemoryError(w, r, err, "list memory snapshots failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *MemoryHandler) PreviewContext(w http.ResponseWriter, r *http.Request) {
	budget, ok := queryInt(w, r, "budget")
	if !ok {
		return
	}
	result, err := h.service.PreviewContext(r.Context(), chi.URLParam(r, "projectId"), memory.ContextPreviewRequest{Purpose: r.URL.Query().Get("purpose"), Budget: budget, ContentItemID: r.URL.Query().Get("content_item_id")})
	if err != nil {
		writeMemoryError(w, r, err, "preview context failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *MemoryHandler) AssembleContext(w http.ResponseWriter, r *http.Request) {
	var req memory.AssembleContextRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeMemoryError(w, r, memory.ErrValidation, "invalid assemble context request")
		return
	}
	result, err := h.service.AssembleContext(r.Context(), chi.URLParam(r, "projectId"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeMemoryError(w, r, err, "assemble context failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *MemoryHandler) UpdateDynamicState(w http.ResponseWriter, r *http.Request) {
	var req memory.UpdateDynamicStateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeMemoryError(w, r, memory.ErrValidation, "invalid dynamic state request")
		return
	}
	result, err := h.service.UpdateDynamicState(r.Context(), chi.URLParam(r, "id"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeMemoryError(w, r, err, "update dynamic state failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *MemoryHandler) CreateConsistencyReport(w http.ResponseWriter, r *http.Request) {
	var req memory.CreateConsistencyReportRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeMemoryError(w, r, memory.ErrValidation, "invalid consistency report request")
		return
	}
	result, err := h.service.CreateConsistencyReport(r.Context(), chi.URLParam(r, "projectId"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeMemoryError(w, r, err, "create consistency report failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusAccepted, result)
}

func (h *MemoryHandler) ListConsistencyReports(w http.ResponseWriter, r *http.Request) {
	page, ok := queryInt(w, r, "page")
	if !ok {
		return
	}
	pageSize, ok := queryInt(w, r, "page_size")
	if !ok {
		return
	}
	result, err := h.service.ListConsistencyReports(r.Context(), chi.URLParam(r, "projectId"), memory.ListConsistencyReportsRequest{Status: r.URL.Query().Get("status"), Page: page, PageSize: pageSize, Sort: r.URL.Query().Get("sort"), Order: r.URL.Query().Get("order")})
	if err != nil {
		writeMemoryError(w, r, err, "list consistency reports failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *MemoryHandler) GetConsistencyReport(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetConsistencyReport(r.Context(), chi.URLParam(r, "projectId"), chi.URLParam(r, "reportId"))
	if err != nil {
		writeMemoryError(w, r, err, "get consistency report failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func queryInt(w http.ResponseWriter, r *http.Request, name string) (int, bool) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return 0, true
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		writeMemoryError(w, r, memory.ErrValidation, "invalid query parameter")
		return 0, false
	}
	return parsed, true
}

func writeMemoryError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, memory.ErrForbidden):
		api.WriteError(w, r, http.StatusForbidden, api.ErrorForbidden, message, nil)
	case errors.Is(err, memory.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, memory.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, memory.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	case errors.Is(err, memory.ErrIdempotencyConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorIdempotencyConflict, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}
