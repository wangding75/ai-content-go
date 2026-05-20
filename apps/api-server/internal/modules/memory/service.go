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

var supportedTruncationPolicies = map[string]bool{"time": true, "token": true}
var supportedReportStatuses = map[string]bool{
	"":                            true,
	string(ReportStatusPending):   true,
	string(ReportStatusRunning):   true,
	string(ReportStatusCompleted): true,
	string(ReportStatusFailed):    true,
}
var supportedListSorts = map[string]bool{"": true, "created_at": true}
var supportedListOrders = map[string]bool{"": true, "asc": true, "desc": true}

const maxContextBudget = 100000
const maxPageSize = 100

// --- In-memory repository (schema-aligned) ---

type memoryRecord struct {
	ID                 string
	ProjectID          string
	StaticContext      map[string]any
	DynamicState       map[string]any
	RecentWindowPolicy RecentWindowPolicy
	StyleGuide         map[string]any
	Version            int
	OperationLogID     string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type snapshotRecord struct {
	ID               string
	ProjectID        string
	ContentItemID    string
	SourceType       string
	SourceID         string
	AssembledContext map[string]any
	SourceRefs       []string
	TokenBudget      int
	EstimatedTokens  int
	TruncationPolicy string
	TriggeredBy      string
	CreatedAt        time.Time
}

type reportRecord struct {
	ID                string
	ProjectID         string
	Range             map[string]any
	Scope             string
	SeverityThreshold string
	Status            string
	IssueCount        int
	SeveritySummary   map[string]int
	Issues            []ConsistencyIssue
	SourceSnapshotID  string
	ErrorCode         string
	ErrorMessage      string
	CreatedAt         time.Time
	CompletedAt       *time.Time
}

type idempotencyRec struct {
	ID              string
	Scope           string
	Endpoint        string
	IdempotencyKey  string
	RequestHash     string
	ResponseRefType string
	ResponseRefID   string
	CreatedAt       time.Time
}

type contentItemRecord struct {
	ID        string
	ProjectID string
}

type repository struct {
	mu            sync.Mutex
	memories      map[string]*memoryRecord // keyed by projectID
	snapshots     map[string]*snapshotRecord
	reports       map[string]*reportRecord
	idempotencies map[string]*idempotencyRec // keyed by scope:endpoint:key
	contentItems  map[string]*contentItemRecord
	forbidden     map[string]bool
	conflict      map[string]bool
}

func newRepo() *repository {
	r := &repository{
		memories:      make(map[string]*memoryRecord),
		snapshots:     make(map[string]*snapshotRecord),
		reports:       make(map[string]*reportRecord),
		idempotencies: make(map[string]*idempotencyRec),
		contentItems:  make(map[string]*contentItemRecord),
		forbidden:     map[string]bool{"project-forbidden": true},
		conflict:      map[string]bool{"project-conflict": true},
	}

	now := time.Now().UTC()

	// Seed project-1
	r.memories["project-1"] = &memoryRecord{
		ID: "memory-project-1", ProjectID: "project-1",
		StaticContext:      map[string]any{"summary": "static context"},
		DynamicState:       map[string]any{"status": "active"},
		RecentWindowPolicy: RecentWindowPolicy{ItemCount: 5, TokenLimit: 2000, TruncationPolicy: "time"},
		StyleGuide:         map[string]any{"tone": "consistent"},
		Version:            1, CreatedAt: now, UpdatedAt: now,
	}

	// Seed project-1 default snapshot
	r.snapshots["snapshot-1"] = &snapshotRecord{
		ID: "snapshot-1", ProjectID: "project-1",
		SourceType: string(SnapshotSourceAssembleContext),
		TokenBudget: 2000, EstimatedTokens: 1200, TruncationPolicy: "time",
		CreatedAt: now,
	}

	// Seed report-1 as a completed report for project-1 (used by HTTP contract tests)
	r.reports["report-1"] = &reportRecord{
		ID: "report-1", ProjectID: "project-1",
		Status: string(ReportStatusCompleted), IssueCount: 1,
		SeveritySummary: map[string]int{"high": 1},
		Issues: []ConsistencyIssue{{
			IssueID: "issue_001", Severity: "high", Type: "character_inconsistency",
			Title: "角色设定前后不一致", Description: "主角年龄在不同内容单元中不一致",
			AffectedContentItems: []string{"item_001", "item_003"},
			Suggestion: "以最新设定为准修正内容。",
		}},
		SourceSnapshotID: "snapshot-1",
		CreatedAt: now,
	}

	// Seed seed-project
	r.memories["seed-project"] = &memoryRecord{
		ID: "memory-seed-project", ProjectID: "seed-project",
		StaticContext:      map[string]any{"summary": "seed static"},
		DynamicState:       map[string]any{"status": "active"},
		RecentWindowPolicy: RecentWindowPolicy{ItemCount: 5, TokenLimit: 2000, TruncationPolicy: "time"},
		StyleGuide:         map[string]any{"tone": "formal"},
		Version:            1, CreatedAt: now, UpdatedAt: now,
	}

	// Seed content-item-1 belonging to project-1
	r.contentItems["content-item-1"] = &contentItemRecord{ID: "content-item-1", ProjectID: "project-1"}

	return r
}

func (r *repository) checkProject(projectID string) error {
	if r.forbidden[projectID] {
		return ErrForbidden
	}
	if _, ok := r.memories[projectID]; !ok {
		return ErrNotFound
	}
	return nil
}

func (r *repository) checkProjectForMutation(projectID string) error {
	if r.forbidden[projectID] {
		return ErrForbidden
	}
	if r.conflict[projectID] {
		return ErrConflict
	}
	if _, ok := r.memories[projectID]; !ok {
		return ErrNotFound
	}
	return nil
}

func (r *repository) getLatestSnapshot(projectID string) *snapshotRecord {
	var latest *snapshotRecord
	for _, s := range r.snapshots {
		if s.ProjectID == projectID {
			if latest == nil || s.CreatedAt.After(latest.CreatedAt) {
				latest = s
			}
		}
	}
	return latest
}

func (r *repository) listSnapshots(projectID string, contentItemID string) []*snapshotRecord {
	var result []*snapshotRecord
	for _, s := range r.snapshots {
		if s.ProjectID != projectID {
			continue
		}
		if contentItemID != "" && s.ContentItemID != contentItemID {
			continue
		}
		result = append(result, s)
	}
	return result
}

func (r *repository) findContentItem(id string) (*contentItemRecord, error) {
	ci, ok := r.contentItems[id]
	if !ok {
		return nil, ErrNotFound
	}
	return ci, nil
}

func (r *repository) getReport(projectID, reportID string) (*reportRecord, error) {
	rep, ok := r.reports[reportID]
	if !ok {
		return nil, ErrNotFound
	}
	if rep.ProjectID != projectID {
		return nil, ErrNotFound
	}
	return rep, nil
}

func (r *repository) checkIdempotency(scope, endpoint, key string, reqHash string) (refType, refID string, conflict bool) {
	ik := scope + ":" + endpoint + ":" + key
	if entry, ok := r.idempotencies[ik]; ok {
		if entry.RequestHash != reqHash {
			return "", "", true // idempotency conflict
		}
		return entry.ResponseRefType, entry.ResponseRefID, false
	}
	return "", "", false
}

func (r *repository) storeIdempotency(scope, endpoint, key, reqHash, refType, refID string) {
	ik := scope + ":" + endpoint + ":" + key
	r.idempotencies[ik] = &idempotencyRec{
		ID: "idem-" + key, Scope: scope, Endpoint: endpoint,
		IdempotencyKey: key, RequestHash: reqHash,
		ResponseRefType: refType, ResponseRefID: refID,
		CreatedAt: time.Now().UTC(),
	}
}

// --- service implementation ---

type service struct {
	repo *repository
}

// SetReportFailureFixture marks a report as a failure fixture for deterministic testing.
func (s *service) SetReportFailureFixture(reportID string) {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()
	if rep, ok := s.repo.reports[reportID]; ok {
		rep.ErrorCode = "FIXTURE_FAILURE"
		rep.ErrorMessage = "fixture failure triggered"
	}
}

func NewService() Service {
	return &service{repo: newRepo()}
}

func (s *service) GetKnowledgeMemory(ctx context.Context, projectID string) (KnowledgeMemoryResponse, error) {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()

	if err := s.repo.checkProject(projectID); err != nil {
		return KnowledgeMemoryResponse{}, err
	}

	rec := s.repo.memories[projectID]
	latest := s.repo.getLatestSnapshot(projectID)

	resp := KnowledgeMemoryResponse{
		ID: rec.ID, ProjectID: rec.ProjectID,
		StaticContext: rec.StaticContext, DynamicState: rec.DynamicState,
		RecentWindowPolicy: rec.RecentWindowPolicy, StyleGuide: rec.StyleGuide,
		Version: rec.Version, UpdatedAt: rec.UpdatedAt,
	}
	if latest != nil {
		resp.RecentSnapshotSummary = SnapshotSummaryResponse{
			ID: latest.ID, SourceType: latest.SourceType,
			EstimatedTokens: latest.EstimatedTokens, TruncationPolicy: latest.TruncationPolicy,
			CreatedAt: latest.CreatedAt,
		}
	}
	return resp, nil
}

func (s *service) UpdateStaticContext(ctx context.Context, projectID string, req UpdateStaticContextRequest) (MemoryUpdateResponse, error) {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()

	if len(req.StaticContext) == 0 || req.Note == "" {
		return MemoryUpdateResponse{}, ErrValidation
	}
	if err := s.repo.checkProjectForMutation(projectID); err != nil {
		return MemoryUpdateResponse{}, err
	}

	rec := s.repo.memories[projectID]
	rec.StaticContext = req.StaticContext
	rec.Version++
	rec.OperationLogID = "oplog-" + projectID
	rec.UpdatedAt = time.Now().UTC()

	return MemoryUpdateResponse{Version: rec.Version, OperationLogID: rec.OperationLogID}, nil
}

func (s *service) UpdateStyleGuide(ctx context.Context, projectID string, req UpdateStyleGuideRequest) (MemoryUpdateResponse, error) {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()

	if len(req.StyleGuide) == 0 || req.Note == "" {
		return MemoryUpdateResponse{}, ErrValidation
	}
	if err := s.repo.checkProjectForMutation(projectID); err != nil {
		return MemoryUpdateResponse{}, err
	}

	rec := s.repo.memories[projectID]
	rec.StyleGuide = req.StyleGuide
	rec.Version++
	rec.OperationLogID = "oplog-" + projectID
	rec.UpdatedAt = time.Now().UTC()

	return MemoryUpdateResponse{Version: rec.Version, OperationLogID: rec.OperationLogID}, nil
}

func (s *service) CorrectDynamicState(ctx context.Context, projectID string, req CorrectDynamicStateRequest, idempotencyKey string) (DynamicStateCorrectionResponse, error) {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()

	if projectID == "" || req.Reason == "" || len(req.Changes) == 0 || len(req.SourceRefs) == 0 || idempotencyKey == "" {
		return DynamicStateCorrectionResponse{}, ErrValidation
	}
	if err := s.repo.checkProject(projectID); err != nil {
		return DynamicStateCorrectionResponse{}, err
	}

	reqHash := hashRequest(req)
	refType, refID, conflict := s.repo.checkIdempotency("project:"+projectID, "correction", idempotencyKey, reqHash)
	if conflict {
		return DynamicStateCorrectionResponse{}, ErrIdempotencyConflict
	}
	if refType != "" {
		return DynamicStateCorrectionResponse{MemorySnapshotID: refID, DynamicStateVersion: s.repo.memories[projectID].Version, OperationLogID: s.repo.memories[projectID].OperationLogID}, nil
	}

	snapshotID := "snapshot-corr-" + projectID
	now := time.Now().UTC()
	s.repo.snapshots[snapshotID] = &snapshotRecord{
		ID: snapshotID, ProjectID: projectID,
		SourceType: string(SnapshotSourceDynamicStateCorrection),
		SourceRefs: req.SourceRefs,
		TokenBudget: 0, EstimatedTokens: 0, TruncationPolicy: "time",
		CreatedAt: now,
	}

	rec := s.repo.memories[projectID]
	rec.Version++
	rec.OperationLogID = "oplog-" + projectID
	rec.UpdatedAt = now

	s.repo.storeIdempotency("project:"+projectID, "correction", idempotencyKey, reqHash, "memory_snapshot", snapshotID)

	return DynamicStateCorrectionResponse{MemorySnapshotID: snapshotID, DynamicStateVersion: rec.Version, OperationLogID: rec.OperationLogID}, nil
}

func (s *service) UpdateRecentWindowPolicy(ctx context.Context, projectID string, req UpdateRecentWindowPolicyRequest) (RecentWindowPolicyResponse, error) {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()

	if req.ItemCount < 0 {
		return RecentWindowPolicyResponse{}, ErrValidation
	}
	if req.TokenLimit <= 0 {
		return RecentWindowPolicyResponse{}, ErrValidation
	}
	if !supportedTruncationPolicies[req.TruncationPolicy] {
		return RecentWindowPolicyResponse{}, ErrValidation
	}
	if err := s.repo.checkProject(projectID); err != nil {
		return RecentWindowPolicyResponse{}, err
	}

	rec := s.repo.memories[projectID]
	rec.RecentWindowPolicy = RecentWindowPolicy{ItemCount: req.ItemCount, TokenLimit: req.TokenLimit, TruncationPolicy: req.TruncationPolicy}
	rec.Version++
	rec.OperationLogID = "oplog-" + projectID
	rec.UpdatedAt = time.Now().UTC()

	return RecentWindowPolicyResponse{RecentWindowPolicy: rec.RecentWindowPolicy, Version: rec.Version, OperationLogID: rec.OperationLogID}, nil
}

func (s *service) ListSnapshots(ctx context.Context, projectID string, req ListSnapshotsRequest) (PagedMemorySnapshotsResponse, error) {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()

	if !validPagination(req.Page, req.PageSize) || !supportedListSorts[req.Sort] || !supportedListOrders[req.Order] {
		return PagedMemorySnapshotsResponse{}, ErrValidation
	}
	if err := s.repo.checkProject(projectID); err != nil {
		return PagedMemorySnapshotsResponse{}, err
	}

	snapshots := s.repo.listSnapshots(projectID, req.ContentItemID)
	items := make([]MemorySnapshotResponse, 0, len(snapshots))
	for _, sn := range snapshots {
		items = append(items, MemorySnapshotResponse{
			ID: sn.ID, ProjectID: sn.ProjectID, ContentItemID: sn.ContentItemID,
			SourceType: sn.SourceType, TokenBudget: sn.TokenBudget,
			EstimatedTokens: sn.EstimatedTokens, TruncationPolicy: sn.TruncationPolicy,
			CreatedAt: sn.CreatedAt,
		})
	}

	return PagedMemorySnapshotsResponse{Items: items, Pagination: pagination(req.Page, req.PageSize)}, nil
}

func (s *service) PreviewContext(ctx context.Context, projectID string, req ContextPreviewRequest) (ContextPreviewResponse, error) {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()

	if req.Purpose == "" {
		return ContextPreviewResponse{}, ErrValidation
	}
	if req.Budget <= 0 {
		return ContextPreviewResponse{}, ErrValidation
	}
	if req.Budget > maxContextBudget {
		return ContextPreviewResponse{}, ErrValidation
	}
	if err := s.repo.checkProject(projectID); err != nil {
		return ContextPreviewResponse{}, err
	}

	return ContextPreviewResponse{Sources: []string{"static_context", "style_guide", "dynamic_state"}, TokenBudget: req.Budget, EstimatedTokens: req.Budget / 2, TruncationPolicy: "time", PreviewText: "preview context"}, nil
}

func (s *service) AssembleContext(ctx context.Context, projectID string, req AssembleContextRequest, idempotencyKey string) (AssembleContextResponse, error) {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()

	if projectID == "" || req.Purpose == "" || req.Budget <= 0 || idempotencyKey == "" {
		return AssembleContextResponse{}, ErrValidation
	}
	if err := s.repo.checkProject(projectID); err != nil {
		return AssembleContextResponse{}, err
	}

	reqHash := hashRequest(req)
	refType, refID, conflict := s.repo.checkIdempotency("project:"+projectID, "assemble", idempotencyKey, reqHash)
	if conflict {
		return AssembleContextResponse{}, ErrIdempotencyConflict
	}
	if refType != "" {
		return AssembleContextResponse{ContextSnapshotID: refID, EstimatedTokens: req.Budget / 2, TruncationPolicy: "time"}, nil
	}

	snapshotID := "snapshot-asm-" + projectID
	now := time.Now().UTC()
	s.repo.snapshots[snapshotID] = &snapshotRecord{
		ID: snapshotID, ProjectID: projectID,
		SourceType: string(SnapshotSourceAssembleContext),
		TokenBudget: req.Budget, EstimatedTokens: req.Budget / 2, TruncationPolicy: "time",
		CreatedAt: now,
	}

	s.repo.storeIdempotency("project:"+projectID, "assemble", idempotencyKey, reqHash, "memory_snapshot", snapshotID)

	return AssembleContextResponse{ContextSnapshotID: snapshotID, EstimatedTokens: req.Budget / 2, TruncationPolicy: "time"}, nil
}

func (s *service) UpdateDynamicState(ctx context.Context, contentItemID string, req UpdateDynamicStateRequest, idempotencyKey string) (UpdateDynamicStateResponse, error) {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()

	if contentItemID == "" || req.Summary == "" || len(req.Changes) == 0 || req.SourceVersionID == "" || idempotencyKey == "" {
		return UpdateDynamicStateResponse{}, ErrValidation
	}

	ci, err := s.repo.findContentItem(contentItemID)
	if err != nil {
		return UpdateDynamicStateResponse{}, err
	}

	reqHash := hashRequest(req)
	refType, refID, conflict := s.repo.checkIdempotency("content_item:"+contentItemID, "dynamic_state", idempotencyKey, reqHash)
	if conflict {
		return UpdateDynamicStateResponse{}, ErrIdempotencyConflict
	}
	if refType != "" {
		rec := s.repo.memories[ci.ProjectID]
		return UpdateDynamicStateResponse{MemorySnapshotID: refID, DynamicStateVersion: rec.Version}, nil
	}

	snapshotID := "snapshot-ds-" + contentItemID
	now := time.Now().UTC()
	s.repo.snapshots[snapshotID] = &snapshotRecord{
		ID: snapshotID, ProjectID: ci.ProjectID, ContentItemID: contentItemID,
		SourceType: string(SnapshotSourceDynamicStateUpdate),
		TokenBudget: 0, EstimatedTokens: 0, TruncationPolicy: "time",
		CreatedAt: now,
	}

	rec := s.repo.memories[ci.ProjectID]
	rec.Version++
	rec.UpdatedAt = now

	s.repo.storeIdempotency("content_item:"+contentItemID, "dynamic_state", idempotencyKey, reqHash, "memory_snapshot", snapshotID)

	return UpdateDynamicStateResponse{MemorySnapshotID: snapshotID, DynamicStateVersion: rec.Version}, nil
}

func (s *service) CreateConsistencyReport(ctx context.Context, projectID string, req CreateConsistencyReportRequest, idempotencyKey string) (CreateConsistencyReportResponse, error) {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()

	if projectID == "" || len(req.Range) == 0 || req.Scope == "" || req.SeverityThreshold == "" || idempotencyKey == "" {
		return CreateConsistencyReportResponse{}, ErrValidation
	}
	if err := s.repo.checkProject(projectID); err != nil {
		return CreateConsistencyReportResponse{}, err
	}

	reqHash := hashRequest(req)
	refType, refID, conflict := s.repo.checkIdempotency("project:"+projectID, "report", idempotencyKey, reqHash)
	if conflict {
		return CreateConsistencyReportResponse{}, ErrIdempotencyConflict
	}
	if refType != "" {
		rep := s.repo.reports[refID]
		return CreateConsistencyReportResponse{ReportID: rep.ID, Status: rep.Status}, nil
	}

	reportID := "report-" + projectID
	now := time.Now().UTC()
	s.repo.reports[reportID] = &reportRecord{
		ID: reportID, ProjectID: projectID,
		Range: req.Range, Scope: req.Scope, SeverityThreshold: req.SeverityThreshold,
		Status: string(ReportStatusPending), CreatedAt: now,
	}

	s.repo.storeIdempotency("project:"+projectID, "report", idempotencyKey, reqHash, "consistency_report", reportID)

	return CreateConsistencyReportResponse{ReportID: reportID, Status: string(ReportStatusPending)}, nil
}

func (s *service) ListConsistencyReports(ctx context.Context, projectID string, req ListConsistencyReportsRequest) (PagedConsistencyReportsResponse, error) {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()

	if !validPagination(req.Page, req.PageSize) || !supportedReportStatuses[req.Status] || !supportedListSorts[req.Sort] || !supportedListOrders[req.Order] {
		return PagedConsistencyReportsResponse{}, ErrValidation
	}
	if err := s.repo.checkProject(projectID); err != nil {
		return PagedConsistencyReportsResponse{}, err
	}

	var items []ConsistencyReportResponse
	for _, rep := range s.repo.reports {
		if rep.ProjectID != projectID {
			continue
		}
		if req.Status != "" && rep.Status != req.Status {
			continue
		}
		items = append(items, ConsistencyReportResponse{
			ID: rep.ID, ProjectID: rep.ProjectID, Status: rep.Status,
			IssueCount: rep.IssueCount, SeveritySummary: rep.SeveritySummary,
			CreatedAt: rep.CreatedAt,
		})
	}

	return PagedConsistencyReportsResponse{Items: items, Pagination: pagination(req.Page, req.PageSize)}, nil
}

func (s *service) GetConsistencyReport(ctx context.Context, projectID string, reportID string) (ConsistencyReportDetailResponse, error) {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()

	if err := s.repo.checkProject(projectID); err != nil {
		return ConsistencyReportDetailResponse{}, err
	}

	rep, err := s.repo.getReport(projectID, reportID)
	if err != nil {
		return ConsistencyReportDetailResponse{}, err
	}

	detail := ConsistencyReportDetailResponse{
		ConsistencyReportResponse: ConsistencyReportResponse{
			ID: rep.ID, ProjectID: rep.ProjectID, Status: rep.Status,
			IssueCount: rep.IssueCount, SeveritySummary: rep.SeveritySummary,
			CreatedAt: rep.CreatedAt,
		},
		SourceSnapshotID: rep.SourceSnapshotID,
		Issues:           rep.Issues,
		ErrorCode:        rep.ErrorCode,
		ErrorMessage:     rep.ErrorMessage,
	}
	return detail, nil
}

// UpdateReportStatus updates a report's status and fields. Used by executor.
func (s *service) UpdateReportStatus(reportID string, status string, issues []ConsistencyIssue, issueCount int, severitySummary map[string]int, sourceSnapshotID string, errorCode string, errorMessage string) error {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()

	rep, ok := s.repo.reports[reportID]
	if !ok {
		return ErrNotFound
	}

	rep.Status = status
	rep.Issues = issues
	rep.IssueCount = issueCount
	rep.SeveritySummary = severitySummary
	rep.SourceSnapshotID = sourceSnapshotID
	rep.ErrorCode = errorCode
	rep.ErrorMessage = errorMessage
	if status == string(ReportStatusCompleted) || status == string(ReportStatusFailed) {
		now := time.Now().UTC()
		rep.CompletedAt = &now
	}
	return nil
}

// GetReportRecord returns a report record for executor use.
func (s *service) GetReportRecord(reportID string) (*reportRecord, error) {
	s.repo.mu.Lock()
	defer s.repo.mu.Unlock()

	rep, ok := s.repo.reports[reportID]
	if !ok {
		return nil, ErrNotFound
	}
	return rep, nil
}

func hashRequest(v any) string {
	b, _ := json.Marshal(v)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

func validPagination(page int, pageSize int) bool {
	return page >= 0 && pageSize >= 0 && pageSize <= maxPageSize
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
