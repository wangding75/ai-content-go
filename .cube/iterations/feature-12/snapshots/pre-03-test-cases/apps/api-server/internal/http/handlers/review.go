package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/review"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type ReviewHandler struct {
	service   reviewHandlerService
	workflow  workflow.Service
	submitter engine.Submitter
	logger    *slog.Logger
}

type reviewHandlerService interface {
	ListReviews(ctx context.Context, req review.ListReviewsRequest) (review.PagedContentReviewsResponse, error)
	CreateReview(ctx context.Context, contentItemID string, req review.CreateReviewRequest, idempotencyKey string) (review.CreateReviewResponse, error)
	GetReview(ctx context.Context, id string) (review.ContentReviewDetailResponse, error)
	TriggerAIReport(ctx context.Context, id string, req review.TriggerAIReportRequest, workflowRunID string, idempotencyKey string) (review.TriggerAIReportResponse, error)
	GetAIReport(ctx context.Context, id string) (review.ReviewReportResponse, error)
	ApproveReview(ctx context.Context, id string, req review.ApproveReviewRequest) (review.ApproveReviewResponse, error)
	RejectReview(ctx context.Context, id string, req review.RejectReviewRequest, regenerationRunID string) (review.RejectReviewResponse, error)
	ApproveWithEdit(ctx context.Context, id string, req review.ApproveWithEditRequest) (review.ApproveWithEditResponse, error)
	ListContentVersions(ctx context.Context, contentItemID string, req review.ListReviewsRequest) (review.PagedContentVersionsResponse, error)
}

func NewReviewHandler(service reviewHandlerService, workflowService workflow.Service, submitter engine.Submitter, logger *slog.Logger) *ReviewHandler {
	return &ReviewHandler{service: service, workflow: workflowService, submitter: submitter, logger: logger}
}

func (h *ReviewHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ListReviews(r.Context(), review.ListReviewsRequest{PaginationRequest: paginationRequestFromQuery(r), ProjectID: r.URL.Query().Get("project_id"), Status: r.URL.Query().Get("status")})
	if err != nil {
		writeReviewError(w, r, err, "list content reviews failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	var req review.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeReviewError(w, r, review.ErrValidation, "invalid create review request")
		return
	}
	result, err := h.service.CreateReview(r.Context(), chi.URLParam(r, "id"), req, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeReviewError(w, r, err, "create content review failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, result)
}

func (h *ReviewHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetReview(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeReviewError(w, r, err, "get content review failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *ReviewHandler) TriggerAIReport(w http.ResponseWriter, r *http.Request) {
	var req review.TriggerAIReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeReviewError(w, r, review.ErrValidation, "invalid ai report request")
		return
	}
	workflowRunID, err := h.createReviewWorkflowRun(r, chi.URLParam(r, "id"), req.Config, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeReviewError(w, r, review.ErrWorkflowRunFailed, "create ai report run failed")
		return
	}
	result, err := h.service.TriggerAIReport(r.Context(), chi.URLParam(r, "id"), req, workflowRunID, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeReviewError(w, r, err, "trigger ai report failed")
		return
	}
	if err := h.submitReviewWorkflow(workflowRunID); err != nil {
		writeReviewError(w, r, err, "submit ai report run failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusAccepted, result)
}

func (h *ReviewHandler) GetAIReport(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetAIReport(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeReviewError(w, r, err, "get ai report failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *ReviewHandler) ApproveReview(w http.ResponseWriter, r *http.Request) {
	var req review.ApproveReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeReviewError(w, r, review.ErrValidation, "invalid approve request")
		return
	}
	result, err := h.service.ApproveReview(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		writeReviewError(w, r, err, "approve review failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *ReviewHandler) RejectReview(w http.ResponseWriter, r *http.Request) {
	var req review.RejectReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeReviewError(w, r, review.ErrValidation, "invalid reject request")
		return
	}
	regenerationRunID := ""
	if req.TriggerRegeneration {
		var err error
		regenerationRunID, err = h.createReviewWorkflowRun(r, chi.URLParam(r, "id"), map[string]any{"reason": req.Reason, "instruction": req.RegenerateInstruction}, "reject:"+api.RequestID(r))
		if err != nil {
			writeReviewError(w, r, review.ErrWorkflowRunFailed, "create regeneration run failed")
			return
		}
	}
	result, err := h.service.RejectReview(r.Context(), chi.URLParam(r, "id"), req, regenerationRunID)
	if err != nil {
		writeReviewError(w, r, err, "reject review failed")
		return
	}
	if req.TriggerRegeneration {
		if err := h.submitReviewWorkflow(regenerationRunID); err != nil {
			writeReviewError(w, r, err, "submit regeneration run failed")
			return
		}
	}
	api.WriteSuccess(w, r, http.StatusAccepted, result)
}

func (h *ReviewHandler) ApproveWithEdit(w http.ResponseWriter, r *http.Request) {
	var req review.ApproveWithEditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeReviewError(w, r, review.ErrValidation, "invalid edit approve request")
		return
	}
	result, err := h.service.ApproveWithEdit(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		writeReviewError(w, r, err, "approve with edit failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *ReviewHandler) ListContentVersions(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ListContentVersions(r.Context(), chi.URLParam(r, "id"), review.ListReviewsRequest{PaginationRequest: paginationRequestFromQuery(r)})
	if err != nil {
		writeReviewError(w, r, err, "list content versions failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, result)
}

func (h *ReviewHandler) createReviewWorkflowRun(r *http.Request, reviewID string, input map[string]any, idempotencyKey string) (string, error) {
	if h.workflow == nil {
		return "wfr-" + api.RequestID(r) + "-" + reviewID, nil
	}
	detail, err := h.service.GetReview(r.Context(), reviewID)
	if err != nil {
		return "", err
	}
	if detail.ProjectID == "" {
		return "", review.ErrValidation
	}
	result, err := h.workflow.CreateRun(r.Context(), workflow.CreateWorkflowRunRequest{ProjectID: detail.ProjectID, TemplateVersionID: "review-quality-control", Input: input}, idempotencyKey)
	if err != nil {
		return "", err
	}
	return result.WorkflowRunID, nil
}

func (h *ReviewHandler) submitReviewWorkflow(workflowRunID string) error {
	if h.submitter != nil && !h.submitter.Submit(workflowRunID) {
		return review.ErrWorkflowRunFailed
	}
	return nil
}

func writeReviewError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, review.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, review.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, review.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	case errors.Is(err, review.ErrIdempotencyConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorIdempotencyConflict, message, nil)
	case errors.Is(err, review.ErrWorkflowRunFailed):
		api.WriteError(w, r, http.StatusUnprocessableEntity, api.ErrorWorkflowRunFailed, message, nil)
	case errors.Is(err, review.ErrLLMProviderError):
		api.WriteError(w, r, http.StatusBadGateway, api.ErrorLLMProviderError, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}
