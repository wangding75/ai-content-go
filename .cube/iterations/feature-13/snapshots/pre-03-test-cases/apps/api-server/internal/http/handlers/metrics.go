package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/metrics"
)

type metricCoreService interface {
	CreateTemplate(ctx context.Context, req metrics.CreateMetricTemplateRequest) (metrics.CreateMetricTemplateResponse, error)
	ListTemplates(ctx context.Context, req metrics.ListMetricTemplatesRequest) (metrics.PagedMetricTemplatesResponse, error)
	CreateRecord(ctx context.Context, req metrics.CreateMetricRecordRequest, idempotencyKey string) (metrics.CreateMetricRecordResponse, error)
	BatchCreateRecords(ctx context.Context, req metrics.BatchCreateMetricRecordsRequest, idempotencyKey string) (metrics.BatchCreateMetricRecordsResponse, error)
	ListRecords(ctx context.Context, req metrics.ListMetricRecordsRequest) (metrics.PagedMetricRecordsResponse, error)
	GetSummary(ctx context.Context, projectID string, req metrics.MetricSummaryRequest) (metrics.MetricSummaryResponse, error)
	GetTrends(ctx context.Context, projectID string, req metrics.MetricTrendRequest) (metrics.MetricTrendResponse, error)
	GetMissingDates(ctx context.Context, projectID string, req metrics.MissingMetricDatesRequest) (metrics.MissingMetricDatesResponse, error)
}

type collectLogService interface {
	SubmitPlatformCollectLog(ctx context.Context, req metrics.SubmitPlatformCollectLogRequest, auth metrics.PlatformCollectLogAuth, idempotencyKey string) (metrics.SubmitPlatformCollectLogResponse, error)
	ListPlatformCollectLogs(ctx context.Context, req metrics.ListPlatformCollectLogsRequest) (metrics.PagedPlatformCollectLogsResponse, error)
	GetPlatformCollectLog(ctx context.Context, collectLogID string) (metrics.PlatformCollectLogDetailResponse, error)
	ConfirmPlatformCollectLogMetrics(ctx context.Context, collectLogID string, req metrics.ConfirmPlatformCollectLogMetricsRequest, idempotencyKey string) (metrics.ConfirmPlatformCollectLogMetricsResponse, error)
}

type MetricsHandler struct {
	coreService metricCoreService
	collectSvc  collectLogService
	logger      *slog.Logger
}

func NewMetricsHandler(coreService metricCoreService, collectSvc collectLogService, logger *slog.Logger) *MetricsHandler {
	return &MetricsHandler{coreService: coreService, collectSvc: collectSvc, logger: logger}
}

func (h *MetricsHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req metrics.CreateMetricTemplateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeMetricsError(w, r, metrics.ErrValidation, "invalid metric template request")
		return
	}
	data, err := h.coreService.CreateTemplate(r.Context(), req)
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
	data, err := h.coreService.ListTemplates(r.Context(), req)
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
	data, err := h.coreService.CreateRecord(r.Context(), req, r.Header.Get("Idempotency-Key"))
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
	data, err := h.coreService.BatchCreateRecords(r.Context(), req, r.Header.Get("Idempotency-Key"))
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
	data, err := h.coreService.ListRecords(r.Context(), req)
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
	data, err := h.coreService.GetSummary(r.Context(), chi.URLParam(r, "projectId"), req)
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
		TargetID:    r.URL.Query().Get("target_id"),
	}
	data, err := h.coreService.GetTrends(r.Context(), chi.URLParam(r, "projectId"), req)
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
		TargetID:    r.URL.Query().Get("target_id"),
		DateFrom:   r.URL.Query().Get("date_from"),
		DateTo:     r.URL.Query().Get("date_to"),
	}
	data, err := h.coreService.GetMissingDates(r.Context(), chi.URLParam(r, "projectId"), req)
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

func (h *MetricsHandler) SubmitPlatformCollectLog(w http.ResponseWriter, r *http.Request) {
	var req metrics.SubmitPlatformCollectLogRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeMetricsError(w, r, metrics.ErrValidation, "invalid platform collect log request")
		return
	}
	auth := metrics.PlatformCollectLogAuth{
		SourceType:         r.Header.Get("X-Source-Type"),
		PluginToken:        pluginBearerToken(r),
		BindingID:          r.Header.Get("X-External-Binding-Id"),
		CallbackAuthHeader: r.Header.Get("X-Callback-Auth"),
		SignatureHeader:    r.Header.Get("X-Signature"),
	}
	if auth.SourceType == "" {
		auth.SourceType = req.SourceType
	}
	data, err := h.collectSvc.SubmitPlatformCollectLog(r.Context(), req, auth, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeMetricsError(w, r, err, "submit platform collect log failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, data)
}

func (h *MetricsHandler) ListPlatformCollectLogs(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	pagination, err := parsePagination(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid query parameters", nil)
		return
	}
	sortParam := r.URL.Query().Get("sort")
	validSorts := map[string]bool{"": true, "collected_at": true, "status": true}
	if !validSorts[sortParam] {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid sort field", nil)
		return
	}
	req := metrics.ListPlatformCollectLogsRequest{
		PaginationRequest: pagination,
		ProjectID:         r.URL.Query().Get("project_id"),
		Platform:          r.URL.Query().Get("platform"),
		Status:            r.URL.Query().Get("status"),
	}
	data, err := h.collectSvc.ListPlatformCollectLogs(r.Context(), req)
	if err != nil {
		writeMetricsError(w, r, err, "list platform collect logs failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *MetricsHandler) GetPlatformCollectLog(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	data, err := h.collectSvc.GetPlatformCollectLog(r.Context(), chi.URLParam(r, "collectLogId"))
	if err != nil {
		writeMetricsError(w, r, err, "get platform collect log failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *MetricsHandler) ConfirmPlatformCollectLogMetrics(w http.ResponseWriter, r *http.Request) {
	var req metrics.ConfirmPlatformCollectLogMetricsRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeMetricsError(w, r, metrics.ErrValidation, "invalid confirm platform collect log metrics request")
		return
	}
	data, err := h.collectSvc.ConfirmPlatformCollectLogMetrics(r.Context(), chi.URLParam(r, "collectLogId"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeMetricsError(w, r, err, "confirm platform collect log metrics failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}