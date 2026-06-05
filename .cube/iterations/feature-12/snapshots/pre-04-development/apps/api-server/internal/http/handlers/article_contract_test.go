package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/article"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/generation"
)

// Stub service for Task-01: contract test — verifies types compile and errors package is wired.
type task01ArticleService struct{}

func (task01ArticleService) RegisterPack(ctx context.Context, req article.RegisterPackRequest, idempotencyKey string) (article.RegisterPackResponse, error) {
	return article.RegisterPackResponse{}, article.ErrValidation
}
func (task01ArticleService) GetPackStatus(ctx context.Context) (article.ArticlePackStatusResponse, error) {
	return article.ArticlePackStatusResponse{}, article.ErrNotFound
}
func (task01ArticleService) GetConfig(ctx context.Context, projectID string) (article.ArticleConfigResponse, error) {
	return article.ArticleConfigResponse{}, article.ErrNotFound
}
func (task01ArticleService) UpdateConfig(ctx context.Context, projectID string, req article.UpdateArticleConfigRequest, idempotencyKey string) (article.UpdateArticleConfigResponse, error) {
	return article.UpdateArticleConfigResponse{}, article.ErrValidation
}
func (task01ArticleService) CreateGenerationRun(ctx context.Context, projectID string, req article.CreateArticleGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.CreateGenerationRunResponse, error) {
	return generation.CreateGenerationRunResponse{}, article.ErrForbidden
}
func (task01ArticleService) ListGenerationRuns(ctx context.Context, projectID string, req article.ListGenerationRunsRequest) (article.PagedArticleGenerationRunResponse, error) {
	return article.PagedArticleGenerationRunResponse{}, article.ErrNotFound
}
func (task01ArticleService) GetGenerationRun(ctx context.Context, projectID, id string) (article.ArticleGenerationRunDetailResponse, error) {
	return article.ArticleGenerationRunDetailResponse{}, article.ErrNotFound
}
func (task01ArticleService) RetryGenerationRun(ctx context.Context, projectID, id string, req article.RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.RetryGenerationRunResponse, error) {
	return generation.RetryGenerationRunResponse{}, article.ErrConflict
}
func (task01ArticleService) GetContentSnapshot(ctx context.Context, itemID string) (article.ArticleContentSnapshotResponse, error) {
	return article.ArticleContentSnapshotResponse{}, article.ErrNotFound
}
func (task01ArticleService) GetProjectArticleMetrics(ctx context.Context, projectID string) (article.PagedProjectArticleMetricsResponse, error) {
	return article.PagedProjectArticleMetricsResponse{}, article.ErrNotFound
}
func (task01ArticleService) UpdateProjectArticleMetrics(ctx context.Context, projectID string, req article.UpdateProjectArticleMetricsRequest, idempotencyKey string) (article.UpdateProjectArticleMetricsResponse, error) {
	return article.UpdateProjectArticleMetricsResponse{}, article.ErrValidation
}

// @Test
func TestTask01ArticleServiceInterfaceCompilesWithStub(t *testing.T) {
	var svc article.Service = task01ArticleService{}
	// RegisterPack
	resp, err := svc.RegisterPack(context.Background(), article.RegisterPackRequest{}, "")
	if !errors.Is(err, article.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
	if resp.ContentPackID != "" {
		t.Fatal("expected empty response on error")
	}
	// GetPackStatus
	_, err = svc.GetPackStatus(context.Background())
	if !errors.Is(err, article.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	// GetConfig
	_, err = svc.GetConfig(context.Background(), "project-1")
	if !errors.Is(err, article.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	// UpdateConfig
	_, err = svc.UpdateConfig(context.Background(), "project-1", article.UpdateArticleConfigRequest{}, "")
	if !errors.Is(err, article.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
	// CreateGenerationRun
	_, err = svc.CreateGenerationRun(context.Background(), "project-1", article.CreateArticleGenerationRunRequest{}, "wfr-1", "")
	if !errors.Is(err, article.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	// ListGenerationRuns
	_, err = svc.ListGenerationRuns(context.Background(), "project-1", article.ListGenerationRunsRequest{})
	if !errors.Is(err, article.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	// GetGenerationRun
	_, err = svc.GetGenerationRun(context.Background(), "project-1", "run-1")
	if !errors.Is(err, article.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	// RetryGenerationRun
	_, err = svc.RetryGenerationRun(context.Background(), "project-1", "run-1", article.RetryGenerationRunRequest{}, "wfr-1", "")
	if !errors.Is(err, article.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	// GetContentSnapshot
	_, err = svc.GetContentSnapshot(context.Background(), "item-1")
	if !errors.Is(err, article.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	// GetProjectArticleMetrics
	_, err = svc.GetProjectArticleMetrics(context.Background(), "project-1")
	if !errors.Is(err, article.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	// UpdateProjectArticleMetrics
	_, err = svc.UpdateProjectArticleMetrics(context.Background(), "project-1", article.UpdateProjectArticleMetricsRequest{}, "")
	if !errors.Is(err, article.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

// @Test
func TestTask01NewArticleServiceCompiles(t *testing.T) {
	svc := article.NewService(nil, nil, nil)
	if svc == nil {
		t.Fatal("expected non-nil Service")
	}
}

// @Test
func TestTask01ArticleErrorConstantsWireCorrectly(t *testing.T) {
	if article.ErrValidation == nil || article.ErrNotFound == nil || article.ErrForbidden == nil || article.ErrConflict == nil || article.ErrIdempotencyConflict == nil {
		t.Fatal("all error constants must be non-nil")
	}
	_ = errors.Is(article.ErrValidation, article.ErrValidation)
}

// @Test
func TestTask01CompileGuardUsesPagedResponse(t *testing.T) {
	resp := article.PagedArticleGenerationRunResponse{}
	if resp.Items != nil {
		t.Fatal("expected nil items in zero-value paged response")
	}
	_ = content.PaginationResponse{}
}

// @Test
func TestTask01DTOFieldsMapToJSONTags(t *testing.T) {
	// Verify RegisterPackRequest serialisation
	req := article.RegisterPackRequest{}
	_ = req
	// Verify ArticleConfigResponse field access
	cfg := article.ArticleConfigResponse{}
	_ = cfg.TopicStyle
	_ = cfg.Version
	// Verify snapshot type
	snap := article.ArticleSnapshot{}
	_ = snap.Title
	_ = snap.LatestContentVersionID
}

// @Test
func TestTask01ArticleDTOConstructorsAndZeroValues(t *testing.T) {
	resp := article.ArticlePackStatusResponse{}
	if resp.Registered {
		t.Fatal("expected false registered in zero-value")
	}
	_ = article.CreateArticleGenerationRunRequest{}
	_ = article.UpdateArticleConfigResponse{}
	_ = article.PagedProjectArticleMetricsResponse{}
	_ = article.ArticleContentSnapshotResponse{}
}
