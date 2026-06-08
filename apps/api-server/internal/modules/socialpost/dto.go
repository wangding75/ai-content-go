package socialpost

import "github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"

// --- Social Post Pack Registration ---

type RegisterSocialPostPackRequest struct {
	Schema    SocialPostPackSchema    `json:"schema"`
	Workflows []SocialPostPackWorkflow `json:"workflows"`
	Metrics   []SocialPostPackMetric   `json:"metrics"`
	Version   string                   `json:"version"`
}

type SocialPostPackSchema struct {
	ContentTypeCode string         `json:"content_type_code"`
	Name            string         `json:"name"`
	ProjectSchema   map[string]any `json:"project_schema"`
}

type SocialPostPackWorkflow struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type SocialPostPackMetric struct {
	MetricCode string `json:"metric_code"`
	MetricName string `json:"metric_name"`
	Unit       string `json:"unit"`
}

type RegisterSocialPostPackResponse struct {
	ContentPackID     string `json:"content_pack_id"`
	ContentTypeID     string `json:"content_type_id"`
	RegisteredVersion string `json:"registered_version"`
}

type SocialPostPackStatusResponse struct {
	ContentPackID string                       `json:"content_pack_id"`
	ContentType   *content.ContentTypeResponse `json:"content_type,omitempty"`
	Schema        map[string]any               `json:"schema"`
	Workflows     []PackWorkflowSummary        `json:"workflows"`
	Metrics       []PackMetricSummary          `json:"metrics"`
	CurrentVersion string                      `json:"current_version"`
}

type PackWorkflowSummary struct {
	TemplateID     string `json:"template_id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	CurrentVersion string `json:"current_version"`
}

type PackMetricSummary struct {
	MetricCode string `json:"metric_code"`
	MetricName string `json:"metric_name"`
	Unit       string `json:"unit"`
	Platform   string `json:"platform"`
}

// --- Social Post Project Config ---

type SocialPostConfigResponse struct {
	TargetPlatforms     []string       `json:"target_platforms"`
	DefaultVariantCount int            `json:"default_variant_count"`
	CaptionLengthPolicy string         `json:"caption_length_policy"`
	HashtagPolicy       map[string]any `json:"hashtag_policy"`
	CoverCopyPolicy     map[string]any `json:"cover_copy_policy"`
	ToneStyle           string         `json:"tone_style"`
	ForbiddenTerms      []string       `json:"forbidden_terms"`
	ConfigVersion       int            `json:"config_version"`
}

type UpdateSocialPostConfigRequest struct {
	TargetPlatforms     []string       `json:"target_platforms"`
	DefaultVariantCount int            `json:"default_variant_count"`
	CaptionLengthPolicy string         `json:"caption_length_policy"`
	HashtagPolicy       map[string]any `json:"hashtag_policy"`
	CoverCopyPolicy     map[string]any `json:"cover_copy_policy"`
	ToneStyle           string         `json:"tone_style"`
	ForbiddenTerms      []string       `json:"forbidden_terms"`
}

type UpdateSocialPostConfigResponse struct {
	VersionID      string `json:"version_id"`
	OperationLogID string `json:"operation_log_id"`
}

// --- Social Post Generation Run ---

type CreateSocialPostGenerationRunRequest struct {
	Topic               string               `json:"topic"`
	SourceContentItemID string               `json:"source_content_item_id"`
	Platform            string               `json:"platform"`
	VersionCount        int                  `json:"version_count"`
	ToneStyle           string               `json:"tone_style"`
	AssetOptions        SocialPostAssetOptions `json:"asset_options"`
}

type SocialPostAssetOptions struct {
	GenerateTags      bool `json:"generate_tags"`
	GenerateCoverCopy bool `json:"generate_cover_copy"`
}

type CreateSocialPostGenerationRunResponse struct {
	GenerationRunID string `json:"generation_run_id"`
	WorkflowRunID   string `json:"workflow_run_id"`
	Status          string `json:"status"`
}

type SocialPostGenerationRunDetailResponse struct {
	GenerationRunID string                      `json:"generation_run_id"`
	WorkflowRunID   string                      `json:"workflow_run_id"`
	Status          string                      `json:"status"`
	ContentItemID   string                      `json:"content_item_id"`
	Trace           *GenerationTrace            `json:"trace,omitempty"`
	Variants        []SocialPostVariantResponse `json:"variants"`
	Error           string                      `json:"error"`
}

type GenerationTrace struct {
	AgentTaskIDs  []string `json:"agent_task_ids"`
	LLMCallLogIDs []string `json:"llm_call_log_ids"`
}

// --- Social Post Variants ---

type ListSocialPostVariantsRequest struct {
	ContentItemID string
	Status        string
	Platform      string
	Page          int
	PageSize      int
}

type SocialPostVariantResponse struct {
	ID               string   `json:"id"`
	ContentItemID    string   `json:"content_item_id"`
	GenerationRunID  string   `json:"generation_run_id"`
	VariantIndex     int      `json:"variant_index"`
	Platform         string   `json:"platform"`
	Title            string   `json:"title"`
	Body             string   `json:"body"`
	Hashtags         []string `json:"hashtags"`
	CoverCopy        string   `json:"cover_copy"`
	ToneStyle        string   `json:"tone_style"`
	Status           string   `json:"status"`
	ContentVersionID string   `json:"content_version_id"`
	CreatedAt        string   `json:"created_at"`
}

type PagedSocialPostVariantsResponse struct {
	Items      []SocialPostVariantResponse  `json:"items"`
	Pagination content.PaginationResponse   `json:"pagination"`
}

type SelectSocialPostVariantRequest struct {
	ContentItemID string `json:"content_item_id"`
	Note          string `json:"note"`
}

type SelectSocialPostVariantResponse struct {
	SelectedVariantID string `json:"selected_variant_id"`
	ContentVersionID  string `json:"content_version_id"`
	OperationLogID    string `json:"operation_log_id"`
}

// --- Social Post Assets ---

type GenerateSocialPostTagsRequest struct {
	ContentItemID string `json:"content_item_id"`
	VariantID     string `json:"variant_id"`
	Platform      string `json:"platform"`
	Count         int    `json:"count"`
	Style         string `json:"style"`
}

type GenerateSocialPostCoverCopyRequest struct {
	ContentItemID string `json:"content_item_id"`
	VariantID     string `json:"variant_id"`
	Platform      string `json:"platform"`
	Count         int    `json:"count"`
	Style         string `json:"style"`
}

type GenerateSocialPostAssetResponse struct {
	GenerationRunID string `json:"generation_run_id"`
	WorkflowRunID   string `json:"workflow_run_id"`
	Status          string `json:"status"`
}

type GetSocialPostAssetsRequest struct {
	ContentItemID string
	Platform      string
	VariantID     string
}

type SocialPostAssetItem struct {
	ID              string `json:"id"`
	Platform        string `json:"platform"`
	SourceVariantID string `json:"source_variant_id"`
	GenerationRunID string `json:"generation_run_id"`
	Result          map[string]any `json:"result"`
	CreatedAt       string `json:"created_at"`
}

type SocialPostAssetsResponse struct {
	Tags             []SocialPostAssetItem `json:"tags"`
	CoverCopy        []SocialPostAssetItem `json:"cover_copy"`
	AssetSuggestions []string              `json:"asset_suggestions"`
	SourceRuns       []string              `json:"source_runs"`
}