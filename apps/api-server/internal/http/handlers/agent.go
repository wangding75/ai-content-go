package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/agent"
)

type AgentHandler struct {
	service agent.Service
	logger  *slog.Logger
}

func NewAgentHandler(service agent.Service, logger *slog.Logger) *AgentHandler {
	return &AgentHandler{service: service, logger: logger}
}

func (h *AgentHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	resp, err := h.service.ListTasks(r.Context(), agent.ListAgentTasksRequest{
		WorkflowRunID: r.URL.Query().Get("workflow_run_id"),
		StepRunID:     r.URL.Query().Get("step_run_id"),
		AgentCode:     r.URL.Query().Get("agent_code"),
		Status:        r.URL.Query().Get("status"),
	})
	resp.Pagination.Page = page
	if pageSize > 0 {
		resp.Pagination.PageSize = pageSize
	}
	if err != nil {
		writeAgentError(w, r, err, "list tasks failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *AgentHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	resp, err := h.service.GetTask(r.Context(), id)
	if err != nil {
		writeAgentError(w, r, err, "task not found")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func writeAgentError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, agent.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, agent.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, agent.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}
