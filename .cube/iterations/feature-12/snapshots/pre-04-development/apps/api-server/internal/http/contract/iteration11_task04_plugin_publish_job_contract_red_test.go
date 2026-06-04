package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTask04PluginPublishJobHTTPRequiresPluginBearerAndReturnsLockEnvelope(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/plugin/publish-jobs", nil)
	rr := httptest.NewRecorder()
	iteration11Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("list plugin publish jobs without bearer = %d, want 401", rr.Code)
	}

	rr = iteration11PluginRequest(http.MethodGet, "/api/v1/plugin/publish-jobs?page=1&page_size=20&status=queued", nil, "plugin-dev", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("list plugin publish jobs = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env := decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("list plugin publish jobs must return success envelope with request_id: %s", rr.Body.String())
	}
	var listData map[string]any
	if err := json.Unmarshal(env.Data, &listData); err != nil {
		t.Fatalf("decode plugin publish jobs data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"items", "pagination"} {
		if listData[field] == nil {
			t.Fatalf("list plugin publish jobs response missing %s: %s", field, rr.Body.String())
		}
	}

	lockBody := []byte(`{"lock_ttl_seconds":300}`)
	rr = iteration11PluginRequest(http.MethodPost, "/api/v1/plugin/publish-jobs/job-001/lock", lockBody, "plugin-dev", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("lock plugin publish job = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("lock plugin publish job must return success envelope with request_id: %s", rr.Body.String())
	}
	var lockData map[string]any
	if err := json.Unmarshal(env.Data, &lockData); err != nil {
		t.Fatalf("decode plugin publish lock data: %v data=%s", err, string(env.Data))
	}
	if lockData["lock_id"] == nil || lockData["lock_id"] == "" {
		t.Fatalf("lock plugin publish job response missing lock_id: %s", rr.Body.String())
	}
	if lockData["locked_until"] == nil || lockData["locked_until"] == "" {
		t.Fatalf("lock plugin publish job response missing locked_until: %s", rr.Body.String())
	}
	if lockData["payload"] == nil {
		t.Fatalf("lock plugin publish job response missing payload: %s", rr.Body.String())
	}
	if lockData["payload_hash"] == nil || lockData["payload_hash"] == "" {
		t.Fatalf("lock plugin publish job response missing payload_hash: %s", rr.Body.String())
	}
	if lockData["content_version_id"] == nil || lockData["content_version_id"] == "" {
		t.Fatalf("lock plugin publish job response missing content_version_id: %s", rr.Body.String())
	}
	if lockData["adapter_config_id"] == nil || lockData["adapter_config_id"] == "" {
		t.Fatalf("lock plugin publish job response missing adapter_config_id: %s", rr.Body.String())
	}
	if adapterVersion, ok := lockData["adapter_version"].(float64); !ok || adapterVersion <= 0 {
		t.Fatalf("lock plugin publish job response must expose positive adapter_version: %s", rr.Body.String())
	}
	lockID, _ := lockData["lock_id"].(string)
	payloadHash, _ := lockData["payload_hash"].(string)

	filledBody := []byte(`{"lock_id":"` + lockID + `","payload_hash":"` + payloadHash + `","note":"payload filled in browser extension"}`)
	rr = iteration11PluginRequest(http.MethodPost, "/api/v1/plugin/publish-jobs/job-001/filled", filledBody, "plugin-dev", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("mark plugin publish job filled = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("mark plugin publish job filled must return success envelope with request_id: %s", rr.Body.String())
	}
	var filledData map[string]any
	if err := json.Unmarshal(env.Data, &filledData); err != nil {
		t.Fatalf("decode plugin filled data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"event_id", "current_status"} {
		if filledData[field] == nil || filledData[field] == "" {
			t.Fatalf("mark plugin publish job filled response missing %s: %s", field, rr.Body.String())
		}
	}

	if filledStatus, ok := filledData["current_status"].(string); ok && filledStatus == "published" {
		t.Fatalf("filled must not advance directly to published, got status=%s", filledStatus)
	}

	// HIGH-1: idempotent published request with same payload must return same result
	publishedBody2 := []byte(`{"lock_id":"` + lockID + `","external_url":"https://example.com/post/1","published_at":"2026-06-02T10:05:00Z","payload_hash":"` + payloadHash + `","note":"published by plugin"}`)
	rr = iteration11PluginRequest(http.MethodPost, "/api/v1/plugin/publish-jobs/job-001/published", publishedBody2, "plugin-dev", "plugin-published-idem")
	if rr.Code != http.StatusOK {
		t.Fatalf("idempotent published request must return 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil {
		t.Fatalf("idempotent published request must return success: %s", rr.Body.String())
	}

	// HIGH-1: idempotent request with different payload must return IDEMPOTENCY_CONFLICT
	differentPublishedBody := []byte(`{"lock_id":"` + lockID + `","external_url":"https://example.com/post/2","published_at":"2026-06-02T10:10:00Z","payload_hash":"` + payloadHash + `","note":"different payload"}`)
	rr = iteration11PluginRequest(http.MethodPost, "/api/v1/plugin/publish-jobs/job-001/published", differentPublishedBody, "plugin-dev", "plugin-published-idem")
	if rr.Code != http.StatusConflict {
		t.Fatalf("idempotent request with different payload must return 409, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "IDEMPOTENCY_CONFLICT" {
		t.Fatalf("idempotent request with different payload must return error code IDEMPOTENCY_CONFLICT: %s", rr.Body.String())
	}

	publishedBody := []byte(`{"lock_id":"` + lockID + `","external_url":"https://example.com/post/1","published_at":"2026-06-02T10:05:00Z","payload_hash":"` + payloadHash + `","note":"published by plugin"}`)
	rr = iteration11PluginRequest(http.MethodPost, "/api/v1/plugin/publish-jobs/job-001/published", publishedBody, "plugin-dev", "plugin-published-idem")
	if rr.Code != http.StatusOK {
		t.Fatalf("mark plugin publish job published = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("mark plugin publish job published must return success envelope with request_id: %s", rr.Body.String())
	}
	var publishedData map[string]any
	if err := json.Unmarshal(env.Data, &publishedData); err != nil {
		t.Fatalf("decode plugin published data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"publish_job_id", "current_status", "operation_log_id"} {
		if publishedData[field] == nil || publishedData[field] == "" {
			t.Fatalf("mark plugin publish job published response missing %s: %s", field, rr.Body.String())
		}
	}

	failedBody := []byte(`{"lock_id":"` + lockID + `","reason":"browser validation failed","retryable":true,"platform_error_summary":"selector missing"}`)
	rr = iteration11PluginRequest(http.MethodPost, "/api/v1/plugin/publish-jobs/job-001/failed", failedBody, "plugin-dev", "plugin-failed-idem")
	if rr.Code != http.StatusOK {
		t.Fatalf("mark plugin publish job failed = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("mark plugin publish job failed must return success envelope with request_id: %s", rr.Body.String())
	}
	var failedData map[string]any
	if err := json.Unmarshal(env.Data, &failedData); err != nil {
		t.Fatalf("decode plugin failed data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"publish_job_id", "current_status", "operation_log_id"} {
		if failedData[field] == nil || failedData[field] == "" {
			t.Fatalf("mark plugin publish job failed response missing %s: %s", field, rr.Body.String())
		}
	}

	// HIGH-1: idempotent failed request with different payload must return IDEMPOTENCY_CONFLICT
	differentFailedBody := []byte(`{"lock_id":"` + lockID + `","reason":"different reason","retryable":false,"platform_error_summary":"other error"}`)
	rr = iteration11PluginRequest(http.MethodPost, "/api/v1/plugin/publish-jobs/job-001/failed", differentFailedBody, "plugin-dev", "plugin-failed-idem")
	if rr.Code != http.StatusConflict {
		t.Fatalf("idempotent failed request with different payload must return 409, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "IDEMPOTENCY_CONFLICT" {
		t.Fatalf("idempotent failed request with different payload must return error code IDEMPOTENCY_CONFLICT: %s", rr.Body.String())
	}

	secondLockBody := []byte(`{"lock_ttl_seconds":300}`)
	rr = iteration11PluginRequest(http.MethodPost, "/api/v1/plugin/publish-jobs/job-001/lock", secondLockBody, "plugin-dev", "plugin-lock-retry-idem")
	if rr.Code != http.StatusConflict {
		t.Fatalf("second lock on already-locked job must return 409, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "CONFLICT" {
		t.Fatalf("second lock must return error code CONFLICT: %s", rr.Body.String())
	}
}