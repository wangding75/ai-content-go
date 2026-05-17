package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// @Test
func TestListLLMCallLogsEndpointReturns200WithPaginatedResponse(t *testing.T) {
	router := newTestRouter()
	req := authedRequest(http.MethodGet, "/api/v1/llm-call-logs?workflow_run_id=wfr-1&page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if body["success"] != true {
		t.Fatalf("expected success=true, got: %v", body)
	}
	data, _ := body["data"].(map[string]any)
	if data["items"] == nil || data["pagination"] == nil {
		t.Fatalf("expected items and pagination, got: %v", body)
	}
}

// @Test
func TestGetLLMCallLogEndpointReturns404ForMissingID(t *testing.T) {
	router := newTestRouter()
	req := authedRequest(http.MethodGet, "/api/v1/llm-call-logs/nonexistent-log-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	errObj, _ := body["error"].(map[string]any)
	if errObj["code"] != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND error code, got: %v", errObj)
	}
}

// @Test
func TestListLLMCallLogsFiltersByProviderQueryParam(t *testing.T) {
	router := newTestRouter()
	req := authedRequest(http.MethodGet, "/api/v1/llm-call-logs?provider=openai&model=gpt-4o&status=success", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if body["success"] != true {
		t.Fatalf("expected success=true, got: %v", body)
	}
}
