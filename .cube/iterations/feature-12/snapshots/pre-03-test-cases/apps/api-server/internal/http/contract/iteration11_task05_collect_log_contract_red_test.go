package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTask05PlatformCollectLogHTTPAcceptsPluginBearerAndReturnsConfirmEnvelope(t *testing.T) {
	collectBody := []byte(`{"project_id":"project-001","platform":"wechat","publish_job_id":"job-001","source_type":"extension","raw_payload":{"views":"123"},"parsed_metrics":[{"metric_code":"views","raw_value":"123","unit":"count"}],"collected_at":"2026-06-02T10:00:00Z"}`)
	rr := iteration11PluginRequest(http.MethodPost, "/api/v1/platform-collect-logs", collectBody, "plugin-dev", "")
	if rr.Code != http.StatusCreated {
		t.Fatalf("submit platform collect log = %d, want 201 body=%s", rr.Code, rr.Body.String())
	}
	env := decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("submit platform collect log must return success envelope with request_id: %s", rr.Body.String())
	}
	var collectData map[string]any
	if err := json.Unmarshal(env.Data, &collectData); err != nil {
		t.Fatalf("decode collect log data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"collect_log_id", "status"} {
		if collectData[field] == nil || collectData[field] == "" {
			t.Fatalf("submit platform collect log response missing %s: %s", field, rr.Body.String())
		}
	}
	collectLogID, _ := collectData["collect_log_id"].(string)

	// HIGH-6: invalid publish_job_id must return VALIDATION_ERROR and not persist
	invalidJobCollectBody := []byte(`{"project_id":"project-001","platform":"wechat","publish_job_id":"job-nonexistent","source_type":"extension","raw_payload":{"views":"123"},"parsed_metrics":[{"metric_code":"views","raw_value":"123","unit":"count"}],"collected_at":"2026-06-02T10:00:00Z"}`)
	rr = iteration11PluginRequest(http.MethodPost, "/api/v1/platform-collect-logs", invalidJobCollectBody, "plugin-dev", "")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("invalid publish_job_id must return 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("invalid publish_job_id must return VALIDATION_ERROR: %s", rr.Body.String())
	}

	externalBody := []byte(`{"project_id":"project-001","platform":"wechat","publish_job_id":"job-001","binding_id":"binding-001","source_type":"external_callback","raw_payload":{"views":"123"},"parsed_metrics":[{"metric_code":"views","raw_value":"123","unit":"count"}],"collected_at":"2026-06-02T10:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/platform-collect-logs", bytes.NewReader(externalBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Id", "req-iteration-11-collect-external")
	req.Header.Set("X-External-Binding-Id", "binding-001")
	req.Header.Set("Authorization", "Bearer callback-token")
	rr = httptest.NewRecorder()
	iteration11Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("submit external callback collect log = %d, want 201 body=%s", rr.Code, rr.Body.String())
	}

	rr = iteration11Request(http.MethodGet, "/api/v1/platform-collect-logs?page=1&page_size=20&status=ready", nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("list platform collect logs = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("list platform collect logs must return success envelope with request_id: %s", rr.Body.String())
	}
	var listData map[string]any
	if err := json.Unmarshal(env.Data, &listData); err != nil {
		t.Fatalf("decode collect log list data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"items", "pagination"} {
		if listData[field] == nil {
			t.Fatalf("list platform collect logs response missing %s: %s", field, rr.Body.String())
		}
	}

	rr = iteration11Request(http.MethodGet, "/api/v1/platform-collect-logs?sort=invalid_field&order=asc", nil, "")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("collect log invalid sort must return 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("collect log invalid sort must return VALIDATION_ERROR: %s", rr.Body.String())
	}

	rr = iteration11Request(http.MethodGet, "/api/v1/platform-collect-logs/"+collectLogID, nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("get platform collect log = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("get platform collect log must return success envelope with request_id: %s", rr.Body.String())
	}
	var detailData map[string]any
	if err := json.Unmarshal(env.Data, &detailData); err != nil {
		t.Fatalf("decode collect log detail data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"raw_payload", "parsed_metrics", "related", "publish_job_id", "content_item_id", "content_version_id", "target_id", "content_type"} {
		if detailData[field] == nil || detailData[field] == "" {
			t.Fatalf("get platform collect log response missing %s: %s", field, rr.Body.String())
		}
	}

	confirmBody := []byte(`{"records":[{"metric_template_id":"metric-template-views","metric_code":"views","metric_date":"2026-06-02","period":"day","raw_value":"123","normalized_value":123,"unit":"count"}],"note":"confirmed by admin"}`)
	rr = iteration11Request(http.MethodPost, "/api/v1/platform-collect-logs/"+collectLogID+"/confirm-metrics", confirmBody, "collect-confirm-idem")
	if rr.Code != http.StatusOK {
		t.Fatalf("confirm platform collect log metrics = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("confirm platform collect log metrics must return success envelope with request_id: %s", rr.Body.String())
	}
	var confirmData map[string]any
	if err := json.Unmarshal(env.Data, &confirmData); err != nil {
		t.Fatalf("decode collect log confirm data: %v data=%s", err, string(env.Data))
	}
	metricRecordIDs, ok := confirmData["metric_record_ids"].([]any)
	if !ok || len(metricRecordIDs) == 0 {
		t.Fatalf("confirm platform collect log metrics response must expose non-empty metric_record_ids: %s", rr.Body.String())
	}
	if confirmData["operation_log_id"] == nil || confirmData["operation_log_id"] == "" {
		t.Fatalf("confirm platform collect log metrics response missing operation_log_id: %s", rr.Body.String())
	}

	// HIGH-4: duplicate confirm must return IDEMPOTENCY_CONFLICT
	dupConfirmBody := []byte(`{"records":[{"metric_template_id":"metric-template-views","metric_code":"views","metric_date":"2026-06-02","period":"day","raw_value":"123","normalized_value":123,"unit":"count"}],"note":"confirmed again"}`)
	rr = iteration11Request(http.MethodPost, "/api/v1/platform-collect-logs/"+collectLogID+"/confirm-metrics", dupConfirmBody, "collect-confirm-dup-idem")
	if rr.Code != http.StatusConflict {
		t.Fatalf("duplicate confirm must return 409, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "IDEMPOTENCY_CONFLICT" {
		t.Fatalf("duplicate confirm must return error code IDEMPOTENCY_CONFLICT: %s", rr.Body.String())
	}
}