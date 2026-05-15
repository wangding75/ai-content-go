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
	panic("not implemented")
}

func (h *LLMLogHandler) GetCallLog(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
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
