package article

import (
	"context"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/generation"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/metrics"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

// Integration test: ArticleService -> ContentService (Task-02: RegisterPack creates ContentType with code=article)
// Component chain: ArticleService.RegisterPack -> content.Service.CreateContentType
// Covers: Task-02 (Article Pack registration)
// @Test
func TestTask02IntegrationRegisterPackCreatesContentType(t *testing.T) {
	contentSvc := content.NewService()
	wfSvc := workflow.NewService()
	metricsSvc := metrics.NewService()
	svc := NewService(contentSvc, wfSvc, metricsSvc)

	_, err := svc.RegisterPack(context.Background(), RegisterPackRequest{}, "idem-int-ct")
	if err != nil {
		t.Fatalf("register: expected no error, got %v", err)
	}

	ctResp, err := contentSvc.ListContentTypes(context.Background(), content.ListContentTypesRequest{})
	if err != nil {
		t.Fatalf("list content types: %v", err)
	}
	var found bool
	for _, ct := range ctResp.Items {
		if ct.Code == "article" {
			found = true
			if ct.ID == "" {
				t.Fatal("expected non-empty content type ID")
			}
			break
		}
	}
	if !found {
		t.Fatal("expected article content type to be created")
	}
}

// Integration test: ArticleService -> WorkflowService (Task-02: RegisterPack creates default workflow templates)
// Component chain: ArticleService.RegisterPack -> workflow.Service.CreateTemplate + CreateVersion
// Covers: Task-02 (Article Pack registration)
// @Test
func TestTask02IntegrationRegisterPackCreatesWorkflowTemplates(t *testing.T) {
	contentSvc := content.NewService()
	wfSvc := workflow.NewService()
	metricsSvc := metrics.NewService()
	svc := NewService(contentSvc, wfSvc, metricsSvc)

	_, err := svc.RegisterPack(context.Background(), RegisterPackRequest{}, "idem-int-wf")
	if err != nil {
		t.Fatalf("register: expected no error, got %v", err)
	}

	templates, err := wfSvc.ListTemplates(context.Background(), workflow.ListWorkflowTemplatesRequest{ContentType: "article"})
	if err != nil {
		t.Fatalf("list templates: %v", err)
	}
	if len(templates.Items) == 0 {
		t.Fatal("expected at least one workflow template with content_type article")
	}
	_ = templates.Items[0].ID
}

// Integration test: ArticleService -> MetricsService (Task-02: RegisterPack creates default metric templates)
// Component chain: ArticleService.RegisterPack -> metrics.Service.CreateTemplate
// Covers: Task-02 (Article Pack registration)
// @Test
func TestTask02IntegrationRegisterPackCreatesMetricTemplates(t *testing.T) {
	contentSvc := content.NewService()
	wfSvc := workflow.NewService()
	metricsSvc := metrics.NewService()
	svc := NewService(contentSvc, wfSvc, metricsSvc)

	_, err := svc.RegisterPack(context.Background(), RegisterPackRequest{}, "idem-int-mt")
	if err != nil {
		t.Fatalf("register: expected no error, got %v", err)
	}

	mtResp, err := metricsSvc.ListTemplates(context.Background(), metrics.ListMetricTemplatesRequest{})
	if err != nil {
		t.Fatalf("list metric templates: %v", err)
	}
	if len(mtResp.Items) == 0 {
		t.Fatal("expected at least one metric template to be created")
	}
}

// Integration test: ArticleService -> GenerationService (Task-04: CreateGenerationRun creates WorkflowRun + GenerationRun)
// Component chain: ArticleService.CreateGenerationRun -> workflow.Service.CreateRun + generation.Service
// Covers: Task-04 (Article generation run)
// @Test
func TestTask04IntegrationCreateGenerationRunWithWorkflow(t *testing.T) {
	contentSvc := content.NewService()
	wfSvc := workflow.NewService()
	metricsSvc := metrics.NewService()
	svc := NewService(contentSvc, wfSvc, metricsSvc)

	// Setup: register pack and configure project
	packResp, err := svc.RegisterPack(context.Background(), RegisterPackRequest{}, "idem-int-gen")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if len(packResp.RegisteredWorkflowVersionIDs) == 0 {
		t.Fatal("expected workflow version IDs from register")
	}
	defaultWfVersionID := packResp.RegisteredWorkflowVersionIDs[0]

	_, err = svc.UpdateConfig(context.Background(), "project-int-gen", UpdateArticleConfigRequest{
		TopicStyle:                      "tech",
		DefaultWorkflowTemplateVersionID: defaultWfVersionID,
	}, "idem-int-gen-cfg")
	if err != nil {
		t.Fatalf("update config: %v", err)
	}

	// Act: create generation run
	resp, err := svc.CreateGenerationRun(context.Background(), "project-int-gen",
		CreateArticleGenerationRunRequest{
			Topic:          "Go Integration Test",
			Audience:       "developers",
			TargetPlatform: "blog",
		}, "wfr-int-1", "idem-int-genrun")
	if err != nil {
		t.Fatalf("create generation run: %v", err)
	}
	if resp.GenerationRunID == "" {
		t.Fatal("expected non-empty generation_run_id")
	}
	if resp.WorkflowRunID == "" {
		t.Fatal("expected non-empty workflow_run_id")
	}
}

// Integration test: ArticleService -> GenerationService (Task-04: RetryGenerationRun creates new run linked to original)
// Component chain: ArticleService.RetryGenerationRun -> generation.Service
// Covers: Task-04 (Article generation run retry)
// @Test
func TestTask04IntegrationRetryGenerationRun(t *testing.T) {
	contentSvc := content.NewService()
	wfSvc := workflow.NewService()
	metricsSvc := metrics.NewService()
	svc := NewService(contentSvc, wfSvc, metricsSvc)

	_, err := svc.RegisterPack(context.Background(), RegisterPackRequest{}, "idem-int-retry")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	svc.UpdateConfig(context.Background(), "project-int-retry", UpdateArticleConfigRequest{
		TopicStyle: "tech",
	}, "idem-int-retry-cfg")

	createResp, err := svc.CreateGenerationRun(context.Background(), "project-int-retry",
		CreateArticleGenerationRunRequest{Topic: "retry test", TargetPlatform: "web"},
		"wfr-retry-1", "idem-int-retry-gen")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Retry with explicit workflow run ID
	retryResp, err := svc.RetryGenerationRun(context.Background(), "project-int-retry",
		createResp.GenerationRunID,
		RetryGenerationRunRequest{Reason: "integration retry"},
		"wfr-retry-2", "idem-int-retry-retry")
	if err != nil {
		t.Fatalf("retry: %v", err)
	}
	if retryResp.NewGenerationRunID == "" {
		t.Fatal("expected new generation_run_id from retry")
	}
	if retryResp.NewGenerationRunID == createResp.GenerationRunID {
		t.Fatal("expected different generation_run_id for retry")
	}
}

// Compile guard for generation types used in integration
// @Test
func TestTaskCompileGuardGenerationIntegration(t *testing.T) {
	_ = generation.CreateGenerationRunResponse{}
	_ = generation.RetryGenerationRunResponse{}
}
