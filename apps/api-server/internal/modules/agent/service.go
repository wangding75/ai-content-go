package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
)

type Service interface {
	CreateTask(ctx context.Context, req CreateAgentTaskRequest) (AgentTaskResponse, error)
	UpdateTask(ctx context.Context, id string, update UpdateAgentTaskRequest) error
	ListTasks(ctx context.Context, req ListAgentTasksRequest) (PagedAgentTasksResponse, error)
	GetTask(ctx context.Context, id string) (AgentTaskDetailResponse, error)
}

type agentTaskRecord struct {
	AgentTaskDetailResponse
	llmCallLogIDs []string
}

type agentService struct {
	mu      sync.RWMutex
	tasks   []agentTaskRecord
	taskNext int
}

func NewService() Service {
	return &agentService{}
}

func (s *agentService) CreateTask(_ context.Context, req CreateAgentTaskRequest) (AgentTaskResponse, error) {
	if req.WorkflowRunID == "" || req.StepRunID == "" {
		return AgentTaskResponse{}, ErrValidation
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.taskNext++
	id := fmt.Sprintf("at-%d", s.taskNext)
	now := time.Now()
	rec := agentTaskRecord{
		AgentTaskDetailResponse: AgentTaskDetailResponse{
			AgentTaskResponse: AgentTaskResponse{
				ID:            id,
				WorkflowRunID: req.WorkflowRunID,
				StepRunID:     req.StepRunID,
				AgentCode:     req.AgentCode,
				Status:        "pending",
				CreatedAt:     now,
			},
			Input:         req.Input,
			LLMCallLogIDs: []string{},
		},
		llmCallLogIDs: []string{},
	}
	s.tasks = append(s.tasks, rec)
	return rec.AgentTaskResponse, nil
}

func (s *agentService) UpdateTask(_ context.Context, id string, update UpdateAgentTaskRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.tasks {
		if t.ID == id {
			if update.Status != "" {
				s.tasks[i].Status = update.Status
			}
			if update.Output != nil {
				s.tasks[i].Output = update.Output
			}
			if update.Error != "" {
				s.tasks[i].Error = update.Error
			}
			return nil
		}
	}
	return ErrNotFound
}

func (s *agentService) ListTasks(_ context.Context, req ListAgentTasksRequest) (PagedAgentTasksResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var items []AgentTaskResponse
	for _, t := range s.tasks {
		if req.WorkflowRunID != "" && t.WorkflowRunID != req.WorkflowRunID {
			continue
		}
		if req.StepRunID != "" && t.StepRunID != req.StepRunID {
			continue
		}
		if req.AgentCode != "" && t.AgentCode != req.AgentCode {
			continue
		}
		if req.Status != "" && t.Status != req.Status {
			continue
		}
		items = append(items, t.AgentTaskResponse)
	}
	if items == nil {
		items = []AgentTaskResponse{}
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	total := len(items)
	return PagedAgentTasksResponse{
		Items:      items,
		Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: total, HasNext: false},
	}, nil
}

func (s *agentService) GetTask(_ context.Context, id string) (AgentTaskDetailResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tasks {
		if t.ID == id {
			detail := t.AgentTaskDetailResponse
			detail.LLMCallLogCount = len(t.llmCallLogIDs)
			if len(t.llmCallLogIDs) > 0 {
				detail.LLMCallLogIDs = append([]string(nil), t.llmCallLogIDs...)
			} else {
				detail.LLMCallLogIDs = []string{}
			}
			return detail, nil
		}
	}
	return AgentTaskDetailResponse{}, ErrNotFound
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
