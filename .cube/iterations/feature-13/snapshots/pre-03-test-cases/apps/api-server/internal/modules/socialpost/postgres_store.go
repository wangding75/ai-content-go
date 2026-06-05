package socialpost

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

func (p *postgresStore) GetExtensionByProjectID(_ context.Context, _ string) (*SocialPostConfigRow, error) {
	return nil, ErrInternal
}

func (p *postgresStore) UpsertExtension(_ context.Context, _ SocialPostConfigRow) error {
	return ErrInternal
}

func (p *postgresStore) InsertVariant(_ context.Context, _ SocialPostVariantRow) error {
	return ErrInternal
}

func (p *postgresStore) ListVariants(_ context.Context, _ string, _ ListSocialPostVariantsRequest) ([]SocialPostVariantResponse, int, error) {
	return nil, 0, ErrInternal
}

func (p *postgresStore) GetVariantByID(_ context.Context, _ string) (*SocialPostVariantRow, error) {
	return nil, ErrInternal
}

func (p *postgresStore) SelectVariantInTx(_ context.Context, _ SelectVariantTxInput) (string, string, error) {
	return "", "", ErrInternal
}

func (p *postgresStore) InsertAsset(_ context.Context, _ SocialPostAssetRow) error {
	return ErrInternal
}

func (p *postgresStore) ListAssets(_ context.Context, _ string, _ GetSocialPostAssetsRequest) ([]SocialPostAssetItem, error) {
	return nil, ErrInternal
}

func (p *postgresStore) InsertOperationLog(_ context.Context, _ OperationLogRow) (string, error) {
	return "", ErrInternal
}

func (p *postgresStore) CheckIdempotency(_ context.Context, _, _, _, _ string) (string, string, bool, error) {
	return "", "", false, ErrInternal
}

func (p *postgresStore) StoreIdempotency(_ context.Context, _, _, _, _, _, _ string) error {
	return ErrInternal
}