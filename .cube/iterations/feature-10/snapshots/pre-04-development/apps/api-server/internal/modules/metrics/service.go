package metrics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

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

type service struct {
	store Store
}

func NewService(stores ...Store) Service {
	var store Store
	if len(stores) > 0 {
		store = stores[0]
	} else {
		store = NewMemoryStore()
	}
	return &service{store: store}
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

func validateTemplateRequest(req CreateMetricTemplateRequest) error {
	if req.ContentType == "" || req.Platform == "" || req.MetricCode == "" || req.MetricName == "" || req.Unit == "" || req.ValueType == "" || req.AggregationMethod == "" || req.Period == "" {
		return ErrValidation
	}
	if !validMetricValueType(req.ValueType) || metricAggregationMethods[req.AggregationMethod] == "" || metricBucketColumns[req.Period] == "" {
		return ErrValidation
	}
	return nil
}

func validMetricValueType(value string) bool {
	switch value {
	case ValueTypeInteger, ValueTypeDecimal, ValueTypePercentage, ValueTypeCurrency, ValueTypeDuration:
		return true
	default:
		return false
	}
}

func validateRecordRequest(req CreateMetricRecordRequest, idempotencyKey string) error {
	if idempotencyKey == "" || req.ProjectID == "" || req.ContentItemID == "" || req.ContentVersionID == "" || req.PublishJobID == "" || req.TargetID == "" || req.Platform == "" || req.MetricCode == "" || req.MetricDate == "" || req.Period == "" || req.RawValue == "" || req.SourceType == "" {
		return ErrValidation
	}
	if metricBucketColumns[req.Period] == "" {
		return ErrValidation
	}
	switch req.SourceType {
	case SourceManual, SourceImport, SourceExtension, SourceExternalCallback:
		return nil
	default:
		return ErrValidation
	}
}

func normalizeMetricValue(raw string) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, ErrValidation
	}
	return value, nil
}

func templateKey(contentType string, platform string, metricCode string) string {
	return strings.Join([]string{contentType, platform, metricCode}, "|")
}

func recordUniqueKey(projectID, platform, targetID, contentVersionID, metricCode, metricDate, period string) string {
	return strings.Join([]string{projectID, platform, targetID, contentVersionID, metricCode, metricDate, period}, "|")
}

func requestHash(value any) string {
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func pagination(page int, pageSize int, total int) content.PaginationResponse {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return content.PaginationResponse{Page: page, PageSize: pageSize, Total: total, HasNext: page*pageSize < total}
}

func metricBatchErrorCode(err error) string {
	switch err {
	case ErrNotFound:
		return "NOT_FOUND"
	case ErrConflict:
		return "CONFLICT"
	case ErrInternal:
		return "INTERNAL_ERROR"
	default:
		return "VALIDATION_ERROR"
	}
}

func (s *service) CreateTemplate(ctx context.Context, req CreateMetricTemplateRequest) (CreateMetricTemplateResponse, error) {
	if err := validateTemplateRequest(req); err != nil {
		return CreateMetricTemplateResponse{}, err
	}
	existing, err := s.store.FindTemplateByKey(ctx, req.ContentType, req.Platform, req.MetricCode)
	if err != nil {
		return CreateMetricTemplateResponse{}, ErrInternal
	}
	if existing != nil {
		return CreateMetricTemplateResponse{}, ErrConflict
	}
	id := "metric-template-" + req.ContentType + "-" + req.Platform + "-" + req.MetricCode
	t := MetricTemplateResponse{
		ID:                id,
		ContentType:       req.ContentType,
		Platform:          req.Platform,
		MetricCode:        req.MetricCode,
		MetricName:        req.MetricName,
		Unit:              req.Unit,
		ValueType:         req.ValueType,
		AggregationMethod: req.AggregationMethod,
		Period:            req.Period,
		Required:          req.Required,
		Enabled:           req.Enabled,
		UpdatedAt:         time.Now().UTC(),
	}
	if err := s.store.InsertTemplate(ctx, t); err != nil {
		return CreateMetricTemplateResponse{}, ErrInternal
	}
	return CreateMetricTemplateResponse{MetricTemplateID: id}, nil
}

func (s *service) ListTemplates(ctx context.Context, req ListMetricTemplatesRequest) (PagedMetricTemplatesResponse, error) {
	items, err := s.store.ListTemplates(ctx, req)
	if err != nil {
		return PagedMetricTemplatesResponse{}, ErrInternal
	}
	total := len(items)
	page := req.Page
	pageSize := req.PageSize
	start, end := pageBounds(page, pageSize, total)
	return PagedMetricTemplatesResponse{Items: items[start:end], Pagination: pagination(page, pageSize, total)}, nil
}

func (s *service) CreateRecord(ctx context.Context, req CreateMetricRecordRequest, idempotencyKey string) (CreateMetricRecordResponse, error) {
	if err := validateRecordRequest(req, idempotencyKey); err != nil {
		return CreateMetricRecordResponse{}, err
	}
	normalized, err := normalizeMetricValue(req.RawValue)
	if err != nil {
		return CreateMetricRecordResponse{}, err
	}

	scope := "metrics:" + "record:" + req.ProjectID
	hash := requestHash(req)
	refType, refID, conflict, err := s.store.CheckIdempotency(ctx, scope, "create_record", idempotencyKey, hash)
	if err != nil {
		return CreateMetricRecordResponse{}, ErrInternal
	}
	if conflict {
		return CreateMetricRecordResponse{}, ErrIdempotencyConflict
	}
	if refType != "" {
		return CreateMetricRecordResponse{MetricRecordID: refID, NormalizedValue: normalized, OperationLogID: "operation-log-" + refID}, nil
	}

	contentType := "article"
	existing, err := s.store.FindTemplateByKey(ctx, contentType, req.Platform, req.MetricCode)
	if err != nil {
		return CreateMetricRecordResponse{}, ErrInternal
	}
	if existing == nil {
		return CreateMetricRecordResponse{}, ErrNotFound
	}
	if !existing.Enabled || existing.Period != req.Period {
		return CreateMetricRecordResponse{}, ErrValidation
	}

	rec, err := s.store.FindRecordByUniqueKey(ctx, req.ProjectID, req.Platform, req.TargetID, req.ContentVersionID, req.MetricCode, req.MetricDate, req.Period)
	if err != nil {
		return CreateMetricRecordResponse{}, ErrInternal
	}
	if rec != nil {
		if rec.RawValue != req.RawValue {
			return CreateMetricRecordResponse{}, ErrConflict
		}
		return CreateMetricRecordResponse{MetricRecordID: rec.ID, NormalizedValue: rec.NormalizedValue, OperationLogID: "operation-log-" + rec.ID}, nil
	}

	id := "metric-record-" + req.ProjectID + "-" + req.TargetID + "-" + req.ContentVersionID + "-" + req.MetricCode + "-" + req.MetricDate
	now := time.Now().UTC()
	record := MetricRecordResponse{
		ID:               id,
		ProjectID:        req.ProjectID,
		ContentItemID:    req.ContentItemID,
		ContentVersionID: req.ContentVersionID,
		PublishJobID:     req.PublishJobID,
		TargetID:         req.TargetID,
		ContentType:      contentType,
		MetricTemplateID: existing.ID,
		Platform:         req.Platform,
		ExternalURL:      req.ExternalURL,
		MetricCode:       req.MetricCode,
		MetricDate:       req.MetricDate,
		Period:           req.Period,
		RawValue:         req.RawValue,
		NormalizedValue:  normalized,
		SourceType:       req.SourceType,
		SourceRef:        req.SourceRef,
		CollectedAt:      now,
		UpdatedAt:        now,
	}
	if err := s.store.InsertRecord(ctx, record); err != nil {
		return CreateMetricRecordResponse{}, ErrInternal
	}
	opLogID := "operation-log-" + id
	if err := s.store.StoreIdempotency(ctx, scope, "create_record", idempotencyKey, hash, "metric_record", id); err != nil {
		return CreateMetricRecordResponse{}, ErrInternal
	}
	return CreateMetricRecordResponse{MetricRecordID: id, NormalizedValue: normalized, OperationLogID: opLogID}, nil
}

func (s *service) BatchCreateRecords(ctx context.Context, req BatchCreateMetricRecordsRequest, idempotencyKey string) (BatchCreateMetricRecordsResponse, error) {
	if idempotencyKey == "" || len(req.Records) == 0 {
		return BatchCreateMetricRecordsResponse{}, ErrValidation
	}
	if len(req.Records) > 100 {
		return BatchCreateMetricRecordsResponse{}, ErrValidation
	}
	projectID := req.Records[0].ProjectID
	if projectID == "" {
		return BatchCreateMetricRecordsResponse{}, ErrValidation
	}

	scope := "metrics:" + "batch:" + projectID
	hash := requestHash(req)
	refType, refID, conflict, err := s.store.CheckIdempotency(ctx, scope, "batch_create", idempotencyKey, hash)
	if err != nil {
		return BatchCreateMetricRecordsResponse{}, ErrInternal
	}
	if conflict {
		return BatchCreateMetricRecordsResponse{}, ErrIdempotencyConflict
	}
	if refType != "" {
		return BatchCreateMetricRecordsResponse{CreatedCount: 0, FailedCount: 0, Errors: []BatchMetricRecordError{}, OperationLogID: refID}, nil
	}

	response := BatchCreateMetricRecordsResponse{
		Errors:         []BatchMetricRecordError{},
		OperationLogID: "operation-log-metric-batch-" + projectID,
	}
	for index, record := range req.Records {
		if record.ProjectID != projectID {
			response.FailedCount++
			response.Errors = append(response.Errors, BatchMetricRecordError{Index: index, MetricCode: record.MetricCode, Field: "project_id", Code: "VALIDATION_ERROR", Message: "records must belong to the same project", SourceRef: record.SourceRef})
			continue
		}
		_, err := s.CreateRecord(ctx, record, idempotencyKey+":"+strconv.Itoa(index))
		if err != nil {
			response.FailedCount++
			response.Errors = append(response.Errors, BatchMetricRecordError{Index: index, MetricCode: record.MetricCode, Field: "record", Code: metricBatchErrorCode(err), Message: err.Error(), SourceRef: record.SourceRef})
			continue
		}
		response.CreatedCount++
	}

	if err := s.store.StoreIdempotency(ctx, scope, "batch_create", idempotencyKey, hash, "operation_log", response.OperationLogID); err != nil {
		return BatchCreateMetricRecordsResponse{}, ErrInternal
	}

	if response.CreatedCount == 0 && response.FailedCount > 0 {
		return response, ErrValidation
	}
	return response, nil
}

func (s *service) ListRecords(ctx context.Context, req ListMetricRecordsRequest) (PagedMetricRecordsResponse, error) {
	if req.ProjectID == "" {
		return PagedMetricRecordsResponse{}, ErrValidation
	}
	items, total, err := s.store.ListRecords(ctx, req)
	if err != nil {
		return PagedMetricRecordsResponse{}, ErrInternal
	}
	page := req.Page
	pageSize := req.PageSize
	start, end := pageBounds(page, pageSize, total)
	return PagedMetricRecordsResponse{Items: items[start:end], Pagination: pagination(page, pageSize, total)}, nil
}

func (s *service) GetSummary(ctx context.Context, projectID string, req MetricSummaryRequest) (MetricSummaryResponse, error) {
	if projectID == "" || req.DateFrom == "" || req.DateTo == "" || len(req.MetricCodes) == 0 {
		return MetricSummaryResponse{}, ErrValidation
	}
	items, total, err := s.store.QuerySummary(ctx, projectID, req)
	if err != nil {
		return MetricSummaryResponse{}, ErrInternal
	}
	snapshotID := fmt.Sprintf("metric-summary-snapshot-%s-%s-%s", projectID, req.DateFrom, req.DateTo)
	snap := SummarySnapshotRow{
		ID:                snapshotID,
		ProjectID:         projectID,
		DateFrom:          req.DateFrom,
		DateTo:            req.DateTo,
		Platform:          req.Platform,
		TargetID:          req.TargetID,
		MetricCodes:       "[" + strings.Join(req.MetricCodes, ",") + "]",
		AggregationMethod: "mixed",
		Summary:           "[]",
		SourceRecordCount: total,
	}
	if err := s.store.InsertSummarySnapshot(ctx, snap); err != nil {
		return MetricSummaryResponse{}, ErrInternal
	}
	return MetricSummaryResponse{
		ProjectID:         projectID,
		DateFrom:          req.DateFrom,
		DateTo:            req.DateTo,
		Platform:          req.Platform,
		TargetID:          req.TargetID,
		Items:             items,
		SummarySnapshotID: snapshotID,
		SourceRecordCount: total,
	}, nil
}

func (s *service) GetTrends(ctx context.Context, projectID string, req MetricTrendRequest) (MetricTrendResponse, error) {
	if projectID == "" || req.MetricCode == "" || req.DateFrom == "" || req.DateTo == "" || metricBucketColumns[req.Bucket] == "" {
		return MetricTrendResponse{}, ErrValidation
	}
	series, missingPoints, signature, sourceCount, err := s.store.QueryTrends(ctx, projectID, req)
	if err != nil {
		return MetricTrendResponse{}, ErrInternal
	}
	aggregation := AggregationSum
	tpl, err := s.store.FindTemplateByKey(ctx, "article", req.Platform, req.MetricCode)
	if err != nil {
		return MetricTrendResponse{}, ErrInternal
	}
	if tpl != nil {
		aggregation = tpl.AggregationMethod
	}
	return MetricTrendResponse{
		ProjectID:         projectID,
		MetricCode:        req.MetricCode,
		Bucket:            req.Bucket,
		AggregationMethod: aggregation,
		QuerySignature:    signature,
		SourceRecordCount: sourceCount,
		Series:            series,
		MissingPoints:     missingPoints,
	}, nil
}

func (s *service) GetMissingDates(ctx context.Context, projectID string, req MissingMetricDatesRequest) (MissingMetricDatesResponse, error) {
	if projectID == "" || req.DateFrom == "" || req.DateTo == "" {
		return MissingMetricDatesResponse{}, ErrValidation
	}
	items, err := s.store.QueryMissingDates(ctx, projectID, req)
	if err != nil {
		return MissingMetricDatesResponse{}, ErrInternal
	}
	return MissingMetricDatesResponse{ProjectID: projectID, Items: items}, nil
}

func pageBounds(page, pageSize, total int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return start, end
}
