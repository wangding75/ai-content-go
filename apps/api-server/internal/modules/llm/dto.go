package llm

import "github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"

type ListProvidersRequest struct {
	content.PaginationRequest
}

type CreateProviderRequest struct {
	ProviderType string `json:"provider_type"`
	BaseURL      string `json:"base_url"`
	APIKey       string `json:"api_key"`
}

type ProviderResponse struct {
	ID           string `json:"id"`
	ProviderType string `json:"provider_type"`
	BaseURL      string `json:"base_url"`
	APIKeyMasked string `json:"api_key_masked"`
	Enabled      bool   `json:"enabled"`
}

type PagedProvidersResponse struct {
	Items      []ProviderResponse         `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type CreateProviderResponse struct {
	ProviderID   string `json:"provider_id"`
	APIKeyMasked string `json:"api_key_masked"`
}

// LLMCallLog DTOs

type CreateLLMCallLogRequest struct {
	WorkflowRunID string  `json:"workflow_run_id"`
	StepRunID     string  `json:"step_run_id"`
	AgentTaskID   string  `json:"agent_task_id"`
	Provider      string  `json:"provider"`
	Model         string  `json:"model"`
	InputTokens   int     `json:"input_tokens"`
	OutputTokens  int     `json:"output_tokens"`
	Cost          float64 `json:"cost"`
	Currency      string  `json:"currency"`
	LatencyMS     int     `json:"latency_ms"`
	Status        string  `json:"status"`
	Error         string  `json:"error,omitempty"`
	RequestID     string  `json:"request_id"`
}

type LLMCallLogResponse struct {
	ID            string  `json:"id"`
	WorkflowRunID string  `json:"workflow_run_id"`
	StepRunID     string  `json:"step_run_id"`
	AgentTaskID   string  `json:"agent_task_id"`
	Provider      string  `json:"provider"`
	Model         string  `json:"model"`
	InputTokens   int     `json:"input_tokens"`
	OutputTokens  int     `json:"output_tokens"`
	Cost          float64 `json:"cost"`
	Currency      string  `json:"currency"`
	LatencyMS     int     `json:"latency_ms"`
	Status        string  `json:"status"`
}

type LLMCallLogDetailResponse struct {
	LLMCallLogResponse
	Error     string `json:"error,omitempty"`
	RequestID string `json:"request_id"`
}

type ListLLMCallLogsRequest struct {
	content.PaginationRequest
	WorkflowRunID string `json:"workflow_run_id"`
	AgentTaskID   string `json:"agent_task_id"`
	Provider      string `json:"provider"`
	Model         string `json:"model"`
	Status        string `json:"status"`
}

type PagedLLMCallLogsResponse struct {
	Items      []LLMCallLogResponse      `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

