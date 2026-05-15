package engine

import (
	"context"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/agent"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/llm"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

const engineChannelSize = 100

// Submitter is the interface the HTTP handler uses to enqueue workflow runs.
type Submitter interface {
	Submit(runID string) bool
}

// Engine drives asynchronous WorkflowRun execution.
type Engine struct {
	runCh       chan string // buffered, size engineChannelSize
	workflowSvc workflow.EnginePort
	agentSvc    agent.Service
	llmSvc      llm.Service
}

// New creates a new Engine. The workflowSvc must implement workflow.EnginePort.
func New(wf workflow.EnginePort, ag agent.Service, lm llm.Service) *Engine {
	return &Engine{
		runCh:       make(chan string, engineChannelSize),
		workflowSvc: wf,
		agentSvc:    ag,
		llmSvc:      lm,
	}
}

// Start launches the background goroutine worker. Call once after New.
func (e *Engine) Start(ctx context.Context) {
	go e.worker(ctx)
}

// Submit enqueues runID for async execution. Non-blocking: returns false if channel is full.
func (e *Engine) Submit(runID string) bool {
	select {
	case e.runCh <- runID:
		return true
	default:
		return false
	}
}

func (e *Engine) worker(ctx context.Context) {
	panic("not implemented")
}
