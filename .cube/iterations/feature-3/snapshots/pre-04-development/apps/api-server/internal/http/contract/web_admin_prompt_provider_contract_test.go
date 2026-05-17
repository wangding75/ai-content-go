package contract_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// @Test
func TestPromptAndProviderPagesImplementPrototypeNavigationStateAndMutationFeedback(t *testing.T) {
	promptPage, err := os.ReadFile(filepath.Join(webAdminRoot(t), "app/prompt/page.tsx"))
	if err != nil {
		t.Fatalf("read prompt page: %v", err)
	}
	providerPage, err := os.ReadFile(filepath.Join(webAdminRoot(t), "app/provider/page.tsx"))
	if err != nil {
		t.Fatalf("read provider page: %v", err)
	}
	combined := string(promptPage) + "\n" + string(providerPage)

	for _, required := range []string{
		"Prompt 模板管理",
		"模型 Provider 管理",
		"新建 Prompt",
		"新增 Provider",
		"await fetchPromptTemplates(",
		"await createPromptTemplate(",
		"await fetchLLMProviders(",
		"await createLLMProvider(",
		"data-testid=\"prompt-loading\"",
		"data-testid=\"provider-loading\"",
		"data-testid=\"provider-empty\"",
		"role=\"alert\"",
		"role=\"status\"",
		"VALIDATION_ERROR",
		"request_id",
	} {
		if !strings.Contains(combined, required) {
			t.Fatalf("prompt/provider UI contract must include %q", required)
		}
	}
	for _, forbidden := range []string{
		"将在开发阶段连接 API",
		"只展示 api_key_masked，不展示明文 api_key",
	} {
		if strings.Contains(combined, forbidden) {
			t.Fatalf("prompt/provider pages must not satisfy UI contract with placeholder text %q", forbidden)
		}
	}
}

// @Test
func TestProviderPageDisplaysMaskedKeyAndNeverPlaintextAPIKey(t *testing.T) {
	content, err := os.ReadFile(filepath.Join(webAdminRoot(t), "app/provider/page.tsx"))
	if err != nil {
		t.Fatalf("read provider page: %v", err)
	}
	page := string(content)
	if !strings.Contains(page, "api_key_masked") {
		t.Fatalf("provider page must display api_key_masked")
	}
	if strings.Contains(page, "api_key:") {
		t.Fatalf("provider page must not display plaintext api_key")
	}
}

// @Test
func TestPromptProviderE2ESpecCoversRealRoundtripEmptyValidationAndSecretMasking(t *testing.T) {
	spec := readWebAdminFile(t, filepath.Join(webAdminRoot(t), "e2e/iteration1-ui.spec.ts"))

	for _, required := range []string{
		"/api/v1/prompt-templates",
		"/api/v1/llm-providers",
		"GET /api/v1/prompt-templates",
		"POST /api/v1/prompt-templates",
		"GET /api/v1/llm-providers",
		"POST /api/v1/llm-providers",
		"prompt-loading",
		"provider-loading",
		"provider-empty",
		"provider-key-masked",
		"Prompt 创建成功",
		"Provider 创建成功",
		"VALIDATION_ERROR",
		"not.toContainText('sk-live-red-contract-secret')",
	} {
		if !strings.Contains(spec, required) {
			t.Fatalf("prompt/provider e2e spec must cover frontend-ui requirement %q", required)
		}
	}
}
