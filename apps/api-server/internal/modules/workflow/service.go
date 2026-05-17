package workflow

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

var (
	ErrValidation          = errors.New("validation error")
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrIdempotencyConflict = errors.New("idempotency conflict")
)

// Service is the public interface for workflow template and run management.
type Service interface {
	ListTemplates(ctx context.Context, req ListWorkflowTemplatesRequest) (PagedWorkflowTemplatesResponse, error)
	CreateTemplate(ctx context.Context, req CreateWorkflowTemplateRequest) (CreateWorkflowTemplateResponse, error)
	GetTemplate(ctx context.Context, id string) (WorkflowTemplateResponse, error)
	ListVersions(ctx context.Context, templateID string, req PaginationRequest) (PagedVersionsResponse, error)
	CreateVersion(ctx context.Context, templateID string, req CreateVersionRequest) (CreateVersionResponse, error)
	GetVersion(ctx context.Context, id string) (WorkflowTemplateVersionDetailResponse, error)
	PublishVersion(ctx context.Context, id string, req PublishVersionRequest, idempotencyKey string) (PublishVersionResponse, error)
	ListRuns(ctx context.Context, req ListWorkflowRunsRequest) (PagedWorkflowRunsResponse, error)
	CreateRun(ctx context.Context, req CreateWorkflowRunRequest, idempotencyKey string) (CreateWorkflowRunResponse, error)
	GetRun(ctx context.Context, id string) (WorkflowRunDetailResponse, error)
	GetRunSteps(ctx context.Context, runID string) (ListStepRunsResponse, error)
	CancelRun(ctx context.Context, id string, req CancelRunRequest, idempotencyKey string) (CancelRunResponse, error)
	RetryRun(ctx context.Context, id string, req RetryRunRequest, idempotencyKey string) (RetryRunResponse, error)
}

// EnginePort is the exported interface for the engine package to call workflow internals cross-package.
type EnginePort interface {
	UpdateRunStatus(ctx context.Context, id, status string, output map[string]any, errMsg string) error
	CreateStepRun(ctx context.Context, req CreateStepRunRequest) (WorkflowStepRunResponse, error)
	UpdateStepRunStatus(ctx context.Context, id, status string, output map[string]any, errMsg string) error
	GetRunStepTemplates(ctx context.Context, templateVersionID string) ([]WorkflowStepTemplateResponse, error)
	GetRunForEngine(ctx context.Context, id string) (WorkflowRunResponse, error)
}

// PaginationRequest is a local alias to avoid import cycle (workflow has its own pagination need).
type PaginationRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Sort     string `json:"sort"`
	Order    string `json:"order"`
}

type workflowTemplate struct {
	WorkflowTemplateResponse
}

type workflowVersion struct {
	WorkflowTemplateVersionResponse
	steps []WorkflowStepTemplateResponse
}

type workflowRun struct {
	WorkflowRunDetailResponse
	idempotencyKey string
}

type workflowStepRun struct {
	WorkflowStepRunResponse
}

type idempotentOperation struct {
	kind     string
	payload  any
	response any
}

type workflowService struct {
	mu sync.RWMutex

	templates     []workflowTemplate
	versions      []workflowVersion
	runs          []workflowRun
	stepRuns      []workflowStepRun
	idempotentOps map[string]idempotentOperation

	tmplNext     int
	versionNext  int
	stepTmplNext int
	runNext      int
	stepRunNext  int
	oplogNext    int
}

// NewService returns a Service + EnginePort implementation backed by in-memory storage.
func NewService() interface {
	Service
	EnginePort
} {
	return &workflowService{
		idempotentOps: make(map[string]idempotentOperation),
	}
}

func (s *workflowService) nextOplogID() string {
	s.oplogNext++
	return fmt.Sprintf("oplog-%d", s.oplogNext)
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return page, pageSize
}

func (s *workflowService) idempotentResponse(key string, kind string, payload any) (any, bool) {
	if key == "" {
		return nil, false
	}
	cached, ok := s.idempotentOps[key]
	if !ok || cached.kind != kind || !reflect.DeepEqual(cached.payload, payload) {
		return nil, false
	}
	return cached.response, true
}

// --- Template methods ---

func (s *workflowService) ListTemplates(_ context.Context, req ListWorkflowTemplatesRequest) (PagedWorkflowTemplatesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var items []WorkflowTemplateResponse
	for _, t := range s.templates {
		if req.ContentType != "" && t.ContentType != req.ContentType {
			continue
		}
		if req.Category != "" && t.Category != req.Category {
			continue
		}
		if req.Status != "" && t.Status != req.Status {
			continue
		}
		items = append(items, t.WorkflowTemplateResponse)
	}
	if items == nil {
		items = []WorkflowTemplateResponse{}
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	total := len(items)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return PagedWorkflowTemplatesResponse{
		Items:      items[start:end],
		Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: total, HasNext: end < total},
	}, nil
}

func (s *workflowService) CreateTemplate(_ context.Context, req CreateWorkflowTemplateRequest) (CreateWorkflowTemplateResponse, error) {
	if req.Code == "" || req.Name == "" {
		return CreateWorkflowTemplateResponse{}, ErrValidation
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.templates {
		if t.Code == req.Code {
			return CreateWorkflowTemplateResponse{}, ErrConflict
		}
	}
	s.tmplNext++
	id := fmt.Sprintf("wft-%d", s.tmplNext)
	s.templates = append(s.templates, workflowTemplate{WorkflowTemplateResponse{
		ID: id, Code: req.Code, Name: req.Name, ContentType: req.ContentType,
		Category: req.Category, Description: req.Description, Status: "draft", CreatedAt: time.Now(),
	}})
	return CreateWorkflowTemplateResponse{WorkflowTemplateID: id, Status: "draft"}, nil
}

func (s *workflowService) GetTemplate(_ context.Context, id string) (WorkflowTemplateResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.templates {
		if t.ID == id {
			return t.WorkflowTemplateResponse, nil
		}
	}
	return WorkflowTemplateResponse{}, ErrNotFound
}

func (s *workflowService) ListVersions(_ context.Context, templateID string, req PaginationRequest) (PagedVersionsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var items []WorkflowTemplateVersionResponse
	for _, v := range s.versions {
		if v.TemplateID == templateID {
			items = append(items, v.WorkflowTemplateVersionResponse)
		}
	}
	if items == nil {
		items = []WorkflowTemplateVersionResponse{}
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	total := len(items)
	return PagedVersionsResponse{
		Items:      items,
		Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: total, HasNext: false},
	}, nil
}

func (s *workflowService) CreateVersion(_ context.Context, templateID string, req CreateVersionRequest) (CreateVersionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for _, t := range s.templates {
		if t.ID == templateID {
			found = true
			break
		}
	}
	if !found {
		return CreateVersionResponse{}, ErrNotFound
	}

	versionNum := 1
	for _, v := range s.versions {
		if v.TemplateID == templateID {
			versionNum++
		}
	}

	s.versionNext++
	id := fmt.Sprintf("wftv-%d", s.versionNext)
	ver := workflowVersion{
		WorkflowTemplateVersionResponse: WorkflowTemplateVersionResponse{
			ID:           id,
			TemplateID:   templateID,
			Version:      versionNum,
			InputSchema:  req.InputSchema,
			OutputSchema: req.OutputSchema,
			Status:       "draft",
			CreatedAt:    time.Now(),
		},
	}
	for _, step := range req.Steps {
		s.stepTmplNext++
		ver.steps = append(ver.steps, WorkflowStepTemplateResponse{
			ID:                fmt.Sprintf("wfst-%d", s.stepTmplNext),
			TemplateVersionID: id,
			StepCode:          step.StepCode,
			StepType:          step.StepType,
			AgentCode:         step.AgentCode,
			OrderIndex:        step.OrderIndex,
			InputMapping:      step.InputMapping,
			OutputMapping:     step.OutputMapping,
		})
	}
	s.versions = append(s.versions, ver)
	return CreateVersionResponse{TemplateVersionID: id, StepCount: len(req.Steps), Status: "draft"}, nil
}

func (s *workflowService) GetVersion(_ context.Context, id string) (WorkflowTemplateVersionDetailResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.versions {
		if v.ID == id {
			return WorkflowTemplateVersionDetailResponse{
				WorkflowTemplateVersionResponse: v.WorkflowTemplateVersionResponse,
				Steps:                           append([]WorkflowStepTemplateResponse(nil), v.steps...),
			}, nil
		}
	}
	return WorkflowTemplateVersionDetailResponse{}, ErrNotFound
}

func (s *workflowService) PublishVersion(_ context.Context, id string, req PublishVersionRequest, idempotencyKey string) (PublishVersionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	payload := struct {
		ID      string
		Request PublishVersionRequest
	}{ID: id, Request: req}
	if cached, ok := s.idempotentResponse(idempotencyKey, "publish_version", payload); ok {
		return cached.(PublishVersionResponse), nil
	} else if idempotencyKey != "" {
		if _, exists := s.idempotentOps[idempotencyKey]; exists {
			return PublishVersionResponse{}, ErrIdempotencyConflict
		}
	}

	for i, v := range s.versions {
		if v.ID == id {
			if v.Status != "draft" {
				return PublishVersionResponse{}, ErrConflict
			}
			prev := v.Status
			s.versions[i].Status = "published"
			resp := PublishVersionResponse{
				PreviousStatus: prev,
				CurrentStatus:  "published",
				OperationLogID: s.nextOplogID(),
			}
			if idempotencyKey != "" {
				s.idempotentOps[idempotencyKey] = idempotentOperation{kind: "publish_version", payload: payload, response: resp}
			}
			return resp, nil
		}
	}
	return PublishVersionResponse{}, ErrNotFound
}

// --- Run methods ---

func (s *workflowService) ListRuns(_ context.Context, req ListWorkflowRunsRequest) (PagedWorkflowRunsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var items []WorkflowRunResponse
	for _, r := range s.runs {
		if req.ProjectID != "" && r.ProjectID != req.ProjectID {
			continue
		}
		if req.TemplateVersionID != "" && r.TemplateVersionID != req.TemplateVersionID {
			continue
		}
		if req.Status != "" && r.Status != req.Status {
			continue
		}
		items = append(items, r.WorkflowRunResponse)
	}
	if items == nil {
		items = []WorkflowRunResponse{}
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	total := len(items)
	return PagedWorkflowRunsResponse{
		Items:      items,
		Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: total, HasNext: false},
	}, nil
}

func (s *workflowService) CreateRun(_ context.Context, req CreateWorkflowRunRequest, idempotencyKey string) (CreateWorkflowRunResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	payload := req
	if cached, ok := s.idempotentResponse(idempotencyKey, "create_run", payload); ok {
		return cached.(CreateWorkflowRunResponse), nil
	} else if idempotencyKey != "" {
		if _, exists := s.idempotentOps[idempotencyKey]; exists {
			return CreateWorkflowRunResponse{}, ErrIdempotencyConflict
		}
	}

	s.runNext++
	id := fmt.Sprintf("wfr-%d", s.runNext)
	now := time.Now()

	stepCount := 0
	for _, v := range s.versions {
		if v.ID == req.TemplateVersionID {
			stepCount = len(v.steps)
			break
		}
	}

	run := workflowRun{
		WorkflowRunDetailResponse: WorkflowRunDetailResponse{
			WorkflowRunResponse: WorkflowRunResponse{
				ID:                id,
				ProjectID:         req.ProjectID,
				TemplateVersionID: req.TemplateVersionID,
				Status:            "pending",
				Source:            "manual",
				CreatedAt:         now,
				UpdatedAt:         now,
			},
			Input:     req.Input,
			StepCount: stepCount,
		},
		idempotencyKey: idempotencyKey,
	}
	s.runs = append(s.runs, run)

	resp := CreateWorkflowRunResponse{WorkflowRunID: id, Status: "pending"}
	if idempotencyKey != "" {
		s.idempotentOps[idempotencyKey] = idempotentOperation{kind: "create_run", payload: payload, response: resp}
	}
	return resp, nil
}

func (s *workflowService) GetRun(_ context.Context, id string) (WorkflowRunDetailResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.runs {
		if r.ID == id {
			// count step runs
			stepCount := 0
			for _, sr := range s.stepRuns {
				if sr.WorkflowRunID == id {
					stepCount++
				}
			}
			detail := r.WorkflowRunDetailResponse
			detail.StepCount = stepCount
			return detail, nil
		}
	}
	return WorkflowRunDetailResponse{}, ErrNotFound
}

func (s *workflowService) GetRunSteps(_ context.Context, runID string) (ListStepRunsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, r := range s.runs {
		if r.ID == runID {
			var items []WorkflowStepRunResponse
			for _, sr := range s.stepRuns {
				if sr.WorkflowRunID == runID {
					items = append(items, sr.WorkflowStepRunResponse)
				}
			}
			if items == nil {
				items = []WorkflowStepRunResponse{}
			}
			return ListStepRunsResponse{Items: items}, nil
		}
	}
	return ListStepRunsResponse{}, ErrNotFound
}

func (s *workflowService) CancelRun(_ context.Context, id string, req CancelRunRequest, idempotencyKey string) (CancelRunResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	payload := struct {
		ID      string
		Request CancelRunRequest
	}{ID: id, Request: req}
	if cached, ok := s.idempotentResponse(idempotencyKey, "cancel_run", payload); ok {
		return cached.(CancelRunResponse), nil
	} else if idempotencyKey != "" {
		if _, exists := s.idempotentOps[idempotencyKey]; exists {
			return CancelRunResponse{}, ErrIdempotencyConflict
		}
	}

	for i, r := range s.runs {
		if r.ID == id {
			if r.Status != "pending" && r.Status != "running" {
				return CancelRunResponse{}, ErrConflict
			}
			prev := r.Status
			s.runs[i].Status = "cancelled"
			s.runs[i].UpdatedAt = time.Now()
			resp := CancelRunResponse{
				PreviousStatus: prev,
				CurrentStatus:  "cancelled",
				OperationLogID: s.nextOplogID(),
			}
			if idempotencyKey != "" {
				s.idempotentOps[idempotencyKey] = idempotentOperation{kind: "cancel_run", payload: payload, response: resp}
			}
			return resp, nil
		}
	}
	return CancelRunResponse{}, ErrNotFound
}

func (s *workflowService) RetryRun(_ context.Context, id string, req RetryRunRequest, idempotencyKey string) (RetryRunResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	payload := struct {
		ID      string
		Request RetryRunRequest
	}{ID: id, Request: req}
	if cached, ok := s.idempotentResponse(idempotencyKey, "retry_run", payload); ok {
		return cached.(RetryRunResponse), nil
	} else if idempotencyKey != "" {
		if _, exists := s.idempotentOps[idempotencyKey]; exists {
			return RetryRunResponse{}, ErrIdempotencyConflict
		}
	}

	var original *workflowRun
	for i := range s.runs {
		if s.runs[i].ID == id {
			original = &s.runs[i]
			break
		}
	}
	if original == nil {
		return RetryRunResponse{}, ErrNotFound
	}
	if original.Status != "failed" {
		return RetryRunResponse{}, ErrConflict
	}

	s.runNext++
	newID := fmt.Sprintf("wfr-%d", s.runNext)
	now := time.Now()

	input := original.Input
	if req.InputOverride != nil {
		input = req.InputOverride
	}

	newRun := workflowRun{
		WorkflowRunDetailResponse: WorkflowRunDetailResponse{
			WorkflowRunResponse: WorkflowRunResponse{
				ID:                newID,
				ProjectID:         original.ProjectID,
				TemplateVersionID: original.TemplateVersionID,
				Status:            "pending",
				Source:            "retry",
				ParentRunID:       id,
				CreatedAt:         now,
				UpdatedAt:         now,
			},
			Input:     input,
			StepCount: original.StepCount,
		},
	}
	s.runs = append(s.runs, newRun)

	resp := RetryRunResponse{NewWorkflowRunID: newID, Status: "pending"}
	if idempotencyKey != "" {
		s.idempotentOps[idempotencyKey] = idempotentOperation{kind: "retry_run", payload: payload, response: resp}
	}
	return resp, nil
}

// --- EnginePort methods ---

func (s *workflowService) UpdateRunStatus(_ context.Context, id, status string, output map[string]any, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, r := range s.runs {
		if r.ID == id {
			s.runs[i].Status = status
			s.runs[i].UpdatedAt = time.Now()
			if output != nil {
				s.runs[i].Output = output
			}
			if errMsg != "" {
				s.runs[i].Error = errMsg
			}
			return nil
		}
	}
	return ErrNotFound
}

func (s *workflowService) CreateStepRun(_ context.Context, req CreateStepRunRequest) (WorkflowStepRunResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stepRunNext++
	id := fmt.Sprintf("wfsr-%d", s.stepRunNext)
	sr := workflowStepRun{WorkflowStepRunResponse{
		ID:             id,
		WorkflowRunID:  req.WorkflowRunID,
		StepTemplateID: req.StepTemplateID,
		Status:         "pending",
		Input:          req.Input,
	}}
	s.stepRuns = append(s.stepRuns, sr)
	return sr.WorkflowStepRunResponse, nil
}

func (s *workflowService) UpdateStepRunStatus(_ context.Context, id, status string, output map[string]any, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sr := range s.stepRuns {
		if sr.ID == id {
			s.stepRuns[i].Status = status
			if output != nil {
				s.stepRuns[i].Output = output
			}
			if errMsg != "" {
				s.stepRuns[i].Error = errMsg
			}
			return nil
		}
	}
	return ErrNotFound
}

func (s *workflowService) GetRunStepTemplates(_ context.Context, templateVersionID string) ([]WorkflowStepTemplateResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.versions {
		if v.ID == templateVersionID {
			return append([]WorkflowStepTemplateResponse(nil), v.steps...), nil
		}
	}
	return nil, ErrNotFound
}

func (s *workflowService) GetRunForEngine(_ context.Context, id string) (WorkflowRunResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.runs {
		if r.ID == id {
			return r.WorkflowRunResponse, nil
		}
	}
	return WorkflowRunResponse{}, ErrNotFound
}
