package novel

import (
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

type CreatePlanningRunRequest struct {
	Genre             string         `json:"genre"`
	Audience          string         `json:"audience"`
	Count             int            `json:"count"`
	TemplateVersionID string         `json:"template_version_id"`
	InputOverride     map[string]any `json:"input_override"`
}

type CreatePlanningRunResponse struct {
	PlanningRunID string `json:"planning_run_id"`
	WorkflowRunID string `json:"workflow_run_id"`
	Status        string `json:"status"`
}

type ListPlanningRunsRequest struct {
	content.PaginationRequest
	Status string `json:"status"`
}

type PlanningRunResponse struct {
	ID                string    `json:"id"`
	ProjectID         string    `json:"project_id"`
	WorkflowRunID     string    `json:"workflow_run_id"`
	TemplateVersionID string    `json:"template_version_id"`
	Status            string    `json:"status"`
	Genre             string    `json:"genre"`
	Audience          string    `json:"audience"`
	CandidateCount    int       `json:"candidate_count"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type PagedPlanningRunsResponse struct {
	Items      []PlanningRunResponse      `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type TopicCandidateResponse struct {
	CandidateID      string  `json:"candidate_id"`
	PlanningRunID    string  `json:"planning_run_id"`
	SnapshotID       string  `json:"snapshot_id"`
	Title            string  `json:"title"`
	Logline          string  `json:"logline"`
	Status           string  `json:"status"`
	Score            float64 `json:"score"`
	Reason           string  `json:"reason"`
	ConfirmedTopicID string  `json:"confirmed_topic_id,omitempty"`
}

type StepRunSummary struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type AgentTaskSummary struct {
	ID        string `json:"id"`
	AgentCode string `json:"agent_code"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

type LLMCallLogSummary struct {
	ID     string  `json:"id"`
	Model  string  `json:"model"`
	Status string  `json:"status"`
	Cost   float64 `json:"cost"`
}

type PlanningRunDetailResponse struct {
	PlanningRunResponse
	Topics      []TopicCandidateResponse `json:"topics"`
	StepRuns    []StepRunSummary         `json:"step_runs"`
	AgentTasks  []AgentTaskSummary       `json:"agent_tasks"`
	LLMCallLogs []LLMCallLogSummary      `json:"llm_call_logs"`
}

type ConfirmTopicRequest struct {
	Note string `json:"note"`
}

type ConfirmTopicResponse struct {
	ConfirmedTopicID string `json:"confirmed_topic_id"`
	PreviousStatus   string `json:"previous_status"`
	CurrentStatus    string `json:"current_status"`
	OperationLogID   string `json:"operation_log_id"`
}

type WorldviewResponse struct {
	ProjectID      string         `json:"project_id"`
	VersionID      string         `json:"version_id"`
	Version        int            `json:"version"`
	Worldview      map[string]any `json:"worldview"`
	ForbiddenRules []string       `json:"forbidden_rules"`
	PlanningRunID  string         `json:"planning_run_id,omitempty"`
	SnapshotID     string         `json:"snapshot_id,omitempty"`
}

type UpdateWorldviewRequest struct {
	Worldview      map[string]any `json:"worldview"`
	ForbiddenRules []string       `json:"forbidden_rules"`
	Note           string         `json:"note"`
}

type UpdateWorldviewResponse struct {
	VersionID      string `json:"version_id"`
	OperationLogID string `json:"operation_log_id"`
}

type ListCharactersRequest struct {
	content.PaginationRequest
	Role string `json:"role"`
}

type CharacterResponse struct {
	CharacterID   string         `json:"character_id"`
	ProjectID     string         `json:"project_id"`
	Name          string         `json:"name"`
	Role          string         `json:"role"`
	Profile       map[string]any `json:"profile"`
	PlanningRunID string         `json:"planning_run_id,omitempty"`
	SnapshotID    string         `json:"snapshot_id,omitempty"`
}

type PagedCharactersResponse struct {
	Items      []CharacterResponse        `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type CreateCharacterRequest struct {
	Name    string         `json:"name"`
	Role    string         `json:"role"`
	Profile map[string]any `json:"profile"`
	Note    string         `json:"note"`
}

type CreateCharacterResponse struct {
	CharacterID    string `json:"character_id"`
	OperationLogID string `json:"operation_log_id"`
}

type ListArcsRequest struct {
	content.PaginationRequest
}

type ArcResponse struct {
	ArcID         string `json:"arc_id"`
	ProjectID     string `json:"project_id"`
	Title         string `json:"title"`
	Summary       string `json:"summary"`
	OrderIndex    int    `json:"order_index"`
	PlanningRunID string `json:"planning_run_id,omitempty"`
	SnapshotID    string `json:"snapshot_id,omitempty"`
}

type PagedArcsResponse struct {
	Items      []ArcResponse              `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}
