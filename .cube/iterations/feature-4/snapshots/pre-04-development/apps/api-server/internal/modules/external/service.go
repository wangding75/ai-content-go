package external

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

var (
	ErrValidation         = errors.New("validation error")
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrExternalAutomation = errors.New("external automation error")
)

type Service interface {
	ListProviders(ctx context.Context, req ListProvidersRequest) (PagedProvidersResponse, error)
	CreateProvider(ctx context.Context, req CreateProviderRequest) (CreateProviderResponse, error)
	ListBindings(ctx context.Context, req ListBindingsRequest) (PagedBindingsResponse, error)
	CreateBinding(ctx context.Context, req CreateBindingRequest) (CreateBindingResponse, error)
}

type service struct {
	mu           sync.RWMutex
	providers    []ProviderResponse
	bindings     []BindingResponse
	providerNext int
	bindingNext  int
}

func NewService() Service {
	return &service{}
}

func (s *service) ListProviders(_ context.Context, req ListProvidersRequest) (PagedProvidersResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]ProviderResponse, 0, len(s.providers))
	for _, item := range s.providers {
		if req.ProviderType != "" && item.ProviderType != req.ProviderType {
			continue
		}
		items = append(items, item)
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	return PagedProvidersResponse{Items: items, Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: len(items), HasNext: false}}, nil
}

func (s *service) CreateProvider(_ context.Context, req CreateProviderRequest) (CreateProviderResponse, error) {
	if req.ProviderType == "" || req.BaseURL == "" || req.Token == "" {
		return CreateProviderResponse{}, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.providers {
		if item.ProviderType == req.ProviderType && item.BaseURL == req.BaseURL {
			return CreateProviderResponse{}, ErrConflict
		}
	}
	s.providerNext++
	id := fmt.Sprintf("external-provider-%d", s.providerNext)
	masked := maskToken(req.Token)
	s.providers = append(s.providers, ProviderResponse{ID: id, ProviderType: req.ProviderType, BaseURL: req.BaseURL, TokenMasked: masked, Enabled: true})
	return CreateProviderResponse{ProviderID: id, TokenMasked: masked}, nil
}

func (s *service) ListBindings(_ context.Context, req ListBindingsRequest) (PagedBindingsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]BindingResponse, 0, len(s.bindings))
	for _, item := range s.bindings {
		if req.ProviderID != "" && item.ProviderID != req.ProviderID {
			continue
		}
		if req.TriggerEvent != "" && item.TriggerEvent != req.TriggerEvent {
			continue
		}
		items = append(items, item)
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	return PagedBindingsResponse{Items: items, Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: len(items), HasNext: false}}, nil
}

func (s *service) CreateBinding(_ context.Context, req CreateBindingRequest) (CreateBindingResponse, error) {
	if req.ProviderID == "" || req.TriggerEvent == "" || !strings.HasPrefix(req.WebhookURL, "https://") {
		return CreateBindingResponse{}, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	foundProvider := false
	for _, item := range s.providers {
		if item.ID == req.ProviderID {
			foundProvider = true
			break
		}
	}
	if !foundProvider {
		return CreateBindingResponse{}, ErrNotFound
	}
	s.bindingNext++
	id := fmt.Sprintf("external-binding-%d", s.bindingNext)
	s.bindings = append(s.bindings, BindingResponse{ID: id, ProviderID: req.ProviderID, TriggerEvent: req.TriggerEvent, WebhookURL: req.WebhookURL, Enabled: true})
	return CreateBindingResponse{BindingID: id}, nil
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

func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:3] + "****" + token[len(token)-4:]
}
