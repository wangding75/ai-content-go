package worker

import (
	"context"
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
