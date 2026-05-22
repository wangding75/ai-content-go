package contract_test

import (
	"os"
	"strings"
	"testing"
)

// @Test
func TestLLMLogsPageDeclaresFilterFieldsAndCostColumns(t *testing.T) {
	raw, err := os.ReadFile("../../../../../apps/web-admin/app/llm/logs/page.tsx")
	if err != nil {
		t.Fatalf("read llm logs page: %v", err)
	}
	page := string(raw)

	for _, required := range []string{
		"fetchLLMCallLogs",
		"LLMCallLogResponse",
		"provider",
		"model",
		"input_tokens",
		"output_tokens",
		"cost",
		"latency_ms",
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("expected llm logs page to reference %s", required)
		}
	}
}
