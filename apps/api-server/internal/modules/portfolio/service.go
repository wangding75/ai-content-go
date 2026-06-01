package portfolio

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
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

func newID(prefix string) string {
	b := make([]byte, 8)
	rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

func (s *service) CreatePortfolio(ctx context.Context, req CreatePortfolioRequest, idempotencyKey string) (PortfolioDetailResponse, error) {
	if err := validateCreatePortfolio(req); err != nil {
		return PortfolioDetailResponse{}, err
	}
	if idempotencyKey != "" {
		_, refID, conflict, err := s.store.CheckIdempotency(ctx, "portfolio", "create", idempotencyKey, "")
		if err != nil {
			return PortfolioDetailResponse{}, err
		}
		if conflict {
			return PortfolioDetailResponse{}, ErrIdempotencyConflict
		}
		if refID != "" {
			existing, err := s.store.GetPortfolio(ctx, refID)
			if err != nil {
				return PortfolioDetailResponse{}, err
			}
			if existing != nil {
				return *existing, nil
			}
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	item := PortfolioDetailResponse{
		ID: newID("pf_"), Name: req.Name, Description: req.Description,
		ScopeType: req.ScopeType, OwnerID: req.OwnerID, HealthPolicy: req.HealthPolicy,
		Status: req.Status, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.store.CreatePortfolio(ctx, item); err != nil {
		return PortfolioDetailResponse{}, err
	}
	if idempotencyKey != "" {
		s.store.StoreIdempotency(ctx, "portfolio", "create", idempotencyKey, "", refTypePortfolio, item.ID)
	}
	return item, nil
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

func (s *service) UpdatePortfolio(ctx context.Context, portfolioID string, req UpdatePortfolioRequest, idempotencyKey string) (PortfolioDetailResponse, error) {
	existing, err := s.store.GetPortfolio(ctx, portfolioID)
	if err != nil {
		return PortfolioDetailResponse{}, err
	}
	if existing == nil {
		return PortfolioDetailResponse{}, ErrNotFound
	}
	if idempotencyKey != "" {
		_, _, conflict, err := s.store.CheckIdempotency(ctx, "portfolio", "update", idempotencyKey, "")
		if err != nil {
			return PortfolioDetailResponse{}, err
		}
		if conflict {
			return PortfolioDetailResponse{}, ErrIdempotencyConflict
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	item := *existing
	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Description != "" {
		item.Description = req.Description
	}
	if req.Status != "" {
		item.Status = req.Status
	}
	if req.ScopeType != "" {
		item.ScopeType = req.ScopeType
	}
	if req.OwnerID != "" {
		item.OwnerID = req.OwnerID
	}
	if req.HealthPolicy != nil {
		item.HealthPolicy = req.HealthPolicy
	}
	item.UpdatedAt = now
	if err := s.store.UpdatePortfolio(ctx, item); err != nil {
		return PortfolioDetailResponse{}, err
	}
	if idempotencyKey != "" {
		s.store.StoreIdempotency(ctx, "portfolio", "update", idempotencyKey, "", refTypePortfolio, item.ID)
	}
	return item, nil
}

func (s *service) AddProject(ctx context.Context, portfolioID string, req AddPortfolioProjectRequest, idempotencyKey string) (PortfolioProjectResponse, error) {
	if req.ProjectID == "" {
		return PortfolioProjectResponse{}, ErrValidation
	}
	existing, err := s.store.GetPortfolio(ctx, portfolioID)
	if err != nil {
		return PortfolioProjectResponse{}, err
	}
	if existing == nil {
		return PortfolioProjectResponse{}, ErrNotFound
	}
	dup, _ := s.store.GetProject(ctx, portfolioID, req.ProjectID)
	if dup != nil {
		return PortfolioProjectResponse{}, ErrConflict
	}
	if idempotencyKey != "" {
		_, _, conflict, err := s.store.CheckIdempotency(ctx, "portfolio", "add_project", idempotencyKey, "")
		if err != nil {
			return PortfolioProjectResponse{}, err
		}
		if conflict {
			return PortfolioProjectResponse{}, ErrIdempotencyConflict
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	item := PortfolioProjectResponse{
		PortfolioID: portfolioID, ProjectID: req.ProjectID, Role: req.Role,
		Priority: req.Priority, Weight: req.Weight, Note: req.Note,
		AddedBy: req.AddedBy, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.store.AddProject(ctx, item); err != nil {
		return PortfolioProjectResponse{}, err
	}
	if idempotencyKey != "" {
		s.store.StoreIdempotency(ctx, "portfolio", "add_project", idempotencyKey, "", refTypeProject, item.ProjectID)
	}
	return item, nil
}

func (s *service) ListProjects(ctx context.Context, portfolioID string, req ListPortfolioProjectsRequest) (PagedPortfolioProjectsResponse, error) {
	items, total, err := s.store.ListProjects(ctx, portfolioID, req)
	if err != nil {
		return PagedPortfolioProjectsResponse{}, err
	}
	return PagedPortfolioProjectsResponse{Items: items, Pagination: pagination(req.Page, req.PageSize, total)}, nil
}

func (s *service) UpdateProjectPriority(ctx context.Context, portfolioID string, projectID string, req UpdatePortfolioProjectPriorityRequest, idempotencyKey string) (PortfolioProjectResponse, error) {
	existing, err := s.store.GetProject(ctx, portfolioID, projectID)
	if err != nil {
		return PortfolioProjectResponse{}, err
	}
	if existing == nil {
		return PortfolioProjectResponse{}, ErrNotFound
	}
	if idempotencyKey != "" {
		_, _, conflict, err := s.store.CheckIdempotency(ctx, "portfolio", "update_priority", idempotencyKey, "")
		if err != nil {
			return PortfolioProjectResponse{}, err
		}
		if conflict {
			return PortfolioProjectResponse{}, ErrIdempotencyConflict
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	item := *existing
	if req.Role != "" {
		item.Role = req.Role
	}
	if req.Priority > 0 {
		item.Priority = req.Priority
	}
	if req.Weight > 0 {
		item.Weight = req.Weight
	}
	if req.Note != "" {
		item.Note = req.Note
	}
	item.UpdatedAt = now
	if err := s.store.UpdateProject(ctx, item); err != nil {
		return PortfolioProjectResponse{}, err
	}
	if idempotencyKey != "" {
		s.store.StoreIdempotency(ctx, "portfolio", "update_priority", idempotencyKey, "", refTypeProject, item.ProjectID)
	}
	return item, nil
}

func (s *service) RemoveProject(ctx context.Context, portfolioID string, projectID string, req RemovePortfolioProjectRequest) (RemovePortfolioProjectResponse, error) {
	if err := s.store.RemoveProject(ctx, portfolioID, projectID, req); err != nil {
		return RemovePortfolioProjectResponse{}, err
	}
	return RemovePortfolioProjectResponse{
		PortfolioID: portfolioID, ProjectID: projectID,
		Removed: true, OperationID: newID("ppo_"),
	}, nil
}

func (s *service) RecalculateStatusSnapshot(_ context.Context, portfolioID string, _ RecalculatePortfolioStatusSnapshotRequest, idempotencyKey string) (RecalculatePortfolioStatusSnapshotResponse, error) {
	id := newID("pss_")
	return RecalculatePortfolioStatusSnapshotResponse{
		PortfolioID: portfolioID, SnapshotID: id, JobID: id,
		CalculationStatus: SnapshotStatusQueued,
	}, nil
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

const (
	refTypePortfolio = "portfolio"
	refTypeProject   = "project"
)

func validateCreatePortfolio(req CreatePortfolioRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return ErrValidation
	}
	if req.Status != PortfolioStatusActive && req.Status != PortfolioStatusArchived {
		return ErrValidation
	}
	if req.ScopeType != PortfolioScopeManual {
		return ErrValidation
	}
	return nil
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
