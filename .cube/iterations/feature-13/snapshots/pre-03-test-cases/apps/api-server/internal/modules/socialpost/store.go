package socialpost

import "context"

type Store interface {
	GetExtensionByProjectID(ctx context.Context, projectID string) (*SocialPostConfigRow, error)
	UpsertExtension(ctx context.Context, row SocialPostConfigRow) error

	InsertVariant(ctx context.Context, row SocialPostVariantRow) error
	ListVariants(ctx context.Context, projectID string, req ListSocialPostVariantsRequest) ([]SocialPostVariantResponse, int, error)
	GetVariantByID(ctx context.Context, variantID string) (*SocialPostVariantRow, error)
	SelectVariantInTx(ctx context.Context, input SelectVariantTxInput) (contentVersionID string, operationLogID string, err error)

	InsertAsset(ctx context.Context, row SocialPostAssetRow) error
	ListAssets(ctx context.Context, projectID string, req GetSocialPostAssetsRequest) ([]SocialPostAssetItem, error)

	InsertOperationLog(ctx context.Context, row OperationLogRow) (string, error)
	CheckIdempotency(ctx context.Context, scope, endpoint, key, hash string) (refType string, refID string, conflict bool, err error)
	StoreIdempotency(ctx context.Context, scope, endpoint, key, hash, refType, refID string) error
}

// --- Row types for Store operations ---

type SocialPostConfigRow struct {
	ID                  string
	ProjectID           string
	TargetPlatforms     string
	DefaultVariantCount int
	CaptionLengthPolicy string
	HashtagPolicy       string
	CoverCopyPolicy     string
	ToneStyle           string
	ForbiddenTerms      string
	ConfigVersion       int
	OperationLogID      string
}

type SocialPostVariantRow struct {
	ID               string
	ProjectID        string
	ContentItemID    string
	GenerationRunID  string
	WorkflowRunID    string
	VariantIndex     int
	Platform         string
	Title            string
	Body             string
	Hashtags         string
	CoverCopy        string
	ToneStyle        string
	Status           string
	ContentVersionID string
	SelectedAt       string
	OperationLogID   string
}

type SelectVariantTxInput struct {
	VariantID      string
	ContentItemID  string
	ProjectID      string
	Note           string
	IdempotencyKey string
	RequestHash    string
}

type SocialPostAssetRow struct {
	ID               string
	ProjectID        string
	ContentItemID    string
	SourceVariantID  string
	AssetType        string
	Platform         string
	GenerationRunID  string
	WorkflowRunID    string
	Result           string
	AssetSuggestions string
	OperationLogID   string
}

type OperationLogRow struct {
	Actor      string
	Resource   string
	FromState  string
	ToState    string
	Reason     string
	ResourceID string
}