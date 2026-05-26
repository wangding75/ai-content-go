package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/metrics"
)

type MetricsHandler struct {
	service metrics.Service
	logger  *slog.Logger
}

func NewMetricsHandler(service metrics.Service, logger *slog.Logger) *MetricsHandler {
	return &MetricsHandler{service: service, logger: logger}
}

func (h *MetricsHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req metrics.CreateMetricTemplateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeMetricsError(w, r, metrics.ErrValidation, "invalid metric template request")
		return
	}
	data, err := h.service.CreateTemplate(r.Context(), req)
	if err != nil {
		writeMetricsError(w, r, err, "create metric template failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, data)
}

func (h *MetricsHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	req, err := parseListMetricTemplatesRequest(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid query parameters", nil)
		return
	}
	data, err := h.service.ListTemplates(r.Context(), req)
	if err != nil {
		writeMetricsError(w, r, err, "list metric templates failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *MetricsHandler) CreateRecord(w http.ResponseWriter, r *http.Request) {
	var req metrics.CreateMetricRecordRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeMetricsError(w, r, metrics.ErrValidation, "invalid metric record request")
		return
	}
	data, err := h.service.CreateRecord(r.Context(), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeMetricsError(w, r, err, "create metric record failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, data)
}

func (h *MetricsHandler) BatchCreateRecords(w http.ResponseWriter, r *http.Request) {
	var req metrics.BatchCreateMetricRecordsRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeMetricsError(w, r, metrics.ErrValidation, "invalid metric records batch request")
		return
	}
	data, err := h.service.BatchCreateRecords(r.Context(), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		if errors.Is(err, metrics.ErrValidation) && data.FailedCount > 0 {
			details := make([]api.ErrorDetail, 0, len(data.Errors))
			for _, e := range data.Errors {
				details = append(details, api.ErrorDetail{
					Field:  e.Field,
					Reason: e.Code + ": " + e.Message,
				})
			}
			api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "batch create metric records failed", details)
			return
		}
		writeMetricsError(w, r, err, "batch create metric records failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *MetricsHandler) ListRecords(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	req, err := parseListMetricRecordsRequest(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid query parameters", nil)
		return
	}
	data, err := h.service.ListRecords(r.Context(), req)
	if err != nil {
		writeMetricsError(w, r, err, "list metric records failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *MetricsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	req := metrics.MetricSummaryRequest{
		DateFrom:    r.URL.Query().Get("date_from"),
		DateTo:      r.URL.Query().Get("date_to"),
		Platform:    r.URL.Query().Get("platform"),
		TargetID:    r.URL.Query().Get("target_id"),
		MetricCodes: splitCSV(r.URL.Query().Get("metric_codes")),
	}
	data, err := h.service.GetSummary(r.Context(), chi.URLParam(r, "projectId"), req)
	if err != nil {
		writeMetricsError(w, r, err, "get metric summary failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *MetricsHandler) GetTrends(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	req := metrics.MetricTrendRequest{
		MetricCode: r.URL.Query().Get("metric_code"),
		DateFrom:   r.URL.Query().Get("date_from"),
		DateTo:     r.URL.Query().Get("date_to"),
		Bucket:     r.URL.Query().Get("bucket"),
		Platform:   r.URL.Query().Get("platform"),
		TargetID:   r.URL.Query().Get("target_id"),
	}
	data, err := h.service.GetTrends(r.Context(), chi.URLParam(r, "projectId"), req)
	if err != nil {
		writeMetricsError(w, r, err, "get metric trends failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *MetricsHandler) GetMissingDates(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	req := metrics.MissingMetricDatesRequest{
		MetricCode: r.URL.Query().Get("metric_code"),
		Platform:   r.URL.Query().Get("platform"),
		TargetID:   r.URL.Query().Get("target_id"),
		DateFrom:   r.URL.Query().Get("date_from"),
		DateTo:     r.URL.Query().Get("date_to"),
	}
	data, err := h.service.GetMissingDates(r.Context(), chi.URLParam(r, "projectId"), req)
	if err != nil {
		writeMetricsError(w, r, err, "get missing metric dates failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func parseListMetricTemplatesRequest(r *http.Request) (metrics.ListMetricTemplatesRequest, error) {
	pagination, err := parsePagination(r)
	if err != nil {
		return metrics.ListMetricTemplatesRequest{}, err
	}
	req := metrics.ListMetricTemplatesRequest{
		PaginationRequest: pagination,
		ContentType:       r.URL.Query().Get("content_type"),
		Platform:          r.URL.Query().Get("platform"),
	}
	if value := r.URL.Query().Get("enabled"); value != "" {
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return metrics.ListMetricTemplatesRequest{}, err
		}
		req.Enabled = &enabled
	}
	return req, nil
}

func parseListMetricRecordsRequest(r *http.Request) (metrics.ListMetricRecordsRequest, error) {
	pagination, err := parsePagination(r)
	if err != nil {
		return metrics.ListMetricRecordsRequest{}, err
	}
	return metrics.ListMetricRecordsRequest{
		PaginationRequest: pagination,
		ProjectID:         r.URL.Query().Get("project_id"),
		Platform:          r.URL.Query().Get("platform"),
		TargetID:          r.URL.Query().Get("target_id"),
		ContentItemID:     r.URL.Query().Get("content_item_id"),
		MetricCode:        r.URL.Query().Get("metric_code"),
		DateFrom:          r.URL.Query().Get("date_from"),
		DateTo:            r.URL.Query().Get("date_to"),
	}, nil
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func writeMetricsError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, metrics.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, metrics.ErrForbidden):
		api.WriteError(w, r, http.StatusForbidden, api.ErrorForbidden, message, nil)
	case errors.Is(err, metrics.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, metrics.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	case errors.Is(err, metrics.ErrIdempotencyConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorIdempotencyConflict, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}
