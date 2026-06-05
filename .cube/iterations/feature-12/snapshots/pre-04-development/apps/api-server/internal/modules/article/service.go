package article

import (
	"context"
	"fmt"
	"sync"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/generation"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/metrics"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type idempotencyRecord struct {
	key       string
	hash      string
	contentID string
}

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

	contentSvc content.Service
	wfSvc      workflow.Service
	metricsSvc metrics.Service

	packRegistered bool
	packID         string
	contentTypeID  string
	wfVersionIDs   []string
	metricTmplIDs  []string

	// Article configs per project
	projectConfigs     map[string]ArticleConfigResponse
	projectConfigVer   int
	projectMetricsCfgs map[string]projectMetricsConfig

	// Generation runs per project
	projectGenRuns  map[string][]generationRunRecord
	genRunSeq       int

	// Idempotency tracking
	idempotency map[string]idempotencyRecord
}

type projectMetricsConfig struct {
	EnabledCodes []string
	VersionID    string
}

type generationRunRecord struct {
	ID              string
	WorkflowRunID    string
	Status          string
	Topic           string
	TargetPlatform  string
	CreatedAt       string
	RetryOf         string
	Error           string
	ArticleSnapshot *ArticleSnapshot
	ContentItemID   string
	ContentVersionID string
}

func NewService(contentSvc content.Service, wfSvc workflow.Service, metricsSvc metrics.Service) Service {
	return &service{
		contentSvc:          contentSvc,
		wfSvc:               wfSvc,
		metricsSvc:          metricsSvc,
		projectConfigs:      make(map[string]ArticleConfigResponse),
		projectMetricsCfgs:  make(map[string]projectMetricsConfig),
		projectGenRuns:      make(map[string][]generationRunRecord),
		idempotency:         make(map[string]idempotencyRecord),
	}
}

func newID(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, idCounter)
}

var idCounter int

func nextID(prefix string) string {
	idCounter++
	return fmt.Sprintf("%s%d", prefix, idCounter)
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
