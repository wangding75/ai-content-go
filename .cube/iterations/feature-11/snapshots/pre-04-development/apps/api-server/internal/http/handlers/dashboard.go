package handlers

import (
	"log/slog"
	"net/http"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/dashboard"
)

type DashboardHandler struct {
	service dashboard.Service
	logger  *slog.Logger
}

func NewDashboardHandler(service dashboard.Service, logger *slog.Logger) *DashboardHandler {
	return &DashboardHandler{service: service, logger: logger}
}

func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.Summary(r.Context())
	if err != nil {
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, "dashboard summary failed", nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}
