package handlers

import (
	"log/slog"
	"net/http"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/system"
)

type SystemHandler struct {
	service system.Service
	logger  *slog.Logger
}

func NewSystemHandler(service system.Service, logger *slog.Logger) *SystemHandler {
	return &SystemHandler{service: service, logger: logger}
}

func (h *SystemHandler) Health(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.Health(r.Context())
	if err != nil {
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, "health check failed", nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *SystemHandler) Info(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.Info(r.Context())
	if err != nil {
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, "system info failed", nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *SystemHandler) ConfigCheck(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.ConfigCheck(r.Context())
	if err != nil {
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, "config check failed", nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *SystemHandler) DBCheck(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.DBCheck(r.Context())
	if err != nil {
		api.WriteError(w, r, http.StatusServiceUnavailable, api.ErrorDependencyUnavailable, "database check failed", nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *SystemHandler) MigrationStatus(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.MigrationStatus(r.Context())
	if err != nil {
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorMigrationReadFailed, "migration status failed", nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}
