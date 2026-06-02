package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/novel"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type NovelHandler struct {
	service   novel.Service
	workflow  workflow.Service
	submitter engine.Submitter
	logger    *slog.Logger
}

func NewNovelHandler(service novel.Service, workflowService workflow.Service, submitter engine.Submitter, logger *slog.Logger) *NovelHandler {
	return &NovelHandler{service: service, workflow: workflowService, submitter: submitter, logger: logger}
}

func (h *NovelHandler) CreatePlanningRun(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(idempotencyKey(r))
	if key == "" || len(key) > 200 {
		writeNovelError(w, r, novel.ErrValidation, "create planning run failed")
		return
	}
	var req novel.CreatePlanningRunRequest
	if err := decodeNovelJSON(w, r, &req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	projectID := novelProjectID(r)
	wf, err := h.workflow.CreateRun(r.Context(), workflow.CreateWorkflowRunRequest{ProjectID: projectID, TemplateVersionID: req.TemplateVersionID, Input: req.InputOverride}, key)
	if err != nil {
		writeNovelError(w, r, novel.ErrWorkflowRunFailed, "create planning run failed")
		return
	}
	resp, err := h.service.CreatePlanningRun(r.Context(), projectID, req, wf.WorkflowRunID, key)
	if err != nil {
		if !errors.Is(err, novel.ErrIdempotencyConflict) {
			_, _ = h.workflow.CancelRun(r.Context(), wf.WorkflowRunID, workflow.CancelRunRequest{Reason: "planning run creation failed"}, key+":cancel")
		}
		writeNovelError(w, r, err, "create planning run failed")
		return
	}
	if h.submitter != nil {
		h.submitter.Submit(resp.WorkflowRunID)
	}
	api.WriteSuccess(w, r, http.StatusAccepted, resp)
}

func (h *NovelHandler) ListPlanningRuns(w http.ResponseWriter, r *http.Request) {
	pagination, err := parseNovelPagination(r)
	if err != nil {
		writeNovelError(w, r, err, "list planning runs failed")
		return
	}
	resp, err := h.service.ListPlanningRuns(r.Context(), novelProjectID(r), novel.ListPlanningRunsRequest{PaginationRequest: pagination, Status: r.URL.Query().Get("status")})
	if err != nil {
		writeNovelError(w, r, err, "list planning runs failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *NovelHandler) GetPlanningRun(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GetPlanningRun(r.Context(), novelProjectID(r), chi.URLParam(r, "runId"))
	if err != nil {
		writeNovelError(w, r, err, "planning run not found")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *NovelHandler) ConfirmTopic(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(idempotencyKey(r))
	if key == "" || len(key) > 200 {
		writeNovelError(w, r, novel.ErrValidation, "confirm topic failed")
		return
	}
	var req novel.ConfirmTopicRequest
	if err := decodeNovelJSON(w, r, &req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	resp, err := h.service.ConfirmTopic(r.Context(), novelProjectID(r), chi.URLParam(r, "topicId"), req, key)
	if err != nil {
		writeNovelError(w, r, err, "confirm topic failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *NovelHandler) GetWorldview(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GetWorldview(r.Context(), novelProjectID(r))
	if err != nil {
		writeNovelError(w, r, err, "worldview not found")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *NovelHandler) UpdateWorldview(w http.ResponseWriter, r *http.Request) {
	var req novel.UpdateWorldviewRequest
	if err := decodeNovelJSON(w, r, &req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	resp, err := h.service.UpdateWorldview(r.Context(), novelProjectID(r), req)
	if err != nil {
		writeNovelError(w, r, err, "update worldview failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *NovelHandler) ListCharacters(w http.ResponseWriter, r *http.Request) {
	pagination, err := parseNovelPagination(r)
	if err != nil {
		writeNovelError(w, r, err, "list characters failed")
		return
	}
	resp, err := h.service.ListCharacters(r.Context(), novelProjectID(r), novel.ListCharactersRequest{PaginationRequest: pagination, Role: r.URL.Query().Get("role")})
	if err != nil {
		writeNovelError(w, r, err, "list characters failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *NovelHandler) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	var req novel.CreateCharacterRequest
	if err := decodeNovelJSON(w, r, &req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	resp, err := h.service.CreateCharacter(r.Context(), novelProjectID(r), req)
	if err != nil {
		writeNovelError(w, r, err, "create character failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, resp)
}

func (h *NovelHandler) ListArcs(w http.ResponseWriter, r *http.Request) {
	pagination, err := parseNovelPagination(r)
	if err != nil {
		writeNovelError(w, r, err, "list arcs failed")
		return
	}
	resp, err := h.service.ListArcs(r.Context(), novelProjectID(r), novel.ListArcsRequest{PaginationRequest: pagination})
	if err != nil {
		writeNovelError(w, r, err, "list arcs failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func decodeNovelJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func novelProjectID(r *http.Request) string {
	if projectID := chi.URLParam(r, "projectId"); projectID != "" {
		return projectID
	}
	return chi.URLParam(r, "id")
}

func parseNovelPagination(r *http.Request) (content.PaginationRequest, error) {
	page, err := parseOptionalPositiveInt(r.URL.Query().Get("page"))
	if err != nil {
		return content.PaginationRequest{}, novel.ErrValidation
	}
	pageSize, err := parseOptionalPositiveInt(r.URL.Query().Get("page_size"))
	if err != nil {
		return content.PaginationRequest{}, novel.ErrValidation
	}
	if pageSize > 100 {
		return content.PaginationRequest{}, novel.ErrValidation
	}
	sort := r.URL.Query().Get("sort")
	if sort != "" && sort != "created_at" && sort != "order_index" {
		return content.PaginationRequest{}, novel.ErrValidation
	}
	order := r.URL.Query().Get("order")
	if order != "" && order != "asc" && order != "desc" {
		return content.PaginationRequest{}, novel.ErrValidation
	}
	return content.PaginationRequest{Page: page, PageSize: pageSize, Sort: sort, Order: order}, nil
}

func parseOptionalPositiveInt(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, novel.ErrValidation
	}
	return parsed, nil
}

func writeNovelError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, novel.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, novel.ErrForbidden):
		api.WriteError(w, r, http.StatusForbidden, api.ErrorForbidden, message, nil)
	case errors.Is(err, novel.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, novel.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	case errors.Is(err, novel.ErrIdempotencyConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorIdempotencyConflict, message, nil)
	case errors.Is(err, novel.ErrWorkflowRunFailed):
		api.WriteError(w, r, http.StatusUnprocessableEntity, api.ErrorWorkflowRunFailed, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}
