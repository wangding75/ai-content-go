package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/publish"
)

type PublishHandler struct {
	service publish.Service
	logger  *slog.Logger
}

func NewPublishHandler(service publish.Service, logger *slog.Logger) *PublishHandler {
	return &PublishHandler{service: service, logger: logger}
}

func (h *PublishHandler) ListTargets(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	req, err := parseListPublishTargetsRequest(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid query parameters", nil)
		return
	}
	data, err := h.service.ListTargets(r.Context(), chi.URLParam(r, "projectId"), req)
	if err != nil {
		writePublishError(w, r, err, "list publish targets failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PublishHandler) CreateTarget(w http.ResponseWriter, r *http.Request) {
	var req publish.CreatePublishTargetRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writePublishError(w, r, publish.ErrValidation, "invalid publish target request")
		return
	}
	data, err := h.service.CreateTarget(r.Context(), chi.URLParam(r, "projectId"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writePublishError(w, r, err, "create publish target failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, data)
}

func (h *PublishHandler) UpdateTarget(w http.ResponseWriter, r *http.Request) {
	var req publish.UpdatePublishTargetRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writePublishError(w, r, publish.ErrValidation, "invalid publish target update request")
		return
	}
	data, err := h.service.UpdateTarget(r.Context(), chi.URLParam(r, "projectId"), chi.URLParam(r, "id"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writePublishError(w, r, err, "update publish target failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PublishHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req publish.CreatePublishJobRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writePublishError(w, r, publish.ErrValidation, "invalid publish job request")
		return
	}
	data, err := h.service.CreateJob(r.Context(), chi.URLParam(r, "projectId"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writePublishError(w, r, err, "create publish job failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, data)
}

func (h *PublishHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	req, err := parseListPublishJobsRequest(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid query parameters", nil)
		return
	}
	data, err := h.service.ListJobs(r.Context(), chi.URLParam(r, "projectId"), req)
	if err != nil {
		writePublishError(w, r, err, "list publish jobs failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PublishHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	data, err := h.service.GetJob(r.Context(), chi.URLParam(r, "projectId"), chi.URLParam(r, "id"))
	if err != nil {
		writePublishError(w, r, err, "get publish job failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PublishHandler) GetCopyPayload(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	data, err := h.service.GetCopyPayload(r.Context(), chi.URLParam(r, "projectId"), chi.URLParam(r, "id"))
	if err != nil {
		writePublishError(w, r, err, "get publish copy payload failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PublishHandler) CopyPayload(w http.ResponseWriter, r *http.Request) {
	var req publish.CopyPublishPayloadRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writePublishError(w, r, publish.ErrValidation, "invalid copy publish payload request")
		return
	}
	data, err := h.service.CopyPayload(r.Context(), chi.URLParam(r, "projectId"), chi.URLParam(r, "id"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writePublishError(w, r, err, "copy publish payload failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PublishHandler) MarkPublished(w http.ResponseWriter, r *http.Request) {
	var req publish.MarkPublishedRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writePublishError(w, r, publish.ErrValidation, "invalid mark published request")
		return
	}
	data, err := h.service.MarkPublished(r.Context(), chi.URLParam(r, "projectId"), chi.URLParam(r, "id"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writePublishError(w, r, err, "mark publish job published failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PublishHandler) MarkFailed(w http.ResponseWriter, r *http.Request) {
	var req publish.MarkFailedRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writePublishError(w, r, publish.ErrValidation, "invalid mark failed request")
		return
	}
	data, err := h.service.MarkFailed(r.Context(), chi.URLParam(r, "projectId"), chi.URLParam(r, "id"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writePublishError(w, r, err, "mark publish job failed failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PublishHandler) Requeue(w http.ResponseWriter, r *http.Request) {
	var req publish.RequeuePublishJobRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writePublishError(w, r, publish.ErrValidation, "invalid requeue request")
		return
	}
	data, err := h.service.Requeue(r.Context(), chi.URLParam(r, "projectId"), chi.URLParam(r, "id"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writePublishError(w, r, err, "requeue publish job failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func parseListPublishTargetsRequest(r *http.Request) (publish.ListPublishTargetsRequest, error) {
	pagination, err := parsePagination(r)
	if err != nil {
		return publish.ListPublishTargetsRequest{}, err
	}
	req := publish.ListPublishTargetsRequest{PaginationRequest: pagination}
	if value := r.URL.Query().Get("enabled"); value != "" {
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return publish.ListPublishTargetsRequest{}, err
		}
		req.Enabled = &enabled
	}
	return req, nil
}

func parseListPublishJobsRequest(r *http.Request) (publish.ListPublishJobsRequest, error) {
	pagination, err := parsePagination(r)
	if err != nil {
		return publish.ListPublishJobsRequest{}, err
	}
	req := publish.ListPublishJobsRequest{PaginationRequest: pagination, TargetID: r.URL.Query().Get("target_id"), Status: r.URL.Query().Get("status")}
	if value := r.URL.Query().Get("scheduled_from"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return publish.ListPublishJobsRequest{}, err
		}
		req.ScheduledFrom = &parsed
	}
	return req, nil
}

func noStore(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
}

func writePublishError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, publish.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, publish.ErrForbidden):
		api.WriteError(w, r, http.StatusForbidden, api.ErrorForbidden, message, nil)
	case errors.Is(err, publish.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, publish.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	case errors.Is(err, publish.ErrIdempotencyConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorIdempotencyConflict, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}
