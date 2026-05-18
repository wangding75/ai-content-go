package handlers

import (
	"log/slog"
	"net/http"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/llm"
)

type LLMHandler struct {
	service llm.Service
	logger  *slog.Logger
}

func NewLLMHandler(service llm.Service, logger *slog.Logger) *LLMHandler {
	return &LLMHandler{service: service, logger: logger}
}

func (h *LLMHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	pagination, err := parsePagination(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid query parameters", nil)
		return
	}
	data, err := h.service.ListProviders(r.Context(), llm.ListProvidersRequest{PaginationRequest: pagination})
	if err != nil {
		writeLLMError(w, r, err, "llm providers failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *LLMHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	var req llm.CreateProviderRequest
	if err := decodeJSON(w, r, &req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	data, err := h.service.CreateProvider(r.Context(), req)
	if err != nil {
		writeLLMError(w, r, err, "create llm provider failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, data)
}
