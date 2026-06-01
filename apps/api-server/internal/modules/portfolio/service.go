package portfolio

import (
	"context"
	"fmt"
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

type Service interface {
	CreatePortfolio(ctx context.Context, req CreatePortfolioRequest, idempotencyKey string) (PortfolioDetailResponse, error)
	ListPortfolios(ctx context.Context, req ListPortfoliosRequest) (PagedPortfoliosResponse, error)
	GetPortfolio(ctx context.Context, portfolioID string) (PortfolioDetailResponse, error)
	UpdatePortfolio(ctx context.Context, portfolioID string, req UpdatePortfolioRequest, idempotencyKey string) (PortfolioDetailResponse, error)
	AddProject(ctx context.Context, portfolioID string, req AddPortfolioProjectRequest, idempotencyKey string) (PortfolioProjectResponse, error)
	ListProjects(ctx context.Context, portfolioID string, req ListPortfolioProjectsRequest) (PagedPortfolioProjectsResponse, error)
	UpdateProjectPriority(ctx context.Context, portfolioID string, projectID string, req UpdatePortfolioProjectPriorityRequest, idempotencyKey string) (PortfolioProjectResponse, error)
	RemoveProject(ctx context.Context, portfolioID string, projectID string, req RemovePortfolioProjectRequest) (RemovePortfolioProjectResponse, error)
	RecalculateStatusSnapshot(ctx context.Context, portfolioID string, req RecalculatePortfolioStatusSnapshotRequest, idempotencyKey string) (RecalculatePortfolioStatusSnapshotResponse, error)
	ListStatusSnapshots(ctx context.Context, portfolioID string, req ListPortfolioStatusSnapshotsRequest) (PagedPortfolioStatusSnapshotsResponse, error)
	GetHealthSummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioHealthSummaryResponse, error)
	GetCostSummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioCostSummaryResponse, error)
	GetStrategySummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioStrategySummaryResponse, error)
}

type service struct {
	store Store
}

func NewService(stores ...Store) Service {
	store := Store(newMemoryStore())
	if len(stores) > 0 && stores[0] != nil {
		store = stores[0]
	}
	return &service{store: store}
}

func (s *service) CreatePortfolio(ctx context.Context, req CreatePortfolioRequest, _ string) (PortfolioDetailResponse, error) {
	now := time.Now().UTC()
	item := PortfolioDetailResponse{ID: fmt.Sprintf("pf_%d", now.UnixNano()), Name: req.Name, Description: req.Description, ScopeType: req.ScopeType, OwnerID: req.OwnerID, HealthPolicy: req.HealthPolicy, Status: req.Status, CreatedAt: now.Format(time.RFC3339), UpdatedAt: now.Format(time.RFC3339)}
	return item, s.store.CreatePortfolio(ctx, item)
}

func (s *service) ListPortfolios(ctx context.Context, req ListPortfoliosRequest) (PagedPortfoliosResponse, error) {
	items, total, err := s.store.ListPortfolios(ctx, req)
	if err != nil {
		return PagedPortfoliosResponse{}, err
	}
	return PagedPortfoliosResponse{Items: items, Pagination: pagination(req.Page, req.PageSize, total)}, nil
}

func (s *service) GetPortfolio(ctx context.Context, portfolioID string) (PortfolioDetailResponse, error) {
	item, err := s.store.GetPortfolio(ctx, portfolioID)
	if err != nil {
		return PortfolioDetailResponse{}, err
	}
	if item == nil {
		return PortfolioDetailResponse{}, ErrNotFound
	}
	return *item, nil
}

func (s *service) UpdatePortfolio(ctx context.Context, portfolioID string, req UpdatePortfolioRequest, _ string) (PortfolioDetailResponse, error) {
	now := time.Now().UTC()
	item := PortfolioDetailResponse{ID: portfolioID, Name: req.Name, Description: req.Description, ScopeType: req.ScopeType, OwnerID: req.OwnerID, HealthPolicy: req.HealthPolicy, Status: req.Status, UpdatedAt: now.Format(time.RFC3339)}
	return item, s.store.UpdatePortfolio(ctx, item)
}

func (s *service) AddProject(ctx context.Context, portfolioID string, req AddPortfolioProjectRequest, _ string) (PortfolioProjectResponse, error) {
	item := PortfolioProjectResponse{PortfolioID: portfolioID, ProjectID: req.ProjectID, Role: req.Role, Priority: req.Priority, Weight: req.Weight, Note: req.Note, AddedBy: req.AddedBy}
	return item, s.store.AddProject(ctx, item)
}

func (s *service) ListProjects(ctx context.Context, portfolioID string, req ListPortfolioProjectsRequest) (PagedPortfolioProjectsResponse, error) {
	items, total, err := s.store.ListProjects(ctx, portfolioID, req)
	if err != nil {
		return PagedPortfolioProjectsResponse{}, err
	}
	return PagedPortfolioProjectsResponse{Items: items, Pagination: pagination(req.Page, req.PageSize, total)}, nil
}

func (s *service) UpdateProjectPriority(ctx context.Context, portfolioID string, projectID string, req UpdatePortfolioProjectPriorityRequest, _ string) (PortfolioProjectResponse, error) {
	item := PortfolioProjectResponse{PortfolioID: portfolioID, ProjectID: projectID, Role: req.Role, Priority: req.Priority, Weight: req.Weight, Note: req.Note}
	return item, s.store.UpdateProject(ctx, item)
}

func (s *service) RemoveProject(ctx context.Context, portfolioID string, projectID string, req RemovePortfolioProjectRequest) (RemovePortfolioProjectResponse, error) {
	if err := s.store.RemoveProject(ctx, portfolioID, projectID, req); err != nil {
		return RemovePortfolioProjectResponse{}, err
	}
	return RemovePortfolioProjectResponse{PortfolioID: portfolioID, ProjectID: projectID, Removed: true, OperationID: fmt.Sprintf("ppo_%d", time.Now().UnixNano())}, nil
}

func (s *service) RecalculateStatusSnapshot(_ context.Context, portfolioID string, _ RecalculatePortfolioStatusSnapshotRequest, _ string) (RecalculatePortfolioStatusSnapshotResponse, error) {
	id := fmt.Sprintf("pss_%d", time.Now().UnixNano())
	return RecalculatePortfolioStatusSnapshotResponse{PortfolioID: portfolioID, SnapshotID: id, JobID: id, CalculationStatus: SnapshotStatusQueued}, nil
}

func (s *service) ListStatusSnapshots(ctx context.Context, portfolioID string, req ListPortfolioStatusSnapshotsRequest) (PagedPortfolioStatusSnapshotsResponse, error) {
	items, total, err := s.store.ListStatusSnapshots(ctx, portfolioID, req)
	if err != nil {
		return PagedPortfolioStatusSnapshotsResponse{}, err
	}
	return PagedPortfolioStatusSnapshotsResponse{Items: items, Pagination: pagination(req.Page, req.PageSize, total)}, nil
}

func (s *service) GetHealthSummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioHealthSummaryResponse, error) {
	return s.store.QueryHealthSummary(ctx, portfolioID, req)
}

func (s *service) GetCostSummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioCostSummaryResponse, error) {
	return s.store.QueryCostSummary(ctx, portfolioID, req)
}

func (s *service) GetStrategySummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioStrategySummaryResponse, error) {
	return s.store.QueryStrategySummary(ctx, portfolioID, req)
}

func pagination(page int, pageSize int, total int) content.PaginationResponse {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return content.PaginationResponse{Page: page, PageSize: pageSize, Total: total, HasNext: page*pageSize < total}
}
