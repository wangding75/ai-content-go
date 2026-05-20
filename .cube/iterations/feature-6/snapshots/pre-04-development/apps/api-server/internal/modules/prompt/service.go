package prompt

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

var (
	ErrValidation = errors.New("validation error")
	ErrConflict   = errors.New("conflict")
)

type Service interface {
	ListTemplates(ctx context.Context, req ListTemplatesRequest) (PagedTemplatesResponse, error)
	CreateTemplate(ctx context.Context, req CreateTemplateRequest) (CreateTemplateResponse, error)
}

type service struct {
	mu        sync.RWMutex
	templates []TemplateResponse
}

func NewService() Service {
	return &service{templates: []TemplateResponse{{
		ID:        "template-1",
		Code:      "outline_seed",
		AgentCode: "outline_agent",
		Template:  "Write an outline for {{topic}}",
		Variables: []string{"topic"},
	}}}
}

func (s *service) ListTemplates(ctx context.Context, req ListTemplatesRequest) (PagedTemplatesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]TemplateResponse, 0, len(s.templates))
	for _, item := range s.templates {
		if req.AgentCode != "" && item.AgentCode != req.AgentCode {
			continue
		}
		items = append(items, item)
	}
	if req.Sort == "code" && req.Order == "asc" {
		sort.Slice(items, func(i, j int) bool { return items[i].Code < items[j].Code })
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	return PagedTemplatesResponse{Items: items, Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: len(items), HasNext: len(items) > page*pageSize}}, nil
}

func (s *service) CreateTemplate(ctx context.Context, req CreateTemplateRequest) (CreateTemplateResponse, error) {
	if req.Code == "" || req.Template == "" || req.Variables == nil {
		return CreateTemplateResponse{}, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.templates {
		if item.Code == req.Code {
			return CreateTemplateResponse{}, ErrConflict
		}
	}
	id := "prompt-template-new"
	s.templates = append(s.templates, TemplateResponse{ID: id, Code: req.Code, Template: req.Template, Variables: req.Variables})
	return CreateTemplateResponse{PromptTemplateID: id}, nil
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
