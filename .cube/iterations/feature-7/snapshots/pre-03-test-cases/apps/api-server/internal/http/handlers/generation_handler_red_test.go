package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/generation"
)

type generationHandlerStubService struct{}

func (generationHandlerStubService) CreateGenerationRun(ctx context.Context, projectID string, req generation.CreateGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.CreateGenerationRunResponse, error) {
	if idempotencyKey == "" || req.TargetCount <= 0 {
		return generation.CreateGenerationRunResponse{}, generation.ErrValidation
	}
	return generation.CreateGenerationRunResponse{GenerationRunID: "genrun-1", WorkflowRunID: workflowRunID, Status: generation.GenerationRunPending}, nil
}

func (generationHandlerStubService) CreateBatchGenerationRuns(ctx context.Context, projectID string, req generation.CreateBatchGenerationRunsRequest, workflowRunIDs []string, idempotencyKey string) (generation.CreateBatchGenerationRunsResponse, error) {
	if idempotencyKey == "" || req.BatchSize <= 0 {
		return generation.CreateBatchGenerationRunsResponse{}, generation.ErrValidation
	}
	return generation.CreateBatchGenerationRunsResponse{GenerationRunIDs: []string{"genrun-1"}, WorkflowRunIDs: workflowRunIDs, AcceptedCount: len(workflowRunIDs)}, nil
}

func (generationHandlerStubService) ListGenerationRuns(ctx context.Context, projectID string, req generation.ListGenerationRunsRequest) (generation.PagedGenerationRunsResponse, error) {
	return generation.PagedGenerationRunsResponse{Items: []generation.GenerationRunResponse{{ID: "genrun-1", ProjectID: projectID, WorkflowRunID: "wfr-1", Status: generation.GenerationRunPending}}, Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: 1}}, nil
}

func (generationHandlerStubService) GetGenerationRun(ctx context.Context, id string) (generation.GenerationRunDetailResponse, error) {
	return generation.GenerationRunDetailResponse{GenerationRunResponse: generation.GenerationRunResponse{ID: id, WorkflowRunID: "wfr-1", Status: generation.GenerationRunFailed}, ContentItems: []generation.ContentItemResponse{{ID: "ci-1", GenerationRunID: id, Status: generation.ContentItemGenerationFailed}}}, nil
}

func (generationHandlerStubService) RetryGenerationRun(ctx context.Context, id string, req generation.RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.RetryGenerationRunResponse, error) {
	if idempotencyKey == "" {
		return generation.RetryGenerationRunResponse{}, generation.ErrValidation
	}
	return generation.RetryGenerationRunResponse{NewGenerationRunID: "genrun-2", WorkflowRunID: workflowRunID, OperationLogID: "oplog-1"}, nil
}

func (generationHandlerStubService) ListContentItems(ctx context.Context, projectID string, req generation.ListContentItemsRequest) (generation.PagedContentItemsResponse, error) {
	return generation.PagedContentItemsResponse{Items: []generation.ContentItemResponse{{ID: "ci-1", ProjectID: projectID, GenerationRunID: "genrun-1", Status: generation.ContentItemPendingReview}}, Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: 1}}, nil
}

func (generationHandlerStubService) GetContentItem(ctx context.Context, id string) (generation.ContentItemDetailResponse, error) {
	return generation.ContentItemDetailResponse{ContentItemResponse: generation.ContentItemResponse{ID: id, ProjectID: "project-1", GenerationRunID: "genrun-1", Status: generation.ContentItemPendingReview}, Body: "generated body", Extension: generation.NovelChapterExtensionResponse{ChapterNo: 1}}, nil
}

type generationHandlerWorkflowStub struct{}

type generationHandlerSubmitterStub struct{ submitted []string }

func (s *generationHandlerSubmitterStub) Submit(workflowRunID string) bool {
	s.submitted = append(s.submitted, workflowRunID)
	return true
}

var _ engine.Submitter = (*generationHandlerSubmitterStub)(nil)

// @Test
func TestTask03CreateGenerationRunReturnsAcceptedEnvelopeAndSubmitsWorkflow(t *testing.T) {
	submitter := &generationHandlerSubmitterStub{}
	handler := NewGenerationHandler(generationHandlerStubService{}, nil, submitter, nil)
	body := bytes.NewBufferString(`{"confirmed_topic_id":"topic-1","worldview_version_id":"worldview-v1","arc_id":"arc-1","target_count":1,"start_sequence_no":1,"template_version_id":"wftv-generation","generation_config":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/project-1/generation-runs", body)
	req.Header.Set("Idempotency-Key", "idem-1")
	req = withChiParam(req, "projectId", "project-1")
	w := httptest.NewRecorder()

	handler.CreateGenerationRun(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusAccepted, w.Body.String())
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			GenerationRunID string `json:"generation_run_id"`
			WorkflowRunID   string `json:"workflow_run_id"`
			Status          string `json:"status"`
		} `json:"data"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.Success || envelope.Data.GenerationRunID != "genrun-1" || envelope.Data.WorkflowRunID == "" || envelope.Data.Status != generation.GenerationRunPending || envelope.RequestID == "" {
		t.Fatalf("unexpected envelope: %#v", envelope)
	}
	if len(submitter.submitted) != 1 || submitter.submitted[0] != envelope.Data.WorkflowRunID {
		t.Fatalf("expected workflow submit, got %#v", submitter.submitted)
	}
}

// @Test
func TestTask03CreateGenerationRunValidationFailureReturnsRequestIDAndErrorCode(t *testing.T) {
	handler := NewGenerationHandler(generationHandlerStubService{}, nil, &generationHandlerSubmitterStub{}, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/project-1/generation-runs", bytes.NewBufferString(`{"target_count":0}`))
	req = withChiParam(req, "projectId", "project-1")
	w := httptest.NewRecorder()

	handler.CreateGenerationRun(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	assertErrorCode(t, w.Body.Bytes(), "VALIDATION_ERROR")
}

// @Test
func TestTask03ReadEndpointsReturnGenerationRunAndContentItemPagination(t *testing.T) {
	handler := NewGenerationHandler(generationHandlerStubService{}, nil, &generationHandlerSubmitterStub{}, nil)
	for _, tc := range []struct {
		name   string
		handle func(http.ResponseWriter, *http.Request)
		path   string
		param  string
		value  string
		want   string
	}{
		{"generation list", handler.ListGenerationRuns, "/api/v1/projects/project-1/generation-runs", "projectId", "project-1", "genrun-1"},
		{"generation detail", handler.GetGenerationRun, "/api/v1/generation-runs/genrun-1", "id", "genrun-1", "ci-1"},
		{"content list", handler.ListContentItems, "/api/v1/projects/project-1/content-items", "projectId", "project-1", "pending_review"},
		{"content detail", handler.GetContentItem, "/api/v1/content-items/ci-1", "id", "ci-1", "generated body"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := withChiParam(httptest.NewRequest(http.MethodGet, tc.path, nil), tc.param, tc.value)
			w := httptest.NewRecorder()
			tc.handle(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
			}
			if !bytes.Contains(w.Body.Bytes(), []byte(tc.want)) || !bytes.Contains(w.Body.Bytes(), []byte("request_id")) {
				t.Fatalf("expected response to include %q and request_id, got %s", tc.want, w.Body.String())
			}
		})
	}
}

// @Test
func TestTask03RetryGenerationRunReturnsNewRunAndOperationLog(t *testing.T) {
	handler := NewGenerationHandler(generationHandlerStubService{}, nil, &generationHandlerSubmitterStub{}, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/generation-runs/genrun-1/retry", bytes.NewBufferString(`{"reason":"retry after provider failure","input_override":{}}`))
	req.Header.Set("Idempotency-Key", "idem-retry")
	req = withChiParam(req, "id", "genrun-1")
	w := httptest.NewRecorder()

	handler.RetryGenerationRun(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusAccepted, w.Body.String())
	}
	for _, want := range []string{"new_generation_run_id", "workflow_run_id", "operation_log_id", "request_id"} {
		if !bytes.Contains(w.Body.Bytes(), []byte(want)) {
			t.Fatalf("expected retry response to include %q, got %s", want, w.Body.String())
		}
	}
}

func withChiParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func assertErrorCode(t *testing.T, body []byte, code string) {
	t.Helper()
	var envelope struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Success || envelope.Error.Code != code || envelope.RequestID == "" {
		t.Fatalf("unexpected error envelope: %#v", envelope)
	}
}
