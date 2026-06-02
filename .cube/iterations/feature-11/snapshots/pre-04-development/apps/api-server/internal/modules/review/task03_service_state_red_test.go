package review

import (
	"context"
	"testing"
)

// @Test
func TestTask03CreateReviewReturnsPendingReviewIDForValidPendingContentItem(t *testing.T) {
	service := NewService()

	result, err := service.CreateReview(context.Background(), "content-item-1", CreateReviewRequest{ReviewType: "combined"}, "review-idem-1")
	if err != nil {
		t.Fatalf("CreateReview returned error: %v", err)
	}
	if result.ReviewID == "" || result.Status != ReviewStatusPending {
		t.Fatalf("unexpected create review result: %#v", result)
	}
}

// @Test
func TestTask03ApproveRejectAndEditApproveReturnOperationLogIDs(t *testing.T) {
	service := NewService()

	approved, err := service.ApproveReview(context.Background(), "review-1", ApproveReviewRequest{Note: "approved"})
	if err != nil {
		t.Fatalf("ApproveReview returned error: %v", err)
	}
	if approved.Status != ReviewStatusApproved || approved.OperationLogID == "" {
		t.Fatalf("approve must return approved status and operation log: %#v", approved)
	}

	rejected, err := service.RejectReview(context.Background(), "review-2", RejectReviewRequest{Reason: "needs changes", RegenerateInstruction: "regenerate with stronger structure", TriggerRegeneration: true}, "workflow-run-1")
	if err != nil {
		t.Fatalf("RejectReview returned error: %v", err)
	}
	if rejected.Status != ReviewStatusRejected || rejected.OperationLogID == "" || rejected.RegenerationRunID == "" {
		t.Fatalf("reject with regeneration must return operation log and regeneration run: %#v", rejected)
	}

	edited, err := service.ApproveWithEdit(context.Background(), "review-3", ApproveWithEditRequest{EditableFields: map[string]any{"title": "Edited title", "body": "Edited body"}, Note: "edited"})
	if err != nil {
		t.Fatalf("ApproveWithEdit returned error: %v", err)
	}
	if edited.Status != ReviewStatusApprovedWithEdit || edited.OperationLogID == "" || edited.ContentVersionID == "" {
		t.Fatalf("approve-with-edit must return version and operation log: %#v", edited)
	}
}

// @Test
func TestTask03TriggerAIReportReturnsAsyncTraceIDs(t *testing.T) {
	service := NewService()

	result, err := service.TriggerAIReport(context.Background(), "review-1", TriggerAIReportRequest{ReportType: "default", Config: map[string]any{"strictness": "high"}}, "workflow-run-1", "ai-report-idem-1")
	if err != nil {
		t.Fatalf("TriggerAIReport returned error: %v", err)
	}
	if result.ReportID == "" || result.WorkflowRunID != "workflow-run-1" || result.Status != ReviewReportStatusGenerating {
		t.Fatalf("trigger AI report must return report and workflow trace IDs: %#v", result)
	}
}
