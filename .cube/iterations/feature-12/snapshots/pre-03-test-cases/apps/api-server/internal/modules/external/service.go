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
	ErrForbidden          = errors.New("forbidden")
)

type Service interface {
	ListProviders(ctx context.Context, req ListProvidersRequest) (PagedProvidersResponse, error)
	CreateProvider(ctx context.Context, req CreateProviderRequest) (CreateProviderResponse, error)
	ListBindings(ctx context.Context, req ListBindingsRequest) (PagedBindingsResponse, error)
	CreateBinding(ctx context.Context, req CreateBindingRequest) (CreateBindingResponse, error)
	RotateCallbackToken(ctx context.Context, bindingID string, req RotateCallbackTokenRequest, idempotencyKey string) (RotateCallbackTokenResponse, error)
	UpdateCallbackAuth(ctx context.Context, bindingID string, req UpdateCallbackAuthRequest, idempotencyKey string) (UpdateCallbackAuthResponse, error)
	ReceiveCallback(ctx context.Context, req ExternalCallbackRequest, auth ExternalCallbackAuth, idempotencyKey string) (ExternalCallbackResponse, error)
	ListCallbackLogs(ctx context.Context, req ListCallbackLogsRequest) (PagedExternalCallbackLogsResponse, error)
	TestCallback(ctx context.Context, req TestExternalCallbackRequest) (ExternalCallbackResponse, error)
}

type service struct {
	mu           sync.RWMutex
	providers    []ProviderResponse
	bindings     []BindingResponse
	providerNext int
	bindingNext  int
	callbackLogs []ExternalCallbackLogResponse
}

func NewService() Service {
	return &service{
		bindings: []BindingResponse{
			{ID: "binding-001", ProviderID: "external-provider-1", TriggerEvent: "platform_collect.submitted", WebhookURL: "https://n8n.example.com/webhook/collect", Enabled: true},
		},
	}
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

func (s *service) RotateCallbackToken(_ context.Context, bindingID string, req RotateCallbackTokenRequest, _ string) (RotateCallbackTokenResponse, error) {
	if bindingID == "" {
		return RotateCallbackTokenResponse{}, ErrValidation
	}
	s.mu.RLock()
	found := false
	for _, b := range s.bindings {
		if b.ID == bindingID {
			found = true
			break
		}
	}
	s.mu.RUnlock()
	if !found {
		return RotateCallbackTokenResponse{}, ErrNotFound
	}
	return RotateCallbackTokenResponse{
		BindingID:           bindingID,
		CallbackTokenOnce:   "cbtoken_once_" + bindingID + "_" + fmt.Sprintf("%d", len(s.callbackLogs)),
		CallbackTokenMasked: "cbto****" + bindingID[len(bindingID)-4:],
		OperationLogID:      "oplog-rotate-callback-" + bindingID,
	}, nil
}

func (s *service) UpdateCallbackAuth(_ context.Context, bindingID string, req UpdateCallbackAuthRequest, _ string) (UpdateCallbackAuthResponse, error) {
	if bindingID == "" || req.CallbackAuthType == "" {
		return UpdateCallbackAuthResponse{}, ErrValidation
	}
	validAuthTypes := map[string]bool{"token": true, "signature": true}
	if !validAuthTypes[req.CallbackAuthType] {
		return UpdateCallbackAuthResponse{}, ErrValidation
	}
	s.mu.RLock()
	found := false
	for _, b := range s.bindings {
		if b.ID == bindingID {
			found = true
			break
		}
	}
	s.mu.RUnlock()
	if !found {
		return UpdateCallbackAuthResponse{}, ErrNotFound
	}
	return UpdateCallbackAuthResponse{
		BindingID:        bindingID,
		CallbackAuthType: req.CallbackAuthType,
		OperationLogID:   "oplog-update-callback-auth-" + bindingID,
	}, nil
}

var validCallbackEventTypes = map[string]bool{
	"platform_collect.submitted": true,
	"platform_collect.completed": true,
	"platform_publish.succeeded": true,
	"platform_publish.failed":    true,
}

var boundaryViolationKeys = map[string]bool{
	"workflow_run_id": true,
	"agent_code":      true,
	"content_body":    true,
}

func (s *service) ReceiveCallback(_ context.Context, req ExternalCallbackRequest, auth ExternalCallbackAuth, _ string) (ExternalCallbackResponse, error) {
	if req.BindingID == "" || req.EventType == "" {
		return ExternalCallbackResponse{}, ErrValidation
	}
	if !validCallbackEventTypes[req.EventType] {
		return ExternalCallbackResponse{}, ErrValidation
	}
	if auth.BindingID == "" {
		return ExternalCallbackResponse{}, ErrValidation
	}

	violation := false
	for key := range req.Payload {
		if boundaryViolationKeys[key] {
			violation = true
			break
		}
	}
	if violation {
		s.mu.Lock()
		logID := "callback-log-" + fmt.Sprintf("%d", len(s.callbackLogs))
		s.callbackLogs = append(s.callbackLogs, ExternalCallbackLogResponse{
			ID:                logID,
			ProviderID:        "external-provider-1",
			BindingID:         req.BindingID,
			EventType:         req.EventType,
			Accepted:          false,
			RejectedReason:    "boundary_violation",
			BoundaryViolation: true,
			CreatedAt:         fmt.Sprintf("%d", len(s.callbackLogs)),
		})
		s.mu.Unlock()
		return ExternalCallbackResponse{}, ErrForbidden
	}

	s.mu.Lock()
	logID := "callback-log-" + fmt.Sprintf("%d", len(s.callbackLogs))
	s.callbackLogs = append(s.callbackLogs, ExternalCallbackLogResponse{
		ID:                logID,
		ProviderID:        "external-provider-1",
		BindingID:         req.BindingID,
		EventType:         req.EventType,
		Accepted:          true,
		BoundaryViolation: false,
		CreatedAt:         fmt.Sprintf("%d", len(s.callbackLogs)),
	})
	s.mu.Unlock()

	return ExternalCallbackResponse{
		Accepted:      true,
		CallbackLogID: logID,
	}, nil
}

func (s *service) ListCallbackLogs(_ context.Context, req ListCallbackLogsRequest) (PagedExternalCallbackLogsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]ExternalCallbackLogResponse, 0, len(s.callbackLogs))
	for _, log := range s.callbackLogs {
		if req.BindingID != "" && log.BindingID != req.BindingID {
			continue
		}
		if req.EventType != "" && log.EventType != req.EventType {
			continue
		}
		if req.Accepted != nil && log.Accepted != *req.Accepted {
			continue
		}
		items = append(items, log)
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	start, end := pageBounds(page, pageSize, len(items))
	return PagedExternalCallbackLogsResponse{
		Items:      items[start:end],
		Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: len(items), HasNext: end < len(items)},
	}, nil
}

func (s *service) TestCallback(_ context.Context, req TestExternalCallbackRequest) (ExternalCallbackResponse, error) {
	if req.BindingID == "" {
		return ExternalCallbackResponse{}, ErrValidation
	}
	return ExternalCallbackResponse{
		Accepted:      true,
		CallbackLogID: "test-callback-log-" + req.BindingID,
		BindingID:     req.BindingID,
		StableEventID: "test-stable-event",
		TestResult:    "delivered",
	}, nil
}

func pageBounds(page, pageSize, total int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return start, end
}
