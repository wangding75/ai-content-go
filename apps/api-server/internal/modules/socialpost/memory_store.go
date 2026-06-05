package socialpost

import (
	"context"
	"sync"
)

type memoryStore struct {
	mu sync.RWMutex

	extensions  map[string]*SocialPostConfigRow
	variants    map[string]*SocialPostVariantRow
	assets      map[string]*SocialPostAssetRow
	operationLogs map[string]string
	idempotency  map[string]idempotencyMemRecord
}

type idempotencyMemRecord struct {
	refType string
	refID   string
	hash    string
}

func NewMemoryStore() Store {
	return &memoryStore{
		extensions:    make(map[string]*SocialPostConfigRow),
		variants:      make(map[string]*SocialPostVariantRow),
		assets:        make(map[string]*SocialPostAssetRow),
		operationLogs: make(map[string]string),
		idempotency:   make(map[string]idempotencyMemRecord),
	}
}

func (m *memoryStore) GetExtensionByProjectID(_ context.Context, projectID string) (*SocialPostConfigRow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	row, ok := m.extensions[projectID]
	if !ok {
		return nil, ErrNotFound
	}
	return row, nil
}

func (m *memoryStore) UpsertExtension(_ context.Context, row SocialPostConfigRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.extensions[row.ProjectID] = &row
	return nil
}

func (m *memoryStore) InsertVariant(_ context.Context, row SocialPostVariantRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.variants[row.ID] = &row
	return nil
}

func (m *memoryStore) ListVariants(_ context.Context, _ string, _ ListSocialPostVariantsRequest) ([]SocialPostVariantResponse, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return nil, 0, nil
}

func (m *memoryStore) GetVariantByID(_ context.Context, variantID string) (*SocialPostVariantRow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	row, ok := m.variants[variantID]
	if !ok {
		return nil, ErrNotFound
	}
	return row, nil
}

func (m *memoryStore) SelectVariantInTx(_ context.Context, _ SelectVariantTxInput) (string, string, error) {
	return "", "", ErrInternal
}

func (m *memoryStore) InsertAsset(_ context.Context, row SocialPostAssetRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.assets[row.ID] = &row
	return nil
}

func (m *memoryStore) ListAssets(_ context.Context, _ string, _ GetSocialPostAssetsRequest) ([]SocialPostAssetItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return nil, nil
}

func (m *memoryStore) InsertOperationLog(_ context.Context, _ OperationLogRow) (string, error) {
	return "", ErrInternal
}

func (m *memoryStore) CheckIdempotency(_ context.Context, scope, endpoint, key, hash string) (string, string, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	fullKey := scope + ":" + endpoint + ":" + key
	rec, ok := m.idempotency[fullKey]
	if !ok {
		return "", "", false, nil
	}
	if rec.hash != hash {
		return "", "", true, nil
	}
	return rec.refType, rec.refID, false, nil
}

func (m *memoryStore) StoreIdempotency(_ context.Context, scope, endpoint, key, hash, refType, refID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	fullKey := scope + ":" + endpoint + ":" + key
	m.idempotency[fullKey] = idempotencyMemRecord{refType: refType, refID: refID, hash: hash}
	return nil
}