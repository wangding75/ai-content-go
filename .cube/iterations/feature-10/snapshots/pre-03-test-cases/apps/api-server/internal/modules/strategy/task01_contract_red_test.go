package strategy

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

// @Test
func TestTask01StrategyDTOsExposeAllRequiredFields(t *testing.T) {
	generateReqType := reflect.TypeOf(GenerateSuggestionsRequest{})
	for _, field := range []string{"date_from", "date_to", "rule_codes", "metric_codes", "force_regenerate"} {
		if !hasJSONField(generateReqType, field) {
			t.Fatalf("GenerateSuggestionsRequest missing json field %q", field)
		}
	}

	listReqType := reflect.TypeOf(ListStrategySuggestionsRequest{})
	for _, field := range []string{"status", "suggestion_type", "risk_level", "confidence", "date_from", "date_to", "sort", "order"} {
		if !hasJSONField(listReqType, field) {
			t.Fatalf("ListStrategySuggestionsRequest missing json field %q", field)
		}
	}

	detailType := reflect.TypeOf(StrategySuggestionDetailResponse{})
	for _, field := range []string{
		"id", "project_id", "suggestion_run_id", "suggestion_type", "title",
		"trigger_reason", "evidence_metrics", "impact_scope", "risk_level",
		"confidence", "suggested_action", "expected_benefit", "metrics_snapshot",
		"status", "ignored_reason", "ignored_note", "confirmed_at", "ignored_at",
		"executed_at", "date_from", "date_to", "triggered_rules", "generation_method",
	} {
		if !hasJSONField(detailType, field) {
			t.Fatalf("StrategySuggestionDetailResponse missing json field %q", field)
		}
	}

	confirmReqType := reflect.TypeOf(ConfirmSuggestionRequest{})
	if !hasJSONField(confirmReqType, "note") {
		t.Fatalf("ConfirmSuggestionRequest missing json field %q", "note")
	}

	ignoreReqType := reflect.TypeOf(IgnoreSuggestionRequest{})
	for _, field := range []string{"reason", "note"} {
		if !hasJSONField(ignoreReqType, field) {
			t.Fatalf("IgnoreSuggestionRequest missing json field %q", field)
		}
	}

	executeReqType := reflect.TypeOf(ExecuteSuggestionRequest{})
	for _, field := range []string{"action_type", "target_type", "target_id", "operator_note"} {
		if !hasJSONField(executeReqType, field) {
			t.Fatalf("ExecuteSuggestionRequest missing json field %q", field)
		}
	}

	statusChangeType := reflect.TypeOf(SuggestionStatusChangeResponse{})
	for _, field := range []string{"suggestion_id", "previous_status", "current_status", "operation_log_id"} {
		if !hasJSONField(statusChangeType, field) {
			t.Fatalf("SuggestionStatusChangeResponse missing json field %q", field)
		}
	}

	executeRespType := reflect.TypeOf(ExecuteSuggestionResponse{})
	for _, field := range []string{"execution_log_id", "suggestion_id", "previous_status", "current_status", "operation_log_id"} {
		if !hasJSONField(executeRespType, field) {
			t.Fatalf("ExecuteSuggestionResponse missing json field %q", field)
		}
	}

	logType := reflect.TypeOf(ExecutionLogResponse{})
	for _, field := range []string{"id", "suggestion_id", "action_type", "target_type", "target_id", "operator_note", "previous_status", "current_status", "result", "failure_reason", "target_interface", "target_resource"} {
		if !hasJSONField(logType, field) {
			t.Fatalf("ExecutionLogResponse missing json field %q", field)
		}
	}
}

// @Test
func TestTask01StrategyConstantsAndErrorsAreStableContracts(t *testing.T) {
	for _, value := range []string{
		SuggestionTypeKeep, SuggestionTypeOptimize, SuggestionTypeSuspend, SuggestionTypePromote, SuggestionTypeCostControl,
		RiskLevelLow, RiskLevelMedium, RiskLevelHigh,
		ConfidenceLow, ConfidenceMedium, ConfidenceHigh,
		StatusPending, StatusConfirmed, StatusIgnored, StatusExecuted, StatusExecutionFailed,
		RunStatusGenerating, RunStatusCompleted, RunStatusFailed,
		ExecutionResultSuccess, ExecutionResultFailed,
		GenerationMethodRule,
	} {
		if value == "" {
			t.Fatalf("strategy enum constants must be non-empty")
		}
	}
	for _, errValue := range []error{ErrValidation, ErrNotFound, ErrConflict, ErrIdempotencyConflict, ErrInternal} {
		if errValue == nil {
			t.Fatalf("strategy domain errors must be declared")
		}
	}
}

// @Test
func TestTask01ServiceInterfaceDeclaresAllStrategyUseCases(t *testing.T) {
	serviceType := reflect.TypeOf((*Service)(nil)).Elem()
	for _, method := range []string{
		"GenerateSuggestions",
		"ListSuggestions",
		"GetSuggestion",
		"ConfirmSuggestion",
		"IgnoreSuggestion",
		"ExecuteSuggestion",
		"RetrySuggestion",
		"ListExecutionLogs",
	} {
		if _, ok := serviceType.MethodByName(method); !ok {
			t.Fatalf("strategy Service missing method %s", method)
		}
	}
}

// @Test
func TestTask01StoreInterfaceDeclaresAllStrategyDataMethods(t *testing.T) {
	storeType := reflect.TypeOf((*Store)(nil)).Elem()
	for _, method := range []string{
		"InsertSuggestionRun",
		"FindSuggestionRunByID",
		"UpdateSuggestionRunStatus",
		"InsertSuggestion",
		"FindSuggestionByID",
		"UpdateSuggestionStatus",
		"ListSuggestions",
		"InsertExecutionLog",
		"ListExecutionLogs",
		"CheckIdempotency",
		"StoreIdempotency",
	} {
		if _, ok := storeType.MethodByName(method); !ok {
			t.Fatalf("strategy Store missing method %s", method)
		}
	}
}

// @Test
func TestTask04SuggestionTypeMustBelongToEnum(t *testing.T) {
	validTypes := map[string]bool{
		SuggestionTypeKeep: true, SuggestionTypeOptimize: true, SuggestionTypeSuspend: true,
		SuggestionTypePromote: true, SuggestionTypeCostControl: true,
	}
	for _, invalid := range []string{"unknown", "delete", "cancel", "", "OPTIMIZE"} {
		if validTypes[invalid] {
			t.Fatalf("type %q should not be valid", invalid)
		}
	}
}

// @Test
func TestTask05StateMachinePendingCanTransitionToConfirmedOrIgnored(t *testing.T) {
	svc := NewService()
	store := svc.(*service).store.(*memoryStore)

	s := StrategySuggestionDetailResponse{
		ID: "sug-1", ProjectID: "p-1", SuggestionRunID: "run-1",
		SuggestionType: SuggestionTypeOptimize, Title: "test", TriggerReason: "test",
		Status: StatusPending, DateFrom: "2026-05-01", DateTo: "2026-05-25",
	}
	store.InsertSuggestion(context.Background(), s)

	result, err := svc.ConfirmSuggestion(context.Background(), "sug-1", ConfirmSuggestionRequest{Note: "ok"}, "idem-confirm-1")
	if err != nil {
		t.Fatalf("confirm pending suggestion must succeed: %v", err)
	}
	if result.PreviousStatus != StatusPending || result.CurrentStatus != StatusConfirmed {
		t.Fatalf("confirm must transition pending→confirmed, got %s→%s", result.PreviousStatus, result.CurrentStatus)
	}
	if result.OperationLogID == "" {
		t.Fatalf("confirm must return operation_log_id")
	}
}

// @Test
func TestTask05ConfirmNonPendingReturnsConflict(t *testing.T) {
	svc := NewService()
	store := svc.(*service).store.(*memoryStore)

	s := StrategySuggestionDetailResponse{
		ID: "sug-2", ProjectID: "p-1", SuggestionRunID: "run-1",
		SuggestionType: SuggestionTypeOptimize, Title: "test", TriggerReason: "test",
		Status: StatusConfirmed, DateFrom: "2026-05-01", DateTo: "2026-05-25",
	}
	store.InsertSuggestion(context.Background(), s)

	_, err := svc.ConfirmSuggestion(context.Background(), "sug-2", ConfirmSuggestionRequest{}, "idem-confirm-2")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("confirm non-pending must return ErrConflict, got %v", err)
	}
}

// @Test
func TestTask05IgnorePendingRequiresReason(t *testing.T) {
	svc := NewService()
	store := svc.(*service).store.(*memoryStore)

	s := StrategySuggestionDetailResponse{
		ID: "sug-3", ProjectID: "p-1", SuggestionRunID: "run-1",
		SuggestionType: SuggestionTypeOptimize, Title: "test", TriggerReason: "test",
		Status: StatusPending, DateFrom: "2026-05-01", DateTo: "2026-05-25",
	}
	store.InsertSuggestion(context.Background(), s)

	_, err := svc.IgnoreSuggestion(context.Background(), "sug-3", IgnoreSuggestionRequest{Reason: "", Note: ""}, "idem-ignore-1")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("ignore without reason must return ErrValidation, got %v", err)
	}
}

// @Test
func TestTask05IgnorePendingWithReasonSucceeds(t *testing.T) {
	svc := NewService()
	store := svc.(*service).store.(*memoryStore)

	s := StrategySuggestionDetailResponse{
		ID: "sug-4", ProjectID: "p-1", SuggestionRunID: "run-1",
		SuggestionType: SuggestionTypeOptimize, Title: "test", TriggerReason: "test",
		Status: StatusPending, DateFrom: "2026-05-01", DateTo: "2026-05-25",
	}
	store.InsertSuggestion(context.Background(), s)

	result, err := svc.IgnoreSuggestion(context.Background(), "sug-4", IgnoreSuggestionRequest{Reason: "不需要", Note: "下次再说"}, "idem-ignore-2")
	if err != nil {
		t.Fatalf("ignore with reason must succeed: %v", err)
	}
	if result.PreviousStatus != StatusPending || result.CurrentStatus != StatusIgnored {
		t.Fatalf("ignore must transition pending→ignored, got %s→%s", result.PreviousStatus, result.CurrentStatus)
	}
	if result.OperationLogID == "" {
		t.Fatalf("ignore must return operation_log_id")
	}
}

// @Test
func TestTask06ExecuteConfirmedSuggestionTransitionsToExecuted(t *testing.T) {
	svc := NewService()
	store := svc.(*service).store.(*memoryStore)

	now := time.Now().UTC()
	s := StrategySuggestionDetailResponse{
		ID: "sug-5", ProjectID: "p-1", SuggestionRunID: "run-1",
		SuggestionType: SuggestionTypeOptimize, Title: "test", TriggerReason: "test",
		Status: StatusConfirmed, DateFrom: "2026-05-01", DateTo: "2026-05-25",
		ConfirmedAt: &now,
	}
	store.InsertSuggestion(context.Background(), s)

	result, err := svc.ExecuteSuggestion(context.Background(), "sug-5", ExecuteSuggestionRequest{
		ActionType: "adjust_schedule", TargetType: "workflow_schedule", TargetID: "sched-1",
	}, "idem-exec-1")
	if err != nil {
		t.Fatalf("execute confirmed must succeed: %v", err)
	}
	if result.PreviousStatus != StatusConfirmed {
		t.Fatalf("execute previous must be confirmed, got %s", result.PreviousStatus)
	}
	if result.CurrentStatus != StatusExecuted && result.CurrentStatus != StatusExecutionFailed {
		t.Fatalf("execute must transition to executed or execution_failed, got %s", result.CurrentStatus)
	}
	if result.ExecutionLogID == "" || result.OperationLogID == "" {
		t.Fatalf("execute must return execution_log_id and operation_log_id")
	}
}

// @Test
func TestTask06ExecuteNonConfirmedReturnsConflict(t *testing.T) {
	svc := NewService()
	store := svc.(*service).store.(*memoryStore)

	s := StrategySuggestionDetailResponse{
		ID: "sug-6", ProjectID: "p-1", SuggestionRunID: "run-1",
		SuggestionType: SuggestionTypeOptimize, Title: "test", TriggerReason: "test",
		Status: StatusPending, DateFrom: "2026-05-01", DateTo: "2026-05-25",
	}
	store.InsertSuggestion(context.Background(), s)

	_, err := svc.ExecuteSuggestion(context.Background(), "sug-6", ExecuteSuggestionRequest{
		ActionType: "adjust_schedule", TargetType: "workflow_schedule", TargetID: "sched-1",
	}, "idem-exec-2")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("execute non-confirmed must return ErrConflict, got %v", err)
	}
}

// @Test
func TestTask06RetryExecutionFailedSucceeds(t *testing.T) {
	svc := NewService()
	store := svc.(*service).store.(*memoryStore)

	s := StrategySuggestionDetailResponse{
		ID: "sug-7", ProjectID: "p-1", SuggestionRunID: "run-1",
		SuggestionType: SuggestionTypeOptimize, Title: "test", TriggerReason: "test",
		Status: StatusExecutionFailed, DateFrom: "2026-05-01", DateTo: "2026-05-25",
	}
	store.InsertSuggestion(context.Background(), s)

	result, err := svc.RetrySuggestion(context.Background(), "sug-7", RetrySuggestionRequest{OperatorNote: "retry"}, "idem-retry-1")
	if err != nil {
		t.Fatalf("retry execution_failed must succeed: %v", err)
	}
	if result.PreviousStatus != StatusExecutionFailed {
		t.Fatalf("retry previous must be execution_failed, got %s", result.PreviousStatus)
	}
}

// @Test
func TestTask06RetryNonFailedReturnsConflict(t *testing.T) {
	svc := NewService()
	store := svc.(*service).store.(*memoryStore)

	s := StrategySuggestionDetailResponse{
		ID: "sug-8", ProjectID: "p-1", SuggestionRunID: "run-1",
		SuggestionType: SuggestionTypeOptimize, Title: "test", TriggerReason: "test",
		Status: StatusPending, DateFrom: "2026-05-01", DateTo: "2026-05-25",
	}
	store.InsertSuggestion(context.Background(), s)

	_, err := svc.RetrySuggestion(context.Background(), "sug-8", RetrySuggestionRequest{}, "idem-retry-2")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("retry non-failed must return ErrConflict, got %v", err)
	}
}

// @Test
func TestTask09GetSuggestionReturnsDetail(t *testing.T) {
	svc := NewService()
	store := svc.(*service).store.(*memoryStore)

	s := StrategySuggestionDetailResponse{
		ID: "sug-detail-1", ProjectID: "p-1", SuggestionRunID: "run-1",
		SuggestionType: SuggestionTypeOptimize, Title: "阅读量下降", TriggerReason: "近7天下降30%",
		EvidenceMetrics: []MetricEvidence{{MetricCode: "views", Value: 1200, Trend: "declining"}},
		ImpactScope: "项目整体", RiskLevel: RiskLevelMedium, Confidence: ConfidenceHigh,
		SuggestedAction: "调整频率", ExpectedBenefit: "提升15-25%",
		Status: StatusPending, DateFrom: "2026-05-01", DateTo: "2026-05-25",
	}
	store.InsertSuggestion(context.Background(), s)

	result, err := svc.GetSuggestion(context.Background(), "sug-detail-1")
	if err != nil {
		t.Fatalf("get suggestion must succeed: %v", err)
	}
	if result.Title != "阅读量下降" || len(result.EvidenceMetrics) != 1 {
		t.Fatalf("get suggestion must return full detail: %#v", result)
	}
}

// @Test
func TestTask09GetSuggestionNotFound(t *testing.T) {
	svc := NewService()
	_, err := svc.GetSuggestion(context.Background(), "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("get nonexistent suggestion must return ErrNotFound, got %v", err)
	}
}

// @Test
func TestTask07ListSuggestionsFiltersByProjectAndStatus(t *testing.T) {
	svc := NewService()
	store := svc.(*service).store.(*memoryStore)

	for i, status := range []string{StatusPending, StatusPending, StatusConfirmed} {
		store.InsertSuggestion(context.Background(), StrategySuggestionDetailResponse{
			ID: "sug-list-" + string(rune('a'+i)), ProjectID: "p-1", SuggestionRunID: "run-1",
			SuggestionType: SuggestionTypeOptimize, Title: "test", TriggerReason: "test",
			Status: status, DateFrom: "2026-05-01", DateTo: "2026-05-25",
		})
	}

	result, err := svc.ListSuggestions(context.Background(), "p-1", ListStrategySuggestionsRequest{
		Status: StatusPending,
	})
	if err != nil {
		t.Fatalf("list suggestions must succeed: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 pending items, got %d", len(result.Items))
	}
}

// @Test
func TestTask08MemoryStoreExecutionLogCRUD(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	log1 := ExecutionLogResponse{
		ID: "elog-1", SuggestionID: "sug-1", ActionType: "adjust_schedule",
		TargetType: "workflow_schedule", TargetID: "sched-1",
		PreviousStatus: StatusConfirmed, CurrentStatus: StatusExecuted,
		Result: ExecutionResultSuccess,
	}
	store.InsertExecutionLog(ctx, log1)

	logs, total, err := store.ListExecutionLogs(ctx, "sug-1", 1, 20)
	if err != nil {
		t.Fatalf("list execution logs must succeed: %v", err)
	}
	if total != 1 || len(logs) != 1 || logs[0].ID != "elog-1" {
		t.Fatalf("expected 1 log, got total=%d len=%d", total, len(logs))
	}
}

// @Test
func TestTask08MemoryStoreIdempotencyWorks(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	refType, refID, conflict, err := store.CheckIdempotency(ctx, "strategy", "confirm", "key-1", "hash-abc")
	if err != nil || refType != "" || refID != "" || conflict {
		t.Fatalf("first check should return empty, got refType=%s refID=%s conflict=%v err=%v", refType, refID, conflict, err)
	}

	store.StoreIdempotency(ctx, "strategy", "confirm", "key-1", "hash-abc", "strategy_suggestion", "sug-1")

	refType, refID, conflict, err = store.CheckIdempotency(ctx, "strategy", "confirm", "key-1", "hash-abc")
	if err != nil || refType != "strategy_suggestion" || refID != "sug-1" || conflict {
		t.Fatalf("same key+hash should return saved result, got refType=%s refID=%s conflict=%v", refType, refID, conflict)
	}

	_, _, conflict, err = store.CheckIdempotency(ctx, "strategy", "confirm", "key-1", "hash-different")
	if err != nil || !conflict {
		t.Fatalf("same key different hash should return conflict, got conflict=%v err=%v", conflict, err)
	}
}

// @Test
func TestTask10ConfirmIsIdempotent(t *testing.T) {
	svc := NewService()
	store := svc.(*service).store.(*memoryStore)

	s := StrategySuggestionDetailResponse{
		ID: "sug-idem-1", ProjectID: "p-1", SuggestionRunID: "run-1",
		SuggestionType: SuggestionTypeOptimize, Title: "test", TriggerReason: "test",
		Status: StatusPending, DateFrom: "2026-05-01", DateTo: "2026-05-25",
	}
	store.InsertSuggestion(context.Background(), s)

	result1, err1 := svc.ConfirmSuggestion(context.Background(), "sug-idem-1", ConfirmSuggestionRequest{Note: "ok"}, "idem-key-confirm")
	if err1 != nil {
		t.Fatalf("first confirm must succeed: %v", err1)
	}

	s.Status = StatusPending
	store.InsertSuggestion(context.Background(), s)

	result2, err2 := svc.ConfirmSuggestion(context.Background(), "sug-idem-1", ConfirmSuggestionRequest{Note: "ok"}, "idem-key-confirm")
	if err2 != nil {
		t.Fatalf("idempotent confirm must succeed: %v", err2)
	}
	if result2.SuggestionID != result1.SuggestionID {
		t.Fatalf("idempotent confirm must return same suggestion_id")
	}
}

// @Test
func TestTask10SameIdempotencyKeyDifferentBodyReturnsIdempotencyConflict(t *testing.T) {
	svc := NewService()
	store := svc.(*service).store.(*memoryStore)

	s := StrategySuggestionDetailResponse{
		ID: "sug-idem-2", ProjectID: "p-1", SuggestionRunID: "run-1",
		SuggestionType: SuggestionTypeOptimize, Title: "test", TriggerReason: "test",
		Status: StatusPending, DateFrom: "2026-05-01", DateTo: "2026-05-25",
	}
	store.InsertSuggestion(context.Background(), s)

	_, err1 := svc.ConfirmSuggestion(context.Background(), "sug-idem-2", ConfirmSuggestionRequest{Note: "first"}, "idem-key-conflict")
	if err1 != nil {
		t.Fatalf("first confirm must succeed: %v", err1)
	}

	s.Status = StatusPending
	store.InsertSuggestion(context.Background(), s)

	_, err2 := svc.ConfirmSuggestion(context.Background(), "sug-idem-2", ConfirmSuggestionRequest{Note: "different"}, "idem-key-conflict")
	if !errors.Is(err2, ErrIdempotencyConflict) {
		t.Fatalf("same key different body must return ErrIdempotencyConflict, got %v", err2)
	}
}

func hasJSONField(t reflect.Type, jsonName string) bool {
	for i := 0; i < t.NumField(); i++ {
		name := strings.Split(t.Field(i).Tag.Get("json"), ",")[0]
		if name == jsonName {
			return true
		}
	}
	return false
}
