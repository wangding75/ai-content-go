package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/novel"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type task03NovelService struct{}

func (task03NovelService) CreatePlanningRun(ctx context.Context, projectID string, req novel.CreatePlanningRunRequest, workflowRunID string, idempotencyKey string) (novel.CreatePlanningRunResponse, error) {
	return novel.CreatePlanningRunResponse{PlanningRunID: "plan-1", WorkflowRunID: workflowRunID, Status: "pending"}, nil
}
func (task03NovelService) ListPlanningRuns(ctx context.Context, projectID string, req novel.ListPlanningRunsRequest) (novel.PagedPlanningRunsResponse, error) {
	return novel.PagedPlanningRunsResponse{Items: []novel.PlanningRunResponse{{ID: "plan-1", ProjectID: projectID, Status: "pending"}}, Pagination: content.PaginationResponse{Page: req.Page, PageSize: req.PageSize, Total: 1}}, nil
}
func (task03NovelService) GetPlanningRun(ctx context.Context, projectID, runID string) (novel.PlanningRunDetailResponse, error) {
	return novel.PlanningRunDetailResponse{PlanningRunResponse: novel.PlanningRunResponse{ID: runID, ProjectID: projectID, WorkflowRunID: "wfr-1", Status: "pending"}}, nil
}
func (task03NovelService) ConfirmTopic(ctx context.Context, projectID, topicID string, req novel.ConfirmTopicRequest, idempotencyKey string) (novel.ConfirmTopicResponse, error) {
	return novel.ConfirmTopicResponse{ConfirmedTopicID: "topic-1", PreviousStatus: "candidate", CurrentStatus: "confirmed", OperationLogID: "oplog-1"}, nil
}
func (task03NovelService) GetWorldview(ctx context.Context, projectID string) (novel.WorldviewResponse, error) {
	return novel.WorldviewResponse{ProjectID: projectID, VersionID: "world-1", Version: 1, Worldview: map[string]any{"era": "future"}, ForbiddenRules: []string{"no deus ex machina"}}, nil
}
func (task03NovelService) UpdateWorldview(ctx context.Context, projectID string, req novel.UpdateWorldviewRequest) (novel.UpdateWorldviewResponse, error) {
	return novel.UpdateWorldviewResponse{VersionID: "world-2", OperationLogID: "oplog-2"}, nil
}
func (task03NovelService) ListCharacters(ctx context.Context, projectID string, req novel.ListCharactersRequest) (novel.PagedCharactersResponse, error) {
	return novel.PagedCharactersResponse{Items: []novel.CharacterResponse{{CharacterID: "char-1", ProjectID: projectID, Name: "Lin", Role: req.Role}}, Pagination: content.PaginationResponse{Page: req.Page, PageSize: req.PageSize, Total: 1}}, nil
}
func (task03NovelService) CreateCharacter(ctx context.Context, projectID string, req novel.CreateCharacterRequest) (novel.CreateCharacterResponse, error) {
	return novel.CreateCharacterResponse{CharacterID: "char-2", OperationLogID: "oplog-3"}, nil
}
func (task03NovelService) ListArcs(ctx context.Context, projectID string, req novel.ListArcsRequest) (novel.PagedArcsResponse, error) {
	return novel.PagedArcsResponse{Items: []novel.ArcResponse{{ArcID: "arc-1", ProjectID: projectID, Title: "Act 1", OrderIndex: 1}}, Pagination: content.PaginationResponse{Page: req.Page, PageSize: req.PageSize, Total: 1}}, nil
}

type task03WorkflowService struct{ cancelled bool }

func (s *task03WorkflowService) ListTemplates(ctx context.Context, req workflow.ListWorkflowTemplatesRequest) (workflow.PagedWorkflowTemplatesResponse, error) {
	return workflow.PagedWorkflowTemplatesResponse{}, nil
}
func (s *task03WorkflowService) CreateTemplate(ctx context.Context, req workflow.CreateWorkflowTemplateRequest) (workflow.CreateWorkflowTemplateResponse, error) {
	return workflow.CreateWorkflowTemplateResponse{}, nil
}
func (s *task03WorkflowService) GetTemplate(ctx context.Context, id string) (workflow.WorkflowTemplateResponse, error) {
	return workflow.WorkflowTemplateResponse{}, nil
}
func (s *task03WorkflowService) ListVersions(ctx context.Context, templateID string, req workflow.PaginationRequest) (workflow.PagedVersionsResponse, error) {
	return workflow.PagedVersionsResponse{}, nil
}
func (s *task03WorkflowService) CreateVersion(ctx context.Context, templateID string, req workflow.CreateVersionRequest) (workflow.CreateVersionResponse, error) {
	return workflow.CreateVersionResponse{}, nil
}
func (s *task03WorkflowService) GetVersion(ctx context.Context, id string) (workflow.WorkflowTemplateVersionDetailResponse, error) {
	return workflow.WorkflowTemplateVersionDetailResponse{}, nil
}
func (s *task03WorkflowService) PublishVersion(ctx context.Context, id string, req workflow.PublishVersionRequest, idempotencyKey string) (workflow.PublishVersionResponse, error) {
	return workflow.PublishVersionResponse{}, nil
}
func (s *task03WorkflowService) ListRuns(ctx context.Context, req workflow.ListWorkflowRunsRequest) (workflow.PagedWorkflowRunsResponse, error) {
	return workflow.PagedWorkflowRunsResponse{}, nil
}
func (s *task03WorkflowService) CreateRun(ctx context.Context, req workflow.CreateWorkflowRunRequest, idempotencyKey string) (workflow.CreateWorkflowRunResponse, error) {
	return workflow.CreateWorkflowRunResponse{WorkflowRunID: "wfr-1", Status: "pending"}, nil
}
func (s *task03WorkflowService) GetRun(ctx context.Context, id string) (workflow.WorkflowRunDetailResponse, error) {
	return workflow.WorkflowRunDetailResponse{}, nil
}
func (s *task03WorkflowService) GetRunSteps(ctx context.Context, runID string) (workflow.ListStepRunsResponse, error) {
	return workflow.ListStepRunsResponse{}, nil
}
func (s *task03WorkflowService) CancelRun(ctx context.Context, id string, req workflow.CancelRunRequest, idempotencyKey string) (workflow.CancelRunResponse, error) {
	s.cancelled = true
	return workflow.CancelRunResponse{PreviousStatus: "pending", CurrentStatus: "cancelled", OperationLogID: "oplog-cancel"}, nil
}
func (s *task03WorkflowService) RetryRun(ctx context.Context, id string, req workflow.RetryRunRequest, idempotencyKey string) (workflow.RetryRunResponse, error) {
	return workflow.RetryRunResponse{}, nil
}

type task03Submitter struct{ submitted string }

func (s *task03Submitter) Submit(runID string) bool { s.submitted = runID; return true }

var _ workflow.Service = (*task03WorkflowService)(nil)
var _ engine.Submitter = (*task03Submitter)(nil)

// @Test
func TestTask03CreatePlanningRunRequiresIdempotencyKey(t *testing.T) {
	h := NewNovelHandler(task03NovelService{}, &task03WorkflowService{}, &task03Submitter{}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"genre":"fantasy","audience":"ya","count":1,"template_version_id":"wftv-1"}`))
	rec := httptest.NewRecorder()
	h.CreatePlanningRun(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	var envelope api.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error == nil || envelope.Error.Code != api.ErrorValidation {
		t.Fatalf("expected validation envelope, got %#v", envelope)
	}
}

// @Test
func TestTask03CreatePlanningRunReturnsAcceptedEnvelopeAndSubmitsWorkflow(t *testing.T) {
	submitter := &task03Submitter{}
	h := NewNovelHandler(task03NovelService{}, &task03WorkflowService{}, submitter, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"genre":"fantasy","audience":"ya","count":1,"template_version_id":"wftv-1"}`))
	req.Header.Set("Idempotency-Key", "idem-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "project-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.CreatePlanningRun(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body %s", rec.Code, rec.Body.String())
	}
	if submitter.submitted != "wfr-1" {
		t.Fatalf("expected workflow submission wfr-1, got %q", submitter.submitted)
	}
}

// @Test
func TestTask03InvalidPaginationMapsToValidationEnvelope(t *testing.T) {
	h := NewNovelHandler(task03NovelService{}, &task03WorkflowService{}, &task03Submitter{}, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/?page=abc&page_size=20", nil)
	rec := httptest.NewRecorder()
	h.ListPlanningRuns(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid pagination 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), string(api.ErrorValidation)) {
		t.Fatalf("expected validation error code, got %s", rec.Body.String())
	}
}

type task03FailingNovelService struct{ task03NovelService }

func (task03FailingNovelService) CreatePlanningRun(ctx context.Context, projectID string, req novel.CreatePlanningRunRequest, workflowRunID string, idempotencyKey string) (novel.CreatePlanningRunResponse, error) {
	return novel.CreatePlanningRunResponse{}, novel.ErrConflict
}

// @Test
func TestTask03CreatePlanningRunCompensatesWorkflowWhenNovelRecordFails(t *testing.T) {
	wf := &task03WorkflowService{}
	h := NewNovelHandler(task03FailingNovelService{}, wf, &task03Submitter{}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"genre":"fantasy","audience":"ya","count":1,"template_version_id":"wftv-1"}`))
	req.Header.Set("Idempotency-Key", "idem-1")
	rec := httptest.NewRecorder()
	h.CreatePlanningRun(rec, req)
	if !wf.cancelled {
		t.Fatal("expected workflow cancellation compensation")
	}
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected conflict from novel service, got %d", rec.Code)
	}
}

func TestTask03CompileGuardUsesErrorsPackage(t *testing.T) {
	if !errors.Is(novel.ErrConflict, novel.ErrConflict) {
		t.Fatal("unreachable")
	}
}
