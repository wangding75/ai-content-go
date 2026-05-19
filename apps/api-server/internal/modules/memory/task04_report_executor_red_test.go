package memory

import (
	"context"
	"errors"
	"testing"
)

// @Test
func TestTask04ReportExecutorRejectsEmptyReportID(t *testing.T) {
	_, err := NewReportExecutor(NewService()).RunConsistencyReport(context.Background(), "")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("empty report id must return ErrNotFound, got %v", err)
	}
}

// @Test
func TestTask04ReportExecutorProducesCompletedReportWithStructuredIssues(t *testing.T) {
	detail, err := NewReportExecutor(NewService()).RunConsistencyReport(context.Background(), "report-1")
	if err != nil {
		t.Fatalf("RunConsistencyReport returned error: %v", err)
	}
	if detail.Status != string(ReportStatusCompleted) || detail.IssueCount != len(detail.Issues) {
		t.Fatalf("executor must complete report and align issue_count with issues: %+v", detail)
	}
	if detail.SourceSnapshotID == "" {
		t.Fatalf("completed report must record source snapshot id")
	}
	issue := detail.Issues[0]
	if issue.IssueID == "" || issue.Severity == "" || issue.Type == "" || issue.Title == "" || issue.Description == "" || len(issue.AffectedContentItems) == 0 || issue.Suggestion == "" {
		t.Fatalf("issue must be fully structured: %+v", issue)
	}
}

// @Test
func TestTask04ReportExecutorPersistsFailureReasonForFailedReport(t *testing.T) {
	detail, err := NewReportExecutor(NewService()).RunConsistencyReport(context.Background(), "report-fail-fixture")
	if err != nil {
		t.Fatalf("failed report state should be queryable without returning transport error: %v", err)
	}
	if detail.Status != string(ReportStatusFailed) || detail.ErrorCode == "" || detail.ErrorMessage == "" {
		t.Fatalf("executor must persist failed status with error_code and error_message: %+v", detail)
	}
}
