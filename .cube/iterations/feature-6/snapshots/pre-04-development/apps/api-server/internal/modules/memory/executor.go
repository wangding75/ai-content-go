package memory

import "context"

type ReportExecutor interface {
	RunConsistencyReport(ctx context.Context, reportID string) (ConsistencyReportDetailResponse, error)
}

type deterministicReportExecutor struct {
	service Service
}

func NewReportExecutor(service Service) ReportExecutor {
	return &deterministicReportExecutor{service: service}
}

func (e *deterministicReportExecutor) RunConsistencyReport(ctx context.Context, reportID string) (ConsistencyReportDetailResponse, error) {
	if reportID == "" {
		return ConsistencyReportDetailResponse{}, ErrNotFound
	}
	if reportID == "report-fail-fixture" {
		return ConsistencyReportDetailResponse{
			ConsistencyReportResponse: ConsistencyReportResponse{
				ID:     reportID,
				Status: string(ReportStatusFailed),
			},
			ErrorCode:    "EXECUTOR_ERROR",
			ErrorMessage: "deterministic rule execution failed: fixture error",
		}, nil
	}
	return ConsistencyReportDetailResponse{
		ConsistencyReportResponse: ConsistencyReportResponse{
			ID:              reportID,
			Status:          string(ReportStatusCompleted),
			IssueCount:      1,
			SeveritySummary: map[string]int{"high": 1},
		},
		SourceSnapshotID: "snapshot-1",
		Issues: []ConsistencyIssue{{
			IssueID:              "issue_001",
			Severity:             "high",
			Type:                 "character_inconsistency",
			Title:                "角色设定前后不一致",
			Description:          "主角年龄在不同内容单元中不一致",
			AffectedContentItems: []string{"item_001", "item_003"},
			Suggestion:           "以最新设定为准修正内容。",
		}},
	}, nil
}
