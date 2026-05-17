package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type WorkflowHandler struct {
	service   workflow.Service
	submitter engine.Submitter
	logger    *slog.Logger
}

func NewWorkflowHandler(service workflow.Service, submitter engine.Submitter, logger *slog.Logger) *WorkflowHandler {
	return &WorkflowHandler{service: service, submitter: submitter, logger: logger}
}

func (h *WorkflowHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	resp, err := h.service.ListTemplates(r.Context(), workflow.ListWorkflowTemplatesRequest{
		PaginationRequest: content.PaginationRequest{Page: page, PageSize: pageSize},
		ContentType:       r.URL.Query().Get("content_type"),
		Category:          r.URL.Query().Get("category"),
		Status:            r.URL.Query().Get("status"),
	})
	if err != nil {
		writeWorkflowError(w, r, err, "list templates failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *WorkflowHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req workflow.CreateWorkflowTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	resp, err := h.service.CreateTemplate(r.Context(), req)
	if err != nil {
		writeWorkflowError(w, r, err, "create template failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, resp)
}

func (h *WorkflowHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	resp, err := h.service.GetTemplate(r.Context(), id)
	if err != nil {
		writeWorkflowError(w, r, err, "template not found")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *WorkflowHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	resp, err := h.service.ListVersions(r.Context(), id, workflow.PaginationRequest{Page: page, PageSize: pageSize})
	if err != nil {
		writeWorkflowError(w, r, err, "list versions failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *WorkflowHandler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var req workflow.CreateVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	resp, err := h.service.CreateVersion(r.Context(), id, req)
	if err != nil {
		writeWorkflowError(w, r, err, "create version failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, resp)
}

func (h *WorkflowHandler) GetVersionDetail(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	resp, err := h.service.GetVersion(r.Context(), id)
	if err != nil {
		writeWorkflowError(w, r, err, "version not found")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *WorkflowHandler) PublishVersion(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var req workflow.PublishVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	resp, err := h.service.PublishVersion(r.Context(), id, req, idempotencyKey(r))
	if err != nil {
		writeWorkflowError(w, r, err, "publish version failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *WorkflowHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	resp, err := h.service.ListRuns(r.Context(), workflow.ListWorkflowRunsRequest{
		PaginationRequest:  content.PaginationRequest{Page: page, PageSize: pageSize},
		ProjectID:          r.URL.Query().Get("project_id"),
		TemplateVersionID:  r.URL.Query().Get("template_version_id"),
		Status:             r.URL.Query().Get("status"),
	})
	if err != nil {
		writeWorkflowError(w, r, err, "list runs failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *WorkflowHandler) CreateRun(w http.ResponseWriter, r *http.Request) {
	var req workflow.CreateWorkflowRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	resp, err := h.service.CreateRun(r.Context(), req, idempotencyKey(r))
	if err != nil {
		writeWorkflowError(w, r, err, "create run failed")
		return
	}
	if h.submitter != nil {
		h.submitter.Submit(resp.WorkflowRunID)
	}
	api.WriteSuccess(w, r, http.StatusAccepted, resp)
}

func (h *WorkflowHandler) GetRun(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	resp, err := h.service.GetRun(r.Context(), id)
	if err != nil {
		writeWorkflowError(w, r, err, "run not found")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *WorkflowHandler) GetRunSteps(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	resp, err := h.service.GetRunSteps(r.Context(), id)
	if err != nil {
		writeWorkflowError(w, r, err, "get run steps failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *WorkflowHandler) CancelRun(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var req workflow.CancelRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	resp, err := h.service.CancelRun(r.Context(), id, req, idempotencyKey(r))
	if err != nil {
		writeWorkflowError(w, r, err, "cancel run failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *WorkflowHandler) RetryRun(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var req workflow.RetryRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	resp, err := h.service.RetryRun(r.Context(), id, req, idempotencyKey(r))
	if err != nil {
		writeWorkflowError(w, r, err, "retry run failed")
		return
	}
	if h.submitter != nil {
		h.submitter.Submit(resp.NewWorkflowRunID)
	}
	api.WriteSuccess(w, r, http.StatusAccepted, resp)
}

func writeWorkflowError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, workflow.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, workflow.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, workflow.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	case errors.Is(err, workflow.ErrIdempotencyConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorIdempotencyConflict, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}

func idempotencyKey(r *http.Request) string {
	return r.Header.Get("Idempotency-Key")
}

func pathID(r *http.Request) string {
	return chi.URLParam(r, "id")
}
