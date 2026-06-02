package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
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
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	resp, err := h.service.ListProviders(r.Context(), external.ListProvidersRequest{PaginationRequest: content.PaginationRequest{Page: page, PageSize: pageSize}, ProviderType: r.URL.Query().Get("provider_type")})
	if err != nil {
		writeExternalError(w, r, err, "list external providers failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *ExternalHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	var req external.CreateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	resp, err := h.service.CreateProvider(r.Context(), req)
	if err != nil {
		writeExternalError(w, r, err, "create external provider failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, resp)
}

func (h *ExternalHandler) ListBindings(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	resp, err := h.service.ListBindings(r.Context(), external.ListBindingsRequest{PaginationRequest: content.PaginationRequest{Page: page, PageSize: pageSize}, ProviderID: r.URL.Query().Get("provider_id"), TriggerEvent: r.URL.Query().Get("trigger_event")})
	if err != nil {
		writeExternalError(w, r, err, "list external bindings failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *ExternalHandler) CreateBinding(w http.ResponseWriter, r *http.Request) {
	var req external.CreateBindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	resp, err := h.service.CreateBinding(r.Context(), req)
	if err != nil {
		writeExternalError(w, r, err, "create external binding failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, resp)
}

func writeExternalError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, external.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, external.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, external.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	case errors.Is(err, external.ErrForbidden):
		api.WriteError(w, r, http.StatusForbidden, api.ErrorForbidden, message, nil)
	case errors.Is(err, external.ErrExternalAutomation):
		api.WriteError(w, r, http.StatusBadGateway, api.ErrorExternalAutomation, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}

func (h *ExternalHandler) RotateCallbackToken(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *ExternalHandler) UpdateCallbackAuth(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *ExternalHandler) ReceiveCallback(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *ExternalHandler) ListCallbackLogs(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *ExternalHandler) TestCallback(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}
