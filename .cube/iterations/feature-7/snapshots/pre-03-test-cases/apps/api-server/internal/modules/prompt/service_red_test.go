package prompt_test

import (
	"context"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/prompt"
)

// @Test
func TestListTemplatesSupportsPaginationSortingAndAgentFilter(t *testing.T) {
	service := prompt.NewService()

	resp, err := service.ListTemplates(context.Background(), prompt.ListTemplatesRequest{
		PaginationRequest: content.PaginationRequest{Page: 1, PageSize: 20, Sort: "code", Order: "asc"},
		AgentCode:         "outline_agent",
	})
	if err != nil {
		t.Fatalf("list prompt templates should not fail: %v", err)
	}
	if resp.Pagination.Page != 1 || resp.Pagination.PageSize != 20 || resp.Pagination.Total < 0 {
		t.Fatalf("unexpected template page: %#v", resp)
	}
	for _, item := range resp.Items {
		if item.ID == "" || item.Code == "" || item.AgentCode != "outline_agent" || item.Template == "" || item.Variables == nil {
			t.Fatalf("template item does not satisfy response contract: %#v", item)
		}
	}
}

// @Test
func TestCreateTemplateRequiresCodeTemplateAndReturnsIdentifier(t *testing.T) {
	service := prompt.NewService()

	resp, err := service.CreateTemplate(context.Background(), prompt.CreateTemplateRequest{
		Code:      "outline_writer",
		Template:  "Write an outline for {{topic}}",
		Variables: []string{"topic"},
	})
	if err != nil {
		t.Fatalf("create prompt template should not fail: %v", err)
	}
	if resp.PromptTemplateID == "" {
		t.Fatalf("expected prompt_template_id")
	}
}

// @Test
func TestCreateTemplateRejectsMissingFieldsAndDuplicateCode(t *testing.T) {
	service := prompt.NewService()
	if _, err := service.CreateTemplate(context.Background(), prompt.CreateTemplateRequest{Variables: []string{}}); err == nil {
		t.Fatalf("expected validation error for missing code and template")
	}
	req := prompt.CreateTemplateRequest{Code: "outline_writer", Template: "Write an outline for {{topic}}", Variables: []string{"topic"}}
	_, _ = service.CreateTemplate(context.Background(), req)
	if _, err := service.CreateTemplate(context.Background(), req); err == nil {
		t.Fatalf("expected conflict error for duplicate template code")
	}
}
