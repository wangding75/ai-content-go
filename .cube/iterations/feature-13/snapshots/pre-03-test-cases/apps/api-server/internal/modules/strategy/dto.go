package strategy

import (
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

const (
	SuggestionTypeKeep        = "keep"
	SuggestionTypeOptimize    = "optimize"
	SuggestionTypeSuspend     = "suspend"
	SuggestionTypePromote     = "promote"
	SuggestionTypeCostControl = "cost_control"

	RiskLevelLow    = "low"
	RiskLevelMedium = "medium"
	RiskLevelHigh   = "high"

	ConfidenceLow    = "low"
	ConfidenceMedium = "medium"
	ConfidenceHigh   = "high"

	StatusPending         = "pending"
	StatusConfirmed       = "confirmed"
	StatusIgnored         = "ignored"
	StatusExecuted        = "executed"
	StatusExecutionFailed = "execution_failed"

	RunStatusGenerating = "generating"
	RunStatusCompleted  = "completed"
	RunStatusFailed     = "failed"

	ExecutionResultSuccess = "success"
	ExecutionResultFailed  = "failed"

	GenerationMethodRule = "rule"
)

type GenerateSuggestionsRequest struct {
	DateFrom         string   `json:"date_from"`
	DateTo           string   `json:"date_to"`
	RuleCodes        []string `json:"rule_codes"`
	MetricCodes      []string `json:"metric_codes"`
	ForceRegenerate  bool     `json:"force_regenerate"`
}

type GenerateSuggestionsResponse struct {
	SuggestionRunID string `json:"suggestion_run_id"`
	Status          string `json:"status"`
}

type ListStrategySuggestionsRequest struct {
	content.PaginationRequest
	ProjectID      string `json:"project_id"`
	Status         string `json:"status"`
	SuggestionType string `json:"suggestion_type"`
	RiskLevel      string `json:"risk_level"`
	Confidence     string `json:"confidence"`
	DateFrom       string `json:"date_from"`
	DateTo         string `json:"date_to"`
	Sort           string `json:"sort"`
	Order          string `json:"order"`
}

type StrategySuggestionItem struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	SuggestionType string    `json:"suggestion_type"`
	Title          string    `json:"title"`
	TriggerReason  string    `json:"trigger_reason"`
	RiskLevel      string    `json:"risk_level"`
	Confidence     string    `json:"confidence"`
	Status         string    `json:"status"`
	DateFrom       string    `json:"date_from"`
	DateTo         string    `json:"date_to"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PagedStrategySuggestionsResponse struct {
	Items      []StrategySuggestionItem   `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type StrategySuggestionDetailResponse struct {
	ID               string          `json:"id"`
	ProjectID        string          `json:"project_id"`
	SuggestionRunID  string          `json:"suggestion_run_id"`
	SuggestionType   string          `json:"suggestion_type"`
	Title            string          `json:"title"`
	TriggerReason    string          `json:"trigger_reason"`
	EvidenceMetrics  []MetricEvidence `json:"evidence_metrics"`
	ImpactScope      string          `json:"impact_scope"`
	RiskLevel        string          `json:"risk_level"`
	Confidence       string          `json:"confidence"`
	SuggestedAction  string          `json:"suggested_action"`
	ExpectedBenefit  string          `json:"expected_benefit"`
	MetricsSnapshot  MetricsSnapshot `json:"metrics_snapshot"`
	Status           string          `json:"status"`
	IgnoredReason    string          `json:"ignored_reason"`
	IgnoredNote      string          `json:"ignored_note"`
	ConfirmedAt      *time.Time      `json:"confirmed_at"`
	IgnoredAt        *time.Time      `json:"ignored_at"`
	ExecutedAt       *time.Time      `json:"executed_at"`
	DateFrom         string          `json:"date_from"`
	DateTo           string          `json:"date_to"`
	TriggeredRules   []string        `json:"triggered_rules"`
	GenerationMethod string          `json:"generation_method"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type MetricEvidence struct {
	MetricCode string  `json:"metric_code"`
	Value      float64 `json:"value"`
	Trend      string  `json:"trend"`
}

type MetricsSnapshot struct {
	SummarySnapshotID string `json:"summary_snapshot_id"`
}

type ConfirmSuggestionRequest struct {
	Note string `json:"note"`
}

type IgnoreSuggestionRequest struct {
	Reason string `json:"reason"`
	Note   string `json:"note"`
}

type ExecuteSuggestionRequest struct {
	ActionType   string `json:"action_type"`
	TargetType   string `json:"target_type"`
	TargetID     string `json:"target_id"`
	OperatorNote string `json:"operator_note"`
}

type RetrySuggestionRequest struct {
	OperatorNote string `json:"operator_note"`
}

type SuggestionStatusChangeResponse struct {
	SuggestionID   string `json:"suggestion_id"`
	PreviousStatus string `json:"previous_status"`
	CurrentStatus  string `json:"current_status"`
	OperationLogID string `json:"operation_log_id"`
}

type ExecuteSuggestionResponse struct {
	ExecutionLogID string `json:"execution_log_id"`
	SuggestionID   string `json:"suggestion_id"`
	PreviousStatus string `json:"previous_status"`
	CurrentStatus  string `json:"current_status"`
	OperationLogID string `json:"operation_log_id"`
}

type ListExecutionLogsRequest struct {
	content.PaginationRequest
	SuggestionID string `json:"suggestion_id"`
}

type ExecutionLogResponse struct {
	ID             string    `json:"id"`
	SuggestionID   string    `json:"suggestion_id"`
	ActionType     string    `json:"action_type"`
	TargetType     string    `json:"target_type"`
	TargetID       string    `json:"target_id"`
	OperatorNote   string    `json:"operator_note"`
	PreviousStatus string    `json:"previous_status"`
	CurrentStatus  string    `json:"current_status"`
	Result         string    `json:"result"`
	FailureReason  string    `json:"failure_reason"`
	TargetInterface string   `json:"target_interface"`
	TargetResource  string   `json:"target_resource"`
	CreatedAt      time.Time `json:"created_at"`
}

type PagedExecutionLogsResponse struct {
	Items      []ExecutionLogResponse     `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type StrategySuggestionRunResponse struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"project_id"`
	DateFrom        string    `json:"date_from"`
	DateTo          string    `json:"date_to"`
	RuleCodes       []string  `json:"rule_codes"`
	MetricCodes     []string  `json:"metric_codes"`
	ForceRegenerate bool      `json:"force_regenerate"`
	Status          string    `json:"status"`
	FailureReason   string    `json:"failure_reason"`
	SuggestionCount int       `json:"suggestion_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
