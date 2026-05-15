package llm

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
)

type Service interface {
	ListProviders(ctx context.Context, req ListProvidersRequest) (PagedProvidersResponse, error)
	CreateProvider(ctx context.Context, req CreateProviderRequest) (CreateProviderResponse, error)
	CreateCallLog(ctx context.Context, req CreateLLMCallLogRequest) (LLMCallLogResponse, error)
	ListCallLogs(ctx context.Context, req ListLLMCallLogsRequest) (PagedLLMCallLogsResponse, error)
	GetCallLog(ctx context.Context, id string) (LLMCallLogDetailResponse, error)
}

type service struct {
	mu        sync.RWMutex
	providers []ProviderResponse
	callLogs  []LLMCallLogDetailResponse
	callLogNext int
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

func (s *service) CreateCallLog(_ context.Context, req CreateLLMCallLogRequest) (LLMCallLogResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.callLogNext++
	id := fmt.Sprintf("llmlog-%d", s.callLogNext)
	detail := LLMCallLogDetailResponse{
		LLMCallLogResponse: LLMCallLogResponse{
			ID:            id,
			WorkflowRunID: req.WorkflowRunID,
			StepRunID:     req.StepRunID,
			AgentTaskID:   req.AgentTaskID,
			Provider:      req.Provider,
			Model:         req.Model,
			InputTokens:   req.InputTokens,
			OutputTokens:  req.OutputTokens,
			Cost:          req.Cost,
			Currency:      req.Currency,
			LatencyMS:     req.LatencyMS,
			Status:        req.Status,
		},
		Error:     req.Error,
		RequestID: req.RequestID,
	}
	s.callLogs = append(s.callLogs, detail)
	return detail.LLMCallLogResponse, nil
}

func (s *service) ListCallLogs(_ context.Context, req ListLLMCallLogsRequest) (PagedLLMCallLogsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var items []LLMCallLogResponse
	for _, log := range s.callLogs {
		if req.WorkflowRunID != "" && log.WorkflowRunID != req.WorkflowRunID {
			continue
		}
		if req.AgentTaskID != "" && log.AgentTaskID != req.AgentTaskID {
			continue
		}
		if req.Provider != "" && log.Provider != req.Provider {
			continue
		}
		if req.Model != "" && log.Model != req.Model {
			continue
		}
		if req.Status != "" && log.Status != req.Status {
			continue
		}
		items = append(items, log.LLMCallLogResponse)
	}
	if items == nil {
		items = []LLMCallLogResponse{}
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	total := len(items)
	return PagedLLMCallLogsResponse{
		Items:      items,
		Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: total, HasNext: false},
	}, nil
}

func (s *service) GetCallLog(_ context.Context, id string) (LLMCallLogDetailResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, log := range s.callLogs {
		if log.ID == id {
			return log, nil
		}
	}
	return LLMCallLogDetailResponse{}, ErrNotFound
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

