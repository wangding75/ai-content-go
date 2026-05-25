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

const metricSummarySQL = `
SELECT
    r.metric_code,
    t.unit,
    t.aggregation_method,
    CASE
        WHEN t.aggregation_method = 'sum' THEN SUM(r.normalized_value)
        WHEN t.aggregation_method = 'avg' THEN AVG(r.normalized_value)
        WHEN t.aggregation_method = 'max' THEN MAX(r.normalized_value)
        WHEN t.aggregation_method = 'min' THEN MIN(r.normalized_value)
        ELSE (ARRAY_AGG(r.normalized_value ORDER BY r.metric_date DESC))[1]
    END AS value,
    COUNT(*) AS source_record_count
FROM metric_record r
JOIN metric_template t ON t.id = r.metric_template_id AND t.enabled = TRUE
WHERE r.project_id = $1
  AND r.metric_date >= $2
  AND r.metric_date <= $3
  AND ($4::text = '' OR r.platform = $4)
  AND ($5::text = '' OR r.target_id = $5)
  AND r.metric_code = ANY($6::text[])
GROUP BY r.metric_code, t.unit, t.aggregation_method`

const metricTrendSQL = `
SELECT
    DATE_TRUNC($7, r.metric_date::timestamp)::date AS bucket_start,
    r.metric_code,
    t.aggregation_method,
    CASE
        WHEN t.aggregation_method = 'sum' THEN SUM(r.normalized_value)
        WHEN t.aggregation_method = 'avg' THEN AVG(r.normalized_value)
        WHEN t.aggregation_method = 'max' THEN MAX(r.normalized_value)
        WHEN t.aggregation_method = 'min' THEN MIN(r.normalized_value)
        ELSE (ARRAY_AGG(r.normalized_value ORDER BY r.metric_date DESC))[1]
    END AS value,
    COUNT(*) AS source_record_count
FROM metric_record r
JOIN metric_template t ON t.id = r.metric_template_id AND t.enabled = TRUE
WHERE r.project_id = $1
  AND r.metric_code = $2
  AND r.metric_date >= $3
  AND r.metric_date <= $4
  AND ($5::text = '' OR r.platform = $5)
  AND ($6::text = '' OR r.target_id = $6)
GROUP BY bucket_start, r.metric_code, t.aggregation_method
ORDER BY bucket_start ASC`

const metricMissingDatesSQL = `
SELECT
    j.content_item_id,
    j.content_version_id,
    j.id AS publish_job_id,
    j.target_id,
    t.platform,
    t.metric_code,
    t.period,
    expected.metric_date,
    'required_metric_missing' AS missing_reason
FROM publish_job j
JOIN publish_target pt ON pt.id = j.target_id AND pt.enabled = TRUE
JOIN content_item ci ON ci.id = j.content_item_id
JOIN content_project cp ON cp.id = j.project_id
JOIN content_type ct ON ct.id = cp.content_type_id
JOIN metric_template t
  ON t.platform = pt.platform
 AND t.content_type = ct.code
 AND t.required = TRUE
 AND t.enabled = TRUE
JOIN LATERAL generate_series(
    DATE_TRUNC(t.period, $2::date::timestamp)::date,
    DATE_TRUNC(t.period, $3::date::timestamp)::date,
    CASE t.period
        WHEN 'day' THEN '1 day'::interval
        WHEN 'week' THEN '1 week'::interval
        ELSE '1 month'::interval
    END
) AS expected(metric_date) ON TRUE
WHERE j.project_id = $1
  AND j.status = 'published'
  AND j.published_at::date <= expected.metric_date
  AND ($4::text = '' OR t.metric_code = $4)
  AND ($5::text = '' OR t.platform = $5)
  AND ($6::text = '' OR j.target_id = $6)
  AND NOT EXISTS (
      SELECT 1
      FROM metric_record r
      WHERE r.project_id = j.project_id
        AND r.content_version_id = j.content_version_id
        AND r.target_id = j.target_id
        AND r.metric_template_id = t.id
        AND r.period = t.period
        AND r.metric_date = expected.metric_date::date
  )`

var metricSortColumns = map[string]string{
	"":            "metric_date",
	"metric_date": "metric_date",
	"created_at":  "created_at",
	"metric_code": "metric_code",
	"platform":    "platform",
	"updated_at":  "updated_at",
}

var metricBucketColumns = map[string]string{
	PeriodDay:   "day",
	PeriodWeek:  "week",
	PeriodMonth: "month",
}

var metricAggregationMethods = map[string]string{
	AggregationSum:    AggregationSum,
	AggregationAvg:    AggregationAvg,
	AggregationMax:    AggregationMax,
	AggregationMin:    AggregationMin,
	AggregationLatest: AggregationLatest,
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
