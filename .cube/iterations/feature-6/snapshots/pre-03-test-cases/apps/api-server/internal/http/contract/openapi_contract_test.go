package contract_test

import (
	"os"
	"strings"
	"testing"
)

func readOpenAPIContract(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile("../../../../../openapi/openapi.yaml")
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	return string(content)
}

// @Test
func TestOpenAPIContainsAllIteration1EndpointsWithSecurityAndOperationIDs(t *testing.T) {
	doc := readOpenAPIContract(t)
	for _, required := range []string{
		"/api/v1/dashboard/summary:",
		"operationId: getDashboardSummary",
		"/api/v1/content-types:",
		"operationId: listContentTypes",
		"operationId: createContentType",
		"/api/v1/content-types/{id}/project-schema:",
		"operationId: getProjectSchema",
		"/api/v1/projects:",
		"operationId: listProjects",
		"operationId: createProject",
		"/api/v1/projects/{id}/overview:",
		"operationId: getProjectOverview",
		"/api/v1/projects/{id}/pause:",
		"operationId: pauseProject",
		"/api/v1/prompt-templates:",
		"operationId: listPromptTemplates",
		"operationId: createPromptTemplate",
		"/api/v1/llm-providers:",
		"operationId: listLLMProviders",
		"operationId: createLLMProvider",
		"security:",
		"bearerAuth: []",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("expected OpenAPI contract to contain %s", required)
		}
	}
}

// @Test
func TestOpenAPIProviderResponsesExposeOnlyMaskedAPIKey(t *testing.T) {
	doc := readOpenAPIContract(t)
	providerStart := strings.Index(doc, "    LLMProvider:\n")
	providerEnd := strings.Index(doc, "    LLMProvidersPage:\n")
	if providerStart < 0 || providerEnd < 0 || providerEnd <= providerStart {
		t.Fatalf("expected LLMProvider response schema section")
	}
	providerSection := doc[providerStart:providerEnd]
	if !strings.Contains(providerSection, "api_key_masked:") {
		t.Fatalf("LLMProvider response schema must expose api_key_masked")
	}
	if strings.Contains(providerSection, "api_key:") {
		t.Fatalf("LLMProvider response schema must not expose plaintext api_key")
	}
}

// @Test
func TestOpenAPIErrorResponsesCoverIteration1ErrorMatrix(t *testing.T) {
	doc := readOpenAPIContract(t)
	for _, required := range []string{"ValidationError:", "Unauthorized:", "Forbidden:", "NotFound:", "Conflict:", "InternalError:"} {
		if !strings.Contains(doc, required) {
			t.Fatalf("expected OpenAPI error response %s", required)
		}
	}
}
