package external

import "github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"

type ListProvidersRequest struct {
	content.PaginationRequest
	ProviderType string `json:"provider_type"`
}

type CreateProviderRequest struct {
	ProviderType string `json:"provider_type"`
	BaseURL      string `json:"base_url"`
	Token        string `json:"token"`
}

type ProviderResponse struct {
	ID           string `json:"id"`
	ProviderType string `json:"provider_type"`
	BaseURL      string `json:"base_url"`
	TokenMasked  string `json:"token_masked"`
	Enabled      bool   `json:"enabled"`
}

type PagedProvidersResponse struct {
	Items      []ProviderResponse         `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type CreateProviderResponse struct {
	ProviderID  string `json:"provider_id"`
	TokenMasked string `json:"token_masked"`
}

type ListBindingsRequest struct {
	content.PaginationRequest
	ProviderID   string `json:"provider_id"`
	TriggerEvent string `json:"trigger_event"`
}

type CreateBindingRequest struct {
	ProviderID   string `json:"provider_id"`
	TriggerEvent string `json:"trigger_event"`
	WebhookURL   string `json:"webhook_url"`
}

type BindingResponse struct {
	ID           string `json:"id"`
	ProviderID   string `json:"provider_id"`
	TriggerEvent string `json:"trigger_event"`
	WebhookURL   string `json:"webhook_url"`
	Enabled      bool   `json:"enabled"`
}

type PagedBindingsResponse struct {
	Items      []BindingResponse          `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type CreateBindingResponse struct {
	BindingID string `json:"binding_id"`
}
