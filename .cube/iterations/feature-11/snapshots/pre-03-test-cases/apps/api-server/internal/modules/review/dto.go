package review

import (
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

const (
	ReviewStatusPending          = "pending"
	ReviewStatusInReview         = "in_review"
	ReviewStatusApproved         = "approved"
	ReviewStatusRejected         = "rejected"
	ReviewStatusApprovedWithEdit = "approved_with_edit"

	ReviewReportStatusPending    = "pending"
	ReviewReportStatusGenerating = "generating"
	ReviewReportStatusSucceeded  = "succeeded"
	ReviewReportStatusFailed     = "failed"
)

type CreateReviewRequest struct {
	ReviewType string `json:"review_type"`
}

type CreateReviewResponse struct {
	ReviewID string `json:"review_id"`
	Status   string `json:"status"`
}

type ListReviewsRequest struct {
	content.PaginationRequest
	ProjectID string `json:"project_id"`
	Status    string `json:"status"`
}

type ContentReviewResponse struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	ContentItemID string    `json:"content_item_id"`
	ReviewType    string    `json:"review_type"`
	Status        string    `json:"status"`
	Title         string    `json:"title"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PagedContentReviewsResponse struct {
	Items      []ContentReviewResponse    `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type ReviewReportSummaryResponse struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	QualityScore int    `json:"quality_score"`
	RiskLevel    string `json:"risk_level"`
}

type ContentVersionResponse struct {
	ID             string         `json:"id"`
	ContentItemID  string         `json:"content_item_id"`
	VersionNo      int            `json:"version_no"`
	Source         string         `json:"source"`
	Title          string         `json:"title"`
	Body           string         `json:"body"`
	EditableFields map[string]any `json:"editable_fields"`
	Summary        string         `json:"summary"`
	CreatedAt      time.Time      `json:"created_at"`
}

type PagedContentVersionsResponse struct {
	Items      []ContentVersionResponse   `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type ContentReviewDetailResponse struct {
	ContentReviewResponse
	Body          string                      `json:"body"`
	Metadata      map[string]any              `json:"metadata"`
	Extension     map[string]any              `json:"extension"`
	ReportSummary ReviewReportSummaryResponse `json:"report_summary"`
	Versions      []ContentVersionResponse    `json:"versions"`
}

type TriggerAIReportRequest struct {
	ReportType string         `json:"report_type"`
	Config     map[string]any `json:"config"`
}

type TriggerAIReportResponse struct {
	ReportID      string `json:"report_id"`
	JobID         string `json:"job_id"`
	WorkflowRunID string `json:"workflow_run_id"`
	Status        string `json:"status"`
}

type ReviewReportResponse struct {
	ID            string             `json:"id"`
	ReviewID      string             `json:"review_id"`
	ContentItemID string             `json:"content_item_id"`
	Status        string             `json:"status"`
	QualityScore  int                `json:"quality_score"`
	RiskLevel     string             `json:"risk_level"`
	Issues        []ReviewIssue      `json:"issues"`
	Suggestions   []ReviewSuggestion `json:"suggestions"`
	JobID         string             `json:"job_id,omitempty"`
	WorkflowRunID string             `json:"workflow_run_id,omitempty"`
	ErrorCode     string             `json:"error_code,omitempty"`
	ErrorMessage  string             `json:"error_message,omitempty"`
}

type ReviewIssue struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type ReviewSuggestion struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ApproveReviewRequest struct {
	Note string `json:"note"`
}

type ApproveReviewResponse struct {
	ReviewID       string `json:"review_id"`
	Status         string `json:"status"`
	OperationLogID string `json:"operation_log_id"`
}

type RejectReviewRequest struct {
	Reason                string `json:"reason"`
	RegenerateInstruction string `json:"regenerate_instruction"`
	TriggerRegeneration   bool   `json:"trigger_regeneration"`
}

type RejectReviewResponse struct {
	ReviewID          string `json:"review_id"`
	Status            string `json:"status"`
	OperationLogID    string `json:"operation_log_id"`
	RegenerationRunID string `json:"regeneration_run_id,omitempty"`
	JobID             string `json:"job_id,omitempty"`
}

type ApproveWithEditRequest struct {
	EditableFields map[string]any `json:"editable_fields"`
	Note           string         `json:"note"`
}

type ApproveWithEditResponse struct {
	ReviewID         string `json:"review_id"`
	Status           string `json:"status"`
	ContentVersionID string `json:"content_version_id"`
	OperationLogID   string `json:"operation_log_id"`
}
