package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestIteration10PortfolioListPageRendersWithEmptyFixture(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/portfolios/page.tsx")
	if err != nil {
		t.Fatalf("read portfolio list page: %v", err)
	}
	page := string(raw)
	if !strings.Contains(page, "fetchPortfolios") {
		t.Fatalf("portfolio list page must call fetchPortfolios")
	}
}

// @Test
func TestIteration10PortfolioDetailPageUsesScopedPortfolioId(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/portfolios/[portfolioId]/page.tsx")
	if err != nil {
		t.Fatalf("read portfolio detail page: %v", err)
	}
	page := string(raw)
	for _, required := range []string{
		"fetchPortfolio",
		"fetchPortfolioStrategySummary",
		"portfolioId",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("portfolio detail page must reference %s", required)
		}
	}
}

// @Test
func TestIteration10PortfolioHealthPageUsesCorrectApiCalls(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/portfolios/[portfolioId]/health/page.tsx")
	if err != nil {
		t.Fatalf("read portfolio health page: %v", err)
	}
	page := string(raw)
	for _, required := range []string{
		"fetchPortfolioHealthSummary",
		"fetchPortfolioCostSummary",
		"fetchPortfolioStatusSnapshots",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("portfolio health page must call %s", required)
		}
	}
}
