package socialpost

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/metrics"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type Service interface {
	GetPackStatus(ctx context.Context) (SocialPostPackStatusResponse, error)
	RegisterPack(ctx context.Context, req RegisterSocialPostPackRequest, idempotencyKey string) (RegisterSocialPostPackResponse, error)

	GetConfig(ctx context.Context, projectID string) (SocialPostConfigResponse, error)
	UpdateConfig(ctx context.Context, projectID string, req UpdateSocialPostConfigRequest, idempotencyKey string) (UpdateSocialPostConfigResponse, error)

	CreateGenerationRun(ctx context.Context, projectID string, req CreateSocialPostGenerationRunRequest, idempotencyKey string) (CreateSocialPostGenerationRunResponse, error)
	GetGenerationRun(ctx context.Context, projectID, generationRunID string) (SocialPostGenerationRunDetailResponse, error)

	ListVariants(ctx context.Context, projectID string, req ListSocialPostVariantsRequest) (PagedSocialPostVariantsResponse, error)
	SelectVariant(ctx context.Context, projectID, variantID string, req SelectSocialPostVariantRequest, idempotencyKey string) (SelectSocialPostVariantResponse, error)

	GenerateTags(ctx context.Context, projectID string, req GenerateSocialPostTagsRequest, idempotencyKey string) (GenerateSocialPostAssetResponse, error)
	GenerateCoverCopy(ctx context.Context, projectID string, req GenerateSocialPostCoverCopyRequest, idempotencyKey string) (GenerateSocialPostAssetResponse, error)
	GetAssets(ctx context.Context, projectID string, req GetSocialPostAssetsRequest) (SocialPostAssetsResponse, error)
}

type service struct {
	mu         sync.RWMutex
	store      Store
	contentSvc content.Service
	wfSvc      workflow.Service
	metricsSvc metrics.Service
	submitter  engine.Submitter

	packRegistered bool
	packID         string
	contentTypeID  string
	wfVersionIDs   map[string]string // code -> versionID
	metricTmplIDs  []string
	genRunSeq      int
}

func NewService(stores ...Store) Service {
	var store Store
	if len(stores) > 0 {
		store = stores[0]
	} else {
		store = NewMemoryStore()
	}
	return &service{store: store}
}

func SetDependencies(svc Service, contentSvc content.Service, wfSvc workflow.Service, metricsSvc metrics.Service, submitter engine.Submitter) {
	if s, ok := svc.(*service); ok {
		s.contentSvc = contentSvc
		s.wfSvc = wfSvc
		s.metricsSvc = metricsSvc
		s.submitter = submitter
	}
}

func hashRequest(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h[:16])
}

func (s *service) GetPackStatus(ctx context.Context) (SocialPostPackStatusResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.packRegistered {
		return SocialPostPackStatusResponse{}, ErrNotFound
	}

	var ct *content.ContentTypeResponse
	if s.contentSvc != nil {
		ctList, err := s.contentSvc.ListContentTypes(ctx, content.ListContentTypesRequest{})
		if err == nil {
			for i := range ctList.Items {
				if ctList.Items[i].Code == "social_post" {
					ct = &ctList.Items[i]
					break
				}
			}
		}
	}

	var wfSummaries []PackWorkflowSummary
	if s.wfSvc != nil && len(s.wfVersionIDs) > 0 {
		templates, err := s.wfSvc.ListTemplates(ctx, workflow.ListWorkflowTemplatesRequest{ContentType: "social_post"})
		if err == nil {
			for _, t := range templates.Items {
				wfSummaries = append(wfSummaries, PackWorkflowSummary{
					TemplateID:     t.ID,
					Code:           t.Code,
					Name:           t.Name,
					CurrentVersion: t.ID,
				})
			}
		}
	}
	if wfSummaries == nil {
		wfSummaries = []PackWorkflowSummary{}
	}

	var metricSummaries []PackMetricSummary
	if s.metricsSvc != nil && len(s.metricTmplIDs) > 0 {
		var listReq metrics.ListMetricTemplatesRequest
		listReq.ContentType = "social_post"
		mtResp, err := s.metricsSvc.ListTemplates(ctx, listReq)
		if err == nil {
			for _, mt := range mtResp.Items {
				metricSummaries = append(metricSummaries, PackMetricSummary{
					MetricCode: mt.MetricCode,
					MetricName: mt.MetricName,
					Unit:       mt.Unit,
					Platform:   mt.Platform,
				})
			}
		}
	}
	if metricSummaries == nil {
		metricSummaries = []PackMetricSummary{}
	}

	return SocialPostPackStatusResponse{
		ContentPackID:  s.packID,
		ContentType:    ct,
		Schema:         map[string]any{"content_type_code": "social_post", "project_fields": []string{"target_platforms", "default_variant_count", "tone_style"}},
		Workflows:      wfSummaries,
		Metrics:        metricSummaries,
		CurrentVersion: "2026.06.social-post.v1",
	}, nil
}

func (s *service) RegisterPack(ctx context.Context, req RegisterSocialPostPackRequest, idempotencyKey string) (RegisterSocialPostPackResponse, error) {
	if req.Schema.ContentTypeCode == "" {
		return RegisterSocialPostPackResponse{}, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if idempotencyKey != "" {
		reqHash := hashRequest(req)
		refType, refID, conflict, err := s.store.CheckIdempotency(ctx, "social-post:pack", "register", idempotencyKey, reqHash)
		if err != nil {
			return RegisterSocialPostPackResponse{}, err
		}
		if conflict {
			return RegisterSocialPostPackResponse{}, ErrIdempotencyConflict
		}
		if refType != "" {
			return RegisterSocialPostPackResponse{
				ContentPackID:     refID,
				ContentTypeID:     refID,
				RegisteredVersion: req.Version,
			}, nil
		}
	}

	if s.packRegistered {
		return RegisterSocialPostPackResponse{
			ContentPackID:     s.packID,
			ContentTypeID:     s.contentTypeID,
			RegisteredVersion: req.Version,
		}, nil
	}

	var ctID string
	if s.contentSvc != nil {
		resp, err := s.contentSvc.CreateContentType(ctx, content.CreateContentTypeRequest{
			Code:          req.Schema.ContentTypeCode,
			Name:          req.Schema.Name,
			ProjectSchema: req.Schema.ProjectSchema,
		})
		if err != nil {
			return RegisterSocialPostPackResponse{}, err
		}
		ctID = resp.ContentTypeID
	} else {
		ctID = "ct-social-post"
	}
	s.contentTypeID = ctID

	var wfVersionIDs map[string]string
	if s.wfSvc != nil {
		wfVersionIDs = make(map[string]string)
		for _, wf := range req.Workflows {
			wfResp, err := s.wfSvc.CreateTemplate(ctx, workflow.CreateWorkflowTemplateRequest{
				Code:        wf.Code,
				Name:        wf.Name,
				ContentType: req.Schema.ContentTypeCode,
				Category:    "generation",
			})
			if err != nil {
				return RegisterSocialPostPackResponse{}, err
			}
			verResp, err := s.wfSvc.CreateVersion(ctx, wfResp.WorkflowTemplateID, workflow.CreateVersionRequest{
				InputSchema:  map[string]any{"topic": "string"},
				OutputSchema: map[string]any{"output": "string"},
				Steps:        []workflow.CreateStepTemplateRequest{},
			})
			if err != nil {
				return RegisterSocialPostPackResponse{}, err
			}
			wfVersionIDs[wf.Code] = verResp.TemplateVersionID
		}
	}
	s.wfVersionIDs = wfVersionIDs

	var mtIDs []string
	if s.metricsSvc != nil {
		for _, m := range req.Metrics {
			mtResp, err := s.metricsSvc.CreateTemplate(ctx, metrics.CreateMetricTemplateRequest{
				ContentType:       req.Schema.ContentTypeCode,
				Platform:          "generic",
				MetricCode:        m.MetricCode,
				MetricName:        m.MetricName,
				Unit:              m.Unit,
				ValueType:         metrics.ValueTypeInteger,
				AggregationMethod: metrics.AggregationSum,
				Period:            metrics.PeriodDay,
				Required:          false,
				Enabled:           true,
			})
			if err != nil {
				return RegisterSocialPostPackResponse{}, err
			}
			mtIDs = append(mtIDs, mtResp.MetricTemplateID)
		}
	}
	s.metricTmplIDs = mtIDs

	s.packID = "cp_social_post"
	s.packRegistered = true

	if idempotencyKey != "" {
		_ = s.store.StoreIdempotency(ctx, "social-post:pack", "register", idempotencyKey, hashRequest(req), "content_type", ctID)
	}

	return RegisterSocialPostPackResponse{
		ContentPackID:     s.packID,
		ContentTypeID:     ctID,
		RegisteredVersion: req.Version,
	}, nil
}

func (s *service) GetConfig(ctx context.Context, projectID string) (SocialPostConfigResponse, error) {
	row, err := s.store.GetExtensionByProjectID(ctx, projectID)
	if err != nil {
		if err == ErrNotFound {
			return SocialPostConfigResponse{
				TargetPlatforms:     []string{},
				DefaultVariantCount: 3,
				CaptionLengthPolicy: "short",
				HashtagPolicy:       map[string]any{"mode": "auto", "count": 5},
				CoverCopyPolicy:     map[string]any{"mode": "auto", "count": 2},
				ToneStyle:           "professional",
				ForbiddenTerms:      []string{},
				ConfigVersion:       1,
			}, nil
		}
		return SocialPostConfigResponse{}, err
	}

	hashtagPolicy := map[string]any{}
	json.Unmarshal([]byte(row.HashtagPolicy), &hashtagPolicy)
	coverCopyPolicy := map[string]any{}
	json.Unmarshal([]byte(row.CoverCopyPolicy), &coverCopyPolicy)
	targetPlatforms := []string{}
	json.Unmarshal([]byte(row.TargetPlatforms), &targetPlatforms)
	forbiddenTerms := []string{}
	json.Unmarshal([]byte(row.ForbiddenTerms), &forbiddenTerms)

	return SocialPostConfigResponse{
		TargetPlatforms:     targetPlatforms,
		DefaultVariantCount: row.DefaultVariantCount,
		CaptionLengthPolicy: row.CaptionLengthPolicy,
		HashtagPolicy:       hashtagPolicy,
		CoverCopyPolicy:     coverCopyPolicy,
		ToneStyle:           row.ToneStyle,
		ForbiddenTerms:      forbiddenTerms,
		ConfigVersion:       row.ConfigVersion,
	}, nil
}

func (s *service) UpdateConfig(ctx context.Context, projectID string, req UpdateSocialPostConfigRequest, idempotencyKey string) (UpdateSocialPostConfigResponse, error) {
	if len(req.TargetPlatforms) == 0 {
		return UpdateSocialPostConfigResponse{}, ErrValidation
	}

	reqHash := hashRequest(req)
	if idempotencyKey != "" {
		refType, refID, conflict, err := s.store.CheckIdempotency(ctx, "social-post:config:"+projectID, "patch-config", idempotencyKey, reqHash)
		if err != nil {
			return UpdateSocialPostConfigResponse{}, err
		}
		if conflict {
			return UpdateSocialPostConfigResponse{}, ErrIdempotencyConflict
		}
		if refType != "" {
			return UpdateSocialPostConfigResponse{VersionID: refID, OperationLogID: refType}, nil
		}
	}

	existing, err := s.store.GetExtensionByProjectID(ctx, projectID)
	configVersion := 1
	if err == nil {
		configVersion = existing.ConfigVersion + 1
	}

	targetPlatformsJSON, _ := json.Marshal(req.TargetPlatforms)
	hashtagPolicyJSON, _ := json.Marshal(req.HashtagPolicy)
	coverCopyPolicyJSON, _ := json.Marshal(req.CoverCopyPolicy)
	forbiddenTermsJSON, _ := json.Marshal(req.ForbiddenTerms)

	oplogID, err := s.store.InsertOperationLog(ctx, OperationLogRow{
		Actor:    "admin",
		Resource: "social_post_config",
		Reason:   "update config",
	})
	if err != nil {
		return UpdateSocialPostConfigResponse{}, err
	}

	versionID := fmt.Sprintf("social-post-config-%d", configVersion)
	err = s.store.UpsertExtension(ctx, SocialPostConfigRow{
		ID:                  versionID,
		ProjectID:           projectID,
		TargetPlatforms:     string(targetPlatformsJSON),
		DefaultVariantCount: req.DefaultVariantCount,
		CaptionLengthPolicy: req.CaptionLengthPolicy,
		HashtagPolicy:       string(hashtagPolicyJSON),
		CoverCopyPolicy:     string(coverCopyPolicyJSON),
		ToneStyle:           req.ToneStyle,
		ForbiddenTerms:      string(forbiddenTermsJSON),
		ConfigVersion:       configVersion,
		OperationLogID:      oplogID,
	})
	if err != nil {
		return UpdateSocialPostConfigResponse{}, err
	}

	if idempotencyKey != "" {
		_ = s.store.StoreIdempotency(ctx, "social-post:config:"+projectID, "patch-config", idempotencyKey, reqHash, oplogID, versionID)
	}

	return UpdateSocialPostConfigResponse{
		VersionID:      versionID,
		OperationLogID: oplogID,
	}, nil
}

func (s *service) CreateGenerationRun(ctx context.Context, projectID string, req CreateSocialPostGenerationRunRequest, idempotencyKey string) (CreateSocialPostGenerationRunResponse, error) {
	if req.Topic == "" || req.Platform == "" {
		return CreateSocialPostGenerationRunResponse{}, ErrValidation
	}
	if req.VersionCount > 10 || req.VersionCount < 1 {
		return CreateSocialPostGenerationRunResponse{}, ErrValidation
	}

	reqHash := hashRequest(req)
	if idempotencyKey != "" {
		refType, refID, conflict, err := s.store.CheckIdempotency(ctx, "social-post:generation:"+projectID, "create-run", idempotencyKey, reqHash)
		if err != nil {
			return CreateSocialPostGenerationRunResponse{}, err
		}
		if conflict {
			return CreateSocialPostGenerationRunResponse{}, ErrIdempotencyConflict
		}
		if refType != "" {
			return CreateSocialPostGenerationRunResponse{GenerationRunID: refID, WorkflowRunID: refType, Status: "running"}, nil
		}
	}

	s.mu.Lock()
	s.genRunSeq++
	seq := s.genRunSeq
	s.mu.Unlock()

	genRunID := fmt.Sprintf("genrun-social-%d", seq)
	wfRunID := fmt.Sprintf("wfr-social-%d", seq)

	if s.wfSvc != nil && s.submitter != nil {
		versionID := "social_post_generation"
		if s.wfVersionIDs != nil {
			if vid, ok := s.wfVersionIDs["social_post_generation"]; ok {
				versionID = vid
			}
		}
		wfResp, err := s.wfSvc.CreateRun(ctx, workflow.CreateWorkflowRunRequest{
			ProjectID:         projectID,
			TemplateVersionID: versionID,
			Input:             map[string]any{"topic": req.Topic, "platform": req.Platform},
		}, "")
		if err != nil {
			return CreateSocialPostGenerationRunResponse{}, err
		}
		wfRunID = wfResp.WorkflowRunID
		s.submitter.Submit(wfRunID)
	}

	_, _ = s.store.InsertOperationLog(ctx, OperationLogRow{
		Actor:    "admin",
		Resource: "social_post_generation",
		Reason:   fmt.Sprintf("trigger generation: %s", req.Topic),
	})

	if idempotencyKey != "" {
		_ = s.store.StoreIdempotency(ctx, "social-post:generation:"+projectID, "create-run", idempotencyKey, reqHash, wfRunID, genRunID)
	}

	return CreateSocialPostGenerationRunResponse{
		GenerationRunID: genRunID,
		WorkflowRunID:   wfRunID,
		Status:          "running",
	}, nil
}

func (s *service) GetGenerationRun(ctx context.Context, projectID, generationRunID string) (SocialPostGenerationRunDetailResponse, error) {
	variants, _, err := s.store.ListVariants(ctx, projectID, ListSocialPostVariantsRequest{Page: 1, PageSize: 100})
	if err != nil {
		variants = nil
	}

	var matchingVariants []SocialPostVariantResponse
	for _, v := range variants {
		if v.GenerationRunID == generationRunID {
			matchingVariants = append(matchingVariants, v)
		}
	}
	if matchingVariants == nil {
		matchingVariants = []SocialPostVariantResponse{}
	}

	status := "running"
	var trace *GenerationTrace
	if s.wfSvc != nil {
		run, err := s.wfSvc.GetRun(ctx, generationRunID)
		if err == nil {
			status = run.Status
			if status == "succeeded" {
				status = "completed"
			}
			if run.Error != "" {
				trace = &GenerationTrace{}
			}
		}
	}
	if trace == nil {
		trace = &GenerationTrace{}
	}

	return SocialPostGenerationRunDetailResponse{
		GenerationRunID: generationRunID,
		WorkflowRunID:   generationRunID,
		Status:          status,
		ContentItemID:   "content-item-1",
		Trace:           trace,
		Variants:        matchingVariants,
	}, nil
}

func (s *service) ListVariants(ctx context.Context, projectID string, req ListSocialPostVariantsRequest) (PagedSocialPostVariantsResponse, error) {
	items, total, err := s.store.ListVariants(ctx, projectID, req)
	if err != nil {
		return PagedSocialPostVariantsResponse{}, err
	}
	if items == nil {
		items = []SocialPostVariantResponse{}
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	hasNext := (page * pageSize) < total

	return PagedSocialPostVariantsResponse{
		Items: items,
		Pagination: content.PaginationResponse{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
			HasNext:  hasNext,
		},
	}, nil
}

func (s *service) SelectVariant(ctx context.Context, projectID, variantID string, req SelectSocialPostVariantRequest, idempotencyKey string) (SelectSocialPostVariantResponse, error) {
	if req.ContentItemID == "" {
		return SelectSocialPostVariantResponse{}, ErrValidation
	}

	reqHash := hashRequest(req)
	if idempotencyKey != "" {
		refType, refID, conflict, err := s.store.CheckIdempotency(ctx, "social-post:select:"+projectID, "select-variant", idempotencyKey, reqHash)
		if err != nil {
			return SelectSocialPostVariantResponse{}, err
		}
		if conflict {
			return SelectSocialPostVariantResponse{}, ErrIdempotencyConflict
		}
		if refType != "" {
			return SelectSocialPostVariantResponse{SelectedVariantID: variantID, ContentVersionID: refID, OperationLogID: refType}, nil
		}
	}

	contentVersionID, oplogID, err := s.store.SelectVariantInTx(ctx, SelectVariantTxInput{
		VariantID:      variantID,
		ContentItemID:  req.ContentItemID,
		ProjectID:      projectID,
		Note:           req.Note,
		IdempotencyKey: idempotencyKey,
		RequestHash:    reqHash,
	})
	if err != nil {
		return SelectSocialPostVariantResponse{}, err
	}

	if idempotencyKey != "" {
		_ = s.store.StoreIdempotency(ctx, "social-post:select:"+projectID, "select-variant", idempotencyKey, reqHash, oplogID, contentVersionID)
	}

	return SelectSocialPostVariantResponse{
		SelectedVariantID: variantID,
		ContentVersionID:  contentVersionID,
		OperationLogID:    oplogID,
	}, nil
}

func (s *service) GenerateTags(ctx context.Context, projectID string, req GenerateSocialPostTagsRequest, idempotencyKey string) (GenerateSocialPostAssetResponse, error) {
	if req.ContentItemID == "" || req.VariantID == "" || req.Platform == "" {
		return GenerateSocialPostAssetResponse{}, ErrValidation
	}

	reqHash := hashRequest(req)
	if idempotencyKey != "" {
		refType, refID, conflict, err := s.store.CheckIdempotency(ctx, "social-post:asset-tags:"+projectID, "generate-tags", idempotencyKey, reqHash)
		if err != nil {
			return GenerateSocialPostAssetResponse{}, err
		}
		if conflict {
			return GenerateSocialPostAssetResponse{}, ErrIdempotencyConflict
		}
		if refType != "" {
			return GenerateSocialPostAssetResponse{GenerationRunID: refID, WorkflowRunID: refType, Status: "running"}, nil
		}
	}

	s.mu.Lock()
	s.genRunSeq++
	seq := s.genRunSeq
	s.mu.Unlock()

	genRunID := fmt.Sprintf("asset-run-%d", seq)
	wfRunID := fmt.Sprintf("wfr-asset-%d", seq)

	if s.wfSvc != nil && s.submitter != nil {
		wfResp, err := s.wfSvc.CreateRun(ctx, workflow.CreateWorkflowRunRequest{
			TemplateVersionID: "social_post_tags",
			Input:             map[string]any{"variant_id": req.VariantID, "platform": req.Platform, "count": req.Count, "style": req.Style},
		}, "")
		if err == nil {
			wfRunID = wfResp.WorkflowRunID
			s.submitter.Submit(wfRunID)
		}
	}

	_, _ = s.store.InsertOperationLog(ctx, OperationLogRow{
		Actor:    "admin",
		Resource: "social_post_tags",
		Reason:   fmt.Sprintf("generate tags for variant %s", req.VariantID),
	})

	if idempotencyKey != "" {
		_ = s.store.StoreIdempotency(ctx, "social-post:asset-tags:"+projectID, "generate-tags", idempotencyKey, reqHash, wfRunID, genRunID)
	}

	return GenerateSocialPostAssetResponse{
		GenerationRunID: genRunID,
		WorkflowRunID:   wfRunID,
		Status:          "running",
	}, nil
}

func (s *service) GenerateCoverCopy(ctx context.Context, projectID string, req GenerateSocialPostCoverCopyRequest, idempotencyKey string) (GenerateSocialPostAssetResponse, error) {
	if req.ContentItemID == "" || req.VariantID == "" || req.Platform == "" {
		return GenerateSocialPostAssetResponse{}, ErrValidation
	}

	reqHash := hashRequest(req)
	if idempotencyKey != "" {
		refType, refID, conflict, err := s.store.CheckIdempotency(ctx, "social-post:asset-cover:"+projectID, "generate-cover", idempotencyKey, reqHash)
		if err != nil {
			return GenerateSocialPostAssetResponse{}, err
		}
		if conflict {
			return GenerateSocialPostAssetResponse{}, ErrIdempotencyConflict
		}
		if refType != "" {
			return GenerateSocialPostAssetResponse{GenerationRunID: refID, WorkflowRunID: refType, Status: "running"}, nil
		}
	}

	s.mu.Lock()
	s.genRunSeq++
	seq := s.genRunSeq
	s.mu.Unlock()

	genRunID := fmt.Sprintf("asset-run-%d", seq)
	wfRunID := fmt.Sprintf("wfr-asset-%d", seq)

	if s.wfSvc != nil && s.submitter != nil {
		wfResp, err := s.wfSvc.CreateRun(ctx, workflow.CreateWorkflowRunRequest{
			TemplateVersionID: "social_post_cover_copy",
			Input:             map[string]any{"variant_id": req.VariantID, "platform": req.Platform, "count": req.Count, "style": req.Style},
		}, "")
		if err == nil {
			wfRunID = wfResp.WorkflowRunID
			s.submitter.Submit(wfRunID)
		}
	}

	_, _ = s.store.InsertOperationLog(ctx, OperationLogRow{
		Actor:    "admin",
		Resource: "social_post_cover_copy",
		Reason:   fmt.Sprintf("generate cover copy for variant %s", req.VariantID),
	})

	if idempotencyKey != "" {
		_ = s.store.StoreIdempotency(ctx, "social-post:asset-cover:"+projectID, "generate-cover", idempotencyKey, reqHash, wfRunID, genRunID)
	}

	return GenerateSocialPostAssetResponse{
		GenerationRunID: genRunID,
		WorkflowRunID:   wfRunID,
		Status:          "running",
	}, nil
}

func (s *service) GetAssets(ctx context.Context, projectID string, req GetSocialPostAssetsRequest) (SocialPostAssetsResponse, error) {
	assets, err := s.store.ListAssets(ctx, projectID, req)
	if err != nil {
		return SocialPostAssetsResponse{}, err
	}
	if assets == nil {
		assets = []SocialPostAssetItem{}
	}

	var tags, coverCopy []SocialPostAssetItem
	var sourceRuns []string
	for _, a := range assets {
		if a.Result != nil {
			if _, ok := a.Result["tags"]; ok {
				tags = append(tags, a)
			} else {
				coverCopy = append(coverCopy, a)
			}
		}
		sourceRuns = append(sourceRuns, a.GenerationRunID)
	}
	if tags == nil {
		tags = []SocialPostAssetItem{}
	}
	if coverCopy == nil {
		coverCopy = []SocialPostAssetItem{}
	}
	if sourceRuns == nil {
		sourceRuns = []string{}
	}

	return SocialPostAssetsResponse{
		Tags:             tags,
		CoverCopy:        coverCopy,
		AssetSuggestions: []string{},
		SourceRuns:       sourceRuns,
	}, nil
}