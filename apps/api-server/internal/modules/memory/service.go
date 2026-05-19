package memory

import (
	"context"
	"time"
)

type Service interface {
	GetKnowledgeMemory(ctx context.Context, projectID string) (KnowledgeMemoryResponse, error)
	UpdateStaticContext(ctx context.Context, projectID string, req UpdateStaticContextRequest) (MemoryUpdateResponse, error)
	UpdateStyleGuide(ctx context.Context, projectID string, req UpdateStyleGuideRequest) (MemoryUpdateResponse, error)
	CorrectDynamicState(ctx context.Context, projectID string, req CorrectDynamicStateRequest, idempotencyKey string) (DynamicStateCorrectionResponse, error)
	UpdateRecentWindowPolicy(ctx context.Context, projectID string, req UpdateRecentWindowPolicyRequest) (RecentWindowPolicyResponse, error)
	ListSnapshots(ctx context.Context, projectID string, req ListSnapshotsRequest) (PagedMemorySnapshotsResponse, error)
	PreviewContext(ctx context.Context, projectID string, req ContextPreviewRequest) (ContextPreviewResponse, error)
	AssembleContext(ctx context.Context, projectID string, req AssembleContextRequest, idempotencyKey string) (AssembleContextResponse, error)
	UpdateDynamicState(ctx context.Context, contentItemID string, req UpdateDynamicStateRequest, idempotencyKey string) (UpdateDynamicStateResponse, error)
	CreateConsistencyReport(ctx context.Context, projectID string, req CreateConsistencyReportRequest, idempotencyKey string) (CreateConsistencyReportResponse, error)
	ListConsistencyReports(ctx context.Context, projectID string, req ListConsistencyReportsRequest) (PagedConsistencyReportsResponse, error)
	GetConsistencyReport(ctx context.Context, projectID string, reportID string) (ConsistencyReportDetailResponse, error)
}

type service struct{}

func NewService() Service { return &service{} }

func (s *service) GetKnowledgeMemory(ctx context.Context, projectID string) (KnowledgeMemoryResponse, error) {
	if projectID == "" {
		return KnowledgeMemoryResponse{}, ErrNotFound
	}
	now := time.Now().UTC()
	return KnowledgeMemoryResponse{ID: "memory-" + projectID, ProjectID: projectID, StaticContext: map[string]any{"summary": "static context"}, DynamicState: map[string]any{"status": "active"}, RecentWindowPolicy: RecentWindowPolicy{ItemCount: 5, TokenLimit: 2000, TruncationPolicy: "time"}, StyleGuide: map[string]any{"tone": "consistent"}, Version: 1, UpdatedAt: now, RecentSnapshotSummary: SnapshotSummaryResponse{ID: "snapshot-1", SourceType: string(SnapshotSourceAssembleContext), EstimatedTokens: 1200, TruncationPolicy: "time", CreatedAt: now}}, nil
}

func (s *service) UpdateStaticContext(ctx context.Context, projectID string, req UpdateStaticContextRequest) (MemoryUpdateResponse, error) {
	if projectID == "" || len(req.StaticContext) == 0 || req.Note == "" {
		return MemoryUpdateResponse{}, ErrValidation
	}
	return MemoryUpdateResponse{Version: 2, OperationLogID: "oplog-" + projectID}, nil
}

func (s *service) UpdateStyleGuide(ctx context.Context, projectID string, req UpdateStyleGuideRequest) (MemoryUpdateResponse, error) {
	if projectID == "" || len(req.StyleGuide) == 0 || req.Note == "" {
		return MemoryUpdateResponse{}, ErrValidation
	}
	return MemoryUpdateResponse{Version: 2, OperationLogID: "oplog-" + projectID}, nil
}

func (s *service) CorrectDynamicState(ctx context.Context, projectID string, req CorrectDynamicStateRequest, idempotencyKey string) (DynamicStateCorrectionResponse, error) {
	if projectID == "" || req.Reason == "" || len(req.Changes) == 0 || len(req.SourceRefs) == 0 || idempotencyKey == "" {
		return DynamicStateCorrectionResponse{}, ErrValidation
	}
	return DynamicStateCorrectionResponse{MemorySnapshotID: "snapshot-" + projectID, DynamicStateVersion: 2, OperationLogID: "oplog-" + projectID}, nil
}

func (s *service) UpdateRecentWindowPolicy(ctx context.Context, projectID string, req UpdateRecentWindowPolicyRequest) (RecentWindowPolicyResponse, error) {
	if projectID == "" || req.ItemCount <= 0 || req.TokenLimit <= 0 || req.TruncationPolicy == "" {
		return RecentWindowPolicyResponse{}, ErrValidation
	}
	return RecentWindowPolicyResponse{RecentWindowPolicy: RecentWindowPolicy{ItemCount: req.ItemCount, TokenLimit: req.TokenLimit, TruncationPolicy: req.TruncationPolicy}, Version: 2, OperationLogID: "oplog-" + projectID}, nil
}

func (s *service) ListSnapshots(ctx context.Context, projectID string, req ListSnapshotsRequest) (PagedMemorySnapshotsResponse, error) {
	if projectID == "" {
		return PagedMemorySnapshotsResponse{}, ErrValidation
	}
	return PagedMemorySnapshotsResponse{Items: []MemorySnapshotResponse{{ID: "snapshot-1", ProjectID: projectID, SourceType: string(SnapshotSourceAssembleContext), TokenBudget: 2000, EstimatedTokens: 1200, TruncationPolicy: "time", CreatedAt: time.Now().UTC()}}, Pagination: pagination(req.Page, req.PageSize)}, nil
}

func (s *service) PreviewContext(ctx context.Context, projectID string, req ContextPreviewRequest) (ContextPreviewResponse, error) {
	if projectID == "" || req.Purpose == "" || req.Budget <= 0 {
		return ContextPreviewResponse{}, ErrValidation
	}
	return ContextPreviewResponse{Sources: []string{"static_context", "style_guide", "dynamic_state"}, TokenBudget: req.Budget, EstimatedTokens: req.Budget / 2, TruncationPolicy: "time", PreviewText: "preview context"}, nil
}

func (s *service) AssembleContext(ctx context.Context, projectID string, req AssembleContextRequest, idempotencyKey string) (AssembleContextResponse, error) {
	if projectID == "" || req.Purpose == "" || req.Budget <= 0 || idempotencyKey == "" {
		return AssembleContextResponse{}, ErrValidation
	}
	return AssembleContextResponse{ContextSnapshotID: "snapshot-" + projectID, EstimatedTokens: req.Budget / 2, TruncationPolicy: "time"}, nil
}

func (s *service) UpdateDynamicState(ctx context.Context, contentItemID string, req UpdateDynamicStateRequest, idempotencyKey string) (UpdateDynamicStateResponse, error) {
	if contentItemID == "" || req.Summary == "" || len(req.Changes) == 0 || req.SourceVersionID == "" || idempotencyKey == "" {
		return UpdateDynamicStateResponse{}, ErrValidation
	}
	return UpdateDynamicStateResponse{MemorySnapshotID: "snapshot-" + contentItemID, DynamicStateVersion: 2}, nil
}

func (s *service) CreateConsistencyReport(ctx context.Context, projectID string, req CreateConsistencyReportRequest, idempotencyKey string) (CreateConsistencyReportResponse, error) {
	if projectID == "" || len(req.Range) == 0 || req.Scope == "" || req.SeverityThreshold == "" || idempotencyKey == "" {
		return CreateConsistencyReportResponse{}, ErrValidation
	}
	return CreateConsistencyReportResponse{ReportID: "report-" + projectID, Status: string(ReportStatusPending)}, nil
}

func (s *service) ListConsistencyReports(ctx context.Context, projectID string, req ListConsistencyReportsRequest) (PagedConsistencyReportsResponse, error) {
	if projectID == "" {
		return PagedConsistencyReportsResponse{}, ErrValidation
	}
	return PagedConsistencyReportsResponse{Items: []ConsistencyReportResponse{{ID: "report-1", ProjectID: projectID, Status: string(ReportStatusCompleted), IssueCount: 1, SeveritySummary: map[string]int{"high": 1}, CreatedAt: time.Now().UTC()}}, Pagination: pagination(req.Page, req.PageSize)}, nil
}

func (s *service) GetConsistencyReport(ctx context.Context, projectID string, reportID string) (ConsistencyReportDetailResponse, error) {
	if projectID == "" || reportID == "" {
		return ConsistencyReportDetailResponse{}, ErrNotFound
	}
	return ConsistencyReportDetailResponse{ConsistencyReportResponse: ConsistencyReportResponse{ID: reportID, ProjectID: projectID, Status: string(ReportStatusCompleted), IssueCount: 1, SeveritySummary: map[string]int{"high": 1}, CreatedAt: time.Now().UTC()}, SourceSnapshotID: "snapshot-1", Issues: []ConsistencyIssue{{IssueID: "issue_001", Severity: "high", Type: "character_inconsistency", Title: "角色设定前后不一致", Description: "主角年龄在不同内容单元中不一致", AffectedContentItems: []string{"item_001", "item_003"}, Suggestion: "以最新设定为准修正内容。"}}}, nil
}

func pagination(page int, pageSize int) PaginationResponse {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}
	return PaginationResponse{Page: page, PageSize: pageSize, Total: 1, HasNext: false}
}
