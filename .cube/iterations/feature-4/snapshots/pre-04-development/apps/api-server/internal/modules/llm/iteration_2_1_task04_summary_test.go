package llm_test

import (
	"context"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/llm"
)

// @Test
func TestIteration21LLMCostSummaryAggregatesCallsTokensCostAndByModel(t *testing.T) {
	svc := llm.NewService()
	_, _ = svc.CreateCallLog(context.Background(), llm.CreateLLMCallLogRequest{
		WorkflowRunID: "wfr-cost-1", StepRunID: "step-1", AgentTaskID: "task-1",
		Provider: "openai", Model: "gpt-4o", InputTokens: 100, OutputTokens: 40, Cost: 0.05, Currency: "USD", Status: "success",
	})
	_, _ = svc.CreateCallLog(context.Background(), llm.CreateLLMCallLogRequest{
		WorkflowRunID: "wfr-cost-2", StepRunID: "step-2", AgentTaskID: "task-2",
		Provider: "openai", Model: "gpt-4o-mini", InputTokens: 30, OutputTokens: 20, Cost: 0.01, Currency: "USD", Status: "success",
	})

	resp, err := svc.SummaryCallLogs(context.Background(), llm.SummaryCallLogsRequest{Provider: "openai"})
	if err != nil {
		t.Fatalf("summary call logs: %v", err)
	}
	if resp.Calls != 2 || resp.InputTokens != 130 || resp.OutputTokens != 60 || resp.Tokens != 190 || resp.Cost != 0.06 || resp.Currency != "USD" {
		t.Fatalf("unexpected aggregate: %#v", resp)
	}
	if len(resp.ByModel) != 2 {
		t.Fatalf("expected by_model for two models, got %#v", resp.ByModel)
	}
}

// @Test
func TestIteration21LLMCostSummaryRejectsInvalidDateRange(t *testing.T) {
	svc := llm.NewService()

	_, err := svc.SummaryCallLogs(context.Background(), llm.SummaryCallLogsRequest{DateFrom: "2026-05-31", DateTo: "2026-05-01"})
	if err == nil {
		t.Fatalf("expected validation error for inverted date range")
	}
}
