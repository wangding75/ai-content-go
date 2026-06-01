package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestTask07PortfolioListPageDeclaresRequiredContracts(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/portfolios/page.tsx")
	if err != nil {
		t.Fatalf("read portfolio list page: %v", err)
	}
	page := string(raw)
	for _, required := range []string{
		"fetchPortfolios",
		"PortfolioDetailResponse",
		"page-shell",
		"page-header",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected portfolio list page to reference %s", required)
		}
	}
}

// @Test
func TestTask07PortfolioDetailPageDeclaresRequiredContracts(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/portfolios/[portfolioId]/page.tsx")
	if err != nil {
		t.Fatalf("read portfolio detail page: %v", err)
	}
	page := string(raw)
	for _, required := range []string{
		"fetchPortfolio",
		"fetchPortfolioStrategySummary",
		"PortfolioDetailResponse",
		"PortfolioStrategySummaryResponse",
		"page-shell",
		"summary-card",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected portfolio detail page to reference %s", required)
		}
	}
}

// @Test
func TestTask07PortfolioProjectsPageDeclaresRequiredContracts(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/portfolios/[portfolioId]/projects/page.tsx")
	if err != nil {
		t.Fatalf("read portfolio projects page: %v", err)
	}
	page := string(raw)
	for _, required := range []string{
		"fetchPortfolioProjects",
		"PortfolioProjectResponse",
		"page-shell",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected portfolio projects page to reference %s", required)
		}
	}
}

// @Test
func TestTask07PortfolioHealthPageDeclaresRequiredContracts(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/portfolios/[portfolioId]/health/page.tsx")
	if err != nil {
		t.Fatalf("read portfolio health page: %v", err)
	}
	page := string(raw)
	for _, required := range []string{
		"fetchPortfolioHealthSummary",
		"fetchPortfolioCostSummary",
		"fetchPortfolioStatusSnapshots",
		"PortfolioHealthSummaryResponse",
		"PortfolioCostSummaryResponse",
		"PortfolioStatusSnapshotResponse",
		"page-shell",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected portfolio health page to reference %s", required)
		}
	}
}

// @Test
func TestTask07GlobalNavDeclaresPortfolioEntry(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/global-nav.tsx")
	if err != nil {
		t.Fatalf("read global nav: %v", err)
	}
	nav := string(raw)
	if !strings.Contains(nav, "/portfolios") {
		t.Fatalf("global nav must declare /portfolios entry")
	}
	if !strings.Contains(nav, "Portfolio") {
		t.Fatalf("global nav must reference Portfolio label")
	}
}
