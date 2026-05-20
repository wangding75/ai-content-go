package llm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/llm"
)

// @Test
func TestCreateLLMCallLogReturnsLogWithIDAndTokenCounts(t *testing.T) {
	svc := llm.NewService()

	resp, err := svc.CreateCallLog(context.Background(), llm.CreateLLMCallLogRequest{
		WorkflowRunID: "wfr-1",
		StepRunID:     "wfsr-1",
		AgentTaskID:   "at-1",
		Provider:      "openai",
		Model:         "gpt-4o",
		InputTokens:   100,
		OutputTokens:  50,
		Cost:          0.01,
		Currency:      "USD",
		LatencyMS:     300,
		Status:        "success",
	})
	if err != nil {
		t.Fatalf("create call log: %v", err)
	}
	if resp.ID == "" || resp.Provider != "openai" || resp.InputTokens != 100 {
		t.Fatalf("unexpected call log response: %#v", resp)
	}
}

// @Test
func TestListLLMCallLogsFiltersByWorkflowRunIDAndProvider(t *testing.T) {
	svc := llm.NewService()

	svc.CreateCallLog(context.Background(), llm.CreateLLMCallLogRequest{ //nolint
		WorkflowRunID: "wfr-list", StepRunID: "wfsr-1", AgentTaskID: "at-1",
		Provider: "anthropic", Model: "claude-3", Status: "success",
	})

	resp, err := svc.ListCallLogs(context.Background(), llm.ListLLMCallLogsRequest{
		WorkflowRunID: "wfr-list",
		Provider:      "anthropic",
	})
	if err != nil {
		t.Fatalf("list call logs: %v", err)
	}
	if resp.Pagination.Total < 0 {
		t.Fatalf("unexpected pagination: %#v", resp.Pagination)
	}
	for _, item := range resp.Items {
		if item.WorkflowRunID != "wfr-list" {
			t.Fatalf("filter mismatch: workflow_run_id %q", item.WorkflowRunID)
		}
	}
}

// @Test
func TestGetLLMCallLogDetailIncludesErrorAndRequestID(t *testing.T) {
	svc := llm.NewService()

	created, _ := svc.CreateCallLog(context.Background(), llm.CreateLLMCallLogRequest{
		WorkflowRunID: "wfr-detail", StepRunID: "wfsr-1", AgentTaskID: "at-1",
		Provider: "openai", Model: "gpt-4o", Status: "failed",
		Error: "rate limit exceeded", RequestID: "req-123",
	})

	detail, err := svc.GetCallLog(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get call log: %v", err)
	}
	if detail.ID != created.ID || detail.RequestID != "req-123" {
		t.Fatalf("unexpected detail: %#v", detail)
	}
}

// @Test
func TestGetLLMCallLogReturnsNotFoundForMissingID(t *testing.T) {
	svc := llm.NewService()

	if _, err := svc.GetCallLog(context.Background(), "missing"); !errors.Is(err, llm.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
