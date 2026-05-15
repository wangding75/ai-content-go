package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type Queue interface {
	Enqueue(ctx context.Context, req TaskRequest) (TaskReceipt, error)
}

type TaskRequest struct {
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload"`
	RequestID string         `json:"request_id"`
}

type TaskReceipt struct {
	JobID     string    `json:"job_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type memoryQueue struct {
	counter atomic.Uint64
}

func NewMemoryQueue() Queue {
	return &memoryQueue{}
}

func (q *memoryQueue) Enqueue(ctx context.Context, req TaskRequest) (TaskReceipt, error) {
	if strings.TrimSpace(req.Type) == "" {
		return TaskReceipt{}, errors.New("task type is required")
	}
	select {
	case <-ctx.Done():
		return TaskReceipt{}, ctx.Err()
	default:
	}
	id := q.counter.Add(1)
	return TaskReceipt{JobID: fmt.Sprintf("job_%d", id), Status: "queued", CreatedAt: time.Now().UTC()}, nil
}
