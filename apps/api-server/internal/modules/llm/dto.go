package llm

import "github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"

type ListProvidersRequest struct {
	content.PaginationRequest
}

type CreateProviderRequest struct {
	ProviderType string `json:"provider_type"`
	BaseURL      string `json:"base_url"`
	APIKey       string `json:"api_key"`
}

type ProviderResponse struct {
	ID           string `json:"id"`
	ProviderType string `json:"provider_type"`
	BaseURL      string `json:"base_url"`
	APIKeyMasked string `json:"api_key_masked"`
	Enabled      bool   `json:"enabled"`
}

type PagedProvidersResponse struct {
	Items      []ProviderResponse         `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type CreateProviderResponse struct {
	ProviderID   string `json:"provider_id"`
	APIKeyMasked string `json:"api_key_masked"`
}
