package portfolio

import (
	"context"
	"database/sql"
)

type postgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) Store {
	return &postgresStore{db: db}
}

func (p *postgresStore) CreatePortfolio(context.Context, PortfolioDetailResponse) error {
	return ErrInternal
}

func (p *postgresStore) UpdatePortfolio(context.Context, PortfolioDetailResponse) error {
	return ErrInternal
}

func (p *postgresStore) GetPortfolio(context.Context, string) (*PortfolioDetailResponse, error) {
	return nil, ErrInternal
}

func (p *postgresStore) ListPortfolios(context.Context, ListPortfoliosRequest) ([]PortfolioListItem, int, error) {
	return nil, 0, ErrInternal
}

func (p *postgresStore) AddProject(context.Context, PortfolioProjectResponse) error {
	return ErrInternal
}

func (p *postgresStore) UpdateProject(context.Context, PortfolioProjectResponse) error {
	return ErrInternal
}

func (p *postgresStore) RemoveProject(context.Context, string, string, RemovePortfolioProjectRequest) error {
	return ErrInternal
}

func (p *postgresStore) GetProject(context.Context, string, string) (*PortfolioProjectResponse, error) {
	return nil, ErrInternal
}

func (p *postgresStore) ListProjects(context.Context, string, ListPortfolioProjectsRequest) ([]PortfolioProjectResponse, int, error) {
	return nil, 0, ErrInternal
}

func (p *postgresStore) InsertStatusSnapshot(context.Context, PortfolioStatusSnapshotResponse) error {
	return ErrInternal
}

func (p *postgresStore) ListStatusSnapshots(context.Context, string, ListPortfolioStatusSnapshotsRequest) ([]PortfolioStatusSnapshotResponse, int, error) {
	return nil, 0, ErrInternal
}

func (p *postgresStore) GetLatestStatusSnapshot(context.Context, string) (*PortfolioStatusSnapshotResponse, error) {
	return nil, ErrInternal
}

func (p *postgresStore) QueryHealthSummary(context.Context, string, PortfolioSummaryRequest) (PortfolioHealthSummaryResponse, error) {
	return PortfolioHealthSummaryResponse{}, ErrInternal
}

func (p *postgresStore) QueryCostSummary(context.Context, string, PortfolioSummaryRequest) (PortfolioCostSummaryResponse, error) {
	return PortfolioCostSummaryResponse{}, ErrInternal
}

func (p *postgresStore) QueryStrategySummary(context.Context, string, PortfolioSummaryRequest) (PortfolioStrategySummaryResponse, error) {
	return PortfolioStrategySummaryResponse{}, ErrInternal
}

func (p *postgresStore) CheckIdempotency(context.Context, string, string, string, string) (string, string, bool, error) {
	return "", "", false, ErrInternal
}

func (p *postgresStore) StoreIdempotency(context.Context, string, string, string, string, string, string) error {
	return ErrInternal
}
