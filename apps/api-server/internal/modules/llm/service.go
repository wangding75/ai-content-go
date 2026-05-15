package llm

import (
	"context"
	"errors"
	"sync"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

var (
	ErrValidation = errors.New("validation error")
	ErrConflict   = errors.New("conflict")
)

type Service interface {
	ListProviders(ctx context.Context, req ListProvidersRequest) (PagedProvidersResponse, error)
	CreateProvider(ctx context.Context, req CreateProviderRequest) (CreateProviderResponse, error)
}

type service struct {
	mu        sync.RWMutex
	providers []ProviderResponse
}

func NewService() Service {
	return &service{providers: []ProviderResponse{{
		ID:           "provider-1",
		ProviderType: "openai-compatible",
		BaseURL:      "https://seed.example.invalid/v1",
		APIKeyMasked: "see****1234",
		Enabled:      true,
	}}}
}

func (s *service) ListProviders(ctx context.Context, req ListProvidersRequest) (PagedProvidersResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := append([]ProviderResponse(nil), s.providers...)
	page, pageSize := normalizePage(req.Page, req.PageSize)
	return PagedProvidersResponse{Items: items, Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: len(items), HasNext: len(items) > page*pageSize}}, nil
}

func (s *service) CreateProvider(ctx context.Context, req CreateProviderRequest) (CreateProviderResponse, error) {
	if req.ProviderType == "" || req.BaseURL == "" || req.APIKey == "" {
		return CreateProviderResponse{}, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.providers {
		if item.ProviderType == req.ProviderType && item.BaseURL == req.BaseURL {
			return CreateProviderResponse{}, ErrConflict
		}
	}
	id := "provider-new"
	masked := MaskAPIKey(req.APIKey)
	s.providers = append(s.providers, ProviderResponse{ID: id, ProviderType: req.ProviderType, BaseURL: req.BaseURL, APIKeyMasked: masked, Enabled: true})
	return CreateProviderResponse{ProviderID: id, APIKeyMasked: masked}, nil
}

func MaskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:3] + "****" + apiKey[len(apiKey)-4:]
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return page, pageSize
}
