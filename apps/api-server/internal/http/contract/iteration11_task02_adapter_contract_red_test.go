package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTask02PlatformAdapterHTTPRequiresBearerAndCreatesAdapterEnvelope(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/platform-adapters", nil)
	rr := httptest.NewRecorder()
	iteration11Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("list platform adapters without bearer token = %d, want 401", rr.Code)
	}

	body := []byte(`{"platform":"wechat","display_name":"WeChat Article","publish_mode":"manual_plugin","target_type":"article","field_mapping":{"title":"headline"},"fill_rules":{"editor":"rich-text"},"collect_rules":{"views":"read_count"},"credential_ref":"binding/wechat","enabled":true}`)
	rr = iteration11Request(http.MethodPost, "/api/v1/platform-adapters", body, "adapter-create-idem")
	if rr.Code != http.StatusCreated {
		t.Fatalf("create platform adapter = %d, want 201 body=%s", rr.Code, rr.Body.String())
	}
	env := decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("create adapter must return success envelope with request_id: %s", rr.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(env.Data, &created); err != nil {
		t.Fatalf("decode adapter data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"adapter_id", "version", "operation_log_id"} {
		if created[field] == nil || created[field] == "" {
			t.Fatalf("create adapter response missing %s: %s", field, rr.Body.String())
		}
	}
	if v, ok := created["version"].(float64); !ok || v <= 0 {
		t.Fatalf("create adapter version must be positive integer: %s", rr.Body.String())
	}
	adapterID, _ := created["adapter_id"].(string)

	rr = iteration11Request(http.MethodGet, "/api/v1/platform-adapters/"+adapterID, nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("get platform adapter = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("get adapter must return success envelope with request_id: %s", rr.Body.String())
	}
	var detail map[string]any
	if err := json.Unmarshal(env.Data, &detail); err != nil {
		t.Fatalf("decode adapter detail: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"id", "platform", "display_name", "publish_mode", "target_type", "field_mapping", "fill_rules", "collect_rules", "credential_ref", "rule_summary", "version", "enabled", "updated_at"} {
		if detail[field] == nil || detail[field] == "" {
			t.Fatalf("get adapter response missing %s: %s", field, rr.Body.String())
		}
	}

	// HIGH-4: nonexistent adapter must return NOT_FOUND
	rr = iteration11Request(http.MethodGet, "/api/v1/platform-adapters/adapter-nonexistent", nil, "")
	if rr.Code != http.StatusNotFound {
		t.Fatalf("get nonexistent adapter = %d, want 404 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "NOT_FOUND" {
		t.Fatalf("get nonexistent adapter must return error code NOT_FOUND: %s", rr.Body.String())
	}

	updateBody := []byte(`{"display_name":"WeChat Article Updated","publish_mode":"manual_plugin","target_type":"article","field_mapping":{"title":"headline"},"fill_rules":{"editor":"rich-text"},"collect_rules":{"views":"read_count"},"credential_ref":"binding/wechat","enabled":true,"expected_version":1,"change_reason":"refresh rules"}`)
	rr = iteration11Request(http.MethodPatch, "/api/v1/platform-adapters/"+adapterID, updateBody, "adapter-update-idem")
	if rr.Code != http.StatusOK {
		t.Fatalf("update platform adapter = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("update adapter must return success envelope with request_id: %s", rr.Body.String())
	}
	var updated map[string]any
	if err := json.Unmarshal(env.Data, &updated); err != nil {
		t.Fatalf("decode adapter update data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"adapter_id", "version", "operation_log_id"} {
		if updated[field] == nil || updated[field] == "" {
			t.Fatalf("update adapter response missing %s: %s", field, rr.Body.String())
		}
	}

	dupBody := []byte(`{"platform":"wechat","display_name":"WeChat Duplicate","publish_mode":"manual_plugin","target_type":"article","field_mapping":{"title":"headline"},"fill_rules":{"editor":"rich-text"},"collect_rules":{"views":"read_count"},"credential_ref":"binding/wechat","enabled":true}`)
	rr = iteration11Request(http.MethodPost, "/api/v1/platform-adapters", dupBody, "adapter-dup-idem")
	if rr.Code != http.StatusConflict {
		t.Fatalf("duplicate adapter must return 409, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "CONFLICT" {
		t.Fatalf("duplicate adapter must return error code CONFLICT: %s", rr.Body.String())
	}

	// HIGH-1: nonexistent credential_ref must return VALIDATION_ERROR
	badCredBody := []byte(`{"platform":"youtube","display_name":"YouTube","publish_mode":"manual_plugin","target_type":"video","field_mapping":{"title":"title"},"fill_rules":{"editor":"rich-text"},"collect_rules":{"views":"read_count"},"credential_ref":"binding/nonexistent","enabled":true}`)
	rr = iteration11Request(http.MethodPost, "/api/v1/platform-adapters", badCredBody, "adapter-bad-cred-idem")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("nonexistent credential_ref must return 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("nonexistent credential_ref must return VALIDATION_ERROR: %s", rr.Body.String())
	}

	// HIGH-1: credential_ref to binding owned by another project must return FORBIDDEN
	forbiddenCredBody := []byte(`{"platform":"youtube","display_name":"YouTube","publish_mode":"manual_plugin","target_type":"video","field_mapping":{"title":"title"},"fill_rules":{"editor":"rich-text"},"collect_rules":{"views":"read_count"},"credential_ref":"binding/other-project","enabled":true}`)
	rr = iteration11Request(http.MethodPost, "/api/v1/platform-adapters", forbiddenCredBody, "adapter-forbidden-cred-idem")
	if rr.Code != http.StatusForbidden {
		t.Fatalf("credential_ref to other-project binding must return 403, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "FORBIDDEN" {
		t.Fatalf("credential_ref to other-project binding must return error code FORBIDDEN: %s", rr.Body.String())
	}

	// HIGH-2: disabling adapter with active queued jobs must return CONFLICT
	disableBody := []byte(`{"display_name":"WeChat Article","publish_mode":"manual_plugin","target_type":"article","field_mapping":{"title":"headline"},"fill_rules":{"editor":"rich-text"},"collect_rules":{"views":"read_count"},"credential_ref":"binding/wechat","enabled":false,"expected_version":1,"change_reason":"decommission"}`)
	rr = iteration11Request(http.MethodPatch, "/api/v1/platform-adapters/"+adapterID, disableBody, "adapter-disable-idem")
	if rr.Code != http.StatusConflict {
		t.Fatalf("disable adapter with active jobs must return 409, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "CONFLICT" {
		t.Fatalf("disable adapter with active jobs must return error code CONFLICT: %s", rr.Body.String())
	}

	rr = iteration11Request(http.MethodGet, "/api/v1/platform-adapters?sort=id&order=asc", nil, "")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("invalid sort must return 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("invalid sort must return VALIDATION_ERROR: %s", rr.Body.String())
	}
}