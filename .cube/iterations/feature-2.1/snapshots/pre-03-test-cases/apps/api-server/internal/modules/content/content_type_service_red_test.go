package content_test

import (
	"context"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

// @Test
func TestListContentTypesSupportsPaginationFilteringAndSorting(t *testing.T) {
	service := content.NewService()

	resp, err := service.ListContentTypes(context.Background(), content.ListContentTypesRequest{
		PaginationRequest: content.PaginationRequest{Page: 1, PageSize: 20, Sort: "code", Order: "asc"},
		Enabled:           boolPtr(true),
	})
	if err != nil {
		t.Fatalf("list content types should not fail: %v", err)
	}

	if resp.Pagination.Page != 1 || resp.Pagination.PageSize != 20 || resp.Pagination.Total < 0 {
		t.Fatalf("unexpected pagination: %#v", resp.Pagination)
	}
	for _, item := range resp.Items {
		if item.ID == "" || item.Code == "" || item.Name == "" || item.ProjectSchema == nil || !item.Enabled {
			t.Fatalf("content type item does not satisfy response contract: %#v", item)
		}
	}
}

// @Test
func TestCreateContentTypeReturnsIdentifierForValidSchema(t *testing.T) {
	service := content.NewService()

	resp, err := service.CreateContentType(context.Background(), content.CreateContentTypeRequest{
		Code:          "article",
		Name:          "Article",
		ProjectSchema: map[string]any{"title": map[string]any{"type": "string"}},
	})
	if err != nil {
		t.Fatalf("create content type should not fail: %v", err)
	}
	if resp.ContentTypeID == "" {
		t.Fatalf("expected content_type_id")
	}
}

// @Test
func TestCreateContentTypeRejectsMissingRequiredFields(t *testing.T) {
	service := content.NewService()

	if _, err := service.CreateContentType(context.Background(), content.CreateContentTypeRequest{ProjectSchema: map[string]any{}}); err == nil {
		t.Fatalf("expected validation error for missing code and name")
	}
}

// @Test
func TestCreateContentTypeRejectsDuplicateCode(t *testing.T) {
	service := content.NewService()
	req := content.CreateContentTypeRequest{Code: "article", Name: "Article", ProjectSchema: map[string]any{}}

	_, _ = service.CreateContentType(context.Background(), req)
	if _, err := service.CreateContentType(context.Background(), req); err == nil {
		t.Fatalf("expected conflict error for duplicate content type code")
	}
}

// @Test
func TestProjectSchemaReturnsConfiguredSchemaForContentType(t *testing.T) {
	service := content.NewService()

	resp, err := service.ProjectSchema(context.Background(), "1")
	if err != nil {
		t.Fatalf("project schema should not fail: %v", err)
	}
	if resp.ContentTypeID != "1" || resp.ProjectSchema == nil {
		t.Fatalf("unexpected project schema response: %#v", resp)
	}
}

// @Test
func TestProjectSchemaRejectsInvalidOrMissingContentTypeID(t *testing.T) {
	service := content.NewService()

	if _, err := service.ProjectSchema(context.Background(), ""); err == nil {
		t.Fatalf("expected validation error for missing content type id")
	}
	if _, err := service.ProjectSchema(context.Background(), "missing"); err == nil {
		t.Fatalf("expected not found error for missing content type")
	}
}

func boolPtr(v bool) *bool { return &v }
