package article

import (
	"context"
	"sync"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/generation"
)

type Service interface {
	RegisterPack(ctx context.Context, req RegisterPackRequest, idempotencyKey string) (RegisterPackResponse, error)
	GetPackStatus(ctx context.Context) (ArticlePackStatusResponse, error)
	GetConfig(ctx context.Context, projectID string) (ArticleConfigResponse, error)
	UpdateConfig(ctx context.Context, projectID string, req UpdateArticleConfigRequest, idempotencyKey string) (UpdateArticleConfigResponse, error)
	CreateGenerationRun(ctx context.Context, projectID string, req CreateArticleGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.CreateGenerationRunResponse, error)
	ListGenerationRuns(ctx context.Context, projectID string, req ListGenerationRunsRequest) (PagedArticleGenerationRunResponse, error)
	GetGenerationRun(ctx context.Context, projectID, id string) (ArticleGenerationRunDetailResponse, error)
	RetryGenerationRun(ctx context.Context, projectID, id string, req RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.RetryGenerationRunResponse, error)
	GetContentSnapshot(ctx context.Context, itemID string) (ArticleContentSnapshotResponse, error)
	GetProjectArticleMetrics(ctx context.Context, projectID string) (PagedProjectArticleMetricsResponse, error)
	UpdateProjectArticleMetrics(ctx context.Context, projectID string, req UpdateProjectArticleMetricsRequest, idempotencyKey string) (UpdateProjectArticleMetricsResponse, error)
}

type service struct {
	mu sync.RWMutex
}

func NewService() Service {
	return &service{}
}

func (s *service) RegisterPack(ctx context.Context, req RegisterPackRequest, idempotencyKey string) (RegisterPackResponse, error) {
	panic("implement me")
}

func (s *service) GetPackStatus(ctx context.Context) (ArticlePackStatusResponse, error) {
	panic("implement me")
}

func (s *service) GetConfig(ctx context.Context, projectID string) (ArticleConfigResponse, error) {
	panic("implement me")
}

func (s *service) UpdateConfig(ctx context.Context, projectID string, req UpdateArticleConfigRequest, idempotencyKey string) (UpdateArticleConfigResponse, error) {
	panic("implement me")
}

func (s *service) CreateGenerationRun(ctx context.Context, projectID string, req CreateArticleGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.CreateGenerationRunResponse, error) {
	panic("implement me")
}

func (s *service) ListGenerationRuns(ctx context.Context, projectID string, req ListGenerationRunsRequest) (PagedArticleGenerationRunResponse, error) {
	panic("implement me")
}

func (s *service) GetGenerationRun(ctx context.Context, projectID, id string) (ArticleGenerationRunDetailResponse, error) {
	panic("implement me")
}

func (s *service) RetryGenerationRun(ctx context.Context, projectID, id string, req RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.RetryGenerationRunResponse, error) {
	panic("implement me")
}

func (s *service) GetContentSnapshot(ctx context.Context, itemID string) (ArticleContentSnapshotResponse, error) {
	panic("implement me")
}

func (s *service) GetProjectArticleMetrics(ctx context.Context, projectID string) (PagedProjectArticleMetricsResponse, error) {
	panic("implement me")
}

func (s *service) UpdateProjectArticleMetrics(ctx context.Context, projectID string, req UpdateProjectArticleMetricsRequest, idempotencyKey string) (UpdateProjectArticleMetricsResponse, error) {
	panic("implement me")
}
