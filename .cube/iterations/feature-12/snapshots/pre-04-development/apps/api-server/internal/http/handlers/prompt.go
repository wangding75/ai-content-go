package handlers

import (
	"log/slog"
	"net/http"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/prompt"
)

type PromptHandler struct {
	service prompt.Service
	logger  *slog.Logger
}

func NewPromptHandler(service prompt.Service, logger *slog.Logger) *PromptHandler {
	return &PromptHandler{service: service, logger: logger}
}

func (h *PromptHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	pagination, err := parsePagination(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid query parameters", nil)
		return
	}
	data, err := h.service.ListTemplates(r.Context(), prompt.ListTemplatesRequest{PaginationRequest: pagination, AgentCode: r.URL.Query().Get("agent_code")})
	if err != nil {
		writePromptError(w, r, err, "prompt templates failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *PromptHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req prompt.CreateTemplateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	data, err := h.service.CreateTemplate(r.Context(), req)
	if err != nil {
		writePromptError(w, r, err, "create prompt template failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, data)
}
