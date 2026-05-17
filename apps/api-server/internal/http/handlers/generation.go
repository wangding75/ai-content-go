package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/generation"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type GenerationHandler struct {
	service   generation.Service
	workflow  workflow.Service
	submitter engine.Submitter
	logger    *slog.Logger
}

func NewGenerationHandler(service generation.Service, workflowService workflow.Service, submitter engine.Submitter, logger *slog.Logger) *GenerationHandler {
	return &GenerationHandler{service: service, workflow: workflowService, submitter: submitter, logger: logger}
}

func (h *GenerationHandler) CreateGenerationRun(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "generation run skeleton not implemented", nil)
}

func (h *GenerationHandler) CreateBatchGenerationRuns(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "batch generation skeleton not implemented", nil)
}

func (h *GenerationHandler) ListGenerationRuns(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "generation run list skeleton not implemented", nil)
}

func (h *GenerationHandler) GetGenerationRun(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "generation run detail skeleton not implemented", nil)
}

func (h *GenerationHandler) RetryGenerationRun(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "generation retry skeleton not implemented", nil)
}

func (h *GenerationHandler) ListContentItems(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "content item list skeleton not implemented", nil)
}

func (h *GenerationHandler) GetContentItem(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "content item detail skeleton not implemented", nil)
}

func writeGenerationError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, generation.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, generation.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, generation.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	case errors.Is(err, generation.ErrIdempotencyConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorIdempotencyConflict, message, nil)
	case errors.Is(err, generation.ErrWorkflowRunFailed):
		api.WriteError(w, r, http.StatusUnprocessableEntity, api.ErrorWorkflowRunFailed, message, nil)
	case errors.Is(err, generation.ErrLLMProviderError):
		api.WriteError(w, r, http.StatusBadGateway, api.ErrorLLMProviderError, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}
