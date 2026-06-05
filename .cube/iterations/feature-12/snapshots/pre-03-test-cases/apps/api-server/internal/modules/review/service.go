package review

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

var (
	ErrValidation          = errors.New("validation error")
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrIdempotencyConflict = errors.New("idempotency conflict")
	ErrWorkflowRunFailed   = errors.New("workflow run failed")
	ErrLLMProviderError    = errors.New("llm provider error")
)

type Service interface {
	ListReviews(ctx context.Context, req ListReviewsRequest) (PagedContentReviewsResponse, error)
	CreateReview(ctx context.Context, contentItemID string, req CreateReviewRequest, idempotencyKey string) (CreateReviewResponse, error)
	GetReview(ctx context.Context, id string) (ContentReviewDetailResponse, error)
	TriggerAIReport(ctx context.Context, id string, req TriggerAIReportRequest, workflowRunID string, idempotencyKey string) (TriggerAIReportResponse, error)
	GetAIReport(ctx context.Context, id string) (ReviewReportResponse, error)
	ApproveReview(ctx context.Context, id string, req ApproveReviewRequest) (ApproveReviewResponse, error)
	RejectReview(ctx context.Context, id string, req RejectReviewRequest, regenerationRunID string) (RejectReviewResponse, error)
	ApproveWithEdit(ctx context.Context, id string, req ApproveWithEditRequest) (ApproveWithEditResponse, error)
	ListContentVersions(ctx context.Context, contentItemID string, req ListReviewsRequest) (PagedContentVersionsResponse, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) ListReviews(ctx context.Context, req ListReviewsRequest) (PagedContentReviewsResponse, error) {
	if req.ProjectID == "" {
		return PagedContentReviewsResponse{}, ErrValidation
	}
	return PagedContentReviewsResponse{Items: []ContentReviewResponse{{ID: "review-1", ProjectID: req.ProjectID, ContentItemID: "content-item-1", ReviewType: "combined", Status: ReviewStatusPending, Title: "Draft", UpdatedAt: time.Now().UTC()}}, Pagination: contentPagination(req)}, nil
}

func (s *service) CreateReview(ctx context.Context, contentItemID string, req CreateReviewRequest, idempotencyKey string) (CreateReviewResponse, error) {
	if contentItemID == "" || !isReviewType(req.ReviewType) || idempotencyKey == "" {
		return CreateReviewResponse{}, ErrValidation
	}
	return CreateReviewResponse{ReviewID: "review-" + contentItemID, Status: ReviewStatusPending}, nil
}

func (s *service) GetReview(ctx context.Context, id string) (ContentReviewDetailResponse, error) {
	if id == "" {
		return ContentReviewDetailResponse{}, ErrNotFound
	}
	return ContentReviewDetailResponse{
		ContentReviewResponse: ContentReviewResponse{ID: id, ProjectID: "project-1", ContentItemID: "content-item-1", ReviewType: "combined", Status: ReviewStatusInReview, Title: "Draft", UpdatedAt: time.Now().UTC()},
		Body:                  "draft body",
		Metadata:              map[string]any{},
		Extension:             map[string]any{},
		ReportSummary:         ReviewReportSummaryResponse{ID: "report-" + id, Status: ReviewReportStatusSucceeded, QualityScore: 80, RiskLevel: "medium"},
		Versions:              []ContentVersionResponse{contentVersion("content-item-1", 1, "generation")},
	}, nil
}

func (s *service) TriggerAIReport(ctx context.Context, id string, req TriggerAIReportRequest, workflowRunID string, idempotencyKey string) (TriggerAIReportResponse, error) {
	if id == "" || !isReportType(req.ReportType) || workflowRunID == "" || idempotencyKey == "" {
		return TriggerAIReportResponse{}, ErrValidation
	}
	return TriggerAIReportResponse{ReportID: "report-" + id, JobID: "job-" + id, WorkflowRunID: workflowRunID, Status: ReviewReportStatusGenerating}, nil
}

func (s *service) GetAIReport(ctx context.Context, id string) (ReviewReportResponse, error) {
	if id == "" {
		return ReviewReportResponse{}, ErrNotFound
	}
	return ReviewReportResponse{ID: "report-" + id, ReviewID: id, ContentItemID: "content-item-1", Status: ReviewReportStatusSucceeded, QualityScore: 85, RiskLevel: "low", Issues: []ReviewIssue{{Code: "clarity", Severity: "medium", Message: "Improve clarity"}}, Suggestions: []ReviewSuggestion{{Code: "structure", Message: "Strengthen structure"}}}, nil
}

func (s *service) ApproveReview(ctx context.Context, id string, req ApproveReviewRequest) (ApproveReviewResponse, error) {
	if id == "" {
		return ApproveReviewResponse{}, ErrValidation
	}
	return ApproveReviewResponse{ReviewID: id, Status: ReviewStatusApproved, OperationLogID: "oplog-" + id}, nil
}

func (s *service) RejectReview(ctx context.Context, id string, req RejectReviewRequest, regenerationRunID string) (RejectReviewResponse, error) {
	if id == "" || req.Reason == "" {
		return RejectReviewResponse{}, ErrValidation
	}
	result := RejectReviewResponse{ReviewID: id, Status: ReviewStatusRejected, OperationLogID: "oplog-" + id}
	if req.TriggerRegeneration {
		if req.RegenerateInstruction == "" {
			return RejectReviewResponse{}, ErrValidation
		}
		if regenerationRunID == "" {
			return RejectReviewResponse{}, ErrWorkflowRunFailed
		}
		result.RegenerationRunID = regenerationRunID
		result.JobID = "job-" + id
	}
	return result, nil
}

func (s *service) ApproveWithEdit(ctx context.Context, id string, req ApproveWithEditRequest) (ApproveWithEditResponse, error) {
	if id == "" || len(req.EditableFields) == 0 {
		return ApproveWithEditResponse{}, ErrValidation
	}
	return ApproveWithEditResponse{ReviewID: id, Status: ReviewStatusApprovedWithEdit, ContentVersionID: "version-" + id, OperationLogID: "oplog-" + id}, nil
}

func (s *service) ListContentVersions(ctx context.Context, contentItemID string, req ListReviewsRequest) (PagedContentVersionsResponse, error) {
	if contentItemID == "" {
		return PagedContentVersionsResponse{}, ErrValidation
	}
	return PagedContentVersionsResponse{Items: []ContentVersionResponse{contentVersion(contentItemID, 1, "generation")}, Pagination: contentPagination(req)}, nil
}

func isReviewType(reviewType string) bool {
	switch reviewType {
	case "manual", "ai", "combined":
		return true
	default:
		return false
	}
}

func isReportType(reportType string) bool {
	switch reportType {
	case "default", "quality":
		return true
	default:
		return false
	}
}

func contentVersion(contentItemID string, versionNo int, source string) ContentVersionResponse {
	return ContentVersionResponse{ID: fmt.Sprintf("version-%s-%d", contentItemID, versionNo), ContentItemID: contentItemID, VersionNo: versionNo, Source: source, Title: "Draft", Body: "draft body", EditableFields: map[string]any{}, Summary: fmt.Sprintf("v%d", versionNo), CreatedAt: time.Now().UTC()}
}

func contentPagination(req ListReviewsRequest) content.PaginationResponse {
	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	return content.PaginationResponse{Page: page, PageSize: pageSize, Total: 1, HasNext: false}
}
