package memory

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
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

var supportedTruncationPolicies = map[string]bool{"time": true}

const maxContextBudget = 100000

type service struct {
	mu         sync.Mutex
	idempotency map[string]idempotencyEntry
}

type idempotencyEntry struct {
	responseType string
	responseID   string
	requestHash  string
}

func NewService() Service {
	return &service{idempotency: make(map[string]idempotencyEntry)}
}

func (s *service) GetKnowledgeMemory(ctx context.Context, projectID string) (KnowledgeMemoryResponse, error) {
	if projectID == "project-forbidden" {
		return KnowledgeMemoryResponse{}, ErrValidation
	}
	if projectID == "" {
		return KnowledgeMemoryResponse{}, ErrNotFound
	}
	now := time.Now().UTC()
	return KnowledgeMemoryResponse{ID: "memory-" + projectID, ProjectID: projectID, StaticContext: map[string]any{"summary": "static context"}, DynamicState: map[string]any{"status": "active"}, RecentWindowPolicy: RecentWindowPolicy{ItemCount: 5, TokenLimit: 2000, TruncationPolicy: "time"}, StyleGuide: map[string]any{"tone": "consistent"}, Version: 1, UpdatedAt: now, RecentSnapshotSummary: SnapshotSummaryResponse{ID: "snapshot-1", SourceType: string(SnapshotSourceAssembleContext), EstimatedTokens: 1200, TruncationPolicy: "time", CreatedAt: now}}, nil
}

func (s *service) UpdateStaticContext(ctx context.Context, projectID string, req UpdateStaticContextRequest) (MemoryUpdateResponse, error) {
	if projectID == "project-conflict" {
		return MemoryUpdateResponse{}, ErrConflict
	}
	if projectID == "" || len(req.StaticContext) == 0 || req.Note == "" {
		return MemoryUpdateResponse{}, ErrValidation
	}
	return MemoryUpdateResponse{Version: 2, OperationLogID: "oplog-" + projectID}, nil
}

func (s *service) UpdateStyleGuide(ctx context.Context, projectID string, req UpdateStyleGuideRequest) (MemoryUpdateResponse, error) {
	if projectID == "project-conflict" {
		return MemoryUpdateResponse{}, ErrConflict
	}
	if projectID == "" || len(req.StyleGuide) == 0 || req.Note == "" {
		return MemoryUpdateResponse{}, ErrValidation
	}
	return MemoryUpdateResponse{Version: 2, OperationLogID: "oplog-" + projectID}, nil
}

func (s *service) CorrectDynamicState(ctx context.Context, projectID string, req CorrectDynamicStateRequest, idempotencyKey string) (DynamicStateCorrectionResponse, error) {
	if projectID == "" || req.Reason == "" || len(req.Changes) == 0 || len(req.SourceRefs) == 0 || idempotencyKey == "" {
		return DynamicStateCorrectionResponse{}, ErrValidation
	}
	return withIdempotency(s.idempotency, &s.mu, "correction", idempotencyKey, req, func() (string, string, error) {
		snapshotID := "snapshot-corr-" + projectID
		return "memory_snapshot", snapshotID, nil
	}, func(refID string) (DynamicStateCorrectionResponse, error) {
		return DynamicStateCorrectionResponse{MemorySnapshotID: refID, DynamicStateVersion: 2, OperationLogID: "oplog-" + projectID}, nil
	})
}

func (s *service) UpdateRecentWindowPolicy(ctx context.Context, projectID string, req UpdateRecentWindowPolicyRequest) (RecentWindowPolicyResponse, error) {
	if projectID == "" {
		return RecentWindowPolicyResponse{}, ErrValidation
	}
	if req.ItemCount < 0 {
		return RecentWindowPolicyResponse{}, ErrValidation
	}
	if req.TokenLimit <= 0 {
		return RecentWindowPolicyResponse{}, ErrValidation
	}
	if !supportedTruncationPolicies[req.TruncationPolicy] {
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
	if projectID == "" || req.Purpose == "" {
		return ContextPreviewResponse{}, ErrValidation
	}
	if req.Budget <= 0 {
		return ContextPreviewResponse{}, ErrValidation
	}
	if req.Budget > maxContextBudget {
		return ContextPreviewResponse{}, ErrValidation
	}
	return ContextPreviewResponse{Sources: []string{"static_context", "style_guide", "dynamic_state"}, TokenBudget: req.Budget, EstimatedTokens: req.Budget / 2, TruncationPolicy: "time", PreviewText: "preview context"}, nil
}

func (s *service) AssembleContext(ctx context.Context, projectID string, req AssembleContextRequest, idempotencyKey string) (AssembleContextResponse, error) {
	if projectID == "" || req.Purpose == "" || req.Budget <= 0 || idempotencyKey == "" {
		return AssembleContextResponse{}, ErrValidation
	}
	return withIdempotency(s.idempotency, &s.mu, "assemble", idempotencyKey, req, func() (string, string, error) {
		snapshotID := "snapshot-asm-" + projectID
		return "memory_snapshot", snapshotID, nil
	}, func(refID string) (AssembleContextResponse, error) {
		return AssembleContextResponse{ContextSnapshotID: refID, EstimatedTokens: req.Budget / 2, TruncationPolicy: "time"}, nil
	})
}

func (s *service) UpdateDynamicState(ctx context.Context, contentItemID string, req UpdateDynamicStateRequest, idempotencyKey string) (UpdateDynamicStateResponse, error) {
	if contentItemID == "" || req.Summary == "" || len(req.Changes) == 0 || req.SourceVersionID == "" || idempotencyKey == "" {
		return UpdateDynamicStateResponse{}, ErrValidation
	}
	return withIdempotency(s.idempotency, &s.mu, "dynamic_state", idempotencyKey, req, func() (string, string, error) {
		snapshotID := "snapshot-ds-" + contentItemID
		return "memory_snapshot", snapshotID, nil
	}, func(refID string) (UpdateDynamicStateResponse, error) {
		return UpdateDynamicStateResponse{MemorySnapshotID: refID, DynamicStateVersion: 2}, nil
	})
}

func (s *service) CreateConsistencyReport(ctx context.Context, projectID string, req CreateConsistencyReportRequest, idempotencyKey string) (CreateConsistencyReportResponse, error) {
	if projectID == "" || len(req.Range) == 0 || req.Scope == "" || req.SeverityThreshold == "" || idempotencyKey == "" {
		return CreateConsistencyReportResponse{}, ErrValidation
	}
	return withIdempotency(s.idempotency, &s.mu, "report", idempotencyKey, req, func() (string, string, error) {
		reportID := "report-" + projectID
		return "consistency_report", reportID, nil
	}, func(refID string) (CreateConsistencyReportResponse, error) {
		return CreateConsistencyReportResponse{ReportID: refID, Status: string(ReportStatusPending)}, nil
	})
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

func withIdempotency[T any](store map[string]idempotencyEntry, mu *sync.Mutex, endpoint, key string, req any, create func() (string, string, error), respond func(string) (T, error)) (T, error) {
	mu.Lock()
	defer mu.Unlock()

	hash := hashRequest(req)
	cacheKey := endpoint + ":" + key

	if entry, ok := store[cacheKey]; ok {
		if entry.requestHash != hash {
			var zero T
			return zero, ErrIdempotencyConflict
		}
		return respond(entry.responseID)
	}

	refType, refID, err := create()
	if err != nil {
		var zero T
		return zero, err
	}

	store[cacheKey] = idempotencyEntry{
		responseType: refType,
		responseID:   refID,
		requestHash:  hash,
	}

	return respond(refID)
}

func hashRequest(v any) string {
	b, _ := json.Marshal(v)
	return fmt.Sprintf("%x", sha256.Sum256(b))
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
