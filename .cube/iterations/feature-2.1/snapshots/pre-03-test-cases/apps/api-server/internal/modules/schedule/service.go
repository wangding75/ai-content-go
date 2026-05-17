package schedule

import (
	"context"
	"errors"
)

var (
	ErrValidation          = errors.New("validation error")
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrIdempotencyConflict = errors.New("idempotency conflict")
)

type Service interface {
	ListSchedules(ctx context.Context, req ListSchedulesRequest) (PagedSchedulesResponse, error)
	CreateSchedule(ctx context.Context, req CreateScheduleRequest, idempotencyKey string) (CreateScheduleResponse, error)
	EnableSchedule(ctx context.Context, id string, req ToggleScheduleRequest, idempotencyKey string) (ToggleScheduleResponse, error)
	DisableSchedule(ctx context.Context, id string, req ToggleScheduleRequest, idempotencyKey string) (ToggleScheduleResponse, error)
	PrepareTestRun(ctx context.Context, id string, req TestRunScheduleRequest) (PreparedScheduleRun, error)
	CompleteTrigger(ctx context.Context, triggerLogID, workflowRunID, status string) error
	ListTriggers(ctx context.Context, scheduleID string, req ListTriggersRequest) (PagedScheduleTriggersResponse, error)
}

type scheduleService struct{}

func NewService() Service {
	return &scheduleService{}
}

func (s *scheduleService) ListSchedules(context.Context, ListSchedulesRequest) (PagedSchedulesResponse, error) {
	return PagedSchedulesResponse{}, ErrValidation
}

func (s *scheduleService) CreateSchedule(context.Context, CreateScheduleRequest, string) (CreateScheduleResponse, error) {
	return CreateScheduleResponse{}, ErrValidation
}

func (s *scheduleService) EnableSchedule(context.Context, string, ToggleScheduleRequest, string) (ToggleScheduleResponse, error) {
	return ToggleScheduleResponse{}, ErrValidation
}

func (s *scheduleService) DisableSchedule(context.Context, string, ToggleScheduleRequest, string) (ToggleScheduleResponse, error) {
	return ToggleScheduleResponse{}, ErrValidation
}

func (s *scheduleService) PrepareTestRun(context.Context, string, TestRunScheduleRequest) (PreparedScheduleRun, error) {
	return PreparedScheduleRun{}, ErrValidation
}

func (s *scheduleService) CompleteTrigger(context.Context, string, string, string) error {
	return ErrValidation
}

func (s *scheduleService) ListTriggers(context.Context, string, ListTriggersRequest) (PagedScheduleTriggersResponse, error) {
	return PagedScheduleTriggersResponse{}, ErrValidation
}
