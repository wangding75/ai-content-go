package engine_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type mockEnginePort struct {
	mu                   sync.Mutex
	updateRunStatusCalls []string
	stepRuns             []workflow.WorkflowStepRunResponse
}

func (m *mockEnginePort) UpdateRunStatus(_ context.Context, id, status string, _ map[string]any, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateRunStatusCalls = append(m.updateRunStatusCalls, status)
	return nil
}

func (m *mockEnginePort) runStatusCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.updateRunStatusCalls...)
}

func (m *mockEnginePort) CreateStepRun(_ context.Context, req workflow.CreateStepRunRequest) (workflow.WorkflowStepRunResponse, error) {
	sr := workflow.WorkflowStepRunResponse{ID: "wfsr-mock", WorkflowRunID: req.WorkflowRunID, Status: "pending"}
	m.stepRuns = append(m.stepRuns, sr)
	return sr, nil
}

func (m *mockEnginePort) UpdateStepRunStatus(_ context.Context, id, status string, _ map[string]any, _ string) error {
	return nil
}

func (m *mockEnginePort) GetRunStepTemplates(_ context.Context, _ string) ([]workflow.WorkflowStepTemplateResponse, error) {
	return []workflow.WorkflowStepTemplateResponse{
		{ID: "wfst-1", StepCode: "step-1", StepType: "system_task", OrderIndex: 1},
	}, nil
}

func (m *mockEnginePort) GetRunForEngine(_ context.Context, id string) (workflow.WorkflowRunResponse, error) {
	return workflow.WorkflowRunResponse{ID: id, Status: "running", TemplateVersionID: "wftv-1"}, nil
}

// @Test
func TestEngineSubmitReturnsTrueWhenChannelHasCapacity(t *testing.T) {
	mock := &mockEnginePort{}
	eng := engine.New(mock, nil, nil)
	eng.Start(context.Background())

	if !eng.Submit("wfr-1") {
		t.Fatalf("expected Submit to return true when channel has capacity")
	}
}

// @Test
func TestEngineSubmitReturnsFalseWhenChannelIsFull(t *testing.T) {
	mock := &mockEnginePort{}
	eng := engine.New(mock, nil, nil)
	// do not Start — channel drains only when worker is running
	// fill channel to capacity (100) without draining
	for i := 0; i < 100; i++ {
		eng.Submit("wfr-fill")
	}
	if eng.Submit("wfr-overflow") {
		t.Fatalf("expected Submit to return false when channel is at capacity")
	}
}

// @Test
func TestEngineWorkerTransitionsRunToRunningStatus(t *testing.T) {
	mock := &mockEnginePort{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eng := engine.New(mock, nil, nil)
	eng.Start(ctx)
	eng.Submit("wfr-process")

	// wait for worker to process
	deadline := time.Now().Add(500 * time.Millisecond)
	var calls []string
	for time.Now().Before(deadline) {
		calls = mock.runStatusCalls()
		if len(calls) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if len(calls) == 0 || calls[0] != "running" {
		t.Fatalf("expected engine to call UpdateRunStatus(running), got calls: %v", calls)
	}
}
