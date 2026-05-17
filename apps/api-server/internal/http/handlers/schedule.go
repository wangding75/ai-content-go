package handlers

import (
	"log/slog"
	"net/http"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/schedule"
)

type ScheduleHandler struct {
	service schedule.Service
	logger  *slog.Logger
}

func NewScheduleHandler(service schedule.Service, logger *slog.Logger) *ScheduleHandler {
	return &ScheduleHandler{service: service, logger: logger}
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}
