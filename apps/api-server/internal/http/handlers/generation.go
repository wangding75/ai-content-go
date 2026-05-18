package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/generation"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type GenerationHandler struct {
	service   generationHandlerService
	workflow  workflow.Service
	submitter engine.Submitter
	logger    *slog.Logger
}

type generationHandlerService interface {
	CreateGenerationRun(ctx context.Context, projectID string, req generation.CreateGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.CreateGenerationRunResponse, error)
	CreateBatchGenerationRuns(ctx context.Context, projectID string, req generation.CreateBatchGenerationRunsRequest, workflowRunIDs []string, idempotencyKey string) (generation.CreateBatchGenerationRunsResponse, error)
	ListGenerationRuns(ctx context.Context, projectID string, req generation.ListGenerationRunsRequest) (generation.PagedGenerationRunsResponse, error)
	GetGenerationRun(ctx context.Context, id string) (generation.GenerationRunDetailResponse, error)
	RetryGenerationRun(ctx context.Context, id string, req generation.RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.RetryGenerationRunResponse, error)
	ListContentItems(ctx context.Context, projectID string, req generation.ListContentItemsRequest) (generation.PagedContentItemsResponse, error)
	GetContentItem(ctx context.Context, id string) (generation.ContentItemDetailResponse, error)
}

func NewGenerationHandler(service generationHandlerService, workflowService workflow.Service, submitter engine.Submitter, logger *slog.Logger) *GenerationHandler {
	return &GenerationHandler{service: service, workflow: workflowService, submitter: submitter, logger: logger}
}

func (h *GenerationHandler) CreateGenerationRun(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	idempotencyKey := r.Header.Get("Idempotency-Key")
	var req generation.CreateGenerationRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGenerationError(w, r, generation.ErrValidation, "invalid generation run request")
		return
	}
	if replayer, ok := h.service.(interface {
		ReplayGenerationRun(context.Context, string, generation.CreateGenerationRunRequest, string) (generation.CreateGenerationRunResponse, bool, error)
	}); ok {
		result, replayed, err := replayer.ReplayGenerationRun(r.Context(), projectID, req, idempotencyKey)
		if err != nil {
			writeGenerationError(w, r, err, "invalid generation run request")
			return
		}
		if replayed {
			api.WriteSuccess(w, r, http.StatusAccepted, result)
			return
		}
	} else if projectID == "" || idempotencyKey == "" || req.TargetCount <= 0 || req.StartSequenceNo <= 0 || req.TemplateVersionID == "" {
		writeGenerationError(w, r, generation.ErrValidation, "invalid generation run request")
		return
	}
	workflowRunID, err := h.createWorkflowRun(r, projectID, req.TemplateVersionID, req.GenerationConfig, idempotencyKey)
	if err != nil {
		h.releaseGenerationRunReservation(projectID, req, idempotencyKey)
		writeGenerationError(w, r, generation.ErrWorkflowRunFailed, "create workflow run failed")
		return
	}
	result, err := h.service.CreateGenerationRun(r.Context(), projectID, req, workflowRunID, idempotencyKey)
	if err != nil {
		h.releaseGenerationRunReservation(projectID, req, idempotencyKey)
		writeGenerationError(w, r, err, "create generation run failed")
		return
	}
	if err := h.submitWorkflow(workflowRunID); err != nil {
		h.releaseGenerationRunReservation(projectID, req, idempotencyKey)
		writeGenerationError(w, r, err, "submit workflow run failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusAccepted, result)
}

func (h *GenerationHandler) CreateBatchGenerationRuns(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	idempotencyKey := r.Header.Get("Idempotency-Key")
	var req generation.CreateBatchGenerationRunsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGenerationError(w, r, generation.ErrValidation, "invalid batch generation request")
		return
	}
	batchCount := req.Range.EndSequenceNo - req.Range.StartSequenceNo + 1
	if idempotencyKey == "" || projectID == "" || req.TemplateVersionID == "" || req.BatchSize <= 0 || req.Range.StartSequenceNo <= 0 || req.Range.EndSequenceNo < req.Range.StartSequenceNo || req.BatchSize != batchCount {
		writeGenerationError(w, r, generation.ErrValidation, "invalid batch generation request")
		return
	}
	if replayer, ok := h.service.(interface {
		ReplayBatchGenerationRuns(context.Context, string, generation.CreateBatchGenerationRunsRequest, string) (generation.CreateBatchGenerationRunsResponse, bool, error)
	}); ok {
		result, replayed, err := replayer.ReplayBatchGenerationRuns(r.Context(), projectID, req, idempotencyKey)
		if err != nil {
			writeGenerationError(w, r, err, "invalid batch generation request")
			return
		}
		if replayed {
			api.WriteSuccess(w, r, http.StatusAccepted, result)
			return
		}
	}
	workflowRunIDs := make([]string, 0, batchCount)
	for i := 0; i < batchCount; i++ {
		workflowRunID, err := h.createWorkflowRun(r, projectID, req.TemplateVersionID, req.GenerationConfig, fmt.Sprintf("%s:%d", idempotencyKey, i+1))
		if err != nil {
			h.releaseBatchGenerationRunsReservation(projectID, req, idempotencyKey)
			h.cancelWorkflowRuns(workflowRunIDs)
			writeGenerationError(w, r, generation.ErrWorkflowRunFailed, "create workflow run failed")
			return
		}
		workflowRunIDs = append(workflowRunIDs, workflowRunID)
	}
	result, err := h.service.CreateBatchGenerationRuns(r.Context(), projectID, req, workflowRunIDs, idempotencyKey)
	if err != nil {
		h.releaseBatchGenerationRunsReservation(projectID, req, idempotencyKey)
		h.cancelWorkflowRuns(workflowRunIDs)
		writeGenerationError(w, r, err, "create batch generation runs failed")
		return
	}
	for _, workflowRunID := range workflowRunIDs {
		if err := h.submitWorkflow(workflowRunID); err != nil {
			h.releaseBatchGenerationRunsReservation(projectID, req, idempotencyKey)
			h.cancelWorkflowRuns(workflowRunIDs)
			writeGenerationError(w, r, err, "submit workflow run failed")
			return
		}
	}
	api.WriteSuccess(w, r, http.StatusAccepted, result)
}

func (h *GenerationHandler) ListGenerationRuns(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ListGenerationRuns(r.Context(), chi.URLParam(r, "projectId"), generation.ListGenerationRunsRequest{
		PaginationRequest: paginationRequestFromQuery(r),
		Status:            r.URL.Query().Get("status"),
	})
	if err != nil {
		writeGenerationError(w, r, err, "list generation runs failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *GenerationHandler) GetGenerationRun(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetGenerationRun(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeGenerationError(w, r, err, "get generation run failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *GenerationHandler) RetryGenerationRun(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idempotencyKey := r.Header.Get("Idempotency-Key")
	var req generation.RetryGenerationRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGenerationError(w, r, generation.ErrValidation, "invalid retry request")
		return
	}
	if replayer, ok := h.service.(interface {
		ReplayRetryGenerationRun(context.Context, string, generation.RetryGenerationRunRequest, string) (generation.RetryGenerationRunResponse, bool, error)
	}); ok {
		result, replayed, err := replayer.ReplayRetryGenerationRun(r.Context(), id, req, idempotencyKey)
		if err != nil {
			writeGenerationError(w, r, err, "invalid retry request")
			return
		}
		if replayed {
			api.WriteSuccess(w, r, http.StatusAccepted, result)
			return
		}
	} else if id == "" || idempotencyKey == "" {
		writeGenerationError(w, r, generation.ErrValidation, "invalid retry request")
		return
	}
	original, err := h.service.GetGenerationRun(r.Context(), id)
	if err != nil {
		writeGenerationError(w, r, err, "retry generation run failed")
		return
	}
	workflowRunID, err := h.retryWorkflowRun(r, original, req, idempotencyKey)
	if err != nil {
		h.releaseRetryGenerationRunReservation(id, req, idempotencyKey)
		writeGenerationError(w, r, generation.ErrWorkflowRunFailed, "retry workflow run failed")
		return
	}
	result, err := h.service.RetryGenerationRun(r.Context(), id, req, workflowRunID, idempotencyKey)
	if err != nil {
		h.releaseRetryGenerationRunReservation(id, req, idempotencyKey)
		writeGenerationError(w, r, err, "retry generation run failed")
		return
	}
	if err := h.submitWorkflow(workflowRunID); err != nil {
		h.releaseRetryGenerationRunReservation(id, req, idempotencyKey)
		writeGenerationError(w, r, err, "submit workflow run failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusAccepted, result)
}

func (h *GenerationHandler) ListContentItems(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ListContentItems(r.Context(), chi.URLParam(r, "projectId"), generation.ListContentItemsRequest{
		PaginationRequest: paginationRequestFromQuery(r),
		Status:            r.URL.Query().Get("status"),
	})
	if err != nil {
		writeGenerationError(w, r, err, "list content items failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *GenerationHandler) GetContentItem(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetContentItem(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeGenerationError(w, r, err, "get content item failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *GenerationHandler) releaseGenerationRunReservation(projectID string, req generation.CreateGenerationRunRequest, idempotencyKey string) {
	if releaser, ok := h.service.(interface {
		ReleaseGenerationRunReservation(context.Context, string, generation.CreateGenerationRunRequest, string) error
	}); ok {
		if err := releaser.ReleaseGenerationRunReservation(context.Background(), projectID, req, idempotencyKey); err != nil {
			h.logCleanupError("release generation run reservation failed", err)
		}
	}
}

func (h *GenerationHandler) releaseBatchGenerationRunsReservation(projectID string, req generation.CreateBatchGenerationRunsRequest, idempotencyKey string) {
	if releaser, ok := h.service.(interface {
		ReleaseBatchGenerationRunsReservation(context.Context, string, generation.CreateBatchGenerationRunsRequest, string) error
	}); ok {
		if err := releaser.ReleaseBatchGenerationRunsReservation(context.Background(), projectID, req, idempotencyKey); err != nil {
			h.logCleanupError("release batch generation runs reservation failed", err)
		}
	}
}

func (h *GenerationHandler) releaseRetryGenerationRunReservation(id string, req generation.RetryGenerationRunRequest, idempotencyKey string) {
	if releaser, ok := h.service.(interface {
		ReleaseRetryGenerationRunReservation(context.Context, string, generation.RetryGenerationRunRequest, string) error
	}); ok {
		if err := releaser.ReleaseRetryGenerationRunReservation(context.Background(), id, req, idempotencyKey); err != nil {
			h.logCleanupError("release retry generation run reservation failed", err)
		}
	}
}

func (h *GenerationHandler) cancelWorkflowRuns(workflowRunIDs []string) {
	if h.workflow == nil {
		return
	}
	for _, workflowRunID := range workflowRunIDs {
		if _, err := h.workflow.CancelRun(context.Background(), workflowRunID, workflow.CancelRunRequest{Reason: "batch generation setup failed"}, "cancel:"+workflowRunID); err != nil {
			h.logCleanupError("cancel workflow run failed", err)
		}
	}
}

func (h *GenerationHandler) logCleanupError(message string, err error) {
	if h.logger != nil {
		h.logger.Error(message, "error", err)
	}
}

func (h *GenerationHandler) createWorkflowRun(r *http.Request, projectID string, templateVersionID string, input map[string]any, idempotencyKey string) (string, error) {
	if h.workflow == nil {
		return fmt.Sprintf("wfr-%s-%s", api.RequestID(r), idempotencyKey), nil
	}
	result, err := h.workflow.CreateRun(r.Context(), workflow.CreateWorkflowRunRequest{ProjectID: projectID, TemplateVersionID: templateVersionID, Input: input}, idempotencyKey)
	if err != nil {
		return "", err
	}
	return result.WorkflowRunID, nil
}

func (h *GenerationHandler) retryWorkflowRun(r *http.Request, original generation.GenerationRunDetailResponse, req generation.RetryGenerationRunRequest, idempotencyKey string) (string, error) {
	if h.workflow == nil {
		return fmt.Sprintf("wfr-%s-%s", api.RequestID(r), idempotencyKey), nil
	}
	input := req.InputOverride
	if input == nil {
		input = map[string]any{}
	}
	result, err := h.workflow.CreateRun(r.Context(), workflow.CreateWorkflowRunRequest{ProjectID: original.ProjectID, TemplateVersionID: original.TemplateVersionID, Input: input}, idempotencyKey)
	if err != nil {
		return "", err
	}
	return result.WorkflowRunID, nil
}

func (h *GenerationHandler) submitWorkflow(workflowRunID string) error {
	if h.submitter != nil && !h.submitter.Submit(workflowRunID) {
		return generation.ErrWorkflowRunFailed
	}
	return nil
}

func paginationRequestFromQuery(r *http.Request) content.PaginationRequest {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	return content.PaginationRequest{Page: page, PageSize: pageSize, Sort: query.Get("sort"), Order: query.Get("order")}
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
