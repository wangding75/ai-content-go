package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestWebAdminAPIClientDeclaresIteration1EnvelopeAndFunctions(t *testing.T) {
	content, err := os.ReadFile("../../../../../apps/web-admin/lib/api.ts")
	if err != nil {
		t.Fatalf("read web admin api client: %v", err)
	}
	api := string(content)
	for _, required := range []string{
		"APIEnvelope<T>",
		"success: boolean",
		"data: T | null",
		"error:",
		"request_id: string",
		"fetchDashboardSummary",
		"fetchContentTypes",
		"createContentType",
		"fetchProjectSchema",
		"fetchProjects",
		"createProject",
		"fetchProjectOverview",
		"pauseProject",
		"fetchPromptTemplates",
		"createPromptTemplate",
		"fetchLLMProviders",
		"createLLMProvider",
	} {
		if !strings.Contains(api, required) {
			t.Fatalf("expected api client to declare %s", required)
		}
	}
}

// @Test
func TestWebAdminAPIClientProviderTypesExposeMaskedKeyOnlyInResponses(t *testing.T) {
	content, err := os.ReadFile("../../../../../apps/web-admin/lib/api.ts")
	if err != nil {
		t.Fatalf("read web admin api client: %v", err)
	}
	api := string(content)
	providerTypeStart := strings.Index(api, "export type LLMProviderResponse")
	if providerTypeStart < 0 {
		t.Fatalf("expected LLMProviderResponse type")
	}
	providerType := api[providerTypeStart : strings.Index(api[providerTypeStart:], "};")+providerTypeStart]
	if !strings.Contains(providerType, "api_key_masked") {
		t.Fatalf("provider response type must include api_key_masked")
	}
	if strings.Contains(providerType, "api_key:") {
		t.Fatalf("provider response type must not include plaintext api_key")
	}
}
