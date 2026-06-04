package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestTask06PortfolioAPIClientDeclaresTypesAndFunctions(t *testing.T) {
	content, err := os.ReadFile("../../../../../apps/web-admin/lib/api.ts")
	if err != nil {
		t.Fatalf("read web admin api client: %v", err)
	}
	api := string(content)
	for _, required := range []string{
		"export type DateRange",
		"export type SourceRef",
		"export type PortfolioDetailResponse",
		"export type PortfolioProjectResponse",
		"export type RemovePortfolioProjectResponse",
		"export type RecalculatePortfolioStatusSnapshotResponse",
		"export type PortfolioStatusSnapshotResponse",
		"export type PortfolioHealthSummaryResponse",
		"export type PortfolioCostSummaryResponse",
		"export type PortfolioStrategySummaryResponse",
		"export async function createPortfolio",
		"export async function fetchPortfolios",
		"export async function fetchPortfolio",
		"export async function updatePortfolio",
		"export async function addPortfolioProject",
		"export async function fetchPortfolioProjects",
		"export async function updatePortfolioProjectPriority",
		"export async function removePortfolioProject",
		"export async function recalculatePortfolioStatusSnapshot",
		"export async function fetchPortfolioStatusSnapshots",
		"export async function fetchPortfolioHealthSummary",
		"export async function fetchPortfolioCostSummary",
		"export async function fetchPortfolioStrategySummary",
	} {
		if !strings.Contains(api, required) {
			t.Fatalf("portfolio api client missing %s", required)
		}
	}
}

// @Test
func TestTask06PortfolioAPIClientUsesCorrectHTTPMethods(t *testing.T) {
	content, err := os.ReadFile("../../../../../apps/web-admin/lib/api.ts")
	if err != nil {
		t.Fatalf("read web admin api client: %v", err)
	}
	api := string(content)
	for _, required := range []string{
		"method: 'POST'",
		"method: 'PATCH'",
		"method: 'DELETE'",
	} {
		if !strings.Contains(api, required) {
			t.Fatalf("portfolio api client missing HTTP method %s", required)
		}
	}
}

// @Test
func TestTask06PortfolioAPIClientSendsIdempotencyKey(t *testing.T) {
	content, err := os.ReadFile("../../../../../apps/web-admin/lib/api.ts")
	if err != nil {
		t.Fatalf("read web admin api client: %v", err)
	}
	api := string(content)
	for _, required := range []string{
		"Idempotency-Key",
		"/api/v1/portfolios",
		"/api/v1/portfolios/${pathSegment(portfolioID)}",
		"status-snapshots/recalculate",
		"health-summary",
		"cost-summary",
		"strategy-summary",
	} {
		if !strings.Contains(api, required) {
			t.Fatalf("portfolio api client missing contract %s", required)
		}
	}
}

// @Test
func TestTask06PortfolioTypesDeclareRequiredContractFields(t *testing.T) {
	content, err := os.ReadFile("../../../../../apps/web-admin/lib/api.ts")
	if err != nil {
		t.Fatalf("read web admin api client: %v", err)
	}
	api := string(content)
	for _, required := range []string{
		"operation_id:",
		"snapshot_id:",
		"job_id:",
		"calculation_status:",
		"source_refs:",
		"calculated_at:",
		"scope_type:",
		"owner_id:",
		"health_policy:",
		"date_range:",
		"removed:",
	} {
		if !strings.Contains(api, required) {
			t.Fatalf("portfolio types missing contract field %s", required)
		}
	}
}
