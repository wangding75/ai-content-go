package portfolio

import (
	"context"
	"sync"
)

type memoryStore struct {
	mu          sync.Mutex
	portfolios  map[string]PortfolioDetailResponse
	projects    map[string]PortfolioProjectResponse
	snapshots   map[string]PortfolioStatusSnapshotResponse
	idempotency map[string]idempotencyEntry
}

type idempotencyEntry struct {
	hash    string
	refType string
	refID   string
}

func NewMemoryStore() Store {
	return newMemoryStore()
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		portfolios:  map[string]PortfolioDetailResponse{},
		projects:    map[string]PortfolioProjectResponse{},
		snapshots:   map[string]PortfolioStatusSnapshotResponse{},
		idempotency: map[string]idempotencyEntry{},
	}
}

func (m *memoryStore) CreatePortfolio(_ context.Context, item PortfolioDetailResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.portfolios[item.ID] = item
	return nil
}

func (m *memoryStore) UpdatePortfolio(_ context.Context, item PortfolioDetailResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.portfolios[item.ID] = item
	return nil
}

func (m *memoryStore) GetPortfolio(_ context.Context, portfolioID string) (*PortfolioDetailResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.portfolios[portfolioID]
	if !ok {
		return nil, nil
	}
	return &item, nil
}

func (m *memoryStore) ListPortfolios(_ context.Context, _ ListPortfoliosRequest) ([]PortfolioListItem, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := make([]PortfolioListItem, 0, len(m.portfolios))
	for _, item := range m.portfolios {
		items = append(items, item)
	}
	return items, len(items), nil
}

func (m *memoryStore) AddProject(_ context.Context, item PortfolioProjectResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects[projectKey(item.PortfolioID, item.ProjectID)] = item
	return nil
}

func (m *memoryStore) UpdateProject(_ context.Context, item PortfolioProjectResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects[projectKey(item.PortfolioID, item.ProjectID)] = item
	return nil
}

func (m *memoryStore) RemoveProject(_ context.Context, portfolioID string, projectID string, _ RemovePortfolioProjectRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.projects, projectKey(portfolioID, projectID))
	return nil
}

func (m *memoryStore) GetProject(_ context.Context, portfolioID string, projectID string) (*PortfolioProjectResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.projects[projectKey(portfolioID, projectID)]
	if !ok {
		return nil, nil
	}
	return &item, nil
}

func (m *memoryStore) ListProjects(_ context.Context, portfolioID string, _ ListPortfolioProjectsRequest) ([]PortfolioProjectResponse, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := []PortfolioProjectResponse{}
	for _, item := range m.projects {
		if item.PortfolioID == portfolioID {
			items = append(items, item)
		}
	}
	return items, len(items), nil
}

func (m *memoryStore) InsertStatusSnapshot(_ context.Context, item PortfolioStatusSnapshotResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots[item.ID] = item
	return nil
}

func (m *memoryStore) ListStatusSnapshots(_ context.Context, portfolioID string, _ ListPortfolioStatusSnapshotsRequest) ([]PortfolioStatusSnapshotResponse, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := []PortfolioStatusSnapshotResponse{}
	for _, item := range m.snapshots {
		if item.PortfolioID == portfolioID {
			items = append(items, item)
		}
	}
	return items, len(items), nil
}

func (m *memoryStore) GetLatestStatusSnapshot(_ context.Context, portfolioID string) (*PortfolioStatusSnapshotResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range m.snapshots {
		if item.PortfolioID == portfolioID {
			snapshot := item
			return &snapshot, nil
		}
	}
	return nil, nil
}

func (m *memoryStore) QueryHealthSummary(_ context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioHealthSummaryResponse, error) {
	return PortfolioHealthSummaryResponse{PortfolioID: portfolioID, DateRange: req.DateRange}, nil
}

func (m *memoryStore) QueryCostSummary(_ context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioCostSummaryResponse, error) {
	return PortfolioCostSummaryResponse{PortfolioID: portfolioID, DateRange: req.DateRange}, nil
}

func (m *memoryStore) QueryStrategySummary(_ context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioStrategySummaryResponse, error) {
	return PortfolioStrategySummaryResponse{PortfolioID: portfolioID, DateRange: req.DateRange}, nil
}

func (m *memoryStore) CheckIdempotency(_ context.Context, scope string, endpoint string, key string, hash string) (string, string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.idempotency[scope+":"+endpoint+":"+key]
	if !ok {
		return "", "", false, nil
	}
	if entry.hash != hash {
		return "", "", true, nil
	}
	return entry.refType, entry.refID, false, nil
}

func (m *memoryStore) StoreIdempotency(_ context.Context, scope string, endpoint string, key string, hash string, refType string, refID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idempotency[scope+":"+endpoint+":"+key] = idempotencyEntry{hash: hash, refType: refType, refID: refID}
	return nil
}

func projectKey(portfolioID string, projectID string) string {
	return portfolioID + ":" + projectID
}
