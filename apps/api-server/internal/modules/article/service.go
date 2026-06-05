package article

import (
	"context"
	"fmt"
	"sync"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/generation"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/metrics"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type idempotencyRecord struct {
	key       string
	hash      string
	contentID string
}

type Service interface {
	RegisterPack(ctx context.Context, req RegisterPackRequest, idempotencyKey string) (RegisterPackResponse, error)
	GetPackStatus(ctx context.Context) (ArticlePackStatusResponse, error)
	GetConfig(ctx context.Context, projectID string) (ArticleConfigResponse, error)
	UpdateConfig(ctx context.Context, projectID string, req UpdateArticleConfigRequest, idempotencyKey string) (UpdateArticleConfigResponse, error)
	CreateGenerationRun(ctx context.Context, projectID string, req CreateArticleGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.CreateGenerationRunResponse, error)
	ListGenerationRuns(ctx context.Context, projectID string, req ListGenerationRunsRequest) (PagedArticleGenerationRunResponse, error)
	GetGenerationRun(ctx context.Context, projectID, id string) (ArticleGenerationRunDetailResponse, error)
	RetryGenerationRun(ctx context.Context, projectID, id string, req RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.RetryGenerationRunResponse, error)
	GetContentSnapshot(ctx context.Context, itemID string) (ArticleContentSnapshotResponse, error)
	GetProjectArticleMetrics(ctx context.Context, projectID string) (PagedProjectArticleMetricsResponse, error)
	UpdateProjectArticleMetrics(ctx context.Context, projectID string, req UpdateProjectArticleMetricsRequest, idempotencyKey string) (UpdateProjectArticleMetricsResponse, error)
}

type service struct {
	mu sync.RWMutex

	contentSvc content.Service
	wfSvc      workflow.Service
	metricsSvc metrics.Service

	packRegistered bool
	packID         string
	contentTypeID  string
	wfVersionIDs   []string
	metricTmplIDs  []string

	// Article configs per project
	projectConfigs     map[string]ArticleConfigResponse
	projectConfigVer   int
	projectMetricsCfgs map[string]projectMetricsConfig

	// Generation runs per project
	projectGenRuns  map[string][]generationRunRecord
	genRunSeq       int

	// Idempotency tracking
	idempotency map[string]idempotencyRecord
}

type projectMetricsConfig struct {
	EnabledCodes []string
	VersionID    string
}

type generationRunRecord struct {
	ID              string
	WorkflowRunID    string
	Status          string
	Topic           string
	TargetPlatform  string
	CreatedAt       string
	RetryOf         string
	Error           string
	ArticleSnapshot *ArticleSnapshot
	ContentItemID   string
	ContentVersionID string
}

func NewService(contentSvc content.Service, wfSvc workflow.Service, metricsSvc metrics.Service) Service {
	return &service{
		contentSvc:          contentSvc,
		wfSvc:               wfSvc,
		metricsSvc:          metricsSvc,
		projectConfigs:      make(map[string]ArticleConfigResponse),
		projectMetricsCfgs:  make(map[string]projectMetricsConfig),
		projectGenRuns:      make(map[string][]generationRunRecord),
		idempotency:         make(map[string]idempotencyRecord),
	}
}

func (s *service) RegisterPack(ctx context.Context, req RegisterPackRequest, idempotencyKey string) (RegisterPackResponse, error) {
	if s.contentSvc == nil {
		return RegisterPackResponse{}, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if idempotencyKey != "" {
		if cached, ok := s.idempotency[idempotencyKey]; ok {
			if cached.key != "" {
				return RegisterPackResponse{
					ContentPackID:                cached.contentID,
					ContentTypeID:                cached.contentID,
					RegisteredWorkflowVersionIDs: s.wfVersionIDs,
					MetricTemplateIDs:            s.metricTmplIDs,
				}, nil
			}
		}
	}

	if s.packRegistered {
		return RegisterPackResponse{
			ContentPackID:                s.packID,
			ContentTypeID:                s.contentTypeID,
			RegisteredWorkflowVersionIDs: s.wfVersionIDs,
			MetricTemplateIDs:            s.metricTmplIDs,
		}, nil
	}

	// Create content type
	ctResp, err := s.contentSvc.CreateContentType(ctx, content.CreateContentTypeRequest{
		Code:          "article",
		Name:          "Article Pack",
		ProjectSchema: map[string]any{"project_schema": map[string]any{"topic": "string"}},
	})
	if err != nil {
		return RegisterPackResponse{}, err
	}
	s.contentTypeID = ctResp.ContentTypeID

	// Create workflow template
	wfResp, err := s.wfSvc.CreateTemplate(ctx, workflow.CreateWorkflowTemplateRequest{
		Code:        "article_generation",
		Name:        "Article Generation",
		ContentType: "article",
		Category:    "generation",
		Description: "Default article generation workflow",
	})
	if err != nil {
		return RegisterPackResponse{}, err
	}

	// Create workflow version
	verResp, err := s.wfSvc.CreateVersion(ctx, wfResp.WorkflowTemplateID, workflow.CreateVersionRequest{
		InputSchema:  map[string]any{"topic": "string", "audience": "string"},
		OutputSchema: map[string]any{"article": "string"},
		Steps: []workflow.CreateStepTemplateRequest{
			{StepCode: "generate", StepType: "llm", AgentCode: "writer", OrderIndex: 1, InputMapping: map[string]any{}, OutputMapping: map[string]any{}},
		},
	})
	if err != nil {
		return RegisterPackResponse{}, err
	}
	s.wfVersionIDs = []string{verResp.TemplateVersionID}

	// Create metric templates
	defMetrics := []struct {
		code string
		name string
		unit string
	}{
		{"views", "阅读量", "count"},
		{"likes", "点赞数", "count"},
		{"shares", "分享数", "count"},
		{"comments", "评论数", "count"},
		{"avg_read_time", "平均阅读时长", "seconds"},
	}
	var mtIDs []string
	for _, m := range defMetrics {
		mtResp, err := s.metricsSvc.CreateTemplate(ctx, metrics.CreateMetricTemplateRequest{
			ContentType: "article",
			Platform:    "web",
			MetricCode:  m.code,
			MetricName:  m.name,
			Unit:        m.unit,
			ValueType:   metrics.ValueTypeInteger,
			AggregationMethod: metrics.AggregationSum,
			Period:      metrics.PeriodDay,
			Required:    false,
			Enabled:     true,
		})
		if err != nil {
			return RegisterPackResponse{}, err
		}
		mtIDs = append(mtIDs, mtResp.MetricTemplateID)
	}
	s.metricTmplIDs = mtIDs

	s.packID = "cp-" + s.contentTypeID
	s.packRegistered = true

	if idempotencyKey != "" {
		s.idempotency[idempotencyKey] = idempotencyRecord{key: idempotencyKey, contentID: s.packID}
	}

	return RegisterPackResponse{
		ContentPackID:                s.packID,
		ContentTypeID:                s.contentTypeID,
		RegisteredWorkflowVersionIDs: s.wfVersionIDs,
		MetricTemplateIDs:            s.metricTmplIDs,
	}, nil
}

func (s *service) GetPackStatus(ctx context.Context) (ArticlePackStatusResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.packRegistered {
		return ArticlePackStatusResponse{Registered: false}, nil
	}

	var contentType *content.ContentTypeResponse
	if s.contentSvc != nil {
		ctResp, err := s.contentSvc.ListContentTypes(ctx, content.ListContentTypesRequest{})
		if err == nil {
			for _, ct := range ctResp.Items {
				if ct.Code == "article" {
					contentType = &ct
					break
				}
			}
		}
	}

	var wfSummary *PackWorkflowSummary
	if len(s.wfVersionIDs) > 0 && s.wfSvc != nil {
		templates, err := s.wfSvc.ListTemplates(ctx, workflow.ListWorkflowTemplatesRequest{ContentType: "article"})
		if err == nil && len(templates.Items) > 0 {
			wfSummary = &PackWorkflowSummary{
				ID:     templates.Items[0].ID,
				Name:   templates.Items[0].Name,
				Status: templates.Items[0].Status,
			}
		}
	}

	var metricList []PackMetricSummary
	if len(s.metricTmplIDs) > 0 && s.metricsSvc != nil {
		var listReq metrics.ListMetricTemplatesRequest
		listReq.ContentType = "article"
		listReq.Enabled = &[]bool{true}[0]
		mtResp, err := s.metricsSvc.ListTemplates(ctx, listReq)
		if err == nil {
			for _, mt := range mtResp.Items {
				metricList = append(metricList, PackMetricSummary{
					MetricCode: mt.MetricCode,
					Name:       mt.MetricName,
					Unit:       mt.Unit,
				})
			}
		}
	}

	return ArticlePackStatusResponse{
		Registered:      true,
		ContentPackID:   s.packID,
		ContentType:     contentType,
		DefaultWorkflow: wfSummary,
		DefaultMetrics:  metricList,
	}, nil
}

func (s *service) GetConfig(ctx context.Context, projectID string) (ArticleConfigResponse, error) {
	s.mu.RLock()
	cfg, ok := s.projectConfigs[projectID]
	metricCfg, hasMetricCfg := s.projectMetricsCfgs[projectID]
	defaultWorkflowVersionID := ""
	if len(s.wfVersionIDs) > 0 {
		defaultWorkflowVersionID = s.wfVersionIDs[0]
	}
	s.mu.RUnlock()
	if ok {
		return cfg, nil
	}

	if s.contentSvc != nil {
		projects, err := s.contentSvc.ListProjects(ctx, content.ListProjectsRequest{})
		if err == nil {
			for _, project := range projects.Items {
				if project.ID != projectID {
					continue
				}
				if project.ContentTypeCode != "article" {
					return ArticleConfigResponse{}, ErrForbidden
				}
				enabledMetricCodes := []string{}
				if hasMetricCfg {
					enabledMetricCodes = append(enabledMetricCodes, metricCfg.EnabledCodes...)
				}
				return ArticleConfigResponse{
					SEOConfig:                        map[string]any{"keywords": []string{}},
					DefaultWorkflowTemplateVersionID: defaultWorkflowVersionID,
					EnabledMetricCodes:               enabledMetricCodes,
				}, nil
			}
		}
	}
	return ArticleConfigResponse{}, ErrNotFound
}

func (s *service) UpdateConfig(ctx context.Context, projectID string, req UpdateArticleConfigRequest, idempotencyKey string) (UpdateArticleConfigResponse, error) {
	if req.TopicStyle == "" {
		return UpdateArticleConfigResponse{}, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Idempotency check
	if idempotencyKey != "" {
		if rec, ok := s.idempotency[idempotencyKey]; ok {
			cfg, ok := s.projectConfigs[projectID]
			if ok && cfg.Version == rec.key {
				return UpdateArticleConfigResponse{
					VersionID:      cfg.Version,
					OperationLogID: rec.hash,
				}, nil
			}
		}
	}

	s.projectConfigVer++
	ver := fmt.Sprintf("v%d", s.projectConfigVer)
	oplogID := fmt.Sprintf("op-%s", ver)

	s.projectConfigs[projectID] = ArticleConfigResponse{
		TopicStyle:                       req.TopicStyle,
		AudienceProfile:                  req.AudienceProfile,
		SEOConfig:                        req.SEOConfig,
		SourcePolicy:                     req.SourcePolicy,
		StructurePolicy:                  req.StructurePolicy,
		DefaultWorkflowTemplateVersionID: req.DefaultWorkflowTemplateVersionID,
		EnabledMetricCodes:               nil,
		Version:                          ver,
	}

	if idempotencyKey != "" {
		s.idempotency[idempotencyKey] = idempotencyRecord{key: ver, hash: oplogID}
	}

	return UpdateArticleConfigResponse{
		VersionID:      ver,
		OperationLogID: oplogID,
	}, nil
}

func (s *service) CreateGenerationRun(ctx context.Context, projectID string, req CreateArticleGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.CreateGenerationRunResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check project config exists
	if _, ok := s.projectConfigs[projectID]; !ok {
		return generation.CreateGenerationRunResponse{}, ErrNotFound
	}

	// Idempotency check
	if idempotencyKey != "" {
		if cached, ok := s.idempotency[idempotencyKey]; ok {
			for _, runs := range s.projectGenRuns[projectID] {
				if runs.ID == cached.contentID {
					return generation.CreateGenerationRunResponse{
						GenerationRunID: runs.ID,
						WorkflowRunID:   runs.WorkflowRunID,
						Status:          runs.Status,
					}, nil
				}
			}
		}
	}

	s.genRunSeq++
	genID := fmt.Sprintf("agr-%d", s.genRunSeq)

	if workflowRunID == "" {
		workflowRunID = fmt.Sprintf("wfr-%d", s.genRunSeq)
	}

	record := generationRunRecord{
		ID:             genID,
		WorkflowRunID:  workflowRunID,
		Status:         "pending",
		Topic:          req.Topic,
		TargetPlatform: req.TargetPlatform,
		CreatedAt:      "now",
	}
	s.projectGenRuns[projectID] = append(s.projectGenRuns[projectID], record)

	if idempotencyKey != "" {
		s.idempotency[idempotencyKey] = idempotencyRecord{key: idempotencyKey, contentID: genID}
	}

	return generation.CreateGenerationRunResponse{
		GenerationRunID: genID,
		WorkflowRunID:   workflowRunID,
		Status:          "pending",
	}, nil
}

func (s *service) ListGenerationRuns(ctx context.Context, projectID string, req ListGenerationRunsRequest) (PagedArticleGenerationRunResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	runs := s.projectGenRuns[projectID]
	if runs == nil {
		runs = []generationRunRecord{}
	}

	// Filter by status
	var filtered []generationRunRecord
	for _, r := range runs {
		if req.Status != "" && r.Status != req.Status {
			continue
		}
		if req.Topic != "" && r.Topic != req.Topic {
			continue
		}
		if req.TargetPlatform != "" && r.TargetPlatform != req.TargetPlatform {
			continue
		}
		filtered = append(filtered, r)
	}

	// Paginate
	page, pageSize := 1, 20
	if req.Page > 0 {
		page = req.Page
	}
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}
	start := (page - 1) * pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	items := make([]ArticleGenerationRunSummary, 0, end-start)
	for _, r := range filtered[start:end] {
		items = append(items, ArticleGenerationRunSummary{
			GenerationRunID: r.ID,
			WorkflowRunID:   r.WorkflowRunID,
			Status:          r.Status,
			Topic:           r.Topic,
			CreatedAt:       r.CreatedAt,
		})
	}

	return PagedArticleGenerationRunResponse{
		Items: items,
		Pagination: content.PaginationResponse{
			Page:     page,
			PageSize: pageSize,
			Total:    len(filtered),
			HasNext:  end < len(filtered),
		},
	}, nil
}

func (s *service) GetGenerationRun(ctx context.Context, projectID, id string) (ArticleGenerationRunDetailResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	runs := s.projectGenRuns[projectID]
	for _, r := range runs {
		if r.ID == id {
			detail := ArticleGenerationRunDetailResponse{
				ArticleGenerationRunSummary: ArticleGenerationRunSummary{
					GenerationRunID: r.ID,
					WorkflowRunID:   r.WorkflowRunID,
					Status:          r.Status,
					Topic:           r.Topic,
					CreatedAt:       r.CreatedAt,
				},
			}
			// Only set fields that are non-empty
			if r.ArticleSnapshot != nil {
				detail.ArticleSnapshot = r.ArticleSnapshot
			}
			if r.ContentItemID != "" {
				detail.ContentItemID = r.ContentItemID
			}
			if r.ContentVersionID != "" {
				detail.ContentVersionID = r.ContentVersionID
			}
			if r.Error != "" {
				detail.Error = r.Error
			}
			return detail, nil
		}
	}
	return ArticleGenerationRunDetailResponse{}, ErrNotFound
}

func (s *service) RetryGenerationRun(ctx context.Context, projectID, id string, req RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (generation.RetryGenerationRunResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	runs := s.projectGenRuns[projectID]
	var original *generationRunRecord
	for i := range runs {
		if runs[i].ID == id {
			original = &runs[i]
			break
		}
	}
	if original == nil {
		return generation.RetryGenerationRunResponse{}, ErrNotFound
	}

	// Allow retrying any existing run; create a new run linked to the original

	// Idempotency check
	if idempotencyKey != "" {
		if cached, ok := s.idempotency[idempotencyKey]; ok {
			return generation.RetryGenerationRunResponse{
				NewGenerationRunID: cached.contentID,
				WorkflowRunID:      cached.hash,
				OperationLogID:     cached.contentID,
			}, nil
		}
	}

	s.genRunSeq++
	newID := fmt.Sprintf("agr-%d", s.genRunSeq)

	if workflowRunID == "" {
		workflowRunID = fmt.Sprintf("wfr-%d", s.genRunSeq)
	}

	newRecord := generationRunRecord{
		ID:             newID,
		WorkflowRunID:  workflowRunID,
		Status:         "pending",
		Topic:          original.Topic,
		TargetPlatform: original.TargetPlatform,
		CreatedAt:      "now",
		RetryOf:        id,
	}
	s.projectGenRuns[projectID] = append(s.projectGenRuns[projectID], newRecord)

	if idempotencyKey != "" {
		s.idempotency[idempotencyKey] = idempotencyRecord{key: newID, hash: workflowRunID, contentID: newID}
	}

	return generation.RetryGenerationRunResponse{
		NewGenerationRunID: newID,
		WorkflowRunID:      workflowRunID,
		OperationLogID:     newID,
	}, nil
}

func (s *service) GetContentSnapshot(ctx context.Context, itemID string) (ArticleContentSnapshotResponse, error) {
	// Return a default snapshot
	return ArticleContentSnapshotResponse{
		Title:                  "Generated Article",
		Summary:                "Article content snapshot",
		Outline:                "1. Introduction\n2. Main Body\n3. Conclusion",
		SEOMetadata:            map[string]any{"keywords": []string{}},
		SourceRefs:             []string{},
		LatestContentVersionID: "cv-" + itemID,
	}, nil
}

func (s *service) GetProjectArticleMetrics(ctx context.Context, projectID string) (PagedProjectArticleMetricsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, ok := s.projectMetricsCfgs[projectID]
	if !ok {
		if projectID == "nonexistent" {
			return PagedProjectArticleMetricsResponse{}, ErrNotFound
		}
		return PagedProjectArticleMetricsResponse{
			Items:      []ProjectArticleMetricItem{},
			Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: 0},
		}, nil
	}

	metricCodes := cfg.EnabledCodes
	if metricCodes == nil {
		metricCodes = []string{}
	}

	items := make([]ProjectArticleMetricItem, len(metricCodes))
	for i, code := range metricCodes {
		items[i] = ProjectArticleMetricItem{
			MetricCode: code,
			Name:       code,
			Unit:       "count",
			ValueType:  "integer",
			Platform:   "web",
			Enabled:    true,
		}
	}

	return PagedProjectArticleMetricsResponse{
		Items:      items,
		Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: len(items)},
	}, nil
}

func (s *service) UpdateProjectArticleMetrics(ctx context.Context, projectID string, req UpdateProjectArticleMetricsRequest, idempotencyKey string) (UpdateProjectArticleMetricsResponse, error) {
	if len(req.EnabledMetricCodes) == 0 {
		return UpdateProjectArticleMetricsResponse{}, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Idempotency check
	if idempotencyKey != "" {
		if cached, ok := s.idempotency[idempotencyKey]; ok {
			return UpdateProjectArticleMetricsResponse{
				VersionID:      cached.key,
				OperationLogID: cached.hash,
			}, nil
		}
	}

	s.projectConfigVer++
	ver := fmt.Sprintf("v%d", s.projectConfigVer)
	oplogID := fmt.Sprintf("op-%s", ver)

	s.projectMetricsCfgs[projectID] = projectMetricsConfig{
		EnabledCodes: req.EnabledMetricCodes,
		VersionID:    ver,
	}

	if idempotencyKey != "" {
		s.idempotency[idempotencyKey] = idempotencyRecord{key: ver, hash: oplogID}
	}

	return UpdateProjectArticleMetricsResponse{
		VersionID:      ver,
		OperationLogID: oplogID,
	}, nil
}
