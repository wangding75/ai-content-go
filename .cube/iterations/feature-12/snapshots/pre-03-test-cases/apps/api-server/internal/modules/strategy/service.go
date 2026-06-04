package strategy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

type Service interface {
	GenerateSuggestions(ctx context.Context, projectID string, req GenerateSuggestionsRequest, idempotencyKey string) (GenerateSuggestionsResponse, error)
	ListSuggestions(ctx context.Context, projectID string, req ListStrategySuggestionsRequest) (PagedStrategySuggestionsResponse, error)
	GetSuggestion(ctx context.Context, suggestionID string) (StrategySuggestionDetailResponse, error)
	ConfirmSuggestion(ctx context.Context, suggestionID string, req ConfirmSuggestionRequest, idempotencyKey string) (SuggestionStatusChangeResponse, error)
	IgnoreSuggestion(ctx context.Context, suggestionID string, req IgnoreSuggestionRequest, idempotencyKey string) (SuggestionStatusChangeResponse, error)
	ExecuteSuggestion(ctx context.Context, suggestionID string, req ExecuteSuggestionRequest, idempotencyKey string) (ExecuteSuggestionResponse, error)
	RetrySuggestion(ctx context.Context, suggestionID string, req RetrySuggestionRequest, idempotencyKey string) (ExecuteSuggestionResponse, error)
	ListExecutionLogs(ctx context.Context, suggestionID string, req ListExecutionLogsRequest) (PagedExecutionLogsResponse, error)
}

type service struct {
	store Store
}

func NewService(stores ...Store) Service {
	var store Store
	if len(stores) > 0 {
		store = stores[0]
	} else {
		store = NewMemoryStore()
	}
	return &service{store: store}
}

func requestHash(value any) string {
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func pagination(page, pageSize, total int) content.PaginationResponse {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return content.PaginationResponse{Page: page, PageSize: pageSize, Total: total, HasNext: page*pageSize < total}
}

func pageBounds(page, pageSize, total int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return start, end
}

func (s *service) GenerateSuggestions(ctx context.Context, projectID string, req GenerateSuggestionsRequest, idempotencyKey string) (GenerateSuggestionsResponse, error) {
	if projectID == "" || req.DateFrom == "" || req.DateTo == "" {
		return GenerateSuggestionsResponse{}, ErrValidation
	}

	scope := "strategy:" + "generate:" + projectID
	hash := requestHash(req)
	refType, refID, conflict, err := s.store.CheckIdempotency(ctx, scope, "generate", idempotencyKey, hash)
	if err != nil {
		return GenerateSuggestionsResponse{}, ErrInternal
	}
	if conflict {
		return GenerateSuggestionsResponse{}, ErrIdempotencyConflict
	}
	if refType != "" {
		return GenerateSuggestionsResponse{SuggestionRunID: refID, Status: RunStatusGenerating}, nil
	}

	runID := fmt.Sprintf("strategy-run-%s-%s", projectID, req.DateTo)
	now := time.Now().UTC()
	run := StrategySuggestionRunResponse{
		ID:              runID,
		ProjectID:       projectID,
		DateFrom:        req.DateFrom,
		DateTo:          req.DateTo,
		RuleCodes:       req.RuleCodes,
		MetricCodes:     req.MetricCodes,
		ForceRegenerate: req.ForceRegenerate,
		Status:          RunStatusGenerating,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.store.InsertSuggestionRun(ctx, run); err != nil {
		return GenerateSuggestionsResponse{}, ErrInternal
	}

	s.generateRuleBasedSuggestions(ctx, projectID, runID, req)

	if err := s.store.StoreIdempotency(ctx, scope, "generate", idempotencyKey, hash, "strategy_suggestion_run", runID); err != nil {
		return GenerateSuggestionsResponse{}, ErrInternal
	}

	return GenerateSuggestionsResponse{SuggestionRunID: runID, Status: RunStatusGenerating}, nil
}

func (s *service) generateRuleBasedSuggestions(ctx context.Context, projectID, runID string, req GenerateSuggestionsRequest) {
	now := time.Now().UTC()
	suggestions := []StrategySuggestionDetailResponse{
		{
			ID:               fmt.Sprintf("strategy-suggestion-%s-declining_views-%s", projectID, req.DateTo),
			ProjectID:        projectID,
			SuggestionRunID:  runID,
			SuggestionType:   SuggestionTypeOptimize,
			Title:            "阅读量持续下降",
			TriggerReason:    "近7天阅读量下降超过30%",
			EvidenceMetrics:  []MetricEvidence{{MetricCode: "views", Value: 1200, Trend: "declining"}},
			ImpactScope:      "项目整体阅读表现",
			RiskLevel:        RiskLevelMedium,
			Confidence:       ConfidenceHigh,
			SuggestedAction:  "调整发布时间和频率",
			ExpectedBenefit:  "预计可提升阅读量15-25%",
			MetricsSnapshot:  MetricsSnapshot{SummarySnapshotID: fmt.Sprintf("metric-summary-snapshot-%s-%s-%s", projectID, req.DateFrom, req.DateTo)},
			Status:           StatusPending,
			DateFrom:         req.DateFrom,
			DateTo:           req.DateTo,
			TriggeredRules:   []string{"declining_views"},
			GenerationMethod: GenerationMethodRule,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	for _, sug := range suggestions {
		s.store.InsertSuggestion(ctx, sug)
	}

	s.store.UpdateSuggestionRunStatus(ctx, runID, RunStatusCompleted, "", len(suggestions))
}

func (s *service) ListSuggestions(ctx context.Context, projectID string, req ListStrategySuggestionsRequest) (PagedStrategySuggestionsResponse, error) {
	if projectID == "" {
		return PagedStrategySuggestionsResponse{}, ErrValidation
	}
	items, total, err := s.store.ListSuggestions(ctx, projectID, req)
	if err != nil {
		return PagedStrategySuggestionsResponse{}, ErrInternal
	}
	page := req.Page
	pageSize := req.PageSize
	start, end := pageBounds(page, pageSize, total)
	listItems := make([]StrategySuggestionItem, 0, len(items[start:end]))
	for _, item := range items[start:end] {
		listItems = append(listItems, StrategySuggestionItem{
			ID:             item.ID,
			ProjectID:      item.ProjectID,
			SuggestionType: item.SuggestionType,
			Title:          item.Title,
			TriggerReason:  item.TriggerReason,
			RiskLevel:      item.RiskLevel,
			Confidence:     item.Confidence,
			Status:         item.Status,
			DateFrom:       item.DateFrom,
			DateTo:         item.DateTo,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		})
	}
	return PagedStrategySuggestionsResponse{Items: listItems, Pagination: pagination(page, pageSize, total)}, nil
}

func (s *service) GetSuggestion(ctx context.Context, suggestionID string) (StrategySuggestionDetailResponse, error) {
	if suggestionID == "" {
		return StrategySuggestionDetailResponse{}, ErrValidation
	}
	sug, err := s.store.FindSuggestionByID(ctx, suggestionID)
	if err != nil {
		return StrategySuggestionDetailResponse{}, ErrInternal
	}
	if sug == nil {
		return StrategySuggestionDetailResponse{}, ErrNotFound
	}
	return *sug, nil
}

func (s *service) ConfirmSuggestion(ctx context.Context, suggestionID string, req ConfirmSuggestionRequest, idempotencyKey string) (SuggestionStatusChangeResponse, error) {
	sug, err := s.store.FindSuggestionByID(ctx, suggestionID)
	if err != nil {
		return SuggestionStatusChangeResponse{}, ErrInternal
	}
	if sug == nil {
		return SuggestionStatusChangeResponse{}, ErrNotFound
	}
	if sug.Status != StatusPending {
		return SuggestionStatusChangeResponse{}, ErrConflict
	}

	scope := "strategy:" + "confirm:" + suggestionID
	hash := requestHash(req)
	refType, refID, conflict, err := s.store.CheckIdempotency(ctx, scope, "confirm", idempotencyKey, hash)
	if err != nil {
		return SuggestionStatusChangeResponse{}, ErrInternal
	}
	if conflict {
		return SuggestionStatusChangeResponse{}, ErrIdempotencyConflict
	}
	if refType != "" {
		return SuggestionStatusChangeResponse{SuggestionID: refID, PreviousStatus: StatusPending, CurrentStatus: StatusConfirmed, OperationLogID: "operation-log-" + refID}, nil
	}

	previousStatus := sug.Status
	now := time.Now().UTC()
	fields := map[string]any{"confirmed_at": &now}
	s.store.UpdateSuggestionStatus(ctx, suggestionID, StatusConfirmed, fields)

	opLogID := "operation-log-" + suggestionID
	s.store.StoreIdempotency(ctx, scope, "confirm", idempotencyKey, hash, "strategy_suggestion", suggestionID)

	return SuggestionStatusChangeResponse{SuggestionID: suggestionID, PreviousStatus: previousStatus, CurrentStatus: StatusConfirmed, OperationLogID: opLogID}, nil
}

func (s *service) IgnoreSuggestion(ctx context.Context, suggestionID string, req IgnoreSuggestionRequest, idempotencyKey string) (SuggestionStatusChangeResponse, error) {
	if req.Reason == "" {
		return SuggestionStatusChangeResponse{}, ErrValidation
	}

	sug, err := s.store.FindSuggestionByID(ctx, suggestionID)
	if err != nil {
		return SuggestionStatusChangeResponse{}, ErrInternal
	}
	if sug == nil {
		return SuggestionStatusChangeResponse{}, ErrNotFound
	}
	if sug.Status != StatusPending {
		return SuggestionStatusChangeResponse{}, ErrConflict
	}

	scope := "strategy:" + "ignore:" + suggestionID
	hash := requestHash(req)
	refType, refID, conflict, err := s.store.CheckIdempotency(ctx, scope, "ignore", idempotencyKey, hash)
	if err != nil {
		return SuggestionStatusChangeResponse{}, ErrInternal
	}
	if conflict {
		return SuggestionStatusChangeResponse{}, ErrIdempotencyConflict
	}
	if refType != "" {
		return SuggestionStatusChangeResponse{SuggestionID: refID, PreviousStatus: StatusPending, CurrentStatus: StatusIgnored, OperationLogID: "operation-log-" + refID}, nil
	}

	previousStatus := sug.Status
	now := time.Now().UTC()
	fields := map[string]any{"ignored_reason": req.Reason, "ignored_note": req.Note, "ignored_at": &now}
	s.store.UpdateSuggestionStatus(ctx, suggestionID, StatusIgnored, fields)

	opLogID := "operation-log-" + suggestionID
	s.store.StoreIdempotency(ctx, scope, "ignore", idempotencyKey, hash, "strategy_suggestion", suggestionID)

	return SuggestionStatusChangeResponse{SuggestionID: suggestionID, PreviousStatus: previousStatus, CurrentStatus: StatusIgnored, OperationLogID: opLogID}, nil
}

func (s *service) ExecuteSuggestion(ctx context.Context, suggestionID string, req ExecuteSuggestionRequest, idempotencyKey string) (ExecuteSuggestionResponse, error) {
	sug, err := s.store.FindSuggestionByID(ctx, suggestionID)
	if err != nil {
		return ExecuteSuggestionResponse{}, ErrInternal
	}
	if sug == nil {
		return ExecuteSuggestionResponse{}, ErrNotFound
	}
	if sug.Status != StatusConfirmed {
		return ExecuteSuggestionResponse{}, ErrConflict
	}

	scope := "strategy:" + "execute:" + suggestionID
	hash := requestHash(req)
	refType, refID, conflict, err := s.store.CheckIdempotency(ctx, scope, "execute", idempotencyKey, hash)
	if err != nil {
		return ExecuteSuggestionResponse{}, ErrInternal
	}
	if conflict {
		return ExecuteSuggestionResponse{}, ErrIdempotencyConflict
	}
	if refType != "" {
		return ExecuteSuggestionResponse{ExecutionLogID: "execution-log-" + refID, SuggestionID: refID, PreviousStatus: StatusConfirmed, CurrentStatus: StatusExecuted, OperationLogID: "operation-log-" + refID}, nil
	}

	previousStatus := sug.Status
	now := time.Now().UTC()

	execResult := ExecutionResultSuccess
	newStatus := StatusExecuted
	failureReason := ""
	targetInterface := ""
	targetResource := ""

	if req.TargetType == "" || req.TargetID == "" {
		execResult = ExecutionResultFailed
		newStatus = StatusExecutionFailed
		failureReason = "缺少可用的目标接口或目标资源标识"
	}

	execLogID := fmt.Sprintf("execution-log-%s-%d", suggestionID, now.UnixMilli())
	execLog := ExecutionLogResponse{
		ID:              execLogID,
		SuggestionID:    suggestionID,
		ActionType:      req.ActionType,
		TargetType:      req.TargetType,
		TargetID:        req.TargetID,
		OperatorNote:    req.OperatorNote,
		PreviousStatus:  previousStatus,
		CurrentStatus:   newStatus,
		Result:          execResult,
		FailureReason:   failureReason,
		TargetInterface: targetInterface,
		TargetResource:  targetResource,
		CreatedAt:       now,
	}
	s.store.InsertExecutionLog(ctx, execLog)

	fields := map[string]any{"executed_at": &now}
	s.store.UpdateSuggestionStatus(ctx, suggestionID, newStatus, fields)

	opLogID := "operation-log-" + suggestionID
	s.store.StoreIdempotency(ctx, scope, "execute", idempotencyKey, hash, "strategy_suggestion", suggestionID)

	return ExecuteSuggestionResponse{ExecutionLogID: execLogID, SuggestionID: suggestionID, PreviousStatus: previousStatus, CurrentStatus: newStatus, OperationLogID: opLogID}, nil
}

func (s *service) RetrySuggestion(ctx context.Context, suggestionID string, req RetrySuggestionRequest, idempotencyKey string) (ExecuteSuggestionResponse, error) {
	sug, err := s.store.FindSuggestionByID(ctx, suggestionID)
	if err != nil {
		return ExecuteSuggestionResponse{}, ErrInternal
	}
	if sug == nil {
		return ExecuteSuggestionResponse{}, ErrNotFound
	}
	if sug.Status != StatusExecutionFailed {
		return ExecuteSuggestionResponse{}, ErrConflict
	}

	scope := "strategy:" + "retry:" + suggestionID
	hash := requestHash(req)
	refType, refID, conflict, err := s.store.CheckIdempotency(ctx, scope, "retry", idempotencyKey, hash)
	if err != nil {
		return ExecuteSuggestionResponse{}, ErrInternal
	}
	if conflict {
		return ExecuteSuggestionResponse{}, ErrIdempotencyConflict
	}
	if refType != "" {
		return ExecuteSuggestionResponse{ExecutionLogID: "execution-log-" + refID, SuggestionID: refID, PreviousStatus: StatusExecutionFailed, CurrentStatus: StatusExecuted, OperationLogID: "operation-log-" + refID}, nil
	}

	previousStatus := sug.Status
	now := time.Now().UTC()

	newStatus := StatusExecuted
	execResult := ExecutionResultSuccess
	failureReason := ""

	execLogID := fmt.Sprintf("execution-log-%s-retry-%d", suggestionID, now.UnixMilli())
	execLog := ExecutionLogResponse{
		ID:             execLogID,
		SuggestionID:   suggestionID,
		ActionType:     "retry",
		PreviousStatus: previousStatus,
		CurrentStatus:  newStatus,
		Result:         execResult,
		FailureReason:  failureReason,
		OperatorNote:   req.OperatorNote,
		CreatedAt:      now,
	}
	s.store.InsertExecutionLog(ctx, execLog)

	fields := map[string]any{"executed_at": &now}
	s.store.UpdateSuggestionStatus(ctx, suggestionID, newStatus, fields)

	opLogID := "operation-log-" + suggestionID
	s.store.StoreIdempotency(ctx, scope, "retry", idempotencyKey, hash, "strategy_suggestion", suggestionID)

	return ExecuteSuggestionResponse{ExecutionLogID: execLogID, SuggestionID: suggestionID, PreviousStatus: previousStatus, CurrentStatus: newStatus, OperationLogID: opLogID}, nil
}

func (s *service) ListExecutionLogs(ctx context.Context, suggestionID string, req ListExecutionLogsRequest) (PagedExecutionLogsResponse, error) {
	if suggestionID == "" {
		return PagedExecutionLogsResponse{}, ErrValidation
	}
	page := req.Page
	pageSize := req.PageSize
	logs, total, err := s.store.ListExecutionLogs(ctx, suggestionID, page, pageSize)
	if err != nil {
		return PagedExecutionLogsResponse{}, ErrInternal
	}
	return PagedExecutionLogsResponse{Items: logs, Pagination: pagination(page, pageSize, total)}, nil
}
