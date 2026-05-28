package strategy

import "context"

type Store interface {
	InsertSuggestionRun(ctx context.Context, run StrategySuggestionRunResponse) error
	FindSuggestionRunByID(ctx context.Context, id string) (*StrategySuggestionRunResponse, error)
	UpdateSuggestionRunStatus(ctx context.Context, id, status, failureReason string, suggestionCount int) error

	InsertSuggestion(ctx context.Context, s StrategySuggestionDetailResponse) error
	FindSuggestionByID(ctx context.Context, id string) (*StrategySuggestionDetailResponse, error)
	UpdateSuggestionStatus(ctx context.Context, id, status string, fields map[string]any) error
	ListSuggestions(ctx context.Context, projectID string, req ListStrategySuggestionsRequest) ([]StrategySuggestionDetailResponse, int, error)

	InsertExecutionLog(ctx context.Context, log ExecutionLogResponse) error
	ListExecutionLogs(ctx context.Context, suggestionID string, page, pageSize int) ([]ExecutionLogResponse, int, error)

	CheckIdempotency(ctx context.Context, scope, endpoint, key, hash string) (refType string, refID string, conflict bool, err error)
	StoreIdempotency(ctx context.Context, scope, endpoint, key, hash, refType, refID string) error
}
