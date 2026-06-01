package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestWorkflowTemplatesListPageDeclaresAPICallsAndFilterUI(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/workflow/templates/page.tsx")
	if err != nil {
		t.Fatalf("read templates page: %v", err)
	}
	page := string(raw)

	for _, required := range []string{
		"fetchWorkflowTemplates",
		"createWorkflowTemplate",
		"WorkflowTemplateResponse",
		"content_type",
		"category",
		"status",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected workflow templates page to reference %s", required)
		}
	}
}

// @Test
func TestWorkflowTemplateDetailPageDeclaresVersionAndPublishOperations(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/workflow/templates/[id]/page.tsx")
	if err != nil {
		t.Fatalf("read template detail page: %v", err)
	}
	page := string(raw)

	for _, required := range []string{
		"createWorkflowVersion",
		"publishWorkflowVersion",
		"WorkflowTemplateVersionResponse",
		"Idempotency-Key",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected template detail page to reference %s", required)
		}
	}
}
