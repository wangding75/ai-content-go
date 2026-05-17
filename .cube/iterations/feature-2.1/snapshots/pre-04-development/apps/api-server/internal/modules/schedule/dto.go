package schedule

import (
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

type ListSchedulesRequest struct {
	content.PaginationRequest
	ProjectID string `json:"project_id"`
	Enabled   *bool  `json:"enabled"`
}

type CreateScheduleRequest struct {
	ProjectID         string         `json:"project_id"`
	TemplateVersionID string         `json:"template_version_id"`
	CronExpression    string         `json:"cron_expression"`
	DailyContentCount int            `json:"daily_content_count"`
	InputTemplate     map[string]any `json:"input_template"`
}

type WorkflowScheduleResponse struct {
	ID                string    `json:"id"`
	ProjectID         string    `json:"project_id"`
	TemplateVersionID string    `json:"template_version_id"`
	CronExpression    string    `json:"cron_expression"`
	Enabled           bool      `json:"enabled"`
	NextRunAt         time.Time `json:"next_run_at"`
	DailyContentCount int       `json:"daily_content_count"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type PagedSchedulesResponse struct {
	Items      []WorkflowScheduleResponse `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type CreateScheduleResponse struct {
	ScheduleID        string    `json:"schedule_id"`
	NextRunAt         time.Time `json:"next_run_at"`
	DailyContentCount int       `json:"daily_content_count"`
	Status            string    `json:"status"`
}

type ToggleScheduleRequest struct {
	Reason string `json:"reason"`
	Note   string `json:"note"`
}

type ToggleScheduleResponse struct {
	PreviousEnabled bool      `json:"previous_enabled"`
	CurrentEnabled  bool      `json:"current_enabled"`
	NextRunAt       time.Time `json:"next_run_at"`
	OperationLogID  string    `json:"operation_log_id"`
}

type TestRunScheduleRequest struct {
	InputOverride map[string]any `json:"input_override"`
}

type PreparedScheduleRun struct {
	TriggerLogID      string         `json:"trigger_log_id"`
	ProjectID         string         `json:"project_id"`
	TemplateVersionID string         `json:"template_version_id"`
	Input             map[string]any `json:"input"`
}

type TestRunScheduleResponse struct {
	WorkflowRunID string `json:"workflow_run_id"`
	Status        string `json:"status"`
	TriggerLogID  string `json:"trigger_log_id"`
}

type ListTriggersRequest struct {
	content.PaginationRequest
}

type ScheduleTriggerResponse struct {
	ID            string    `json:"id"`
	ScheduleID    string    `json:"schedule_id"`
	TriggerType   string    `json:"trigger_type"`
	WorkflowRunID string    `json:"workflow_run_id,omitempty"`
	Status        string    `json:"status"`
	Error         string    `json:"error,omitempty"`
	TriggeredAt   time.Time `json:"triggered_at"`
}

type PagedScheduleTriggersResponse struct {
	Items      []ScheduleTriggerResponse  `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}
