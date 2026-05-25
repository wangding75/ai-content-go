package metrics

import (
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

const (
	ValueTypeInteger    = "integer"
	ValueTypeDecimal    = "decimal"
	ValueTypePercentage = "percentage"
	ValueTypeCurrency   = "currency"
	ValueTypeDuration   = "duration"

	AggregationSum    = "sum"
	AggregationAvg    = "avg"
	AggregationMax    = "max"
	AggregationMin    = "min"
	AggregationLatest = "latest"

	PeriodDay   = "day"
	PeriodWeek  = "week"
	PeriodMonth = "month"

	SourceManual           = "manual"
	SourceImport           = "import"
	SourceExtension        = "extension"
	SourceExternalCallback = "external_callback"
)

type CreateMetricTemplateRequest struct {
	ContentType       string `json:"content_type"`
	Platform          string `json:"platform"`
	MetricCode        string `json:"metric_code"`
	MetricName        string `json:"metric_name"`
	Unit              string `json:"unit"`
	ValueType         string `json:"value_type"`
	AggregationMethod string `json:"aggregation_method"`
	Period            string `json:"period"`
	Required          bool   `json:"required"`
	Enabled           bool   `json:"enabled"`
}

type CreateMetricTemplateResponse struct {
	MetricTemplateID string `json:"metric_template_id"`
}

type ListMetricTemplatesRequest struct {
	content.PaginationRequest
	ContentType string `json:"content_type"`
	Platform    string `json:"platform"`
	Enabled     *bool  `json:"enabled"`
}

type MetricTemplateResponse struct {
	ID                string    `json:"id"`
	ContentType       string    `json:"content_type"`
	Platform          string    `json:"platform"`
	MetricCode        string    `json:"metric_code"`
	MetricName        string    `json:"metric_name"`
	Unit              string    `json:"unit"`
	ValueType         string    `json:"value_type"`
	AggregationMethod string    `json:"aggregation_method"`
	Period            string    `json:"period"`
	Required          bool      `json:"required"`
	Enabled           bool      `json:"enabled"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type PagedMetricTemplatesResponse struct {
	Items      []MetricTemplateResponse   `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type CreateMetricRecordRequest struct {
	ProjectID        string `json:"project_id"`
	ContentItemID    string `json:"content_item_id"`
	ContentVersionID string `json:"content_version_id"`
	PublishJobID     string `json:"publish_job_id"`
	TargetID         string `json:"target_id"`
	Platform         string `json:"platform"`
	ExternalURL      string `json:"external_url"`
	MetricCode       string `json:"metric_code"`
	MetricDate       string `json:"metric_date"`
	Period           string `json:"period"`
	RawValue         string `json:"raw_value"`
	SourceType       string `json:"source_type"`
	SourceRef        string `json:"source_ref"`
}

type CreateMetricRecordResponse struct {
	MetricRecordID  string  `json:"metric_record_id"`
	NormalizedValue float64 `json:"normalized_value"`
	OperationLogID  string  `json:"operation_log_id"`
}

type BatchCreateMetricRecordsRequest struct {
	Records      []CreateMetricRecordRequest `json:"records"`
	ImportSource string                      `json:"import_source"`
}

type BatchMetricRecordError struct {
	Index      int    `json:"index"`
	MetricCode string `json:"metric_code"`
	Field      string `json:"field"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	SourceRef  string `json:"source_ref"`
}

type BatchCreateMetricRecordsResponse struct {
	CreatedCount   int                      `json:"created_count"`
	FailedCount    int                      `json:"failed_count"`
	Errors         []BatchMetricRecordError `json:"errors"`
	OperationLogID string                   `json:"operation_log_id"`
}

type ListMetricRecordsRequest struct {
	content.PaginationRequest
	ProjectID     string `json:"project_id"`
	Platform      string `json:"platform"`
	TargetID      string `json:"target_id"`
	ContentItemID string `json:"content_item_id"`
	MetricCode    string `json:"metric_code"`
	DateFrom      string `json:"date_from"`
	DateTo        string `json:"date_to"`
}

type MetricRecordResponse struct {
	ID               string    `json:"id"`
	ProjectID        string    `json:"project_id"`
	ContentItemID    string    `json:"content_item_id"`
	ContentVersionID string    `json:"content_version_id"`
	PublishJobID     string    `json:"publish_job_id"`
	TargetID         string    `json:"target_id"`
	ContentType      string    `json:"content_type"`
	MetricTemplateID string    `json:"metric_template_id"`
	Platform         string    `json:"platform"`
	ExternalURL      string    `json:"external_url"`
	MetricCode       string    `json:"metric_code"`
	MetricDate       string    `json:"metric_date"`
	Period           string    `json:"period"`
	RawValue         string    `json:"raw_value"`
	NormalizedValue  float64   `json:"normalized_value"`
	SourceType       string    `json:"source_type"`
	SourceRef        string    `json:"source_ref"`
	CollectedAt      time.Time `json:"collected_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type PagedMetricRecordsResponse struct {
	Items      []MetricRecordResponse     `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type MetricSummaryRequest struct {
	DateFrom    string   `json:"date_from"`
	DateTo      string   `json:"date_to"`
	Platform    string   `json:"platform"`
	TargetID    string   `json:"target_id"`
	MetricCodes []string `json:"metric_codes"`
}

type MetricSummaryItem struct {
	MetricCode        string  `json:"metric_code"`
	Value             float64 `json:"value"`
	Unit              string  `json:"unit"`
	AggregationMethod string  `json:"aggregation_method"`
	SourceRecordCount int     `json:"source_record_count"`
}

type MetricSummaryResponse struct {
	ProjectID         string              `json:"project_id"`
	DateFrom          string              `json:"date_from"`
	DateTo            string              `json:"date_to"`
	Platform          string              `json:"platform"`
	TargetID          string              `json:"target_id"`
	Items             []MetricSummaryItem `json:"items"`
	SummarySnapshotID string              `json:"summary_snapshot_id"`
	SourceRecordCount int                 `json:"source_record_count"`
}

type MetricTrendRequest struct {
	MetricCode string `json:"metric_code"`
	DateFrom   string `json:"date_from"`
	DateTo     string `json:"date_to"`
	Bucket     string `json:"bucket"`
	Platform   string `json:"platform"`
	TargetID   string `json:"target_id"`
}

type MetricTrendPoint struct {
	BucketStart       string  `json:"bucket_start"`
	Value             float64 `json:"value"`
	SourceRecordCount int     `json:"source_record_count"`
	Missing           bool    `json:"missing"`
}

type MetricMissingPoint struct {
	MetricDate string `json:"metric_date"`
	Reason     string `json:"reason"`
}

type MetricTrendResponse struct {
	ProjectID         string               `json:"project_id"`
	MetricCode        string               `json:"metric_code"`
	Bucket            string               `json:"bucket"`
	AggregationMethod string               `json:"aggregation_method"`
	QuerySignature    string               `json:"query_signature"`
	SourceRecordCount int                  `json:"source_record_count"`
	Series            []MetricTrendPoint   `json:"series"`
	MissingPoints     []MetricMissingPoint `json:"missing_points"`
}

type MissingMetricDatesRequest struct {
	MetricCode string `json:"metric_code"`
	Platform   string `json:"platform"`
	TargetID   string `json:"target_id"`
	DateFrom   string `json:"date_from"`
	DateTo     string `json:"date_to"`
}

type MissingMetricDateItem struct {
	ContentItemID    string `json:"content_item_id"`
	ContentVersionID string `json:"content_version_id"`
	PublishJobID     string `json:"publish_job_id"`
	TargetID         string `json:"target_id"`
	Platform         string `json:"platform"`
	MetricCode       string `json:"metric_code"`
	Period           string `json:"period"`
	MetricDate       string `json:"metric_date"`
	MissingReason    string `json:"missing_reason"`
	BackfillHint     string `json:"backfill_hint"`
}

type MissingMetricDatesResponse struct {
	ProjectID string                  `json:"project_id"`
	Items     []MissingMetricDateItem `json:"items"`
}
