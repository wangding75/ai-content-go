package memory

import "context"

type ReportExecutor interface {
	RunConsistencyReport(ctx context.Context, reportID string) (ConsistencyReportDetailResponse, error)
}

type deterministicReportExecutor struct {
	service *service
}

func NewReportExecutor(svc Service) ReportExecutor {
	return &deterministicReportExecutor{service: svc.(*service)}
}

func (e *deterministicReportExecutor) RunConsistencyReport(ctx context.Context, reportID string) (ConsistencyReportDetailResponse, error) {
	if reportID == "" {
		return ConsistencyReportDetailResponse{}, ErrNotFound
	}

	rep, err := e.service.GetReportRecord(reportID)
	if err != nil {
		return ConsistencyReportDetailResponse{}, ErrNotFound
	}

	// Check if this is a failure fixture
	if rep.ErrorCode == "FIXTURE_FAILURE" {
		err := e.service.UpdateReportStatus(reportID, string(ReportStatusFailed), nil, 0, nil, "", "EXECUTOR_ERROR", "deterministic rule execution failed: fixture error")
		if err != nil {
			return ConsistencyReportDetailResponse{}, err
		}
	} else {
		// Completed path: deterministic structured issues
		issues := []ConsistencyIssue{{
			IssueID:              "issue_001",
			Severity:             "high",
			Type:                 "character_inconsistency",
			Title:                "角色设定前后不一致",
			Description:          "主角年龄在不同内容单元中不一致",
			AffectedContentItems: []string{"item_001", "item_003"},
			Suggestion:           "以最新设定为准修正内容。",
		}}
		err := e.service.UpdateReportStatus(reportID, string(ReportStatusCompleted), issues, len(issues), map[string]int{"high": 1}, "snapshot-1", "", "")
		if err != nil {
			return ConsistencyReportDetailResponse{}, err
		}
	}

	// Return updated report detail
	detail, err := e.service.GetConsistencyReport(ctx, rep.ProjectID, reportID)
	if err != nil {
		return ConsistencyReportDetailResponse{}, err
	}
	return detail, nil
}
