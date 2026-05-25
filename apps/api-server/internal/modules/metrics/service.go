package metrics

import "context"

type Service interface {
	CreateTemplate(ctx context.Context, req CreateMetricTemplateRequest) (CreateMetricTemplateResponse, error)
	ListTemplates(ctx context.Context, req ListMetricTemplatesRequest) (PagedMetricTemplatesResponse, error)
	CreateRecord(ctx context.Context, req CreateMetricRecordRequest, idempotencyKey string) (CreateMetricRecordResponse, error)
	BatchCreateRecords(ctx context.Context, req BatchCreateMetricRecordsRequest, idempotencyKey string) (BatchCreateMetricRecordsResponse, error)
	ListRecords(ctx context.Context, req ListMetricRecordsRequest) (PagedMetricRecordsResponse, error)
	GetSummary(ctx context.Context, projectID string, req MetricSummaryRequest) (MetricSummaryResponse, error)
	GetTrends(ctx context.Context, projectID string, req MetricTrendRequest) (MetricTrendResponse, error)
	GetMissingDates(ctx context.Context, projectID string, req MissingMetricDatesRequest) (MissingMetricDatesResponse, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) CreateTemplate(ctx context.Context, req CreateMetricTemplateRequest) (CreateMetricTemplateResponse, error) {
	panic("not implemented")
}

func (s *service) ListTemplates(ctx context.Context, req ListMetricTemplatesRequest) (PagedMetricTemplatesResponse, error) {
	panic("not implemented")
}

func (s *service) CreateRecord(ctx context.Context, req CreateMetricRecordRequest, idempotencyKey string) (CreateMetricRecordResponse, error) {
	panic("not implemented")
}

func (s *service) BatchCreateRecords(ctx context.Context, req BatchCreateMetricRecordsRequest, idempotencyKey string) (BatchCreateMetricRecordsResponse, error) {
	panic("not implemented")
}

func (s *service) ListRecords(ctx context.Context, req ListMetricRecordsRequest) (PagedMetricRecordsResponse, error) {
	panic("not implemented")
}

func (s *service) GetSummary(ctx context.Context, projectID string, req MetricSummaryRequest) (MetricSummaryResponse, error) {
	panic("not implemented")
}

func (s *service) GetTrends(ctx context.Context, projectID string, req MetricTrendRequest) (MetricTrendResponse, error) {
	panic("not implemented")
}

func (s *service) GetMissingDates(ctx context.Context, projectID string, req MissingMetricDatesRequest) (MissingMetricDatesResponse, error) {
	panic("not implemented")
}
