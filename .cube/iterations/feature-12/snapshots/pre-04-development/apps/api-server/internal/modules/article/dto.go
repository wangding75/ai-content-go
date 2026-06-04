package article

import (
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

// --- Article Pack Registration ---

type RegisterPackRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
}

type RegisterPackResponse struct {
	ContentPackID                  string   `json:"content_pack_id"`
	ContentTypeID                  string   `json:"content_type_id"`
	RegisteredWorkflowVersionIDs   []string `json:"registered_workflow_version_ids"`
	MetricTemplateIDs              []string `json:"metric_template_ids"`
}

type ArticlePackStatusResponse struct {
	Registered           bool                    `json:"registered"`
	ContentPackID        string                  `json:"content_pack_id,omitempty"`
	ContentType          *content.ContentTypeResponse `json:"content_type,omitempty"`
	DefaultWorkflow      *PackWorkflowSummary    `json:"default_workflow_template,omitempty"`
	DefaultMetrics       []PackMetricSummary     `json:"default_metrics,omitempty"`
}

type PackWorkflowSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type PackMetricSummary struct {
	MetricCode string `json:"metric_code"`
	Name       string `json:"name"`
	Unit       string `json:"unit"`
}

// --- Article Project Config ---

type ArticleConfigResponse struct {
	TopicStyle                        string   `json:"topic_style"`
	AudienceProfile                   string   `json:"audience_profile"`
	SEOConfig                         map[string]any `json:"seo_config"`
	SourcePolicy                      string   `json:"source_policy"`
	StructurePolicy                   string   `json:"structure_policy"`
	DefaultWorkflowTemplateVersionID  string   `json:"default_workflow_template_version_id"`
	EnabledMetricCodes                []string `json:"enabled_metric_codes"`
	Version                           string   `json:"version"`
}

type UpdateArticleConfigRequest struct {
	TopicStyle                       string         `json:"topic_style"`
	AudienceProfile                  string         `json:"audience_profile"`
	SEOConfig                        map[string]any `json:"seo_config"`
	SourcePolicy                     string         `json:"source_policy"`
	StructurePolicy                  string         `json:"structure_policy"`
	DefaultWorkflowTemplateVersionID string         `json:"default_workflow_template_version_id"`
}

type UpdateArticleConfigResponse struct {
	VersionID      string `json:"version_id"`
	OperationLogID string `json:"operation_log_id"`
}

// --- Article Generation Run ---

type CreateArticleGenerationRunRequest struct {
	Topic           string         `json:"topic"`
	Audience        string         `json:"audience"`
	SourceRefs      []string       `json:"source_refs"`
	SEOKeywords     []string       `json:"seo_keywords"`
	OutlineRequired bool           `json:"outline_required"`
	TargetPlatform  string         `json:"target_platform"`
	GenerationConfig map[string]any `json:"generation_config"`
}

type ListGenerationRunsRequest struct {
	content.PaginationRequest
	Status          string `json:"status"`
	Topic           string `json:"topic"`
	TargetPlatform  string `json:"target_platform"`
}

type ArticleGenerationRunSummary struct {
	GenerationRunID string `json:"generation_run_id"`
	WorkflowRunID   string `json:"workflow_run_id"`
	Status          string `json:"status"`
	Topic           string `json:"topic"`
	CreatedAt       string `json:"created_at"`
}

type PagedArticleGenerationRunResponse struct {
	Items      []ArticleGenerationRunSummary `json:"items"`
	Pagination content.PaginationResponse    `json:"pagination"`
}

type ArticleSnapshot struct {
	Title                 string   `json:"title"`
	Summary               string   `json:"summary"`
	Outline               string   `json:"outline"`
	SEOMetadata           map[string]any `json:"seo_metadata"`
	SourceRefs            []string `json:"source_refs"`
	LatestContentVersionID string  `json:"latest_content_version_id"`
}

type ArticleGenerationRunDetailResponse struct {
	ArticleGenerationRunSummary
	ArticleSnapshot  *ArticleSnapshot `json:"article_snapshot,omitempty"`
	ContentItemID    string           `json:"content_item_id,omitempty"`
	ContentVersionID string           `json:"content_version_id,omitempty"`
	Error            string           `json:"error,omitempty"`
}

type RetryGenerationRunRequest struct {
	Reason        string         `json:"reason"`
	InputOverride map[string]any `json:"input_override,omitempty"`
}

type ArticleContentSnapshotResponse struct {
	Title                  string         `json:"title"`
	Summary                string         `json:"summary"`
	Outline                string         `json:"outline"`
	SEOMetadata            map[string]any `json:"seo_metadata"`
	SourceRefs             []string       `json:"source_refs"`
	LatestContentVersionID string         `json:"latest_content_version_id"`
}

// --- Article Metrics Config ---

type ProjectArticleMetricItem struct {
	MetricCode string `json:"metric_code"`
	Name       string `json:"name"`
	Unit       string `json:"unit"`
	ValueType  string `json:"value_type"`
	Platform   string `json:"platform"`
	Enabled    bool   `json:"enabled"`
}

type PagedProjectArticleMetricsResponse struct {
	Items      []ProjectArticleMetricItem  `json:"items"`
	Pagination content.PaginationResponse  `json:"pagination"`
}

type UpdateProjectArticleMetricsRequest struct {
	EnabledMetricCodes []string              `json:"enabled_metric_codes"`
	PlatformOverrides  map[string][]string   `json:"platform_overrides,omitempty"`
	Note               string                `json:"note,omitempty"`
}

type UpdateProjectArticleMetricsResponse struct {
	VersionID      string `json:"version_id"`
	OperationLogID string `json:"operation_log_id"`
}
