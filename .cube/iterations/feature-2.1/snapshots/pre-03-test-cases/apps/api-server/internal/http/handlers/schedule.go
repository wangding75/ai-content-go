package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
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
	api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "list schedules not implemented", nil)
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "create schedule not implemented", nil)
}

func (h *ScheduleHandler) EnableSchedule(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "enable schedule not implemented", nil)
}

func (h *ScheduleHandler) DisableSchedule(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "disable schedule not implemented", nil)
}

func (h *ScheduleHandler) TestRun(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "test run schedule not implemented", nil)
}

func (h *ScheduleHandler) ListTriggers(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "list schedule triggers not implemented", nil)
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
