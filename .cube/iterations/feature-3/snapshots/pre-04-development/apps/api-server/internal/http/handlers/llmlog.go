package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/llm"
)

type LLMLogHandler struct {
	service llm.Service
	logger  *slog.Logger
}

func NewLLMLogHandler(service llm.Service, logger *slog.Logger) *LLMLogHandler {
	return &LLMLogHandler{service: service, logger: logger}
}

func (h *LLMLogHandler) ListCallLogs(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.ListCallLogs(r.Context(), llm.ListLLMCallLogsRequest{
		WorkflowRunID: r.URL.Query().Get("workflow_run_id"),
		AgentTaskID:   r.URL.Query().Get("agent_task_id"),
		Provider:      r.URL.Query().Get("provider"),
		Model:         r.URL.Query().Get("model"),
		Status:        r.URL.Query().Get("status"),
	})
	if err != nil {
		writeLLMLogError(w, r, err, "list call logs failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *LLMLogHandler) GetCallLog(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	resp, err := h.service.GetCallLog(r.Context(), id)
	if err != nil {
		writeLLMLogError(w, r, err, "call log not found")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *LLMLogHandler) Summary(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.SummaryCallLogs(r.Context(), llm.SummaryCallLogsRequest{
		ProjectID: r.URL.Query().Get("project_id"),
		DateFrom:  r.URL.Query().Get("date_from"),
		DateTo:    r.URL.Query().Get("date_to"),
		Provider:  r.URL.Query().Get("provider"),
		Model:     r.URL.Query().Get("model"),
	})
	if err != nil {
		writeLLMLogError(w, r, err, "summary call logs failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func writeLLMLogError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, llm.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, llm.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, llm.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}
