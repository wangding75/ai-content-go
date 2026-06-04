package contract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readIteration10OpenAPI(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean("../../../../../openapi/openapi.yaml"))
	if err != nil {
		t.Fatalf("read openapi.yaml: %v", err)
	}
	return string(content)
}

// @Test
func TestTask05OpenAPIDeclaresAllPortfolioPaths(t *testing.T) {
	openapi := readIteration10OpenAPI(t)
	for _, want := range []string{
		"/api/v1/portfolios:",
		"/api/v1/portfolios/{portfolioId}:",
		"/api/v1/portfolios/{portfolioId}/projects:",
		"/api/v1/portfolios/{portfolioId}/projects/{projectId}/priority:",
		"/api/v1/portfolios/{portfolioId}/projects/{projectId}:",
		"/api/v1/portfolios/{portfolioId}/status-snapshots/recalculate:",
		"/api/v1/portfolios/{portfolioId}/status-snapshots:",
		"/api/v1/portfolios/{portfolioId}/health-summary:",
		"/api/v1/portfolios/{portfolioId}/cost-summary:",
		"/api/v1/portfolios/{portfolioId}/strategy-summary:",
	} {
		if !strings.Contains(openapi, want) {
			t.Fatalf("openapi missing portfolio path %q", want)
		}
	}
}

// @Test
func TestTask05OpenAPIDeclaresPortfolioOperationIds(t *testing.T) {
	openapi := readIteration10OpenAPI(t)
	for _, want := range []string{
		"operationId: createPortfolio",
		"operationId: listPortfolios",
		"operationId: getPortfolio",
		"operationId: updatePortfolio",
		"operationId: addPortfolioProject",
		"operationId: listPortfolioProjects",
		"operationId: updatePortfolioProjectPriority",
		"operationId: removePortfolioProject",
		"operationId: recalculatePortfolioStatusSnapshot",
		"operationId: listPortfolioStatusSnapshots",
		"operationId: getPortfolioHealthSummary",
		"operationId: getPortfolioCostSummary",
		"operationId: getPortfolioStrategySummary",
	} {
		if !strings.Contains(openapi, want) {
			t.Fatalf("openapi missing %s", want)
		}
	}
}

// @Test
func TestTask05OpenAPIDeclaresCorrectStatusCodes(t *testing.T) {
	openapi := readIteration10OpenAPI(t)
	idx := strings.Index(openapi, "operationId: createPortfolio")
	if idx < 0 {
		t.Fatalf("createPortfolio operationId missing")
	}
	window := openapi[idx:min(len(openapi), idx+300)]
	if !strings.Contains(window, "'201':") {
		t.Fatalf("createPortfolio must declare 201 created")
	}

	idx = strings.Index(openapi, "operationId: recalculatePortfolioStatusSnapshot")
	if idx < 0 {
		t.Fatalf("recalculatePortfolioStatusSnapshot operationId missing")
	}
	window = openapi[idx:min(len(openapi), idx+300)]
	if !strings.Contains(window, "'202':") {
		t.Fatalf("recalculatePortfolioStatusSnapshot must declare 202 accepted")
	}
}
