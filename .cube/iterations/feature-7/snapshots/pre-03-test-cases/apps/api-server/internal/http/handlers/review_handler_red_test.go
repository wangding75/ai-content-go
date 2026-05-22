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
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/review"
)

type reviewHandlerStubService struct{}

func (reviewHandlerStubService) ListReviews(ctx context.Context, req review.ListReviewsRequest) (review.PagedContentReviewsResponse, error) {
	if req.ProjectID == "" {
		return review.PagedContentReviewsResponse{}, review.ErrValidation
	}
	return review.PagedContentReviewsResponse{Items: []review.ContentReviewResponse{{ID: "review-1", ProjectID: req.ProjectID, ContentItemID: "content-item-1", ReviewType: "combined", Status: review.ReviewStatusPending, Title: "Draft"}}, Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: 1}}, nil
}

func (reviewHandlerStubService) CreateReview(ctx context.Context, contentItemID string, req review.CreateReviewRequest, idempotencyKey string) (review.CreateReviewResponse, error) {
	if contentItemID == "" || req.ReviewType == "" || idempotencyKey == "" {
		return review.CreateReviewResponse{}, review.ErrValidation
	}
	return review.CreateReviewResponse{ReviewID: "review-1", Status: review.ReviewStatusPending}, nil
}

func (reviewHandlerStubService) GetReview(ctx context.Context, id string) (review.ContentReviewDetailResponse, error) {
	return review.ContentReviewDetailResponse{ContentReviewResponse: review.ContentReviewResponse{ID: id, ProjectID: "project-1", ContentItemID: "content-item-1", ReviewType: "combined", Status: review.ReviewStatusInReview, Title: "Draft"}, Body: "draft body", Metadata: map[string]any{}, Extension: map[string]any{}, ReportSummary: review.ReviewReportSummaryResponse{ID: "report-1", Status: review.ReviewReportStatusSucceeded, QualityScore: 80, RiskLevel: "medium"}, Versions: []review.ContentVersionResponse{{ID: "version-1", ContentItemID: "content-item-1", VersionNo: 1, Source: "generation", Title: "Draft", Body: "draft body", EditableFields: map[string]any{}, Summary: "v1"}}}, nil
}

func (reviewHandlerStubService) TriggerAIReport(ctx context.Context, id string, req review.TriggerAIReportRequest, workflowRunID string, idempotencyKey string) (review.TriggerAIReportResponse, error) {
	if id == "" || req.ReportType == "" || workflowRunID == "" || idempotencyKey == "" {
		return review.TriggerAIReportResponse{}, review.ErrValidation
	}
	return review.TriggerAIReportResponse{ReportID: "report-1", JobID: "job-1", WorkflowRunID: workflowRunID, Status: review.ReviewReportStatusGenerating}, nil
}

func (reviewHandlerStubService) GetAIReport(ctx context.Context, id string) (review.ReviewReportResponse, error) {
	return review.ReviewReportResponse{ID: "report-1", ReviewID: id, ContentItemID: "content-item-1", Status: review.ReviewReportStatusSucceeded, QualityScore: 85, RiskLevel: "low", Issues: []review.ReviewIssue{{Code: "clarity", Severity: "medium", Message: "Improve clarity"}}, Suggestions: []review.ReviewSuggestion{{Code: "structure", Message: "Strengthen structure"}}}, nil
}

func (reviewHandlerStubService) ApproveReview(ctx context.Context, id string, req review.ApproveReviewRequest) (review.ApproveReviewResponse, error) {
	return review.ApproveReviewResponse{ReviewID: id, Status: review.ReviewStatusApproved, OperationLogID: "oplog-approve"}, nil
}

func (reviewHandlerStubService) RejectReview(ctx context.Context, id string, req review.RejectReviewRequest, regenerationRunID string) (review.RejectReviewResponse, error) {
	if req.Reason == "" {
		return review.RejectReviewResponse{}, review.ErrValidation
	}
	return review.RejectReviewResponse{ReviewID: id, Status: review.ReviewStatusRejected, OperationLogID: "oplog-reject", RegenerationRunID: regenerationRunID, JobID: "job-retry"}, nil
}

func (reviewHandlerStubService) ApproveWithEdit(ctx context.Context, id string, req review.ApproveWithEditRequest) (review.ApproveWithEditResponse, error) {
	if len(req.EditableFields) == 0 {
		return review.ApproveWithEditResponse{}, review.ErrValidation
	}
	return review.ApproveWithEditResponse{ReviewID: id, Status: review.ReviewStatusApprovedWithEdit, ContentVersionID: "version-2", OperationLogID: "oplog-edit"}, nil
}

func (reviewHandlerStubService) ListContentVersions(ctx context.Context, contentItemID string, req review.ListReviewsRequest) (review.PagedContentVersionsResponse, error) {
	return review.PagedContentVersionsResponse{Items: []review.ContentVersionResponse{{ID: "version-1", ContentItemID: contentItemID, VersionNo: 1, Source: "generation", Title: "Draft", Body: "draft body", EditableFields: map[string]any{}, Summary: "v1"}}, Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: 1}}, nil
}

type reviewHandlerSubmitterStub struct{ submitted []string }

func (s *reviewHandlerSubmitterStub) Submit(workflowRunID string) bool {
	s.submitted = append(s.submitted, workflowRunID)
	return true
}

var _ engine.Submitter = (*reviewHandlerSubmitterStub)(nil)

func reviewReqWithChiParam(method string, target string, body *bytes.Buffer, key string, value string) *http.Request {
	var req *http.Request
	if body == nil {
		req = httptest.NewRequest(method, target, nil)
	} else {
		req = httptest.NewRequest(method, target, body)
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func assertReviewEnvelopeContains(t *testing.T, body []byte, fields ...string) {
	t.Helper()
	var envelope struct {
		Success   bool            `json:"success"`
		Data      json.RawMessage `json:"data"`
		Error     json.RawMessage `json:"error"`
		RequestID string          `json:"request_id"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode envelope: %v; body=%s", err, string(body))
	}
	if envelope.RequestID == "" {
		t.Fatalf("response must include request_id: %s", string(body))
	}
	for _, field := range fields {
		if !bytes.Contains(body, []byte(field)) {
			t.Fatalf("response missing %q: %s", field, string(body))
		}
	}
}

// @Test
func TestTask04ReviewListCreateAndDetailEndpointsReturnEnvelopeContracts(t *testing.T) {
	handler := NewReviewHandler(reviewHandlerStubService{}, nil, &reviewHandlerSubmitterStub{}, nil)
	cases := []struct {
		name   string
		handle func(http.ResponseWriter, *http.Request)
		req    *http.Request
		status int
		fields []string
	}{
		{"list", handler.ListReviews, httptest.NewRequest(http.MethodGet, "/api/v1/content-reviews?project_id=project-1", nil), http.StatusOK, []string{"review-1", "pagination"}},
		{"create", handler.CreateReview, reviewReqWithChiParam(http.MethodPost, "/api/v1/content-items/content-item-1/reviews", bytes.NewBufferString(`{"review_type":"combined"}`), "id", "content-item-1"), http.StatusCreated, []string{"review_id", "pending"}},
		{"detail", handler.GetReview, reviewReqWithChiParam(http.MethodGet, "/api/v1/content-reviews/review-1", nil, "id", "review-1"), http.StatusOK, []string{"draft body", "report_summary", "versions"}},
	}
	cases[1].req.Header.Set("Idempotency-Key", "review-idem-1")

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tc.handle(w, tc.req)
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d; body=%s", w.Code, tc.status, w.Body.String())
			}
			assertReviewEnvelopeContains(t, w.Body.Bytes(), tc.fields...)
		})
	}
}

// @Test
func TestTask04AIReportEndpointReturnsAcceptedAndSubmitsWorkflow(t *testing.T) {
	submitter := &reviewHandlerSubmitterStub{}
	handler := NewReviewHandler(reviewHandlerStubService{}, nil, submitter, nil)
	req := reviewReqWithChiParam(http.MethodPost, "/api/v1/content-reviews/review-1/ai-report", bytes.NewBufferString(`{"report_type":"default","config":{}}`), "id", "review-1")
	req.Header.Set("Idempotency-Key", "ai-report-idem-1")
	w := httptest.NewRecorder()

	handler.TriggerAIReport(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusAccepted, w.Body.String())
	}
	assertReviewEnvelopeContains(t, w.Body.Bytes(), "report_id", "workflow_run_id", "generating")
	if len(submitter.submitted) != 1 {
		t.Fatalf("expected workflow submission, got %#v", submitter.submitted)
	}
}

// @Test
func TestTask04ApproveRejectEditAndVersionEndpointsExposeOperationAndVersionContracts(t *testing.T) {
	handler := NewReviewHandler(reviewHandlerStubService{}, nil, &reviewHandlerSubmitterStub{}, nil)
	cases := []struct {
		name   string
		handle func(http.ResponseWriter, *http.Request)
		req    *http.Request
		status int
		fields []string
	}{
		{"approve", handler.ApproveReview, reviewReqWithChiParam(http.MethodPost, "/api/v1/content-reviews/review-1/approve", bytes.NewBufferString(`{"note":"approved"}`), "id", "review-1"), http.StatusOK, []string{"operation_log_id", "approved"}},
		{"reject", handler.RejectReview, reviewReqWithChiParam(http.MethodPost, "/api/v1/content-reviews/review-1/reject", bytes.NewBufferString(`{"reason":"needs changes","regenerate_instruction":"retry","trigger_regeneration":true}`), "id", "review-1"), http.StatusAccepted, []string{"operation_log_id", "regeneration_run_id"}},
		{"edit", handler.ApproveWithEdit, reviewReqWithChiParam(http.MethodPost, "/api/v1/content-reviews/review-1/approve-with-edit", bytes.NewBufferString(`{"editable_fields":{"title":"Edited"},"note":"edited"}`), "id", "review-1"), http.StatusOK, []string{"content_version_id", "operation_log_id"}},
		{"versions", handler.ListContentVersions, reviewReqWithChiParam(http.MethodGet, "/api/v1/content-items/content-item-1/versions", nil, "id", "content-item-1"), http.StatusOK, []string{"version_no", "pagination"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tc.handle(w, tc.req)
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d; body=%s", w.Code, tc.status, w.Body.String())
			}
			assertReviewEnvelopeContains(t, w.Body.Bytes(), tc.fields...)
		})
	}
}

// @Test
func TestTask04ReviewValidationErrorsIncludeRequestIDAndErrorCode(t *testing.T) {
	handler := NewReviewHandler(reviewHandlerStubService{}, nil, &reviewHandlerSubmitterStub{}, nil)
	req := reviewReqWithChiParam(http.MethodPost, "/api/v1/content-reviews/review-1/reject", bytes.NewBufferString(`{"reason":"","regenerate_instruction":"","trigger_regeneration":false}`), "id", "review-1")
	w := httptest.NewRecorder()

	handler.RejectReview(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	assertReviewEnvelopeContains(t, w.Body.Bytes(), "VALIDATION_ERROR")
}
