package socialpost

import (
	"context"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/metrics"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type Service interface {
	GetPackStatus(ctx context.Context) (SocialPostPackStatusResponse, error)
	RegisterPack(ctx context.Context, req RegisterSocialPostPackRequest, idempotencyKey string) (RegisterSocialPostPackResponse, error)

	GetConfig(ctx context.Context, projectID string) (SocialPostConfigResponse, error)
	UpdateConfig(ctx context.Context, projectID string, req UpdateSocialPostConfigRequest, idempotencyKey string) (UpdateSocialPostConfigResponse, error)

	CreateGenerationRun(ctx context.Context, projectID string, req CreateSocialPostGenerationRunRequest, idempotencyKey string) (CreateSocialPostGenerationRunResponse, error)
	GetGenerationRun(ctx context.Context, projectID, generationRunID string) (SocialPostGenerationRunDetailResponse, error)

	ListVariants(ctx context.Context, projectID string, req ListSocialPostVariantsRequest) (PagedSocialPostVariantsResponse, error)
	SelectVariant(ctx context.Context, projectID, variantID string, req SelectSocialPostVariantRequest, idempotencyKey string) (SelectSocialPostVariantResponse, error)

	GenerateTags(ctx context.Context, projectID string, req GenerateSocialPostTagsRequest, idempotencyKey string) (GenerateSocialPostAssetResponse, error)
	GenerateCoverCopy(ctx context.Context, projectID string, req GenerateSocialPostCoverCopyRequest, idempotencyKey string) (GenerateSocialPostAssetResponse, error)
	GetAssets(ctx context.Context, projectID string, req GetSocialPostAssetsRequest) (SocialPostAssetsResponse, error)
}

type service struct {
	store      Store
	contentSvc content.Service
	wfSvc      workflow.Service
	metricsSvc metrics.Service
	submitter  engine.Submitter
}

func NewService(stores ...Store) Service {
	var store Store
	if len(stores) > 0 {
		store = stores[0]
	} else {
		store = NewMemoryStore()
	}
	return &service{store: store}
}

func (s *service) GetPackStatus(ctx context.Context) (SocialPostPackStatusResponse, error) {
	return SocialPostPackStatusResponse{}, ErrInternal
}

func (s *service) RegisterPack(ctx context.Context, req RegisterSocialPostPackRequest, idempotencyKey string) (RegisterSocialPostPackResponse, error) {
	return RegisterSocialPostPackResponse{}, ErrInternal
}

func (s *service) GetConfig(ctx context.Context, projectID string) (SocialPostConfigResponse, error) {
	return SocialPostConfigResponse{}, ErrInternal
}

func (s *service) UpdateConfig(ctx context.Context, projectID string, req UpdateSocialPostConfigRequest, idempotencyKey string) (UpdateSocialPostConfigResponse, error) {
	return UpdateSocialPostConfigResponse{}, ErrInternal
}

func (s *service) CreateGenerationRun(ctx context.Context, projectID string, req CreateSocialPostGenerationRunRequest, idempotencyKey string) (CreateSocialPostGenerationRunResponse, error) {
	return CreateSocialPostGenerationRunResponse{}, ErrInternal
}

func (s *service) GetGenerationRun(ctx context.Context, projectID, generationRunID string) (SocialPostGenerationRunDetailResponse, error) {
	return SocialPostGenerationRunDetailResponse{}, ErrInternal
}

func (s *service) ListVariants(ctx context.Context, projectID string, req ListSocialPostVariantsRequest) (PagedSocialPostVariantsResponse, error) {
	return PagedSocialPostVariantsResponse{}, ErrInternal
}

func (s *service) SelectVariant(ctx context.Context, projectID, variantID string, req SelectSocialPostVariantRequest, idempotencyKey string) (SelectSocialPostVariantResponse, error) {
	return SelectSocialPostVariantResponse{}, ErrInternal
}

func (s *service) GenerateTags(ctx context.Context, projectID string, req GenerateSocialPostTagsRequest, idempotencyKey string) (GenerateSocialPostAssetResponse, error) {
	return GenerateSocialPostAssetResponse{}, ErrInternal
}

func (s *service) GenerateCoverCopy(ctx context.Context, projectID string, req GenerateSocialPostCoverCopyRequest, idempotencyKey string) (GenerateSocialPostAssetResponse, error) {
	return GenerateSocialPostAssetResponse{}, ErrInternal
}

func (s *service) GetAssets(ctx context.Context, projectID string, req GetSocialPostAssetsRequest) (SocialPostAssetsResponse, error) {
	return SocialPostAssetsResponse{}, ErrInternal
}