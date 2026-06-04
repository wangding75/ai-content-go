package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/strategy"
)

type StrategyHandler struct {
	service strategy.Service
	logger  *slog.Logger
}

func NewStrategyHandler(service strategy.Service, logger *slog.Logger) *StrategyHandler {
	return &StrategyHandler{service: service, logger: logger}
}

func (h *StrategyHandler) GenerateSuggestions(w http.ResponseWriter, r *http.Request) {
	var req strategy.GenerateSuggestionsRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeStrategyError(w, r, strategy.ErrValidation, "invalid generate suggestions request")
		return
	}
	projectID := chi.URLParam(r, "projectId")
	data, err := h.service.GenerateSuggestions(r.Context(), projectID, req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeStrategyError(w, r, err, "generate strategy suggestions failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusAccepted, data)
}

func (h *StrategyHandler) ListSuggestions(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	projectID := chi.URLParam(r, "projectId")
	pagination, err := parsePagination(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid query parameters", nil)
		return
	}
	req := strategy.ListStrategySuggestionsRequest{
		PaginationRequest: pagination,
		ProjectID:         projectID,
		Status:            r.URL.Query().Get("status"),
		SuggestionType:    r.URL.Query().Get("suggestion_type"),
		RiskLevel:         r.URL.Query().Get("risk_level"),
		Confidence:        r.URL.Query().Get("confidence"),
		DateFrom:          r.URL.Query().Get("date_from"),
		DateTo:            r.URL.Query().Get("date_to"),
		Sort:              r.URL.Query().Get("sort"),
		Order:             r.URL.Query().Get("order"),
	}
	data, err := h.service.ListSuggestions(r.Context(), projectID, req)
	if err != nil {
		writeStrategyError(w, r, err, "list strategy suggestions failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *StrategyHandler) GetSuggestion(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	suggestionID := chi.URLParam(r, "suggestionId")
	data, err := h.service.GetSuggestion(r.Context(), suggestionID)
	if err != nil {
		writeStrategyError(w, r, err, "get strategy suggestion failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *StrategyHandler) ConfirmSuggestion(w http.ResponseWriter, r *http.Request) {
	var req strategy.ConfirmSuggestionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeStrategyError(w, r, strategy.ErrValidation, "invalid confirm request")
		return
	}
	suggestionID := chi.URLParam(r, "suggestionId")
	data, err := h.service.ConfirmSuggestion(r.Context(), suggestionID, req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeStrategyError(w, r, err, "confirm strategy suggestion failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *StrategyHandler) IgnoreSuggestion(w http.ResponseWriter, r *http.Request) {
	var req strategy.IgnoreSuggestionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeStrategyError(w, r, strategy.ErrValidation, "invalid ignore request")
		return
	}
	suggestionID := chi.URLParam(r, "suggestionId")
	data, err := h.service.IgnoreSuggestion(r.Context(), suggestionID, req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeStrategyError(w, r, err, "ignore strategy suggestion failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *StrategyHandler) ExecuteSuggestion(w http.ResponseWriter, r *http.Request) {
	var req strategy.ExecuteSuggestionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeStrategyError(w, r, strategy.ErrValidation, "invalid execute request")
		return
	}
	suggestionID := chi.URLParam(r, "suggestionId")
	data, err := h.service.ExecuteSuggestion(r.Context(), suggestionID, req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeStrategyError(w, r, err, "execute strategy suggestion failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *StrategyHandler) RetrySuggestion(w http.ResponseWriter, r *http.Request) {
	var req strategy.RetrySuggestionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeStrategyError(w, r, strategy.ErrValidation, "invalid retry request")
		return
	}
	suggestionID := chi.URLParam(r, "suggestionId")
	data, err := h.service.RetrySuggestion(r.Context(), suggestionID, req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeStrategyError(w, r, err, "retry strategy suggestion failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *StrategyHandler) ListExecutionLogs(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	suggestionID := chi.URLParam(r, "suggestionId")
	pagination, err := parsePagination(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid query parameters", nil)
		return
	}
	req := strategy.ListExecutionLogsRequest{
		PaginationRequest: pagination,
		SuggestionID:      suggestionID,
	}
	data, err := h.service.ListExecutionLogs(r.Context(), suggestionID, req)
	if err != nil {
		writeStrategyError(w, r, err, "list execution logs failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func writeStrategyError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, strategy.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, strategy.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, strategy.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	case errors.Is(err, strategy.ErrIdempotencyConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorIdempotencyConflict, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}
