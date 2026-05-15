package content

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"sync"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
)

type Service interface {
	ListContentTypes(ctx context.Context, req ListContentTypesRequest) (PagedContentTypesResponse, error)
	CreateContentType(ctx context.Context, req CreateContentTypeRequest) (CreateContentTypeResponse, error)
	ProjectSchema(ctx context.Context, id string) (ProjectSchemaResponse, error)
	ListProjects(ctx context.Context, req ListProjectsRequest) (PagedProjectsResponse, error)
	CreateProject(ctx context.Context, req CreateProjectRequest) (CreateProjectResponse, error)
	ProjectOverview(ctx context.Context, id string) (ProjectOverviewResponse, error)
	PauseProject(ctx context.Context, id string, req PauseProjectRequest) (PauseProjectResponse, error)
}

type service struct {
	mu               sync.RWMutex
	contentTypes     []ContentTypeResponse
	projects         []ProjectResponse
	operationLogNext int
}

func NewService() Service {
	return &service{
		contentTypes: []ContentTypeResponse{{
			ID:            "1",
			Code:          "blog",
			Name:          "Blog",
			ProjectSchema: map[string]any{"project_schema": map[string]any{"title": "string"}},
			Enabled:       true,
		}, {
			ID:            "seed-content-type",
			Code:          "seed_article",
			Name:          "Article Pack",
			ProjectSchema: map[string]any{"project_schema": map[string]any{"topic": "string"}},
			Enabled:       true,
		}},
		projects: []ProjectResponse{{
			ID:              "project-1",
			Name:            "Daily Articles",
			ContentTypeID:   "1",
			ContentTypeCode: "article",
			Status:          "active",
			ProjectConfig:   map[string]any{"title": "daily briefing"},
		}, {
			ID:              "seed-project",
			Name:            "Seed Project",
			ContentTypeID:   "seed-content-type",
			ContentTypeCode: "article",
			Status:          "active",
			ProjectConfig:   map[string]any{"topic": "seed"},
		}, {
			ID:              "paused-project",
			Name:            "Paused Articles",
			ContentTypeID:   "1",
			ContentTypeCode: "article",
			Status:          "paused",
			ProjectConfig:   map[string]any{"title": "paused briefing"},
		}},
		operationLogNext: 1,
	}
}

func (s *service) ListContentTypes(ctx context.Context, req ListContentTypesRequest) (PagedContentTypesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]ContentTypeResponse, 0, len(s.contentTypes))
	for _, item := range s.contentTypes {
		if req.Enabled != nil && item.Enabled != *req.Enabled {
			continue
		}
		items = append(items, item)
	}
	if req.Sort == "code" && req.Order == "asc" {
		sort.Slice(items, func(i, j int) bool { return items[i].Code < items[j].Code })
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	return PagedContentTypesResponse{Items: items, Pagination: PaginationResponse{Page: page, PageSize: pageSize, Total: len(items), HasNext: len(items) > page*pageSize}}, nil
}

func (s *service) CreateContentType(ctx context.Context, req CreateContentTypeRequest) (CreateContentTypeResponse, error) {
	if req.Code == "" || req.Name == "" || req.ProjectSchema == nil {
		return CreateContentTypeResponse{}, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.contentTypes {
		if item.Code == req.Code {
			return CreateContentTypeResponse{}, ErrConflict
		}
	}
	id := "content-type-new"
	s.contentTypes = append(s.contentTypes, ContentTypeResponse{ID: id, Code: req.Code, Name: req.Name, ProjectSchema: req.ProjectSchema, Enabled: true})
	return CreateContentTypeResponse{ContentTypeID: id}, nil
}

func (s *service) ProjectSchema(ctx context.Context, id string) (ProjectSchemaResponse, error) {
	if id == "" {
		return ProjectSchemaResponse{}, ErrValidation
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.contentTypes {
		if item.ID == id {
			return ProjectSchemaResponse{ContentTypeID: id, ProjectSchema: item.ProjectSchema}, nil
		}
	}
	return ProjectSchemaResponse{}, ErrNotFound
}

func (s *service) ListProjects(ctx context.Context, req ListProjectsRequest) (PagedProjectsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]ProjectResponse, 0, len(s.projects))
	for _, item := range s.projects {
		if req.Status != "" && item.Status != req.Status {
			continue
		}
		if req.ContentType != "" && item.ContentTypeCode != req.ContentType {
			continue
		}
		items = append(items, item)
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	return PagedProjectsResponse{Items: items, Pagination: PaginationResponse{Page: page, PageSize: pageSize, Total: len(items), HasNext: len(items) > page*pageSize}}, nil
}

func (s *service) CreateProject(ctx context.Context, req CreateProjectRequest) (CreateProjectResponse, error) {
	if req.Name == "" || req.ContentTypeID == "" || req.ProjectConfig == nil {
		return CreateProjectResponse{}, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	contentType, ok := s.findContentTypeLocked(req.ContentTypeID)
	if !ok {
		return CreateProjectResponse{}, ErrNotFound
	}
	id := "project-new"
	s.projects = append(s.projects, ProjectResponse{ID: id, Name: req.Name, ContentTypeID: req.ContentTypeID, ContentTypeCode: contentType.Code, Status: "active", ProjectConfig: req.ProjectConfig})
	return CreateProjectResponse{ProjectID: id, Status: "active"}, nil
}

func (s *service) ProjectOverview(ctx context.Context, id string) (ProjectOverviewResponse, error) {
	if id == "" {
		return ProjectOverviewResponse{}, ErrValidation
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.findProjectLocked(id); !ok {
		return ProjectOverviewResponse{}, ErrNotFound
	}
	return ProjectOverviewResponse{ProjectID: id, Progress: 0, PendingActions: 0, Cost: 0}, nil
}

func (s *service) PauseProject(ctx context.Context, id string, req PauseProjectRequest) (PauseProjectResponse, error) {
	if id == "" || req.Reason == "" {
		return PauseProjectResponse{}, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	index := -1
	for i, item := range s.projects {
		if item.ID == id {
			index = i
			break
		}
	}
	if index == -1 {
		return PauseProjectResponse{}, ErrNotFound
	}
	if s.projects[index].Status == "paused" {
		return PauseProjectResponse{}, ErrConflict
	}
	s.projects[index].Status = "paused"
	operationLogID := "operation-log-" + strconv.Itoa(s.operationLogNext)
	s.operationLogNext++
	return PauseProjectResponse{ProjectID: id, Status: "paused", OperationLogID: operationLogID}, nil
}

func (s *service) findContentTypeLocked(id string) (ContentTypeResponse, bool) {
	for _, item := range s.contentTypes {
		if item.ID == id {
			return item, true
		}
	}
	return ContentTypeResponse{}, false
}

func (s *service) findProjectLocked(id string) (ProjectResponse, bool) {
	for _, item := range s.projects {
		if item.ID == id {
			return item, true
		}
	}
	return ProjectResponse{}, false
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
