package generation

import (
	"context"
	"errors"
)

var (
	ErrValidation          = errors.New("validation error")
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrIdempotencyConflict = errors.New("idempotency conflict")
	ErrWorkflowRunFailed   = errors.New("workflow run failed")
	ErrLLMProviderError    = errors.New("llm provider error")
)

type Service interface {
	CreateGenerationRun(ctx context.Context, projectID string, req CreateGenerationRunRequest, workflowRunID string, idempotencyKey string) (CreateGenerationRunResponse, error)
	CreateBatchGenerationRuns(ctx context.Context, projectID string, req CreateBatchGenerationRunsRequest, workflowRunIDs []string, idempotencyKey string) (CreateBatchGenerationRunsResponse, error)
	ListGenerationRuns(ctx context.Context, projectID string, req ListGenerationRunsRequest) (PagedGenerationRunsResponse, error)
	GetGenerationRun(ctx context.Context, id string) (GenerationRunDetailResponse, error)
	RetryGenerationRun(ctx context.Context, id string, req RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (RetryGenerationRunResponse, error)
	ListContentItems(ctx context.Context, projectID string, req ListContentItemsRequest) (PagedContentItemsResponse, error)
	GetContentItem(ctx context.Context, id string) (ContentItemDetailResponse, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) CreateGenerationRun(ctx context.Context, projectID string, req CreateGenerationRunRequest, workflowRunID string, idempotencyKey string) (CreateGenerationRunResponse, error) {
	return CreateGenerationRunResponse{}, ErrWorkflowRunFailed
}

func (s *service) CreateBatchGenerationRuns(ctx context.Context, projectID string, req CreateBatchGenerationRunsRequest, workflowRunIDs []string, idempotencyKey string) (CreateBatchGenerationRunsResponse, error) {
	return CreateBatchGenerationRunsResponse{}, ErrWorkflowRunFailed
}

func (s *service) ListGenerationRuns(ctx context.Context, projectID string, req ListGenerationRunsRequest) (PagedGenerationRunsResponse, error) {
	return PagedGenerationRunsResponse{}, ErrNotFound
}

func (s *service) GetGenerationRun(ctx context.Context, id string) (GenerationRunDetailResponse, error) {
	return GenerationRunDetailResponse{}, ErrNotFound
}

func (s *service) RetryGenerationRun(ctx context.Context, id string, req RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (RetryGenerationRunResponse, error) {
	return RetryGenerationRunResponse{}, ErrWorkflowRunFailed
}

func (s *service) ListContentItems(ctx context.Context, projectID string, req ListContentItemsRequest) (PagedContentItemsResponse, error) {
	return PagedContentItemsResponse{}, ErrNotFound
}

func (s *service) GetContentItem(ctx context.Context, id string) (ContentItemDetailResponse, error) {
	return ContentItemDetailResponse{}, ErrNotFound
}
