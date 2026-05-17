package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/schedule"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type ScheduleHandler struct {
	service         schedule.Service
	workflowService workflow.Service
	submitter       engine.Submitter
	logger          *slog.Logger
}

func NewScheduleHandler(service schedule.Service, workflowService workflow.Service, submitter engine.Submitter, logger *slog.Logger) *ScheduleHandler {
	return &ScheduleHandler{service: service, workflowService: workflowService, submitter: submitter, logger: logger}
}

func (h *ScheduleHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	var enabled *bool
	if value := r.URL.Query().Get("enabled"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid enabled filter", nil)
			return
		}
		enabled = &parsed
	}
	resp, err := h.service.ListSchedules(r.Context(), schedule.ListSchedulesRequest{PaginationRequest: content.PaginationRequest{Page: page, PageSize: pageSize}, ProjectID: r.URL.Query().Get("project_id"), Enabled: enabled})
	if err != nil {
		writeScheduleError(w, r, err, "list schedules failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req schedule.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	resp, err := h.service.CreateSchedule(r.Context(), req, idempotencyKey(r))
	if err != nil {
		writeScheduleError(w, r, err, "create schedule failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, resp)
}

func (h *ScheduleHandler) EnableSchedule(w http.ResponseWriter, r *http.Request) {
	var req schedule.ToggleScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	resp, err := h.service.EnableSchedule(r.Context(), pathID(r), req, idempotencyKey(r))
	if err != nil {
		writeScheduleError(w, r, err, "enable schedule failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *ScheduleHandler) DisableSchedule(w http.ResponseWriter, r *http.Request) {
	var req schedule.ToggleScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	resp, err := h.service.DisableSchedule(r.Context(), pathID(r), req, idempotencyKey(r))
	if err != nil {
		writeScheduleError(w, r, err, "disable schedule failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *ScheduleHandler) TestRun(w http.ResponseWriter, r *http.Request) {
	var req schedule.TestRunScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	prepared, err := h.service.PrepareTestRun(r.Context(), pathID(r), req)
	if err != nil {
		writeScheduleError(w, r, err, "prepare schedule run failed")
		return
	}
	run, err := h.workflowService.CreateRun(r.Context(), workflow.CreateWorkflowRunRequest{ProjectID: prepared.ProjectID, TemplateVersionID: prepared.TemplateVersionID, Input: prepared.Input}, idempotencyKey(r))
	if err != nil {
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorWorkflowRunFailed, "create workflow run failed", nil)
		return
	}
	if err := h.service.CompleteTrigger(r.Context(), prepared.TriggerLogID, run.WorkflowRunID, run.Status); err != nil {
		writeScheduleError(w, r, err, "complete schedule trigger failed")
		return
	}
	if h.submitter != nil {
		h.submitter.Submit(run.WorkflowRunID)
	}
	api.WriteSuccess(w, r, http.StatusAccepted, schedule.TestRunScheduleResponse{WorkflowRunID: run.WorkflowRunID, Status: run.Status, TriggerLogID: prepared.TriggerLogID})
}

func (h *ScheduleHandler) ListTriggers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	resp, err := h.service.ListTriggers(r.Context(), pathID(r), schedule.ListTriggersRequest{PaginationRequest: content.PaginationRequest{Page: page, PageSize: pageSize}})
	if err != nil {
		writeScheduleError(w, r, err, "list schedule triggers failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func writeScheduleError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, schedule.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, schedule.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, schedule.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	case errors.Is(err, schedule.ErrIdempotencyConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorIdempotencyConflict, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}
