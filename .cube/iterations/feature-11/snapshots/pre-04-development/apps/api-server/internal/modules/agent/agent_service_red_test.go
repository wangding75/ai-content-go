package agent_test

import (
	"context"
	"errors"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/agent"
)

// @Test
func TestCreateAgentTaskRequiresWorkflowRunIDAndStepRunID(t *testing.T) {
	svc := agent.NewService()

	if _, err := svc.CreateTask(context.Background(), agent.CreateAgentTaskRequest{}); !errors.Is(err, agent.ErrValidation) {
		t.Fatalf("expected ErrValidation for empty request, got %v", err)
	}
}

// @Test
func TestCreateAgentTaskReturnsTaskWithPendingStatus(t *testing.T) {
	svc := agent.NewService()

	resp, err := svc.CreateTask(context.Background(), agent.CreateAgentTaskRequest{
		WorkflowRunID: "wfr-1",
		StepRunID:     "wfsr-1",
		AgentCode:     "writer",
		Input:         map[string]any{"topic": "AI"},
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if resp.ID == "" || resp.Status != "pending" || resp.AgentCode != "writer" {
		t.Fatalf("unexpected task response: %#v", resp)
	}
}

// @Test
func TestListAgentTasksFiltersByWorkflowRunIDAndStatus(t *testing.T) {
	svc := agent.NewService()

	svc.CreateTask(context.Background(), agent.CreateAgentTaskRequest{ //nolint
		WorkflowRunID: "wfr-filter", StepRunID: "wfsr-1", AgentCode: "writer",
	})

	resp, err := svc.ListTasks(context.Background(), agent.ListAgentTasksRequest{
		WorkflowRunID: "wfr-filter",
		Status:        "pending",
	})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if resp.Pagination.Total < 0 {
		t.Fatalf("unexpected pagination: %#v", resp.Pagination)
	}
	for _, item := range resp.Items {
		if item.WorkflowRunID != "wfr-filter" {
			t.Fatalf("filter mismatch: workflow_run_id %q", item.WorkflowRunID)
		}
	}
}

// @Test
func TestGetAgentTaskDetailIncludesLLMCallLogCountAndIDs(t *testing.T) {
	svc := agent.NewService()

	created, _ := svc.CreateTask(context.Background(), agent.CreateAgentTaskRequest{
		WorkflowRunID: "wfr-detail", StepRunID: "wfsr-2", AgentCode: "reviewer",
	})

	detail, err := svc.GetTask(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if detail.ID != created.ID || detail.LLMCallLogIDs == nil {
		t.Fatalf("detail missing llm_call_log_ids field: %#v", detail)
	}
}

// @Test
func TestGetAgentTaskReturnsNotFoundForMissingID(t *testing.T) {
	svc := agent.NewService()

	if _, err := svc.GetTask(context.Background(), "missing"); !errors.Is(err, agent.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
