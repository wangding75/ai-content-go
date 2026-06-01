package external_test

import (
	"context"
	"strings"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/external"
)

// @Test
func TestIteration21ExternalAutomationProviderMasksTokenAndRejectsDuplicateProvider(t *testing.T) {
	svc := external.NewService()
	secret := "sk-live-red-contract-secret-21"

	created, err := svc.CreateProvider(context.Background(), external.CreateProviderRequest{
		ProviderType: "n8n",
		BaseURL:      "https://n8n.example.invalid",
		Token:        secret,
	})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	if created.ProviderID == "" || created.TokenMasked == "" || strings.Contains(created.TokenMasked, secret) || strings.Contains(created.TokenMasked, "red-contract-secret") {
		t.Fatalf("expected masked token only, got %#v", created)
	}
	if _, err := svc.CreateProvider(context.Background(), external.CreateProviderRequest{ProviderType: "n8n", BaseURL: "https://n8n.example.invalid", Token: "other-secret"}); err == nil {
		t.Fatalf("expected conflict for duplicate provider_type/base_url")
	}
}

// @Test
func TestIteration21ExternalAutomationBindingRequiresExistingProviderAndValidWebhook(t *testing.T) {
	svc := external.NewService()

	if _, err := svc.CreateBinding(context.Background(), external.CreateBindingRequest{
		ProviderID: "missing-provider", TriggerEvent: "workflow_run.completed", WebhookURL: "https://n8n.example.invalid/webhook/run",
	}); err == nil {
		t.Fatalf("expected not found for missing provider")
	}

	provider, err := svc.CreateProvider(context.Background(), external.CreateProviderRequest{ProviderType: "n8n", BaseURL: "https://n8n-bind.example.invalid", Token: "sk-test-bind-secret"})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	binding, err := svc.CreateBinding(context.Background(), external.CreateBindingRequest{ProviderID: provider.ProviderID, TriggerEvent: "workflow_run.completed", WebhookURL: "https://n8n-bind.example.invalid/webhook/run"})
	if err != nil {
		t.Fatalf("create binding: %v", err)
	}
	if binding.BindingID == "" {
		t.Fatalf("expected binding_id")
	}
}
