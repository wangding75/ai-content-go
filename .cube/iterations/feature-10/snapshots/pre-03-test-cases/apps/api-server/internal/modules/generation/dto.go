package generation

import (
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/novel"
)

const (
	GenerationRunPending   = "pending"
	GenerationRunRunning   = "running"
	GenerationRunSucceeded = "succeeded"
	GenerationRunFailed    = "failed"
	GenerationRunRetrying  = "retrying"

	ContentItemPlanned          = "planned"
	ContentItemGenerating       = "generating"
	ContentItemGenerated        = "generated"
	ContentItemGenerationFailed = "generation_failed"
	ContentItemPendingReview    = "pending_review"
)

type CreateGenerationRunRequest struct {
	ConfirmedTopicID   string         `json:"confirmed_topic_id"`
	WorldviewVersionID string         `json:"worldview_version_id"`
	ArcID              string         `json:"arc_id"`
	TargetCount        int            `json:"target_count"`
	StartSequenceNo    int            `json:"start_sequence_no"`
	TemplateVersionID  string         `json:"template_version_id"`
	GenerationConfig   map[string]any `json:"generation_config"`
}

type CreateGenerationRunResponse struct {
	GenerationRunID string `json:"generation_run_id"`
	WorkflowRunID   string `json:"workflow_run_id"`
	Status          string `json:"status"`
}

type GenerationRangeRequest struct {
	StartSequenceNo int `json:"start_sequence_no"`
	EndSequenceNo   int `json:"end_sequence_no"`
}

type CreateBatchGenerationRunsRequest struct {
	Range             GenerationRangeRequest `json:"range"`
	BatchSize         int                    `json:"batch_size"`
	TemplateVersionID string                 `json:"template_version_id"`
	GenerationConfig  map[string]any         `json:"generation_config"`
}

type CreateBatchGenerationRunsResponse struct {
	GenerationRunIDs []string `json:"generation_run_ids"`
	WorkflowRunIDs   []string `json:"workflow_run_ids"`
	AcceptedCount    int      `json:"accepted_count"`
}

type ListGenerationRunsRequest struct {
	content.PaginationRequest
	Status string `json:"status"`
}

type GenerationRunResponse struct {
	ID                     string    `json:"id"`
	ProjectID              string    `json:"project_id"`
	WorkflowRunID          string    `json:"workflow_run_id"`
	TemplateVersionID      string    `json:"template_version_id"`
	Status                 string    `json:"status"`
	TriggerType            string    `json:"trigger_type"`
	RetryOfGenerationRunID string    `json:"retry_of_generation_run_id,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type PagedGenerationRunsResponse struct {
	Items      []GenerationRunResponse    `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type ContentItemResponse struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"project_id"`
	GenerationRunID string    `json:"generation_run_id"`
	ContentTypeCode string    `json:"content_type_code"`
	Status          string    `json:"status"`
	SequenceNo      int       `json:"sequence_no"`
	Title           string    `json:"title"`
	Version         int       `json:"version"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type GenerationRunDetailResponse struct {
	GenerationRunResponse
	StepRuns     []novel.StepRunSummary    `json:"step_runs"`
	AgentTasks   []novel.AgentTaskSummary  `json:"agent_tasks"`
	LLMCallLogs  []novel.LLMCallLogSummary `json:"llm_call_logs"`
	ContentItems []ContentItemResponse     `json:"content_items"`
	Error        string                    `json:"error,omitempty"`
}

type RetryGenerationRunRequest struct {
	Reason        string         `json:"reason"`
	InputOverride map[string]any `json:"input_override"`
}

type RetryGenerationRunResponse struct {
	NewGenerationRunID string `json:"new_generation_run_id"`
	WorkflowRunID      string `json:"workflow_run_id"`
	OperationLogID     string `json:"operation_log_id"`
}

type ListContentItemsRequest struct {
	content.PaginationRequest
	Status string `json:"status"`
}

type PagedContentItemsResponse struct {
	Items      []ContentItemResponse      `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type NovelChapterExtensionResponse struct {
	ConfirmedTopicID   string         `json:"confirmed_topic_id"`
	WorldviewVersionID string         `json:"worldview_version_id"`
	ArcID              string         `json:"arc_id"`
	ChapterNo          int            `json:"chapter_no"`
	Script             map[string]any `json:"script"`
}

type ContentItemDetailResponse struct {
	ContentItemResponse
	Body      string                        `json:"body"`
	Metadata  map[string]any                `json:"metadata"`
	Extension NovelChapterExtensionResponse `json:"extension"`
}
