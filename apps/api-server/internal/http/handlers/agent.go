package handlers

import (
	"errors"
	"log/slog"
	"net/http"

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
	panic("not implemented")
}

func (h *AgentHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
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
