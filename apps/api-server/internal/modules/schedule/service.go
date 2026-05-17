package schedule

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrValidation = errors.New("validation error")

type Service interface {
	CreateSchedule(ctx context.Context, req CreateScheduleRequest) (ScheduleResponse, error)
}

type scheduleService struct {
	mu   sync.Mutex
	next int
}

func NewService() Service {
	return &scheduleService{}
}

func (s *scheduleService) CreateSchedule(_ context.Context, req CreateScheduleRequest) (ScheduleResponse, error) {
	if req.TemplateVersionID == "" {
		return ScheduleResponse{}, ErrValidation
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	return ScheduleResponse{ScheduleID: fmt.Sprintf("sched-%d", s.next), Status: "not_implemented"}, nil
}
