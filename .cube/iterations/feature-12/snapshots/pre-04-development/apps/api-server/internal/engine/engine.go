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
	for {
		select {
		case <-ctx.Done():
			return
		case runID := <-e.runCh:
			e.processRun(ctx, runID)
		}
	}
}

func (e *Engine) processRun(ctx context.Context, runID string) {
	if e.workflowSvc == nil {
		return
	}

	run, err := e.workflowSvc.GetRunForEngine(ctx, runID)
	if err != nil {
		return
	}

	if err := e.workflowSvc.UpdateRunStatus(ctx, runID, "running", nil, ""); err != nil {
		return
	}

	steps, err := e.workflowSvc.GetRunStepTemplates(ctx, run.TemplateVersionID)
	if err != nil {
		e.workflowSvc.UpdateRunStatus(ctx, runID, "failed", nil, err.Error()) //nolint
		return
	}

	for _, step := range steps {
		current, err := e.workflowSvc.GetRunForEngine(ctx, runID)
		if err != nil || current.Status == "cancelled" {
			return
		}

		sr, err := e.workflowSvc.CreateStepRun(ctx, workflow.CreateStepRunRequest{
			WorkflowRunID:  runID,
			StepTemplateID: step.ID,
		})
		if err != nil {
			e.workflowSvc.UpdateRunStatus(ctx, runID, "failed", nil, err.Error()) //nolint
			return
		}

		e.workflowSvc.UpdateStepRunStatus(ctx, sr.ID, "running", nil, "") //nolint

		stepErr := e.executeStep(ctx, runID, sr.ID, step)
		if stepErr != nil {
			e.workflowSvc.UpdateStepRunStatus(ctx, sr.ID, "failed", nil, stepErr.Error()) //nolint
			e.workflowSvc.UpdateRunStatus(ctx, runID, "failed", nil, stepErr.Error())     //nolint
			return
		}
		e.workflowSvc.UpdateStepRunStatus(ctx, sr.ID, "success", nil, "") //nolint
	}

	e.workflowSvc.UpdateRunStatus(ctx, runID, "success", nil, "") //nolint
}

func (e *Engine) executeStep(ctx context.Context, runID, stepRunID string, step workflow.WorkflowStepTemplateResponse) error {
	switch step.StepType {
	case "agent":
		return e.executeAgentStep(ctx, runID, stepRunID, step)
	case "human_review":
		// leave in running state, waiting for external driver
		return nil
	case "condition":
		// simple pass-through
		return nil
	default:
		// system_task and unknown: mock success
		return nil
	}
}

func (e *Engine) executeAgentStep(ctx context.Context, runID, stepRunID string, step workflow.WorkflowStepTemplateResponse) error {
	if e.agentSvc == nil {
		return nil
	}

	task, err := e.agentSvc.CreateTask(ctx, agent.CreateAgentTaskRequest{
		WorkflowRunID: runID,
		StepRunID:     stepRunID,
		AgentCode:     step.AgentCode,
	})
	if err != nil {
		return err
	}

	e.agentSvc.UpdateTask(ctx, task.ID, agent.UpdateAgentTaskRequest{Status: "running"}) //nolint

	if e.llmSvc != nil {
		e.llmSvc.CreateCallLog(ctx, llm.CreateLLMCallLogRequest{ //nolint
			WorkflowRunID: runID,
			StepRunID:     stepRunID,
			AgentTaskID:   task.ID,
			Provider:      "mock",
			Model:         "mock-v1",
			InputTokens:   100,
			OutputTokens:  50,
			Status:        "success",
		})
	}

	e.agentSvc.UpdateTask(ctx, task.ID, agent.UpdateAgentTaskRequest{Status: "success"}) //nolint
	return nil
}
