package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/worker"
)

func TestQueueContractRejectsEmptyTaskType(t *testing.T) {
	queue := worker.NewMemoryQueue()

	_, err := queue.Enqueue(context.Background(), worker.TaskRequest{RequestID: "req_123", Payload: map[string]any{"x": "y"}})
	if err == nil {
		t.Fatalf("expected empty task type to be rejected")
	}
}

func TestQueueContractReturnsJobIDImmediately(t *testing.T) {
	queue := worker.NewMemoryQueue()
	start := time.Now()

	receipt, err := queue.Enqueue(context.Background(), worker.TaskRequest{Type: "system.check", RequestID: "req_123"})
	if err != nil {
		t.Fatalf("enqueue returned error: %v", err)
	}
	if receipt.JobID == "" {
		t.Fatalf("expected job_id")
	}
	if time.Since(start) > time.Second {
		t.Fatalf("expected enqueue to return immediately")
	}
}
