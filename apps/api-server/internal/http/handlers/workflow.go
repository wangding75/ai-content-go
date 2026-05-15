package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
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
	panic("not implemented")
}

func (h *WorkflowHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *WorkflowHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *WorkflowHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *WorkflowHandler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *WorkflowHandler) GetVersionDetail(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *WorkflowHandler) PublishVersion(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *WorkflowHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *WorkflowHandler) CreateRun(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *WorkflowHandler) GetRun(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *WorkflowHandler) GetRunSteps(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *WorkflowHandler) CancelRun(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *WorkflowHandler) RetryRun(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
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
