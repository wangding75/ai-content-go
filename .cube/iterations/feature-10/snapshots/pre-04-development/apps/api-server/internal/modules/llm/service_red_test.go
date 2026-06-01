package llm_test

import (
	"context"
	"strings"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/llm"
)

// @Test
func TestListProvidersReturnsOnlyMaskedAPIKeys(t *testing.T) {
	service := llm.NewService()

	resp, err := service.ListProviders(context.Background(), llm.ListProvidersRequest{})
	if err != nil {
		t.Fatalf("list providers should not fail: %v", err)
	}
	for _, provider := range resp.Items {
		if provider.ID == "" || provider.ProviderType == "" || provider.BaseURL == "" || provider.APIKeyMasked == "" || strings.Contains(provider.APIKeyMasked, "sk-live-secret") {
			t.Fatalf("provider response must only contain masked key, got %#v", provider)
		}
	}
}

// @Test
func TestCreateProviderReturnsProviderIDAndMaskedAPIKeyOnly(t *testing.T) {
	service := llm.NewService()

	resp, err := service.CreateProvider(context.Background(), llm.CreateProviderRequest{
		ProviderType: "openai-compatible",
		BaseURL:      "https://api.example.invalid/v1",
		APIKey:       "sk-live-secret-123456",
	})
	if err != nil {
		t.Fatalf("create provider should not fail: %v", err)
	}
	if resp.ProviderID == "" || resp.APIKeyMasked == "" || strings.Contains(resp.APIKeyMasked, "sk-live-secret-123456") {
		t.Fatalf("unexpected create provider response: %#v", resp)
	}
}

// @Test
func TestCreateProviderRejectsMissingFieldsAndDuplicateProvider(t *testing.T) {
	service := llm.NewService()
	if _, err := service.CreateProvider(context.Background(), llm.CreateProviderRequest{}); err == nil {
		t.Fatalf("expected validation error for missing provider fields")
	}
	req := llm.CreateProviderRequest{ProviderType: "openai-compatible", BaseURL: "https://api.example.invalid/v1", APIKey: "sk-live-secret-123456"}
	_, _ = service.CreateProvider(context.Background(), req)
	if _, err := service.CreateProvider(context.Background(), req); err == nil {
		t.Fatalf("expected conflict error for duplicate provider type and base url")
	}
}

// @Test
func TestMaskAPIKeyPreservesPrefixAndSuffixWithoutLeakingSecret(t *testing.T) {
	masked := llm.MaskAPIKey("sk-live-secret-123456")
	if masked == "" || masked == "sk-live-secret-123456" || strings.Contains(masked, "secret") {
		t.Fatalf("masked key leaked secret material: %q", masked)
	}
	if !strings.HasPrefix(masked, "sk-") || !strings.HasSuffix(masked, "3456") {
		t.Fatalf("masked key should preserve safe prefix and suffix, got %q", masked)
	}
}
