package agent

import (
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

type CreateAgentTaskRequest struct {
	WorkflowRunID string         `json:"workflow_run_id"`
	StepRunID     string         `json:"step_run_id"`
	AgentCode     string         `json:"agent_code"`
	Input         map[string]any `json:"input"`
}

type UpdateAgentTaskRequest struct {
	Status string         `json:"status"`
	Output map[string]any `json:"output"`
	Error  string         `json:"error"`
}

type AgentTaskResponse struct {
	ID            string    `json:"id"`
	WorkflowRunID string    `json:"workflow_run_id"`
	StepRunID     string    `json:"step_run_id"`
	AgentCode     string    `json:"agent_code"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type AgentTaskDetailResponse struct {
	AgentTaskResponse
	Input           map[string]any `json:"input"`
	Output          map[string]any `json:"output"`
	Error           string         `json:"error,omitempty"`
	StartedAt       *time.Time     `json:"started_at,omitempty"`
	FinishedAt      *time.Time     `json:"finished_at,omitempty"`
	LLMCallLogCount int            `json:"llm_call_log_count"`
	LLMCallLogIDs   []string       `json:"llm_call_log_ids"`
}

type ListAgentTasksRequest struct {
	content.PaginationRequest
	WorkflowRunID string `json:"workflow_run_id"`
	StepRunID     string `json:"step_run_id"`
	AgentCode     string `json:"agent_code"`
	Status        string `json:"status"`
}

type PagedAgentTasksResponse struct {
	Items      []AgentTaskResponse       `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}
