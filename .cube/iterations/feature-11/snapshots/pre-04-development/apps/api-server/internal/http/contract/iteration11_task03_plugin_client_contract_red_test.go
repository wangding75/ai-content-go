package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTask03PluginClientAndAuthHTTPExposeOneTimeKeyAndNoAdminBearerRequirement(t *testing.T) {
	body := []byte(`{"name":"Chrome Publisher","client_type":"chrome_extension","version":"1.0.0","scopes":["publish:read","publish:write"]}`)
	rr := iteration11Request(http.MethodPost, "/api/v1/plugin-clients", body, "plugin-client-create-idem")
	if rr.Code != http.StatusCreated {
		t.Fatalf("create plugin client = %d, want 201 body=%s", rr.Code, rr.Body.String())
	}
	env := decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("create plugin client must return success envelope with request_id: %s", rr.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(env.Data, &created); err != nil {
		t.Fatalf("decode plugin client data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"client_id", "api_key_once", "api_key_masked"} {
		if created[field] == nil || created[field] == "" {
			t.Fatalf("create plugin client response missing %s: %s", field, rr.Body.String())
		}
	}
	clientID, _ := created["client_id"].(string)
	apiKeyOnce, _ := created["api_key_once"].(string)

	rr = iteration11Request(http.MethodGet, "/api/v1/plugin-clients?page=1&page_size=20&status=enabled", nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("list plugin clients = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("list plugin clients must return success envelope with request_id: %s", rr.Body.String())
	}
	var listed map[string]any
	if err := json.Unmarshal(env.Data, &listed); err != nil {
		t.Fatalf("decode plugin client list data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"items", "pagination"} {
		if listed[field] == nil {
			t.Fatalf("list plugin clients response missing %s: %s", field, rr.Body.String())
		}
	}
	items, ok := listed["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("list plugin clients must expose non-empty items: %s", rr.Body.String())
	}
	firstItem, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("list plugin client item must be object: %s", rr.Body.String())
	}
	for _, forbid := range []string{"api_key_once", "api_key_hash"} {
		if firstItem[forbid] != nil {
			t.Fatalf("plugin client list item must not expose %q: %s", forbid, rr.Body.String())
		}
	}
	if firstItem["api_key_masked"] == nil || firstItem["api_key_masked"] == "" {
		t.Fatalf("plugin client list item must expose api_key_masked: %s", rr.Body.String())
	}

	// HIGH-5: invalid sort on plugin client list must return VALIDATION_ERROR
	rr = iteration11Request(http.MethodGet, "/api/v1/plugin-clients?sort=invalid_field&order=asc", nil, "")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("plugin client invalid sort must return 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("plugin client invalid sort must return VALIDATION_ERROR: %s", rr.Body.String())
	}

	// MED-5: invalid scope returns VALIDATION_ERROR
	badScopeBody := []byte(`{"scopes":["admin:all"],"change_reason":"bad scope"}`)
	rr = iteration11Request(http.MethodPatch, "/api/v1/plugin-clients/"+clientID, badScopeBody, "plugin-client-bad-scope-idem")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("invalid scope must return 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("invalid scope must return VALIDATION_ERROR: %s", rr.Body.String())
	}

	updateBody := []byte(`{"status":"enabled","scopes":["publish:read","publish:write"],"change_reason":"scope refresh"}`)
	rr = iteration11Request(http.MethodPatch, "/api/v1/plugin-clients/"+clientID, updateBody, "plugin-client-update-idem")
	if rr.Code != http.StatusOK {
		t.Fatalf("update plugin client = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("update plugin client must return success envelope with request_id: %s", rr.Body.String())
	}
	var updated map[string]any
	if err := json.Unmarshal(env.Data, &updated); err != nil {
		t.Fatalf("decode plugin client update data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"client_id", "status", "operation_log_id"} {
		if updated[field] == nil || updated[field] == "" {
			t.Fatalf("update plugin client response missing %s: %s", field, rr.Body.String())
		}
	}

	rotateBody := []byte(`{"reason":"rotate browser extension key"}`)
	rr = iteration11Request(http.MethodPost, "/api/v1/plugin-clients/"+clientID+"/rotate-key", rotateBody, "plugin-client-rotate-idem")
	if rr.Code != http.StatusOK {
		t.Fatalf("rotate plugin client key = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("rotate plugin client key must return success envelope with request_id: %s", rr.Body.String())
	}
	var rotated map[string]any
	if err := json.Unmarshal(env.Data, &rotated); err != nil {
		t.Fatalf("decode plugin client rotate data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"client_id", "api_key_once", "api_key_masked", "operation_log_id"} {
		if rotated[field] == nil || rotated[field] == "" {
			t.Fatalf("rotate plugin client key response missing %s: %s", field, rr.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/plugin-auth/token", bytes.NewReader([]byte(`{"api_key":"`+apiKeyOnce+`","client_version":"1.0.0"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Id", "req-iteration-11-plugin-auth")
	rr = httptest.NewRecorder()
	iteration11Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("plugin auth token without admin bearer = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("plugin auth token must return success envelope with request_id: %s", rr.Body.String())
	}
	var tokenData map[string]any
	if err := json.Unmarshal(env.Data, &tokenData); err != nil {
		t.Fatalf("decode plugin auth data: %v data=%s", err, string(env.Data))
	}
	if tokenData["access_token"] == nil || tokenData["access_token"] == "" {
		t.Fatalf("plugin auth response missing access_token: %s", rr.Body.String())
	}
	if tokenData["expires_at"] == nil || tokenData["expires_at"] == "" {
		t.Fatalf("plugin auth response missing expires_at: %s", rr.Body.String())
	}
	scopes, ok := tokenData["scopes"].([]any)
	if !ok || len(scopes) == 0 {
		t.Fatalf("plugin auth response must expose non-empty scopes array: %s", rr.Body.String())
	}

	// HIGH-3: auth failure must not leak key details
	badAuthReq := httptest.NewRequest(http.MethodPost, "/api/v1/plugin-auth/token", bytes.NewReader([]byte(`{"api_key":"invalid-key","client_version":"1.0.0"}`)))
	badAuthReq.Header.Set("Content-Type", "application/json")
	badAuthReq.Header.Set("X-Request-Id", "req-iteration-11-plugin-auth-bad")
	badAuthRR := httptest.NewRecorder()
	iteration11Router().ServeHTTP(badAuthRR, badAuthReq)
	if badAuthRR.Code != http.StatusUnauthorized {
		t.Fatalf("invalid api_key must return 401, got %d body=%s", badAuthRR.Code, badAuthRR.Body.String())
	}
	badAuthEnv := decodeIteration11Envelope(t, badAuthRR.Body.Bytes())
	if badAuthEnv.Success || badAuthEnv.Error == nil {
		t.Fatalf("invalid api_key must return error envelope: %s", badAuthRR.Body.String())
	}
	forbiddenFields := []string{"api_key_hash", "client_exists", "version_mismatch"}
	for _, f := range forbiddenFields {
		if strings.Contains(badAuthRR.Body.String(), f) {
			t.Fatalf("auth failure must not leak %q in response: %s", f, badAuthRR.Body.String())
		}
	}

	// HIGH-11: disabled client auth failure
	updateDisableBody := []byte(`{"status":"disabled","change_reason":"security audit"}`)
	rr = iteration11Request(http.MethodPatch, "/api/v1/plugin-clients/"+clientID, updateDisableBody, "plugin-client-disable-idem")
	if rr.Code != http.StatusOK {
		t.Fatalf("disable plugin client = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	disabledAuthReq := httptest.NewRequest(http.MethodPost, "/api/v1/plugin-auth/token", bytes.NewReader([]byte(`{"api_key":"`+apiKeyOnce+`","client_version":"1.0.0"}`)))
	disabledAuthReq.Header.Set("Content-Type", "application/json")
	disabledAuthReq.Header.Set("X-Request-Id", "req-iteration-11-disabled-auth")
	disabledAuthRR := httptest.NewRecorder()
	iteration11Router().ServeHTTP(disabledAuthRR, disabledAuthReq)
	if disabledAuthRR.Code != http.StatusUnauthorized {
		t.Fatalf("disabled client auth must return 401, got %d body=%s", disabledAuthRR.Code, disabledAuthRR.Body.String())
	}
}