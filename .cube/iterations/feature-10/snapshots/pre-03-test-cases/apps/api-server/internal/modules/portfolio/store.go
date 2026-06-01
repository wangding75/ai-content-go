package portfolio

import "context"

type Store interface {
	CreatePortfolio(ctx context.Context, item PortfolioDetailResponse) error
	UpdatePortfolio(ctx context.Context, item PortfolioDetailResponse) error
	GetPortfolio(ctx context.Context, portfolioID string) (*PortfolioDetailResponse, error)
	ListPortfolios(ctx context.Context, req ListPortfoliosRequest) ([]PortfolioListItem, int, error)
	AddProject(ctx context.Context, item PortfolioProjectResponse) error
	UpdateProject(ctx context.Context, item PortfolioProjectResponse) error
	RemoveProject(ctx context.Context, portfolioID string, projectID string, req RemovePortfolioProjectRequest) error
	GetProject(ctx context.Context, portfolioID string, projectID string) (*PortfolioProjectResponse, error)
	ListProjects(ctx context.Context, portfolioID string, req ListPortfolioProjectsRequest) ([]PortfolioProjectResponse, int, error)
	InsertStatusSnapshot(ctx context.Context, item PortfolioStatusSnapshotResponse) error
	ListStatusSnapshots(ctx context.Context, portfolioID string, req ListPortfolioStatusSnapshotsRequest) ([]PortfolioStatusSnapshotResponse, int, error)
	GetLatestStatusSnapshot(ctx context.Context, portfolioID string) (*PortfolioStatusSnapshotResponse, error)
	QueryHealthSummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioHealthSummaryResponse, error)
	QueryCostSummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioCostSummaryResponse, error)
	QueryStrategySummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioStrategySummaryResponse, error)
	CheckIdempotency(ctx context.Context, scope string, endpoint string, key string, hash string) (refType string, refID string, conflict bool, err error)
	StoreIdempotency(ctx context.Context, scope string, endpoint string, key string, hash string, refType string, refID string) error
}
