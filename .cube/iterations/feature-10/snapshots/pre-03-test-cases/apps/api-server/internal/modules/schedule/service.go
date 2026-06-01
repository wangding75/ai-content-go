package schedule

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
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

type scheduleService struct {
	mu             sync.RWMutex
	schedules      []WorkflowScheduleResponse
	inputTemplates map[string]map[string]any
	triggers       []ScheduleTriggerResponse
	nextID         int
}

func NewService() Service {
	return &scheduleService{inputTemplates: map[string]map[string]any{}}
}

func (s *scheduleService) ListSchedules(_ context.Context, req ListSchedulesRequest) (PagedSchedulesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]WorkflowScheduleResponse, 0, len(s.schedules))
	for _, item := range s.schedules {
		if req.ProjectID != "" && item.ProjectID != req.ProjectID {
			continue
		}
		if req.Enabled != nil && item.Enabled != *req.Enabled {
			continue
		}
		items = append(items, item)
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	return PagedSchedulesResponse{Items: items, Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: len(items), HasNext: false}}, nil
}

func (s *scheduleService) CreateSchedule(_ context.Context, req CreateScheduleRequest, _ string) (CreateScheduleResponse, error) {
	if req.ProjectID == "" || req.TemplateVersionID == "" || req.CronExpression == "" || req.DailyContentCount < 0 {
		return CreateScheduleResponse{}, ErrValidation
	}
	dailyContentCount := req.DailyContentCount
	if dailyContentCount == 0 {
		dailyContentCount = 5
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	now := time.Now().UTC()
	id := fmt.Sprintf("schedule-%d", s.nextID)
	nextRunAt := now.Add(24 * time.Hour)
	s.schedules = append(s.schedules, WorkflowScheduleResponse{
		ID:                id,
		ProjectID:         req.ProjectID,
		TemplateVersionID: req.TemplateVersionID,
		CronExpression:    req.CronExpression,
		Enabled:           false,
		NextRunAt:         nextRunAt,
		DailyContentCount: dailyContentCount,
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	s.inputTemplates[id] = copyMap(req.InputTemplate)
	return CreateScheduleResponse{ScheduleID: id, NextRunAt: nextRunAt, DailyContentCount: dailyContentCount, Status: "disabled"}, nil
}

func (s *scheduleService) EnableSchedule(_ context.Context, id string, _ ToggleScheduleRequest, _ string) (ToggleScheduleResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.findScheduleIndex(id)
	if index < 0 {
		return ToggleScheduleResponse{}, ErrNotFound
	}
	if s.schedules[index].Enabled {
		return ToggleScheduleResponse{}, ErrConflict
	}
	previous := s.schedules[index].Enabled
	s.schedules[index].Enabled = true
	s.schedules[index].UpdatedAt = time.Now().UTC()
	return ToggleScheduleResponse{PreviousEnabled: previous, CurrentEnabled: true, NextRunAt: s.schedules[index].NextRunAt, OperationLogID: fmt.Sprintf("schedule-op-%d", time.Now().UnixNano())}, nil
}

func (s *scheduleService) DisableSchedule(_ context.Context, id string, _ ToggleScheduleRequest, _ string) (ToggleScheduleResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.findScheduleIndex(id)
	if index < 0 {
		return ToggleScheduleResponse{}, ErrNotFound
	}
	if !s.schedules[index].Enabled {
		return ToggleScheduleResponse{}, ErrConflict
	}
	previous := s.schedules[index].Enabled
	s.schedules[index].Enabled = false
	s.schedules[index].UpdatedAt = time.Now().UTC()
	return ToggleScheduleResponse{PreviousEnabled: previous, CurrentEnabled: false, NextRunAt: s.schedules[index].NextRunAt, OperationLogID: fmt.Sprintf("schedule-op-%d", time.Now().UnixNano())}, nil
}

func (s *scheduleService) PrepareTestRun(_ context.Context, id string, req TestRunScheduleRequest) (PreparedScheduleRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.findScheduleIndex(id)
	if index < 0 {
		return PreparedScheduleRun{}, ErrNotFound
	}
	s.nextID++
	triggerID := fmt.Sprintf("schedule-trigger-%d", s.nextID)
	s.triggers = append(s.triggers, ScheduleTriggerResponse{ID: triggerID, ScheduleID: id, TriggerType: "manual_test", Status: "prepared", TriggeredAt: time.Now().UTC()})
	input := copyMap(s.inputTemplates[id])
	for key, value := range req.InputOverride {
		input[key] = value
	}
	item := s.schedules[index]
	return PreparedScheduleRun{TriggerLogID: triggerID, ProjectID: item.ProjectID, TemplateVersionID: item.TemplateVersionID, Input: input}, nil
}

func (s *scheduleService) CompleteTrigger(_ context.Context, triggerLogID, workflowRunID, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index := range s.triggers {
		if s.triggers[index].ID == triggerLogID {
			s.triggers[index].WorkflowRunID = workflowRunID
			s.triggers[index].Status = status
			return nil
		}
	}
	return ErrNotFound
}

func (s *scheduleService) ListTriggers(_ context.Context, scheduleID string, req ListTriggersRequest) (PagedScheduleTriggersResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.findScheduleIndex(scheduleID) < 0 {
		return PagedScheduleTriggersResponse{}, ErrNotFound
	}
	items := make([]ScheduleTriggerResponse, 0, len(s.triggers))
	for _, item := range s.triggers {
		if item.ScheduleID == scheduleID {
			items = append(items, item)
		}
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	return PagedScheduleTriggersResponse{Items: items, Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: len(items), HasNext: false}}, nil
}

func (s *scheduleService) findScheduleIndex(id string) int {
	for index, item := range s.schedules {
		if item.ID == id {
			return index
		}
	}
	return -1
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return page, pageSize
}

func copyMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	copied := make(map[string]any, len(value))
	for key, item := range value {
		copied[key] = item
	}
	return copied
}
