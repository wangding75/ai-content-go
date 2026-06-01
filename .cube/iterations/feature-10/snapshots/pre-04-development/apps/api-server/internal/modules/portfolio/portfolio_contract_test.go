package portfolio

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"
)

func jsonTagOf(model any, fieldName string) string {
	field, ok := reflect.TypeOf(model).FieldByName(fieldName)
	if !ok {
		return ""
	}
	return field.Tag.Get("json")
}

// @Test
func TestTask01ServiceInterfaceExposesAllPortfolioUseCases(t *testing.T) {
	serviceType := reflect.TypeOf((*Service)(nil)).Elem()
	for _, method := range []string{
		"CreatePortfolio",
		"ListPortfolios",
		"GetPortfolio",
		"UpdatePortfolio",
		"AddProject",
		"ListProjects",
		"UpdateProjectPriority",
		"RemoveProject",
		"RecalculateStatusSnapshot",
		"ListStatusSnapshots",
		"GetHealthSummary",
		"GetCostSummary",
		"GetStrategySummary",
	} {
		if _, ok := serviceType.MethodByName(method); !ok {
			t.Fatalf("portfolio Service missing method %s", method)
		}
	}
}

// @Test
func TestTask01StoreInterfaceExposesPersistenceContracts(t *testing.T) {
	storeType := reflect.TypeOf((*Store)(nil)).Elem()
	for _, method := range []string{
		"CreatePortfolio",
		"UpdatePortfolio",
		"GetPortfolio",
		"ListPortfolios",
		"AddProject",
		"UpdateProject",
		"RemoveProject",
		"GetProject",
		"ListProjects",
		"InsertStatusSnapshot",
		"ListStatusSnapshots",
		"GetLatestStatusSnapshot",
		"QueryHealthSummary",
		"QueryCostSummary",
		"QueryStrategySummary",
		"CheckIdempotency",
		"StoreIdempotency",
	} {
		if _, ok := storeType.MethodByName(method); !ok {
			t.Fatalf("portfolio Store missing method %s", method)
		}
	}
}

// @Test
func TestTask01PortfolioDTOsDeclareJSONContracts(t *testing.T) {
	cases := []struct {
		model any
		field string
		json  string
	}{
		{CreatePortfolioRequest{}, "ScopeType", "scope_type"},
		{PortfolioDetailResponse{}, "HealthPolicy", "health_policy"},
		{PortfolioDetailResponse{}, "LatestHealthStatus", "latest_health_status"},
		{PortfolioDetailResponse{}, "ProjectCount", "project_count"},
		{PortfolioDetailResponse{}, "LatestHealthScore", "latest_health_score"},
		{PortfolioDetailResponse{}, "EstimatedMonthlyCost", "estimated_monthly_cost"},
		{PortfolioDetailResponse{}, "Currency", "currency"},
		{PortfolioProjectResponse{}, "AddedBy", "added_by"},
		{RemovePortfolioProjectResponse{}, "OperationID", "operation_id"},
		{RecalculatePortfolioStatusSnapshotResponse{}, "SnapshotID", "snapshot_id"},
		{RecalculatePortfolioStatusSnapshotResponse{}, "JobID", "job_id"},
		{RecalculatePortfolioStatusSnapshotResponse{}, "CalculationStatus", "calculation_status"},
		{PortfolioStatusSnapshotResponse{}, "SourceRefs", "source_refs"},
		{PortfolioStatusSnapshotResponse{}, "RiskSummary", "risk_summary"},
		{PortfolioStatusSnapshotResponse{}, "CostSummary", "cost_summary"},
		{PortfolioStatusSnapshotResponse{}, "StrategySummary", "strategy_summary"},
		{PortfolioStatusSnapshotResponse{}, "DateRange", "date_range"},
		{PortfolioHealthSummaryResponse{}, "CalculatedAt", "calculated_at"},
		{PortfolioHealthSummaryResponse{}, "LatestSnapshotAt", "latest_snapshot_at"},
		{PortfolioCostSummaryResponse{}, "ProjectCosts", "project_costs"},
		{PortfolioCostSummaryResponse{}, "ByModel", "by_model"},
		{PortfolioStrategySummaryResponse{}, "TopSuggestions", "top_suggestions"},
	}
	for _, tc := range cases {
		if got := jsonTagOf(tc.model, tc.field); got != tc.json {
			t.Fatalf("expected %T.%s json tag %q, got %q", tc.model, tc.field, tc.json, got)
		}
	}
}

// @Test
func TestTask01PortfolioConstantsAndErrorsAreStableContracts(t *testing.T) {
	for _, value := range []string{
		PortfolioStatusActive,
		PortfolioStatusArchived,
		PortfolioScopeManual,
		PortfolioProjectRoleMember,
		SnapshotStatusQueued,
		SnapshotStatusRunning,
		SnapshotStatusCompleted,
		SnapshotStatusFailed,
		HealthStatusHealthy,
		HealthStatusWatch,
		HealthStatusCritical,
		HealthStatusPending,
	} {
		if value == "" {
			t.Fatalf("portfolio constants must be non-empty")
		}
	}
	for _, errValue := range []error{ErrValidation, ErrNotFound, ErrForbidden, ErrConflict, ErrIdempotencyConflict, ErrInternal} {
		if errValue == nil || !errors.Is(errValue, errValue) {
			t.Fatalf("portfolio sentinel errors must be declared and usable")
		}
	}
	if ErrValidation == ErrNotFound || ErrConflict == ErrInternal {
		t.Fatalf("portfolio sentinel errors must remain distinct")
	}
}

// @Test
func TestTask01ConstructorsReturnUsableSkeletons(t *testing.T) {
	var postgresDB *sql.DB
	if NewService() == nil {
		t.Fatalf("NewService must return a non-nil Service")
	}
	if NewMemoryStore() == nil {
		t.Fatalf("NewMemoryStore must return a non-nil Store")
	}
	if NewPostgresStore(postgresDB) == nil {
		t.Fatalf("NewPostgresStore must return a non-nil Store")
	}
}

// @Test
func TestTask01CreatePortfolioReturnsDeclaredIdentityAndAuditFields(t *testing.T) {
	svc := NewService()
	resp, err := svc.CreatePortfolio(context.Background(), CreatePortfolioRequest{
		Name:         "Novel 增长组合",
		Description:  "跨项目增长组合",
		ScopeType:    PortfolioScopeManual,
		OwnerID:      "growth-team",
		HealthPolicy: map[string]any{"warning_threshold": 60, "critical_threshold": 40},
		Status:       PortfolioStatusActive,
	}, "idem-portfolio-create")
	if err != nil {
		t.Fatalf("create portfolio contract must not fail for valid input: %v", err)
	}
	if resp.ID == "" {
		t.Fatalf("create portfolio must return portfolio id: %#v", resp)
	}
	if resp.CreatedAt == "" || resp.UpdatedAt == "" {
		t.Fatalf("create portfolio must return created_at and updated_at: %#v", resp)
	}
	if resp.ScopeType != PortfolioScopeManual || resp.OwnerID != "growth-team" || resp.Status != PortfolioStatusActive {
		t.Fatalf("create portfolio must preserve declared contract fields: %#v", resp)
	}
}

// @Test
func TestTask01UpdatePortfolioPreservesIdentityFields(t *testing.T) {
	svc := NewService()
	created, _ := svc.CreatePortfolio(context.Background(), CreatePortfolioRequest{
		Name: "PF-Update", ScopeType: PortfolioScopeManual, Status: PortfolioStatusActive,
	}, "idem-update-1")
	updated, err := svc.UpdatePortfolio(context.Background(), created.ID, UpdatePortfolioRequest{
		Name:   "PF-Updated",
		Status: PortfolioStatusArchived,
	}, "idem-update-2")
	if err != nil {
		t.Fatalf("update portfolio contract must not fail for valid input: %v", err)
	}
	if updated.ID != created.ID {
		t.Fatalf("update must preserve portfolio id: got %s want %s", updated.ID, created.ID)
	}
	if updated.Name != "PF-Updated" || updated.Status != PortfolioStatusArchived {
		t.Fatalf("update must reflect changed fields: %#v", updated)
	}
}

// @Test
func TestTask01RemoveProjectReturnsOperationReference(t *testing.T) {
	svc := NewService()
	resp, err := svc.RemoveProject(context.Background(), "pf-1", "proj-1", RemovePortfolioProjectRequest{
		Reason: "scope-update",
		Note:   "remove from portfolio only",
	})
	if err != nil {
		t.Fatalf("remove project contract must not fail for valid input: %v", err)
	}
	if !resp.Removed {
		t.Fatalf("remove project must mark removed=true: %#v", resp)
	}
	if resp.OperationID == "" {
		t.Fatalf("remove project must return operation_id for auditability: %#v", resp)
	}
}

// @Test
func TestTask01RecalculateStatusSnapshotReturnsQueuedReferences(t *testing.T) {
	svc := NewService()
	resp, err := svc.RecalculateStatusSnapshot(context.Background(), "pf-1", RecalculatePortfolioStatusSnapshotRequest{
		DateRange: DateRange{Start: "2026-05-01", End: "2026-05-31"},
		Force:     true,
	}, "idem-snapshot-recalc")
	if err != nil {
		t.Fatalf("recalculate snapshot contract must not fail for valid input: %v", err)
	}
	if resp.CalculationStatus != SnapshotStatusQueued {
		t.Fatalf("recalculate snapshot must return queued calculation_status: %#v", resp)
	}
	if resp.SnapshotID == "" || resp.JobID == "" {
		t.Fatalf("recalculate snapshot must return snapshot_id and job_id references: %#v", resp)
	}
}
