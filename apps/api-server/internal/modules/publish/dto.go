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

	EventPluginLocked    = "plugin_locked"
	EventPluginFilled    = "plugin_filled"
	EventPluginPublished = "plugin_published"
	EventPluginFailed    = "plugin_failed"

	PluginScopePublishRead  = "publish:read"
	PluginScopePublishWrite = "publish:write"
	PluginScopeCollectWrite = "collect:write"
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
	TargetType    string    `json:"target_type"`
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
	TargetType  string         `json:"target_type"`
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
	TargetType  string         `json:"target_type"`
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
	AdapterConfigID  string     `json:"adapter_config_id"`
	ScheduledAt      *time.Time `json:"scheduled_at"`
}

type CreatePublishJobResponse struct {
	PublishJobID    string `json:"publish_job_id"`
	Status          string `json:"status"`
	PayloadHash     string `json:"payload_hash"`
	AdapterConfigID string `json:"adapter_config_id"`
	AdapterVersion  int    `json:"adapter_version"`
	OperationLogID  string `json:"operation_log_id"`
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
	AdapterConfigID  string     `json:"adapter_config_id"`
	AdapterVersion   int        `json:"adapter_version"`
	LockedUntil      *time.Time `json:"locked_until,omitempty"`
	PlatformErrorSummary string     `json:"platform_error_summary"`
	PluginLockID         string     `json:"-"`
	CreatedAt            time.Time  `json:"created_at"`
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

type CreatePlatformAdapterRequest struct {
	Platform      string         `json:"platform"`
	DisplayName   string         `json:"display_name"`
	PublishMode   string         `json:"publish_mode"`
	TargetType    string         `json:"target_type"`
	FieldMapping  map[string]any `json:"field_mapping"`
	FillRules     map[string]any `json:"fill_rules"`
	CollectRules  map[string]any `json:"collect_rules"`
	CredentialRef string         `json:"credential_ref"`
	Enabled       bool           `json:"enabled"`
}

type UpdatePlatformAdapterRequest struct {
	DisplayName     string         `json:"display_name"`
	PublishMode     string         `json:"publish_mode"`
	TargetType      string         `json:"target_type"`
	FieldMapping    map[string]any `json:"field_mapping"`
	FillRules       map[string]any `json:"fill_rules"`
	CollectRules    map[string]any `json:"collect_rules"`
	CredentialRef   string         `json:"credential_ref"`
	Enabled         *bool          `json:"enabled"`
	ExpectedVersion int            `json:"expected_version"`
	ChangeReason    string         `json:"change_reason"`
}

type ListPlatformAdaptersRequest struct {
	content.PaginationRequest
	Platform    string `json:"platform"`
	PublishMode string `json:"publish_mode"`
	Enabled     *bool  `json:"enabled"`
	Sort        string `json:"sort"`
	Order       string `json:"order"`
}

type PlatformAdapterResponse struct {
	ID          string    `json:"id"`
	Platform    string    `json:"platform"`
	DisplayName string    `json:"display_name"`
	PublishMode string    `json:"publish_mode"`
	TargetType  string    `json:"target_type"`
	Enabled     bool      `json:"enabled"`
	Version     int       `json:"version"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PlatformAdapterDetailResponse struct {
	PlatformAdapterResponse
	FieldMapping  map[string]any `json:"field_mapping"`
	FillRules     map[string]any `json:"fill_rules"`
	CollectRules  map[string]any `json:"collect_rules"`
	CredentialRef string         `json:"credential_ref"`
	RuleSummary   string         `json:"rule_summary"`
}

type PagedPlatformAdaptersResponse struct {
	Items      []PlatformAdapterResponse  `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type CreatePlatformAdapterResponse struct {
	AdapterID      string `json:"adapter_id"`
	Version        int    `json:"version"`
	OperationLogID string `json:"operation_log_id"`
}

type UpdatePlatformAdapterResponse = CreatePlatformAdapterResponse

type RegisterPluginClientRequest struct {
	Name       string   `json:"name"`
	ClientType string   `json:"client_type"`
	Version    string   `json:"version"`
	Scopes     []string `json:"scopes"`
}

type ListPluginClientsRequest struct {
	content.PaginationRequest
	Status     string `json:"status"`
	ClientType string `json:"client_type"`
}

type UpdatePluginClientRequest struct {
	Status       string   `json:"status"`
	Scopes       []string `json:"scopes"`
	ChangeReason string   `json:"change_reason"`
}

type RotatePluginClientKeyRequest struct {
	Reason string `json:"reason"`
}

type PluginClientResponse struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	ClientType   string     `json:"client_type"`
	Version      string     `json:"version"`
	Scopes       []string   `json:"scopes"`
	Status       string     `json:"status"`
	APIKeyMasked string     `json:"api_key_masked"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
}

type PagedPluginClientsResponse struct {
	Items      []PluginClientResponse     `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type RegisterPluginClientResponse struct {
	ClientID     string `json:"client_id"`
	APIKeyOnce   string `json:"api_key_once"`
	APIKeyMasked string `json:"api_key_masked"`
}

type UpdatePluginClientResponse struct {
	ClientID       string `json:"client_id"`
	Status         string `json:"status"`
	OperationLogID string `json:"operation_log_id"`
}

type RotatePluginClientKeyResponse struct {
	ClientID       string `json:"client_id"`
	APIKeyOnce     string `json:"api_key_once"`
	APIKeyMasked   string `json:"api_key_masked"`
	OperationLogID string `json:"operation_log_id"`
}

type PluginAuthRequest struct {
	APIKey        string `json:"api_key"`
	ClientVersion string `json:"client_version"`
}

type PluginAuthTokenResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	Scopes      []string  `json:"scopes"`
}

type ListPluginPublishJobsRequest struct {
	content.PaginationRequest
	ProjectID string `json:"project_id"`
	Platform  string `json:"platform"`
	Status    string `json:"status"`
}

type PluginPublishJobResponse struct {
	ID              string     `json:"id"`
	ProjectID       string     `json:"project_id"`
	Platform        string     `json:"platform"`
	TargetDisplay   string     `json:"target_display"`
	Status          string     `json:"status"`
	PayloadHash     string     `json:"payload_hash"`
	LockedUntil     *time.Time `json:"locked_until,omitempty"`
	AdapterConfigID string     `json:"adapter_config_id"`
	AdapterVersion  int        `json:"adapter_version"`
}

type PagedPluginPublishJobsResponse struct {
	Items      []PluginPublishJobResponse `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type LockPluginPublishJobRequest struct {
	LockTTLSeconds int `json:"lock_ttl_seconds"`
}

type PluginPublishJobLockResponse struct {
	LockID           string         `json:"lock_id"`
	LockedUntil      time.Time      `json:"locked_until"`
	Payload          map[string]any `json:"payload"`
	PayloadHash      string         `json:"payload_hash"`
	ContentVersionID string         `json:"content_version_id"`
	AdapterConfigID  string         `json:"adapter_config_id"`
	AdapterVersion   int            `json:"adapter_version"`
}

type MarkPluginPublishJobFilledRequest struct {
	LockID      string `json:"lock_id"`
	PayloadHash string `json:"payload_hash"`
	Note        string `json:"note"`
}

type PluginPublishJobFilledResponse struct {
	EventID       string `json:"event_id"`
	CurrentStatus string `json:"current_status"`
}

type MarkPluginPublishJobPublishedRequest struct {
	LockID      string     `json:"lock_id"`
	ExternalURL string     `json:"external_url"`
	PublishedAt *time.Time `json:"published_at"`
	PayloadHash string     `json:"payload_hash"`
	Note        string     `json:"note"`
}

type PluginPublishJobPublishedResponse struct {
	PublishJobID   string `json:"publish_job_id"`
	CurrentStatus  string `json:"current_status"`
	OperationLogID string `json:"operation_log_id"`
}

type MarkPluginPublishJobFailedRequest struct {
	LockID               string `json:"lock_id"`
	Reason               string `json:"reason"`
	Retryable            bool   `json:"retryable"`
	PlatformErrorSummary string `json:"platform_error_summary"`
}

type PluginPublishJobFailedResponse = PluginPublishJobPublishedResponse
