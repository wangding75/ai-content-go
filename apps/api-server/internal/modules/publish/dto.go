package publish

import (
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/review"
)

const (
	JobStatusQueued    = "queued"
	JobStatusCopied    = "copied"
	JobStatusPublished = "published"
	JobStatusFailed    = "failed"
	JobStatusCanceled  = "canceled"

	EventJobCreated      = "job_created"
	EventPayloadCopied   = "payload_copied"
	EventMarkedPublished = "marked_published"
	EventMarkedFailed    = "marked_failed"
	EventRequeued        = "requeued"
)

type ListPublishTargetsRequest struct {
	content.PaginationRequest
	Enabled *bool `json:"enabled"`
}

type PublishTargetResponse struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	Platform      string    `json:"platform"`
	AccountName   string    `json:"account_name"`
	DisplayName   string    `json:"display_name"`
	ConfigSummary string    `json:"config_summary"`
	Enabled       bool      `json:"enabled"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PagedPublishTargetsResponse struct {
	Items      []PublishTargetResponse    `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type CreatePublishTargetRequest struct {
	Platform    string         `json:"platform"`
	AccountName string         `json:"account_name"`
	DisplayName string         `json:"display_name"`
	Config      map[string]any `json:"config"`
	Enabled     bool           `json:"enabled"`
}

type CreatePublishTargetResponse struct {
	TargetID       string `json:"target_id"`
	OperationLogID string `json:"operation_log_id"`
}

type UpdatePublishTargetRequest struct {
	Platform    string         `json:"platform"`
	AccountName string         `json:"account_name"`
	DisplayName string         `json:"display_name"`
	Config      map[string]any `json:"config"`
	Enabled     bool           `json:"enabled"`
	Note        string         `json:"note"`
}

type UpdatePublishTargetResponse struct {
	TargetID       string `json:"target_id"`
	OperationLogID string `json:"operation_log_id"`
}

type CreatePublishJobRequest struct {
	ContentItemID    string     `json:"content_item_id"`
	ContentVersionID string     `json:"content_version_id"`
	TargetID         string     `json:"target_id"`
	ScheduledAt      *time.Time `json:"scheduled_at"`
}

type CreatePublishJobResponse struct {
	PublishJobID   string `json:"publish_job_id"`
	Status         string `json:"status"`
	PayloadHash    string `json:"payload_hash"`
	OperationLogID string `json:"operation_log_id"`
}

type ListPublishJobsRequest struct {
	content.PaginationRequest
	TargetID      string     `json:"target_id"`
	Status        string     `json:"status"`
	ScheduledFrom *time.Time `json:"scheduled_from"`
}

type PublishJobResponse struct {
	ID               string     `json:"id"`
	ProjectID        string     `json:"project_id"`
	ContentItemID    string     `json:"content_item_id"`
	ContentVersionID string     `json:"content_version_id"`
	TargetID         string     `json:"target_id"`
	Title            string     `json:"title"`
	TargetPlatform   string     `json:"target_platform"`
	TargetDisplay    string     `json:"target_display"`
	Status           string     `json:"status"`
	PayloadHash      string     `json:"payload_hash"`
	ScheduledAt      *time.Time `json:"scheduled_at,omitempty"`
	CopiedAt         *time.Time `json:"copied_at,omitempty"`
	PublishedAt      *time.Time `json:"published_at,omitempty"`
	LastError        string     `json:"last_error"`
	RetryCount       int        `json:"retry_count"`
	Actions          []string   `json:"actions"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type PagedPublishJobsResponse struct {
	Items      []PublishJobResponse       `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type PublishLogResponse struct {
	ID              string         `json:"id"`
	PublishJobID    string         `json:"publish_job_id"`
	EventType       string         `json:"event_type"`
	FromStatus      string         `json:"from_status"`
	ToStatus        string         `json:"to_status"`
	ActorID         string         `json:"actor_id"`
	Reason          string         `json:"reason"`
	Note            string         `json:"note"`
	PayloadSnapshot map[string]any `json:"payload_snapshot"`
	CreatedAt       time.Time      `json:"created_at"`
}

type PublishJobDetailResponse struct {
	ID               string                        `json:"id"`
	ProjectID        string                        `json:"project_id"`
	ContentItemID    string                        `json:"content_item_id"`
	ContentVersionID string                        `json:"content_version_id"`
	TargetID         string                        `json:"target_id"`
	Title            string                        `json:"title"`
	TargetPlatform   string                        `json:"target_platform"`
	TargetDisplay    string                        `json:"target_display"`
	Status           string                        `json:"status"`
	PayloadHash      string                        `json:"payload_hash"`
	ScheduledAt      *time.Time                    `json:"scheduled_at,omitempty"`
	CopiedAt         *time.Time                    `json:"copied_at,omitempty"`
	PublishedAt      *time.Time                    `json:"published_at,omitempty"`
	LastError        string                        `json:"last_error"`
	RetryCount       int                           `json:"retry_count"`
	Actions          []string                      `json:"actions"`
	CreatedAt        time.Time                     `json:"created_at"`
	UpdatedAt        time.Time                     `json:"updated_at"`
	Target           PublishTargetResponse         `json:"target"`
	ContentVersion   review.ContentVersionResponse `json:"content_version"`
	Logs             []PublishLogResponse          `json:"logs"`
	ExternalURL      string                        `json:"external_url"`
}

type PublishCopyPayloadResponse struct {
	PublishJobID     string `json:"publish_job_id"`
	Title            string `json:"title"`
	Body             string `json:"body"`
	Format           string `json:"format"`
	Platform         string `json:"platform"`
	TargetID         string `json:"target_id"`
	ContentVersionID string `json:"content_version_id"`
	PayloadHash      string `json:"payload_hash"`
}

type CopyPublishPayloadRequest struct {
	CopyScope string `json:"copy_scope"`
	Note      string `json:"note"`
}

type CopyPublishPayloadResponse struct {
	PublishJobID   string    `json:"publish_job_id"`
	PreviousStatus string    `json:"previous_status"`
	CurrentStatus  string    `json:"current_status"`
	CopiedAt       time.Time `json:"copied_at"`
	PublishLogID   string    `json:"publish_log_id"`
}

type MarkPublishedRequest struct {
	ExternalURL string     `json:"external_url"`
	PublishedAt *time.Time `json:"published_at"`
	Reason      string     `json:"reason"`
	Note        string     `json:"note"`
}

type MarkPublishedResponse struct {
	PublishJobID   string    `json:"publish_job_id"`
	PreviousStatus string    `json:"previous_status"`
	CurrentStatus  string    `json:"current_status"`
	ExternalURL    string    `json:"external_url"`
	PublishedAt    time.Time `json:"published_at"`
	OperationLogID string    `json:"operation_log_id"`
	PublishLogID   string    `json:"publish_log_id"`
}

type MarkFailedRequest struct {
	Reason    string `json:"reason"`
	Retryable bool   `json:"retryable"`
	Note      string `json:"note"`
}

type MarkFailedResponse struct {
	PublishJobID   string    `json:"publish_job_id"`
	PreviousStatus string    `json:"previous_status"`
	CurrentStatus  string    `json:"current_status"`
	FailedAt       time.Time `json:"failed_at"`
	OperationLogID string    `json:"operation_log_id"`
	PublishLogID   string    `json:"publish_log_id"`
}

type RequeuePublishJobRequest struct {
	Reason      string     `json:"reason"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	Note        string     `json:"note"`
}

type RequeuePublishJobResponse struct {
	PublishJobID   string `json:"publish_job_id"`
	PreviousStatus string `json:"previous_status"`
	CurrentStatus  string `json:"current_status"`
	RetryCount     int    `json:"retry_count"`
	OperationLogID string `json:"operation_log_id"`
	PublishLogID   string `json:"publish_log_id"`
}
