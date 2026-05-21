package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// @Test
func TestIteration21ScheduleHTTPCreateReturns201EnvelopeWithDefaultCount(t *testing.T) {
	router := newTestRouter()
	payload, _ := json.Marshal(map[string]any{
		"project_id":          "project-21-http",
		"template_version_id": "wftv-21-http",
		"cron_expression":     "0 9 * * *",
	})
	req := authedRequest(http.MethodPost, "/api/v1/workflow-schedules", payload)
	req.Header.Set("Idempotency-Key", "schedule-http-create-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, _ := body["data"].(map[string]any)
	if body["success"] != true || data["schedule_id"] == "" || data["daily_content_count"] != float64(5) {
		t.Fatalf("expected successful schedule envelope with default count, got %#v", body)
	}
	if body["request_id"] == "" {
		t.Fatalf("expected request_id in envelope")
	}
}

// @Test
func TestIteration21ScheduleHTTPTestRunReturns202AndWorkflowRunID(t *testing.T) {
	router := newTestRouter()
	createPayload, _ := json.Marshal(map[string]any{
		"project_id":          "project-21-test-run",
		"template_version_id": "wftv-21-test-run",
		"cron_expression":     "0 10 * * *",
	})
	createReq := authedRequest(http.MethodPost, "/api/v1/workflow-schedules", createPayload)
	createReq.Header.Set("Idempotency-Key", "schedule-http-run-key")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("setup create schedule expected 201, got %d: %s", createW.Code, createW.Body.String())
	}
	var createBody map[string]any
	json.NewDecoder(createW.Body).Decode(&createBody)
	scheduleID := createBody["data"].(map[string]any)["schedule_id"].(string)

	runPayload, _ := json.Marshal(map[string]any{"input_override": map[string]any{"topic": "manual"}})
	runReq := authedRequest(http.MethodPost, "/api/v1/workflow-schedules/"+scheduleID+"/test-run", runPayload)
	runW := httptest.NewRecorder()
	router.ServeHTTP(runW, runReq)

	if runW.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", runW.Code, runW.Body.String())
	}
	var body map[string]any
	json.NewDecoder(runW.Body).Decode(&body)
	data, _ := body["data"].(map[string]any)
	if data["workflow_run_id"] == "" || data["status"] == "" || data["trigger_log_id"] == "" {
		t.Fatalf("expected workflow_run_id/status/trigger_log_id, got %#v", body)
	}
}

// @Test
func TestIteration21ScheduleHTTPReturnsContractErrorCodes(t *testing.T) {
	router := newTestRouter()
	payload, _ := json.Marshal(map[string]any{"project_id": "missing-required-fields"})
	req := authedRequest(http.MethodPost, "/api/v1/workflow-schedules", payload)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	errObj, _ := body["error"].(map[string]any)
	if errObj["code"] != "VALIDATION_ERROR" || body["request_id"] == "" {
		t.Fatalf("expected VALIDATION_ERROR with request_id, got %#v", body)
	}
}
