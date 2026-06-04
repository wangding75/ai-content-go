package portfolio

import (
	"context"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

// @Test
func TestIntegrationPortfolioCreateThenGetRoundTrip(t *testing.T) {
	store := NewMemoryStore()
	svc := NewService(store)
	created, err := svc.CreatePortfolio(context.Background(), CreatePortfolioRequest{
		Name: "Integration Test Portfolio", ScopeType: PortfolioScopeManual, Status: PortfolioStatusActive,
	}, "idem-integ-1")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := svc.GetPortfolio(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Integration Test Portfolio" || got.Status != PortfolioStatusActive {
		t.Fatalf("round-trip mismatch: %#v", got)
	}
}

// @Test
func TestIntegrationPortfolioAddProjectThenList(t *testing.T) {
	store := NewMemoryStore()
	svc := NewService(store)
	pf, _ := svc.CreatePortfolio(context.Background(), CreatePortfolioRequest{
		Name: "PF-Integ", ScopeType: PortfolioScopeManual, Status: PortfolioStatusActive,
	}, "idem-integ-2")
	_, err := svc.AddProject(context.Background(), pf.ID, AddPortfolioProjectRequest{
		ProjectID: "proj-a", Priority: 1, Weight: 1.5,
	}, "idem-integ-3")
	if err != nil {
		t.Fatalf("add project: %v", err)
	}
	result, err := svc.ListProjects(context.Background(), pf.ID, ListPortfolioProjectsRequest{PaginationRequest: content.PaginationRequest{Page: 1, PageSize: 20}})
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].ProjectID != "proj-a" {
		t.Fatalf("expected 1 project proj-a, got %#v", result.Items)
	}
}

// @Test
func TestIntegrationPortfolioRemoveProjectReturnsOperationRef(t *testing.T) {
	store := NewMemoryStore()
	svc := NewService(store)
	pf, _ := svc.CreatePortfolio(context.Background(), CreatePortfolioRequest{
		Name: "PF-Remove", ScopeType: PortfolioScopeManual, Status: PortfolioStatusActive,
	}, "idem-integ-4")
	svc.AddProject(context.Background(), pf.ID, AddPortfolioProjectRequest{
		ProjectID: "proj-b", Priority: 2, Weight: 1,
	}, "idem-integ-5")
	resp, err := svc.RemoveProject(context.Background(), pf.ID, "proj-b", RemovePortfolioProjectRequest{
		Reason: "scope-change", Note: "reassigned",
	})
	if err != nil {
		t.Fatalf("remove project: %v", err)
	}
	if !resp.Removed || resp.PortfolioID != pf.ID || resp.ProjectID != "proj-b" {
		t.Fatalf("remove response contract mismatch: %#v", resp)
	}
}

// @Test
func TestIntegrationPortfolioRecalculateReturnsQueuedReference(t *testing.T) {
	store := NewMemoryStore()
	svc := NewService(store)
	pf, _ := svc.CreatePortfolio(context.Background(), CreatePortfolioRequest{
		Name: "PF-Recalc", ScopeType: PortfolioScopeManual, Status: PortfolioStatusActive,
	}, "idem-integ-6")
	resp, err := svc.RecalculateStatusSnapshot(context.Background(), pf.ID, RecalculatePortfolioStatusSnapshotRequest{
		DateRange: DateRange{Start: "2026-05-01", End: "2026-05-31"}, Force: true,
	}, "idem-integ-7")
	if err != nil {
		t.Fatalf("recalculate: %v", err)
	}
	if resp.CalculationStatus != SnapshotStatusQueued {
		t.Fatalf("expected queued, got %s", resp.CalculationStatus)
	}
	if resp.PortfolioID != pf.ID {
		t.Fatalf("recalculate response portfolio_id mismatch: %s vs %s", resp.PortfolioID, pf.ID)
	}
}

// @Test
func TestIntegrationPortfolioSummaryQueriesReturnPortfolioID(t *testing.T) {
	store := NewMemoryStore()
	svc := NewService(store)
	pf, _ := svc.CreatePortfolio(context.Background(), CreatePortfolioRequest{
		Name: "PF-Summary", ScopeType: PortfolioScopeManual, Status: PortfolioStatusActive,
	}, "idem-integ-8")
	dr := DateRange{Start: "2026-05-01", End: "2026-05-31"}
	health, err := svc.GetHealthSummary(context.Background(), pf.ID, PortfolioSummaryRequest{DateRange: dr})
	if err != nil || health.PortfolioID != pf.ID {
		t.Fatalf("health summary contract: err=%v id=%s", err, health.PortfolioID)
	}
	cost, err := svc.GetCostSummary(context.Background(), pf.ID, PortfolioSummaryRequest{DateRange: dr})
	if err != nil || cost.PortfolioID != pf.ID {
		t.Fatalf("cost summary contract: err=%v id=%s", err, cost.PortfolioID)
	}
	strategy, err := svc.GetStrategySummary(context.Background(), pf.ID, PortfolioSummaryRequest{DateRange: dr})
	if err != nil || strategy.PortfolioID != pf.ID {
		t.Fatalf("strategy summary contract: err=%v id=%s", err, strategy.PortfolioID)
	}
}
