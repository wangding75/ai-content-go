package metrics

import "context"

type Store interface {
	InsertTemplate(ctx context.Context, t MetricTemplateResponse) error
	FindTemplateByKey(ctx context.Context, contentType, platform, metricCode string) (*MetricTemplateResponse, error)
	ListTemplates(ctx context.Context, req ListMetricTemplatesRequest) ([]MetricTemplateResponse, error)

	InsertRecord(ctx context.Context, r MetricRecordResponse) error
	FindRecordByUniqueKey(ctx context.Context, projectID, platform, targetID, contentVersionID, metricCode, metricDate, period string) (*MetricRecordResponse, error)
	ListRecords(ctx context.Context, req ListMetricRecordsRequest) ([]MetricRecordResponse, int, error)

	InsertSummarySnapshot(ctx context.Context, snap SummarySnapshotRow) error

	CheckIdempotency(ctx context.Context, scope, endpoint, key, hash string) (refType string, refID string, conflict bool, err error)
	StoreIdempotency(ctx context.Context, scope, endpoint, key, hash, refType, refID string) error

	QuerySummary(ctx context.Context, projectID string, req MetricSummaryRequest) ([]MetricSummaryItem, int, error)
	QueryTrends(ctx context.Context, projectID string, req MetricTrendRequest) ([]MetricTrendPoint, []MetricMissingPoint, string, int, error)
	QueryMissingDates(ctx context.Context, projectID string, req MissingMetricDatesRequest) ([]MissingMetricDateItem, error)
}

type SummarySnapshotRow struct {
	ID               string
	ProjectID        string
	DateFrom         string
	DateTo           string
	Platform         string
	TargetID         string
	MetricCodes      string
	AggregationMethod string
	Summary          string
	SourceRecordCount int
}
