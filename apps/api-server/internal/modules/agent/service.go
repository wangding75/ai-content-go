package agent

import (
	"context"
	"errors"
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

type agentService struct{}

func NewService() Service {
	return &agentService{}
}

func (s *agentService) CreateTask(ctx context.Context, req CreateAgentTaskRequest) (AgentTaskResponse, error) {
	panic("not implemented")
}

func (s *agentService) UpdateTask(ctx context.Context, id string, update UpdateAgentTaskRequest) error {
	panic("not implemented")
}

func (s *agentService) ListTasks(ctx context.Context, req ListAgentTasksRequest) (PagedAgentTasksResponse, error) {
	panic("not implemented")
}

func (s *agentService) GetTask(ctx context.Context, id string) (AgentTaskDetailResponse, error) {
	panic("not implemented")
}
