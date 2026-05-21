package memory

import "time"

type SnapshotSourceType string

type ConsistencyReportStatus string

const (
	SnapshotSourceAssembleContext        SnapshotSourceType = "assemble_context"
	SnapshotSourceDynamicStateUpdate     SnapshotSourceType = "dynamic_state_update"
	SnapshotSourceDynamicStateCorrection SnapshotSourceType = "dynamic_state_correction"

	ReportStatusPending   ConsistencyReportStatus = "pending"
	ReportStatusRunning   ConsistencyReportStatus = "running"
	ReportStatusCompleted ConsistencyReportStatus = "completed"
	ReportStatusFailed    ConsistencyReportStatus = "failed"
)

type RecentWindowPolicy struct {
	ItemCount        int    `json:"item_count"`
	TokenLimit       int    `json:"token_limit"`
	TruncationPolicy string `json:"truncation_policy"`
}

type SnapshotSummaryResponse struct {
	ID               string    `json:"id"`
	SourceType       string    `json:"source_type"`
	EstimatedTokens  int       `json:"estimated_tokens"`
	TruncationPolicy string    `json:"truncation_policy"`
	CreatedAt        time.Time `json:"created_at"`
}

type KnowledgeMemoryResponse struct {
	ID                    string                  `json:"id"`
	ProjectID             string                  `json:"project_id"`
	StaticContext         map[string]any          `json:"static_context"`
	DynamicState          map[string]any          `json:"dynamic_state"`
	RecentWindowPolicy    RecentWindowPolicy      `json:"recent_window_policy"`
	StyleGuide            map[string]any          `json:"style_guide"`
	Version               int                     `json:"version"`
	UpdatedAt             time.Time               `json:"updated_at"`
	RecentSnapshotSummary SnapshotSummaryResponse `json:"recent_snapshot_summary"`
}

type UpdateStaticContextRequest struct {
	StaticContext map[string]any `json:"static_context"`
	Note          string         `json:"note"`
}

type UpdateStyleGuideRequest struct {
	StyleGuide map[string]any `json:"style_guide"`
	Note       string         `json:"note"`
}

type MemoryUpdateResponse struct {
	Version        int    `json:"version"`
	OperationLogID string `json:"operation_log_id"`
}

type CorrectDynamicStateRequest struct {
	Reason     string         `json:"reason"`
	Changes    map[string]any `json:"changes"`
	SourceRefs []string       `json:"source_refs"`
}

type DynamicStateCorrectionResponse struct {
	MemorySnapshotID    string `json:"memory_snapshot_id"`
	DynamicStateVersion int    `json:"dynamic_state_version"`
	OperationLogID      string `json:"operation_log_id"`
}

type UpdateRecentWindowPolicyRequest struct {
	ItemCount        int    `json:"item_count"`
	TokenLimit       int    `json:"token_limit"`
	TruncationPolicy string `json:"truncation_policy"`
}

type RecentWindowPolicyResponse struct {
	RecentWindowPolicy
	Version        int    `json:"version"`
	OperationLogID string `json:"operation_log_id"`
}

type ListSnapshotsRequest struct {
	ContentItemID string `json:"content_item_id"`
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
}

type MemorySnapshotResponse struct {
	ID               string    `json:"id"`
	ProjectID        string    `json:"project_id"`
	ContentItemID    string    `json:"content_item_id,omitempty"`
	SourceType       string    `json:"source_type"`
	TokenBudget      int       `json:"token_budget"`
	EstimatedTokens  int       `json:"estimated_tokens"`
	TruncationPolicy string    `json:"truncation_policy"`
	CreatedAt        time.Time `json:"created_at"`
}

type PaginationResponse struct {
	Page     int  `json:"page"`
	PageSize int  `json:"page_size"`
	Total    int  `json:"total"`
	HasNext  bool `json:"has_next"`
}

type PagedMemorySnapshotsResponse struct {
	Items      []MemorySnapshotResponse `json:"items"`
	Pagination PaginationResponse       `json:"pagination"`
}

type ContextPreviewRequest struct {
	Purpose       string `json:"purpose"`
	Budget        int    `json:"budget"`
	ContentItemID string `json:"content_item_id"`
}

type ContextPreviewResponse struct {
	Sources          []string `json:"sources"`
	TokenBudget      int      `json:"token_budget"`
	EstimatedTokens  int      `json:"estimated_tokens"`
	TruncationPolicy string   `json:"truncation_policy"`
	PreviewText      string   `json:"preview_text"`
}

type AssembleContextRequest struct {
	Purpose       string `json:"purpose"`
	Budget        int    `json:"budget"`
	ContentItemID string `json:"content_item_id"`
}

type AssembleContextResponse struct {
	ContextSnapshotID string `json:"context_snapshot_id"`
	EstimatedTokens   int    `json:"estimated_tokens"`
	TruncationPolicy  string `json:"truncation_policy"`
}

type UpdateDynamicStateRequest struct {
	Summary         string         `json:"summary"`
	Changes         map[string]any `json:"changes"`
	SourceVersionID string         `json:"source_version_id"`
}

type UpdateDynamicStateResponse struct {
	MemorySnapshotID    string `json:"memory_snapshot_id"`
	DynamicStateVersion int    `json:"dynamic_state_version"`
}

type CreateConsistencyReportRequest struct {
	Range             map[string]any `json:"range"`
	Scope             string         `json:"scope"`
	SeverityThreshold string         `json:"severity_threshold"`
}

type CreateConsistencyReportResponse struct {
	ReportID string `json:"report_id"`
	Status   string `json:"status"`
}

type ListConsistencyReportsRequest struct {
	Status   string `json:"status"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type ConsistencyIssue struct {
	IssueID              string   `json:"issue_id"`
	Severity             string   `json:"severity"`
	Type                 string   `json:"type"`
	Title                string   `json:"title"`
	Description          string   `json:"description"`
	AffectedContentItems []string `json:"affected_content_items"`
	Suggestion           string   `json:"suggestion"`
}

type ConsistencyReportResponse struct {
	ID              string         `json:"id"`
	ProjectID       string         `json:"project_id"`
	Status          string         `json:"status"`
	IssueCount      int            `json:"issue_count"`
	SeveritySummary map[string]int `json:"severity_summary"`
	CreatedAt       time.Time      `json:"created_at"`
}

type PagedConsistencyReportsResponse struct {
	Items      []ConsistencyReportResponse `json:"items"`
	Pagination PaginationResponse          `json:"pagination"`
}

type ConsistencyReportDetailResponse struct {
	ConsistencyReportResponse
	SourceSnapshotID string             `json:"source_snapshot_id"`
	Issues           []ConsistencyIssue `json:"issues"`
	ErrorCode        string             `json:"error_code,omitempty"`
	ErrorMessage     string             `json:"error_message,omitempty"`
}
