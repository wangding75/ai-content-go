package article

import (
	"context"
	"testing"
)

// @Test
func TestTask03GetConfigReturnsArticleConfig(t *testing.T) {
	// Use NewService instead of directly accessing unexported fields
	svc := NewService(nil, nil, nil)
	// Setup config through the public UpdateConfig API
	_, err := svc.UpdateConfig(context.Background(), "project-1", UpdateArticleConfigRequest{
		TopicStyle:                      "tech",
		AudienceProfile:                 "developers",
		SEOConfig:                       map[string]any{"keywords": []string{"golang"}},
		SourcePolicy:                    "web",
		StructurePolicy:                 "standard",
		DefaultWorkflowTemplateVersionID: "wftv-1",
	}, "idem-init-cfg")
	if err != nil {
		t.Fatalf("setup config: %v", err)
	}

	got, err := svc.GetConfig(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.TopicStyle != "tech" {
		t.Fatalf("expected topic_style tech, got %s", got.TopicStyle)
	}
	if got.Version == "" {
		t.Fatal("expected non-empty version")
	}
}

// @Test
func TestTask03GetConfigReturnsNotFoundForMissingProject(t *testing.T) {
	svc := NewService(nil, nil, nil)
	_, err := svc.GetConfig(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing project")
	}
}

// @Test
func TestTask03UpdateConfigReturnsVersionAndOpLog(t *testing.T) {
	svc := NewService(nil, nil, nil)
	resp, err := svc.UpdateConfig(context.Background(), "project-1", UpdateArticleConfigRequest{
		TopicStyle:  "finance",
		SourcePolicy: "newsletter",
	}, "idem-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.VersionID == "" {
		t.Fatal("expected non-empty version_id")
	}
	if resp.OperationLogID == "" {
		t.Fatal("expected non-empty operation_log_id")
	}
}

// @Test
func TestTask03UpdateConfigModifiesConfig(t *testing.T) {
	svc := NewService(nil, nil, nil)
	_, err := svc.UpdateConfig(context.Background(), "project-1", UpdateArticleConfigRequest{
		TopicStyle:   "science",
		SourcePolicy: "research",
	}, "idem-upd1")
	if err != nil {
		t.Fatalf("update: expected no error, got %v", err)
	}
	got, err := svc.GetConfig(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("get: expected no error, got %v", err)
	}
	if got.TopicStyle != "science" {
		t.Fatalf("expected topic_style science, got %s", got.TopicStyle)
	}
	if got.SourcePolicy != "research" {
		t.Fatalf("expected source_policy research, got %s", got.SourcePolicy)
	}
}

// @Test
func TestTask03UpdateConfigIsIdempotent(t *testing.T) {
	svc := NewService(nil, nil, nil)
	resp1, err := svc.UpdateConfig(context.Background(), "project-1", UpdateArticleConfigRequest{
		TopicStyle: "design",
	}, "idem-same-key")
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	resp2, err := svc.UpdateConfig(context.Background(), "project-1", UpdateArticleConfigRequest{
		TopicStyle: "design",
	}, "idem-same-key")
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if resp1.VersionID != resp2.VersionID {
		t.Fatal("expected same version_id for idempotent update")
	}
}

// @Test
func TestTask03UpdateConfigReturnsValidationErrorForMissingTopicStyle(t *testing.T) {
	svc := NewService(nil, nil, nil)
	_, err := svc.UpdateConfig(context.Background(), "project-1", UpdateArticleConfigRequest{}, "")
	if err == nil {
		t.Fatal("expected validation error for empty request")
	}
}

// @Test
func TestTask03GetConfigReturnsUpdatedConfigAfterUpdate(t *testing.T) {
	svc := NewService(nil, nil, nil)
	svc.UpdateConfig(context.Background(), "project-1", UpdateArticleConfigRequest{
		TopicStyle:      "health",
		AudienceProfile: "patients",
		SEOConfig:       map[string]any{"keywords": []string{"wellness"}},
		SourcePolicy:    "medical",
		StructurePolicy: "tutorial",
	}, "idem-cfg-1")

	got, err := svc.GetConfig(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.AudienceProfile != "patients" {
		t.Fatalf("expected audience_profile patients, got %s", got.AudienceProfile)
	}
	if got.SEOConfig == nil {
		t.Fatal("expected seo_config to be stored")
	}
}

// @Test
func TestTask03GetConfigForDifferentProjectsReturnsDifferentConfigs(t *testing.T) {
	svc := NewService(nil, nil, nil)
	svc.UpdateConfig(context.Background(), "project-a", UpdateArticleConfigRequest{TopicStyle: "tech"}, "idem-a")
	svc.UpdateConfig(context.Background(), "project-b", UpdateArticleConfigRequest{TopicStyle: "finance"}, "idem-b")

	cfgA, _ := svc.GetConfig(context.Background(), "project-a")
	cfgB, _ := svc.GetConfig(context.Background(), "project-b")
	if cfgA.TopicStyle == cfgB.TopicStyle {
		t.Fatal("expected different configs for different projects")
	}
}
