package review

import (
	"encoding/json"
	"strings"
	"testing"
)

// @Test
func TestTask01ReviewStatusConstantsAreContentNeutral(t *testing.T) {
	statuses := []string{ReviewStatusPending, ReviewStatusInReview, ReviewStatusApproved, ReviewStatusRejected, ReviewStatusApprovedWithEdit}
	for _, status := range statuses {
		if status == "" {
			t.Fatalf("review status must not be empty")
		}
		if strings.Contains(status, "book") || strings.Contains(status, "chapter") || strings.Contains(status, "novel") {
			t.Fatalf("review status must remain content-type neutral: %q", status)
		}
	}
	if ReviewReportStatusGenerating != "generating" || ReviewReportStatusSucceeded != "succeeded" || ReviewReportStatusFailed != "failed" {
		t.Fatalf("unexpected review report statuses")
	}
}

// @Test
func TestTask01ReviewDTOJSONContractsUseExpectedFields(t *testing.T) {
	input := RejectReviewRequest{Reason: "needs stronger structure", RegenerateInstruction: "regenerate with clearer arc", TriggerRegeneration: true}
	payload, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal reject request: %v", err)
	}
	for _, want := range []string{"reason", "regenerate_instruction", "trigger_regeneration"} {
		if !strings.Contains(string(payload), want) {
			t.Fatalf("reject request JSON missing %q: %s", want, string(payload))
		}
	}

	response := ApproveWithEditResponse{ReviewID: "review-1", Status: ReviewStatusApprovedWithEdit, ContentVersionID: "version-2", OperationLogID: "oplog-1"}
	payload, err = json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal approve-with-edit response: %v", err)
	}
	for _, want := range []string{"review_id", "status", "content_version_id", "operation_log_id"} {
		if !strings.Contains(string(payload), want) {
			t.Fatalf("approve-with-edit response JSON missing %q: %s", want, string(payload))
		}
	}
}
