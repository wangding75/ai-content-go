package workflow

import (
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

// WorkflowTemplate DTOs

type ListWorkflowTemplatesRequest struct {
	content.PaginationRequest
	ContentType string `json:"content_type"`
	Category    string `json:"category"`
	Status      string `json:"status"`
}

type CreateWorkflowTemplateRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type WorkflowTemplateResponse struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	ContentType string    `json:"content_type"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type PagedWorkflowTemplatesResponse struct {
	Items      []WorkflowTemplateResponse `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type CreateWorkflowTemplateResponse struct {
	WorkflowTemplateID string `json:"workflow_template_id"`
	Status             string `json:"status"`
}

// WorkflowTemplateVersion DTOs

type CreateVersionRequest struct {
	InputSchema  map[string]any              `json:"input_schema"`
	OutputSchema map[string]any              `json:"output_schema"`
	Steps        []CreateStepTemplateRequest `json:"steps"`
}

type CreateStepTemplateRequest struct {
	StepCode      string         `json:"step_code"`
	StepType      string         `json:"step_type"`
	AgentCode     string         `json:"agent_code"`
	OrderIndex    int            `json:"order_index"`
	InputMapping  map[string]any `json:"input_mapping"`
	OutputMapping map[string]any `json:"output_mapping"`
}

type WorkflowTemplateVersionResponse struct {
	ID           string    `json:"id"`
	TemplateID   string    `json:"template_id"`
	Version      int       `json:"version"`
	InputSchema  map[string]any `json:"input_schema"`
	OutputSchema map[string]any `json:"output_schema"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type PagedVersionsResponse struct {
	Items      []WorkflowTemplateVersionResponse `json:"items"`
	Pagination content.PaginationResponse        `json:"pagination"`
}

type CreateVersionResponse struct {
	TemplateVersionID string `json:"template_version_id"`
	StepCount         int    `json:"step_count"`
	Status            string `json:"status"`
}

type WorkflowStepTemplateResponse struct {
	ID                string         `json:"id"`
	TemplateVersionID string         `json:"template_version_id"`
	StepCode          string         `json:"step_code"`
	StepType          string         `json:"step_type"`
	AgentCode         string         `json:"agent_code"`
	OrderIndex        int            `json:"order_index"`
	InputMapping      map[string]any `json:"input_mapping"`
	OutputMapping     map[string]any `json:"output_mapping"`
}

type WorkflowTemplateVersionDetailResponse struct {
	WorkflowTemplateVersionResponse
	Steps []WorkflowStepTemplateResponse `json:"steps"`
}

type PublishVersionRequest struct {
	Note string `json:"note"`
}

type PublishVersionResponse struct {
	PreviousStatus string `json:"previous_status"`
	CurrentStatus  string `json:"current_status"`
	OperationLogID string `json:"operation_log_id"`
}

// WorkflowRun DTOs

type ListWorkflowRunsRequest struct {
	content.PaginationRequest
	ProjectID         string `json:"project_id"`
	TemplateVersionID string `json:"template_version_id"`
	Status            string `json:"status"`
}

type CreateWorkflowRunRequest struct {
	ProjectID         string         `json:"project_id"`
	TemplateVersionID string         `json:"template_version_id"`
	Input             map[string]any `json:"input"`
}

type WorkflowRunResponse struct {
	ID                string    `json:"id"`
	ProjectID         string    `json:"project_id"`
	TemplateVersionID string    `json:"template_version_id"`
	Status            string    `json:"status"`
	Source            string    `json:"source"`
	ParentRunID       string    `json:"parent_run_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type WorkflowRunDetailResponse struct {
	WorkflowRunResponse
	Input          map[string]any `json:"input"`
	Output         map[string]any `json:"output"`
	Error          string         `json:"error,omitempty"`
	StepCount      int            `json:"step_count"`
	AgentTaskCount int            `json:"agent_task_count"`
}

type PagedWorkflowRunsResponse struct {
	Items      []WorkflowRunResponse     `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type CreateWorkflowRunResponse struct {
	WorkflowRunID string `json:"workflow_run_id"`
	Status        string `json:"status"`
}

type CancelRunRequest struct {
	Reason string `json:"reason"`
	Note   string `json:"note"`
}

type CancelRunResponse struct {
	PreviousStatus string `json:"previous_status"`
	CurrentStatus  string `json:"current_status"`
	OperationLogID string `json:"operation_log_id"`
}

type RetryRunRequest struct {
	Reason        string         `json:"reason"`
	InputOverride map[string]any `json:"input_override"`
}

type RetryRunResponse struct {
	NewWorkflowRunID string `json:"new_workflow_run_id"`
	Status           string `json:"status"`
}

// WorkflowStepRun DTOs

type WorkflowStepRunResponse struct {
	ID             string         `json:"id"`
	WorkflowRunID  string         `json:"workflow_run_id"`
	StepTemplateID string         `json:"step_template_id"`
	Status         string         `json:"status"`
	Input          map[string]any `json:"input"`
	Output         map[string]any `json:"output"`
	Error          string         `json:"error,omitempty"`
	StartedAt      *time.Time     `json:"started_at,omitempty"`
	FinishedAt     *time.Time     `json:"finished_at,omitempty"`
}

type ListStepRunsResponse struct {
	Items []WorkflowStepRunResponse `json:"items"`
}

// EnginePort internal request types

type CreateStepRunRequest struct {
	WorkflowRunID  string         `json:"workflow_run_id"`
	StepTemplateID string         `json:"step_template_id"`
	Input          map[string]any `json:"input"`
}
