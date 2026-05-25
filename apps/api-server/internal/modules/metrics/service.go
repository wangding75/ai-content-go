package metrics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"sync"
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
	state *memoryState
}

type memoryState struct {
	mu          sync.Mutex
	templates   map[string]MetricTemplateResponse
	records     map[string]MetricRecordResponse
	idempotency map[string]string
}

func NewService() Service {
	return &service{state: newMemoryState()}
}

func newMemoryState() *memoryState {
	return &memoryState{
		templates:   map[string]MetricTemplateResponse{},
		records:     map[string]MetricRecordResponse{},
		idempotency: map[string]string{},
	}
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

func recordKey(req CreateMetricRecordRequest) string {
	return strings.Join([]string{req.ProjectID, req.Platform, req.TargetID, req.ContentVersionID, req.MetricCode, req.MetricDate, req.Period}, "|")
}

func requestHash(value any) string {
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func (s *service) reserveIdempotency(scope string, key string, req any) error {
	hash := requestHash(req)
	composite := scope + ":" + key
	if existing, ok := s.state.idempotency[composite]; ok && existing != hash {
		return ErrIdempotencyConflict
	}
	s.state.idempotency[composite] = hash
	return nil
}

func contentTypeForRecord(req CreateMetricRecordRequest) string {
	if strings.Contains(req.ContentItemID, "novel") {
		return "novel"
	}
	return "article"
}

func pagination(page int, pageSize int, total int) content.PaginationResponse {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
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
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	key := templateKey(req.ContentType, req.Platform, req.MetricCode)
	if _, exists := s.state.templates[key]; exists {
		return CreateMetricTemplateResponse{}, ErrConflict
	}
	id := "metric-template-" + req.ContentType + "-" + req.Platform + "-" + req.MetricCode
	s.state.templates[key] = MetricTemplateResponse{
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
	return CreateMetricTemplateResponse{MetricTemplateID: id}, nil
}

func (s *service) ListTemplates(ctx context.Context, req ListMetricTemplatesRequest) (PagedMetricTemplatesResponse, error) {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	items := make([]MetricTemplateResponse, 0, len(s.state.templates))
	for _, item := range s.state.templates {
		if req.ContentType != "" && item.ContentType != req.ContentType {
			continue
		}
		if req.Platform != "" && item.Platform != req.Platform {
			continue
		}
		if req.Enabled != nil && item.Enabled != *req.Enabled {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].MetricCode < items[j].MetricCode
	})
	return PagedMetricTemplatesResponse{Items: items, Pagination: pagination(req.Page, req.PageSize, len(items))}, nil
}

func (s *service) CreateRecord(ctx context.Context, req CreateMetricRecordRequest, idempotencyKey string) (CreateMetricRecordResponse, error) {
	if err := validateRecordRequest(req, idempotencyKey); err != nil {
		return CreateMetricRecordResponse{}, err
	}
	normalized, err := normalizeMetricValue(req.RawValue)
	if err != nil {
		return CreateMetricRecordResponse{}, err
	}
	contentType := contentTypeForRecord(req)
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	if err := s.reserveIdempotency("metrics:record:"+req.ProjectID, idempotencyKey, req); err != nil {
		return CreateMetricRecordResponse{}, err
	}
	template, ok := s.state.templates[templateKey(contentType, req.Platform, req.MetricCode)]
	if !ok {
		return CreateMetricRecordResponse{}, ErrNotFound
	}
	if !template.Enabled || template.Period != req.Period {
		return CreateMetricRecordResponse{}, ErrValidation
	}
	key := recordKey(req)
	if existing, exists := s.state.records[key]; exists {
		if existing.RawValue != req.RawValue {
			return CreateMetricRecordResponse{}, ErrConflict
		}
		return CreateMetricRecordResponse{MetricRecordID: existing.ID, NormalizedValue: existing.NormalizedValue, OperationLogID: "operation-log-" + existing.ID}, nil
	}
	id := "metric-record-" + req.ProjectID + "-" + req.MetricCode + "-" + req.MetricDate
	record := MetricRecordResponse{
		ID:               id,
		ProjectID:        req.ProjectID,
		ContentItemID:    req.ContentItemID,
		ContentVersionID: req.ContentVersionID,
		PublishJobID:     req.PublishJobID,
		TargetID:         req.TargetID,
		ContentType:      contentType,
		MetricTemplateID: template.ID,
		Platform:         req.Platform,
		ExternalURL:      req.ExternalURL,
		MetricCode:       req.MetricCode,
		MetricDate:       req.MetricDate,
		Period:           req.Period,
		RawValue:         req.RawValue,
		NormalizedValue:  normalized,
		SourceType:       req.SourceType,
		SourceRef:        req.SourceRef,
		CollectedAt:      time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	s.state.records[key] = record
	return CreateMetricRecordResponse{MetricRecordID: id, NormalizedValue: normalized, OperationLogID: "operation-log-" + id}, nil
}

func (s *service) BatchCreateRecords(ctx context.Context, req BatchCreateMetricRecordsRequest, idempotencyKey string) (BatchCreateMetricRecordsResponse, error) {
	if idempotencyKey == "" || len(req.Records) == 0 {
		return BatchCreateMetricRecordsResponse{}, ErrValidation
	}
	projectID := req.Records[0].ProjectID
	if projectID == "" {
		return BatchCreateMetricRecordsResponse{}, ErrValidation
	}
	s.state.mu.Lock()
	if err := s.reserveIdempotency("metrics:batch:"+projectID, idempotencyKey, req); err != nil {
		s.state.mu.Unlock()
		return BatchCreateMetricRecordsResponse{}, err
	}
	s.state.mu.Unlock()

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
	if response.CreatedCount == 0 && response.FailedCount > 0 {
		return response, ErrValidation
	}
	return response, nil
}

func (s *service) ListRecords(ctx context.Context, req ListMetricRecordsRequest) (PagedMetricRecordsResponse, error) {
	if req.ProjectID == "" {
		return PagedMetricRecordsResponse{}, ErrValidation
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	items := make([]MetricRecordResponse, 0, len(s.state.records))
	for _, item := range s.state.records {
		if item.ProjectID != req.ProjectID {
			continue
		}
		if req.Platform != "" && item.Platform != req.Platform {
			continue
		}
		if req.TargetID != "" && item.TargetID != req.TargetID {
			continue
		}
		if req.ContentItemID != "" && item.ContentItemID != req.ContentItemID {
			continue
		}
		if req.MetricCode != "" && item.MetricCode != req.MetricCode {
			continue
		}
		if req.DateFrom != "" && item.MetricDate < req.DateFrom {
			continue
		}
		if req.DateTo != "" && item.MetricDate > req.DateTo {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].MetricDate > items[j].MetricDate
	})
	return PagedMetricRecordsResponse{Items: items, Pagination: pagination(req.Page, req.PageSize, len(items))}, nil
}

func (s *service) GetSummary(ctx context.Context, projectID string, req MetricSummaryRequest) (MetricSummaryResponse, error) {
	if projectID == "" || req.DateFrom == "" || req.DateTo == "" || len(req.MetricCodes) == 0 {
		return MetricSummaryResponse{}, ErrValidation
	}
	records, err := s.ListRecords(ctx, ListMetricRecordsRequest{
		ProjectID: projectID,
		Platform:  req.Platform,
		TargetID:  req.TargetID,
		DateFrom:  req.DateFrom,
		DateTo:    req.DateTo,
	})
	if err != nil {
		return MetricSummaryResponse{}, err
	}
	codeSet := map[string]bool{}
	for _, code := range req.MetricCodes {
		codeSet[code] = true
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	items := make([]MetricSummaryItem, 0)
	total := 0
	for _, code := range req.MetricCodes {
		var sum float64
		count := 0
		unit := ""
		aggregation := AggregationSum
		for _, record := range records.Items {
			if !codeSet[record.MetricCode] || record.MetricCode != code {
				continue
			}
			template := s.templateByIDLocked(record.MetricTemplateID)
			if template.ID != "" {
				unit = template.Unit
				aggregation = template.AggregationMethod
			}
			sum += record.NormalizedValue
			count++
		}
		if count == 0 {
			continue
		}
		total += count
		items = append(items, MetricSummaryItem{MetricCode: code, Value: sum, Unit: unit, AggregationMethod: aggregation, SourceRecordCount: count})
	}
	return MetricSummaryResponse{
		ProjectID:         projectID,
		DateFrom:          req.DateFrom,
		DateTo:            req.DateTo,
		Platform:          req.Platform,
		TargetID:          req.TargetID,
		Items:             items,
		SummarySnapshotID: "metric-summary-snapshot-" + projectID,
		SourceRecordCount: total,
	}, nil
}

func (s *service) GetTrends(ctx context.Context, projectID string, req MetricTrendRequest) (MetricTrendResponse, error) {
	if projectID == "" || req.MetricCode == "" || req.DateFrom == "" || req.DateTo == "" || metricBucketColumns[req.Bucket] == "" {
		return MetricTrendResponse{}, ErrValidation
	}
	records, err := s.ListRecords(ctx, ListMetricRecordsRequest{
		ProjectID:  projectID,
		Platform:   req.Platform,
		TargetID:   req.TargetID,
		MetricCode: req.MetricCode,
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
	})
	if err != nil {
		return MetricTrendResponse{}, err
	}
	series := make([]MetricTrendPoint, 0, len(records.Items))
	sourceCount := 0
	aggregation := AggregationSum
	s.state.mu.Lock()
	for _, record := range records.Items {
		template := s.templateByIDLocked(record.MetricTemplateID)
		if template.ID != "" {
			aggregation = template.AggregationMethod
		}
		series = append(series, MetricTrendPoint{BucketStart: record.MetricDate, Value: record.NormalizedValue, SourceRecordCount: 1, Missing: false})
		sourceCount++
	}
	s.state.mu.Unlock()
	return MetricTrendResponse{
		ProjectID:         projectID,
		MetricCode:        req.MetricCode,
		Bucket:            req.Bucket,
		AggregationMethod: aggregation,
		QuerySignature:    strings.Join([]string{projectID, req.MetricCode, req.Platform, req.TargetID, req.DateFrom, req.DateTo, req.Bucket}, ":"),
		SourceRecordCount: sourceCount,
		Series:            series,
		MissingPoints:     []MetricMissingPoint{},
	}, nil
}

func (s *service) GetMissingDates(ctx context.Context, projectID string, req MissingMetricDatesRequest) (MissingMetricDatesResponse, error) {
	if projectID == "" || req.DateFrom == "" || req.DateTo == "" {
		return MissingMetricDatesResponse{}, ErrValidation
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	items := make([]MissingMetricDateItem, 0)
	for _, template := range s.state.templates {
		if !template.Enabled || !template.Required {
			continue
		}
		if req.Platform != "" && template.Platform != req.Platform {
			continue
		}
		if req.MetricCode != "" && template.MetricCode != req.MetricCode {
			continue
		}
		exists := false
		for _, record := range s.state.records {
			if record.ProjectID == projectID && record.Platform == template.Platform && record.MetricCode == template.MetricCode && record.MetricDate == req.DateTo {
				exists = true
			}
		}
		if exists {
			continue
		}
		targetID := req.TargetID
		if targetID == "" {
			targetID = "publish-target-1"
		}
		items = append(items, MissingMetricDateItem{
			ContentItemID:    "content-item-1",
			ContentVersionID: "content-version-approved-1",
			PublishJobID:     "publish-job-1",
			TargetID:         targetID,
			Platform:         template.Platform,
			MetricCode:       template.MetricCode,
			Period:           template.Period,
			MetricDate:       req.DateFrom,
			MissingReason:    "required_metric_missing",
			BackfillHint:     "补录 " + template.MetricCode + " " + req.DateFrom,
		})
	}
	return MissingMetricDatesResponse{ProjectID: projectID, Items: items}, nil
}

func (s *service) templateByIDLocked(id string) MetricTemplateResponse {
	for _, template := range s.state.templates {
		if template.ID == id {
			return template
		}
	}
	return MetricTemplateResponse{}
}
