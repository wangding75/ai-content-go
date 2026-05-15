package workflow

import (
	"context"
	"errors"
)

var (
	ErrValidation        = errors.New("validation error")
	ErrNotFound          = errors.New("not found")
	ErrConflict          = errors.New("conflict")
	ErrIdempotencyConflict = errors.New("idempotency conflict")
)

// Service is the public interface for workflow template and run management.
type Service interface {
	ListTemplates(ctx context.Context, req ListWorkflowTemplatesRequest) (PagedWorkflowTemplatesResponse, error)
	CreateTemplate(ctx context.Context, req CreateWorkflowTemplateRequest) (CreateWorkflowTemplateResponse, error)
	GetTemplate(ctx context.Context, id string) (WorkflowTemplateResponse, error)
	ListVersions(ctx context.Context, templateID string, req PaginationRequest) (PagedVersionsResponse, error)
	CreateVersion(ctx context.Context, templateID string, req CreateVersionRequest) (CreateVersionResponse, error)
	GetVersion(ctx context.Context, id string) (WorkflowTemplateVersionDetailResponse, error)
	PublishVersion(ctx context.Context, id string, req PublishVersionRequest, idempotencyKey string) (PublishVersionResponse, error)
	ListRuns(ctx context.Context, req ListWorkflowRunsRequest) (PagedWorkflowRunsResponse, error)
	CreateRun(ctx context.Context, req CreateWorkflowRunRequest, idempotencyKey string) (CreateWorkflowRunResponse, error)
	GetRun(ctx context.Context, id string) (WorkflowRunDetailResponse, error)
	GetRunSteps(ctx context.Context, runID string) (ListStepRunsResponse, error)
	CancelRun(ctx context.Context, id string, req CancelRunRequest, idempotencyKey string) (CancelRunResponse, error)
	RetryRun(ctx context.Context, id string, req RetryRunRequest, idempotencyKey string) (RetryRunResponse, error)
}

// EnginePort is the exported interface for the engine package to call workflow internals cross-package.
type EnginePort interface {
	UpdateRunStatus(ctx context.Context, id, status string, output map[string]any, errMsg string) error
	CreateStepRun(ctx context.Context, req CreateStepRunRequest) (WorkflowStepRunResponse, error)
	UpdateStepRunStatus(ctx context.Context, id, status string, output map[string]any, errMsg string) error
	GetRunStepTemplates(ctx context.Context, templateVersionID string) ([]WorkflowStepTemplateResponse, error)
	GetRunForEngine(ctx context.Context, id string) (WorkflowRunResponse, error)
}

// PaginationRequest is a local alias to avoid import cycle (workflow has its own pagination need).
type PaginationRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Sort     string `json:"sort"`
	Order    string `json:"order"`
}

type workflowService struct{}

// NewService returns a Service + EnginePort implementation backed by in-memory storage.
func NewService() interface {
	Service
	EnginePort
} {
	return &workflowService{}
}

func (s *workflowService) ListTemplates(ctx context.Context, req ListWorkflowTemplatesRequest) (PagedWorkflowTemplatesResponse, error) {
	panic("not implemented")
}

func (s *workflowService) CreateTemplate(ctx context.Context, req CreateWorkflowTemplateRequest) (CreateWorkflowTemplateResponse, error) {
	panic("not implemented")
}

func (s *workflowService) GetTemplate(ctx context.Context, id string) (WorkflowTemplateResponse, error) {
	panic("not implemented")
}

func (s *workflowService) ListVersions(ctx context.Context, templateID string, req PaginationRequest) (PagedVersionsResponse, error) {
	panic("not implemented")
}

func (s *workflowService) CreateVersion(ctx context.Context, templateID string, req CreateVersionRequest) (CreateVersionResponse, error) {
	panic("not implemented")
}

func (s *workflowService) GetVersion(ctx context.Context, id string) (WorkflowTemplateVersionDetailResponse, error) {
	panic("not implemented")
}

func (s *workflowService) PublishVersion(ctx context.Context, id string, req PublishVersionRequest, idempotencyKey string) (PublishVersionResponse, error) {
	panic("not implemented")
}

func (s *workflowService) ListRuns(ctx context.Context, req ListWorkflowRunsRequest) (PagedWorkflowRunsResponse, error) {
	panic("not implemented")
}

func (s *workflowService) CreateRun(ctx context.Context, req CreateWorkflowRunRequest, idempotencyKey string) (CreateWorkflowRunResponse, error) {
	panic("not implemented")
}

func (s *workflowService) GetRun(ctx context.Context, id string) (WorkflowRunDetailResponse, error) {
	panic("not implemented")
}

func (s *workflowService) GetRunSteps(ctx context.Context, runID string) (ListStepRunsResponse, error) {
	panic("not implemented")
}

func (s *workflowService) CancelRun(ctx context.Context, id string, req CancelRunRequest, idempotencyKey string) (CancelRunResponse, error) {
	panic("not implemented")
}

func (s *workflowService) RetryRun(ctx context.Context, id string, req RetryRunRequest, idempotencyKey string) (RetryRunResponse, error) {
	panic("not implemented")
}

// EnginePort methods

func (s *workflowService) UpdateRunStatus(ctx context.Context, id, status string, output map[string]any, errMsg string) error {
	panic("not implemented")
}

func (s *workflowService) CreateStepRun(ctx context.Context, req CreateStepRunRequest) (WorkflowStepRunResponse, error) {
	panic("not implemented")
}

func (s *workflowService) UpdateStepRunStatus(ctx context.Context, id, status string, output map[string]any, errMsg string) error {
	panic("not implemented")
}

func (s *workflowService) GetRunStepTemplates(ctx context.Context, templateVersionID string) ([]WorkflowStepTemplateResponse, error) {
	panic("not implemented")
}

func (s *workflowService) GetRunForEngine(ctx context.Context, id string) (WorkflowRunResponse, error) {
	panic("not implemented")
}
