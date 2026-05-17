package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/external"
)

type ExternalHandler struct {
	service external.Service
	logger  *slog.Logger
}

func NewExternalHandler(service external.Service, logger *slog.Logger) *ExternalHandler {
	return &ExternalHandler{service: service, logger: logger}
}

func (h *ExternalHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "list external providers not implemented", nil)
}

func (h *ExternalHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "create external provider not implemented", nil)
}

func (h *ExternalHandler) ListBindings(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "list external bindings not implemented", nil)
}

func (h *ExternalHandler) CreateBinding(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "create external binding not implemented", nil)
}

func writeExternalError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, external.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, external.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, external.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	case errors.Is(err, external.ErrExternalAutomation):
		api.WriteError(w, r, http.StatusBadGateway, api.ErrorExternalAutomation, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}
