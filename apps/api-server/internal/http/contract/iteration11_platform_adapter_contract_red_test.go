package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	serverhttp "github.com/wangding75/ai-content-go/apps/api-server/internal/http"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/system"
)

type iteration11SystemService struct{}

func (iteration11SystemService) Health(ctx context.Context) (system.HealthResponse, error) {
	return system.HealthResponse{}, nil
}

func (iteration11SystemService) Info(ctx context.Context) (system.InfoResponse, error) {
	return system.InfoResponse{}, nil
}

func (iteration11SystemService) ConfigCheck(ctx context.Context) (system.ConfigCheckResponse, error) {
	return system.ConfigCheckResponse{}, nil
}

func (iteration11SystemService) DBCheck(ctx context.Context) (system.DBCheckResponse, error) {
	return system.DBCheckResponse{}, nil
}

func (iteration11SystemService) MigrationStatus(ctx context.Context) (system.MigrationStatusResponse, error) {
	return system.MigrationStatusResponse{}, nil
}

func iteration11Router() http.Handler {
	return serverhttp.NewRouter(iteration11SystemService{}, nil)
}

func iteration11Request(method, path string, body []byte, idempotencyKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("X-Request-Id", "req-iteration-11")
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	iteration11Router().ServeHTTP(rr, req)
	return rr
}

func iteration11PluginRequest(method, path string, body []byte, bearerToken string, idempotencyKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	req.Header.Set("X-Request-Id", "req-iteration-11-plugin")
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	iteration11Router().ServeHTTP(rr, req)
	return rr
}

func decodeIteration11Envelope(t *testing.T, body []byte) struct {
	Success   bool            `json:"success"`
	Data      json.RawMessage `json:"data"`
	Error     *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	RequestID string `json:"request_id"`
} {
	t.Helper()
	var env struct {
		Success   bool            `json:"success"`
		Data      json.RawMessage `json:"data"`
		Error     *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, string(body))
	}
	return env
}

func readIteration11RepoFile(t *testing.T, elems ...string) string {
	t.Helper()
	pathElems := append([]string{"../../../../../"}, elems...)
	content, err := os.ReadFile(filepath.Clean(filepath.Join(pathElems...)))
	if err != nil {
		t.Fatalf("read repo file %v: %v", elems, err)
	}
	return string(content)
}

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


func TestTask06ExternalCallbackHTTPReturnsOneTimeTokenAndAcceptsCallbackWithoutAdminBearer(t *testing.T) {
	rotateBody := []byte(`{"reason":"rotate after browser extension install"}`)
	rr := iteration11Request(http.MethodPost, "/api/v1/external-automation/bindings/binding-001/rotate-callback-token", rotateBody, "callback-rotate-idem")
	if rr.Code != http.StatusOK {
		t.Fatalf("rotate callback token = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env := decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("rotate callback token must return success envelope with request_id: %s", rr.Body.String())
	}
	var rotateData map[string]any
	if err := json.Unmarshal(env.Data, &rotateData); err != nil {
		t.Fatalf("decode callback token rotate data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"binding_id", "callback_token_once", "callback_token_masked", "operation_log_id"} {
		if rotateData[field] == nil || rotateData[field] == "" {
			t.Fatalf("rotate callback token response missing %s: %s", field, rr.Body.String())
		}
	}

	updateAuthBody := []byte(`{"callback_auth_type":"signature","signing_secret_ref":"secret/n8n-signature","change_reason":"switch to signature mode"}`)
	rr = iteration11Request(http.MethodPatch, "/api/v1/external-automation/bindings/binding-001/callback-auth", updateAuthBody, "callback-auth-update-idem")
	if rr.Code != http.StatusOK {
		t.Fatalf("update callback auth = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("update callback auth must return success envelope with request_id: %s", rr.Body.String())
	}
	var authData map[string]any
	if err := json.Unmarshal(env.Data, &authData); err != nil {
		t.Fatalf("decode callback auth update data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"binding_id", "callback_auth_type", "operation_log_id"} {
		if authData[field] == nil || authData[field] == "" {
			t.Fatalf("update callback auth response missing %s: %s", field, rr.Body.String())
		}
	}

	callbackBody := []byte(`{"binding_id":"binding-001","event_type":"platform_collect.submitted","payload":{"external_url":"https://example.com/post/1"},"stable_event_id":"evt-001"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/external-automation/callbacks", bytes.NewReader(callbackBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Id", "req-iteration-11-external-callback")
	req.Header.Set("X-External-Binding-Id", "binding-001")
	req.Header.Set("Authorization", "Bearer callback-token")
	rr = httptest.NewRecorder()
	iteration11Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("receive external callback without admin bearer = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("receive external callback must return success envelope with request_id: %s", rr.Body.String())
	}
	var callbackData map[string]any
	if err := json.Unmarshal(env.Data, &callbackData); err != nil {
		t.Fatalf("decode external callback data: %v data=%s", err, string(env.Data))
	}
	accepted, ok := callbackData["accepted"].(bool)
	if !ok || !accepted {
		t.Fatalf("receive external callback must return accepted=true: %s", rr.Body.String())
	}
	if callbackData["callback_log_id"] == nil || callbackData["callback_log_id"] == "" {
		t.Fatalf("receive external callback response missing callback_log_id: %s", rr.Body.String())
	}

	rr = iteration11Request(http.MethodGet, "/api/v1/external-automation/callback-logs?page=1&page_size=20&accepted=true", nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("list callback logs = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("list callback logs must return success envelope with request_id: %s", rr.Body.String())
	}
	var logsData map[string]any
	if err := json.Unmarshal(env.Data, &logsData); err != nil {
		t.Fatalf("decode callback logs data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"items", "pagination"} {
		if logsData[field] == nil {
			t.Fatalf("list callback logs response missing %s: %s", field, rr.Body.String())
		}
	}
	cbItems, ok := logsData["items"].([]any)
	if !ok || len(cbItems) == 0 {
		t.Fatalf("list callback logs must expose non-empty items: %s", rr.Body.String())
	}
	cbFirstItem, ok := cbItems[0].(map[string]any)
	if !ok {
		t.Fatalf("callback log list item must be object: %s", rr.Body.String())
	}
	if cbFirstItem["boundary_violation"] == nil {
		t.Fatalf("callback log list item must expose boundary_violation: %s", rr.Body.String())
	}

	// HIGH-5: invalid sort on callback log list must return VALIDATION_ERROR
	rr = iteration11Request(http.MethodGet, "/api/v1/external-automation/callback-logs?sort=id&order=asc", nil, "")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("callback log invalid sort must return 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("callback log invalid sort must return VALIDATION_ERROR: %s", rr.Body.String())
	}

	// HIGH-5: boundary-violating payload must return FORBIDDEN
	boundaryBody := []byte(`{"binding_id":"binding-001","event_type":"platform_collect.submitted","payload":{"workflow_run_id":"run-123","agent_code":"agent-001","content_body":"modified"},"stable_event_id":"evt-boundary-001"}`)
	boundaryReq := httptest.NewRequest(http.MethodPost, "/api/v1/external-automation/callbacks", bytes.NewReader(boundaryBody))
	boundaryReq.Header.Set("Content-Type", "application/json")
	boundaryReq.Header.Set("X-Request-Id", "req-iteration-11-boundary")
	boundaryReq.Header.Set("X-External-Binding-Id", "binding-001")
	boundaryReq.Header.Set("Authorization", "Bearer callback-token")
	boundaryRR := httptest.NewRecorder()
	iteration11Router().ServeHTTP(boundaryRR, boundaryReq)
	if boundaryRR.Code != http.StatusForbidden {
		t.Fatalf("boundary-violating callback must return 403, got %d body=%s", boundaryRR.Code, boundaryRR.Body.String())
	}
	boundaryEnv := decodeIteration11Envelope(t, boundaryRR.Body.Bytes())
	if boundaryEnv.Success || boundaryEnv.Error == nil || boundaryEnv.Error.Code != "FORBIDDEN" {
		t.Fatalf("boundary-violating callback must return error code FORBIDDEN: %s", boundaryRR.Body.String())
	}
	if boundaryEnv.Error.Message == "" {
		t.Fatalf("FORBIDDEN error must include message: %s", boundaryRR.Body.String())
	}

	// HIGH-5: unknown event type must return VALIDATION_ERROR
	unknownEventBody := []byte(`{"binding_id":"binding-001","event_type":"unknown.event","payload":{},"stable_event_id":"evt-unknown"}`)
	unknownEventReq := httptest.NewRequest(http.MethodPost, "/api/v1/external-automation/callbacks", bytes.NewReader(unknownEventBody))
	unknownEventReq.Header.Set("Content-Type", "application/json")
	unknownEventReq.Header.Set("X-Request-Id", "req-iteration-11-unknown-event")
	unknownEventReq.Header.Set("X-External-Binding-Id", "binding-001")
	unknownEventReq.Header.Set("Authorization", "Bearer callback-token")
	unknownEventRR := httptest.NewRecorder()
	iteration11Router().ServeHTTP(unknownEventRR, unknownEventReq)
	if unknownEventRR.Code != http.StatusBadRequest {
		t.Fatalf("unknown event type must return 400, got %d body=%s", unknownEventRR.Code, unknownEventRR.Body.String())
	}
	unknownEventEnv := decodeIteration11Envelope(t, unknownEventRR.Body.Bytes())
	if unknownEventEnv.Success || unknownEventEnv.Error == nil || unknownEventEnv.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("unknown event type must return error code VALIDATION_ERROR: %s", unknownEventRR.Body.String())
	}

	// HIGH-6: testExternalCallback must require bearer auth
	testCbReq := httptest.NewRequest(http.MethodPost, "/api/v1/external-automation/callbacks/test", nil)
	testCbReq.Header.Set("X-Request-Id", "req-iteration-11-test-cb")
	rr = httptest.NewRecorder()
	iteration11Router().ServeHTTP(rr, testCbReq)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("test callback without bearer must return 401, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if env.Success || env.Error == nil || env.Error.Code != "UNAUTHORIZED" {
		t.Fatalf("test callback without bearer must return error code UNAUTHORIZED: %s", rr.Body.String())
	}

	testCbAuthReq := httptest.NewRequest(http.MethodPost, "/api/v1/external-automation/callbacks/test", bytes.NewReader([]byte(`{"binding_id":"binding-001"}`)))
	testCbAuthReq.Header.Set("Content-Type", "application/json")
	testCbAuthReq.Header.Set("X-Request-Id", "req-iteration-11-test-cb-auth")
	testCbAuthReq.Header.Set("Authorization", "Bearer dev")
	testCbAuthReq.Header.Set("Idempotency-Key", "test-cb-idem")
	rr = httptest.NewRecorder()
	iteration11Router().ServeHTTP(rr, testCbAuthReq)
	if rr.Code != http.StatusOK {
		t.Fatalf("test callback with bearer must return 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	env = decodeIteration11Envelope(t, rr.Body.Bytes())
	if !env.Success || env.Error != nil || env.RequestID == "" {
		t.Fatalf("test callback must return success envelope with request_id: %s", rr.Body.String())
	}
	var testCbData map[string]any
	if err := json.Unmarshal(env.Data, &testCbData); err != nil {
		t.Fatalf("decode test callback data: %v data=%s", err, string(env.Data))
	}
	for _, field := range []string{"binding_id", "stable_event_id", "test_result"} {
		if testCbData[field] == nil || testCbData[field] == "" {
			t.Fatalf("test callback response missing %s: %s", field, rr.Body.String())
		}
	}
}



func TestTask07OpenAPIAndWebClientDeclareIteration11Contracts(t *testing.T) {
	openapi := readIteration11RepoFile(t, "openapi", "openapi.yaml")
	for _, want := range []string{
		"/api/v1/platform-adapters:",
		"operationId: createPlatformAdapter",
		"operationId: listPlatformAdapters",
		"operationId: getPlatformAdapter",
		"operationId: updatePlatformAdapter",
		"/api/v1/plugin-clients:",
		"operationId: registerPluginClient",
		"operationId: listPluginClients",
		"operationId: updatePluginClient",
		"operationId: rotatePluginClientKey",
		"/api/v1/plugin-auth/token:",
		"operationId: authenticatePlugin",
		"/api/v1/plugin/publish-jobs:",
		"operationId: listPluginPublishJobs",
		"operationId: lockPluginPublishJob",
		"operationId: markPluginPublishJobFilled",
		"operationId: markPluginPublishJobPublished",
		"operationId: markPluginPublishJobFailed",
		"/api/v1/platform-collect-logs:",
		"operationId: submitPlatformCollectLog",
		"operationId: listPlatformCollectLogs",
		"operationId: getPlatformCollectLog",
		"operationId: confirmPlatformCollectLogMetrics",
		"operationId: rotateCallbackToken",
		"operationId: updateCallbackAuth",
		"operationId: receiveExternalCallback",
		"operationId: listCallbackLogs",
		"operationId: testExternalCallback",
		"Idempotency-Key",
		"X-External-Binding-Id",
		"IDEMPOTENCY_CONFLICT",
	} {
		if !strings.Contains(openapi, want) {
			t.Fatalf("iteration11 openapi missing %q", want)
		}
	}

	apiClient := readIteration11RepoFile(t, "apps", "web-admin", "lib", "api.ts")
	for _, want := range []string{
		"export type PlatformAdapterResponse",
		"export type PlatformAdapterDetailResponse",
		"export type PluginClientResponse",
		"export type PluginPublishJobResponse",
		"export type PlatformCollectLogResponse",
		"export type PlatformCollectLogDetailResponse",
		"export type ExternalCallbackLogResponse",
		"export async function createPlatformAdapter",
		"export async function fetchPlatformAdapters",
		"export async function fetchPlatformAdapter",
		"export async function updatePlatformAdapter",
		"export async function createPluginClient",
		"export async function fetchPluginClients",
		"export async function updatePluginClient",
		"export async function rotatePluginClientKey",
		"export async function authenticatePlugin",
		"export async function fetchPluginPublishJobs",
		"export async function lockPluginPublishJob",
		"export async function markPluginPublishJobFilled",
		"export async function markPluginPublishJobPublished",
		"export async function markPluginPublishJobFailed",
		"export async function submitPlatformCollectLog",
		"export async function fetchPlatformCollectLogs",
		"export async function fetchPlatformCollectLog",
		"export async function confirmPlatformCollectLogMetrics",
		"export async function rotateCallbackToken",
		"export async function updateCallbackAuth",
		"export async function receiveExternalCallback",
		"export async function fetchCallbackLogs",
		"export async function testExternalCallback",
		"Idempotency-Key",
	} {
		if !strings.Contains(apiClient, want) {
			t.Fatalf("iteration11 web api client missing %q", want)
		}
	}
	if strings.Contains(apiClient, "export type PlatformAdapterResponse = any") || strings.Contains(apiClient, "export type PluginClientResponse = any") {
		t.Fatalf("iteration11 web api client must not use any for public response types")
	}
}

func TestTask08PlatformAdapterAndPluginClientPagesDeclareUIContracts(t *testing.T) {
	adapterPage := readIteration11RepoFile(t, "apps", "web-admin", "app", "platform-adapters", "page.tsx")
	for _, want := range []string{
		"fetchPlatformAdapters",
		"createPlatformAdapter",
		"updatePlatformAdapter",
		"PlatformAdapterResponse",
		"PlatformAdapterDetailResponse",
		"loading",
		"暂无",
		"request_id",
		"page-shell",
		"card",
	} {
		if !strings.Contains(adapterPage, want) {
			t.Fatalf("platform adapters page missing %q", want)
		}
	}

	pluginPage := readIteration11RepoFile(t, "apps", "web-admin", "app", "plugin-clients", "page.tsx")
	for _, want := range []string{
		"fetchPluginClients",
		"createPluginClient",
		"updatePluginClient",
		"rotatePluginClientKey",
		"PluginClientResponse",
		"api_key_once",
		"loading",
		"暂无",
		"request_id",
		"page-shell",
		"card",
	} {
		if !strings.Contains(pluginPage, want) {
			t.Fatalf("plugin clients page missing %q", want)
		}
	}

	nav := readIteration11RepoFile(t, "apps", "web-admin", "app", "global-nav.tsx")
	for _, want := range []string{"/platform-adapters", "平台 Adapter", "/plugin-clients", "插件客户端"} {
		if !strings.Contains(nav, want) {
			t.Fatalf("global nav missing %q", want)
		}
	}
}

func TestTask09PlatformCollectLogsAndN8NPagesDeclareUIContracts(t *testing.T) {
	collectPage := readIteration11RepoFile(t, "apps", "web-admin", "app", "platform-collect-logs", "page.tsx")
	for _, want := range []string{
		"fetchPlatformCollectLogs",
		"fetchPlatformCollectLog",
		"confirmPlatformCollectLogMetrics",
		"PlatformCollectLogResponse",
		"PlatformCollectLogDetailResponse",
		"错误摘要",
		"request_id",
		"page-shell",
		"card",
	} {
		if !strings.Contains(collectPage, want) {
			t.Fatalf("platform collect logs page missing %q", want)
		}
	}

	n8nPage := readIteration11RepoFile(t, "apps", "web-admin", "app", "external-automation", "n8n", "page.tsx")
	for _, want := range []string{
		"rotateCallbackToken",
		"updateCallbackAuth",
		"testExternalCallback",
		"fetchCallbackLogs",
		"callback_token_once",
		"callback_log_id",
		"request_id",
		"边界",
		"role=\"alert\"",
		"role=\"status\"",
	} {
		if !strings.Contains(n8nPage, want) {
			t.Fatalf("external automation n8n page missing %q", want)
		}
	}

	nav := readIteration11RepoFile(t, "apps", "web-admin", "app", "global-nav.tsx")
	for _, want := range []string{"/platform-collect-logs", "采集日志", "/external-automation/n8n", "外部自动化 / n8n"} {
		if !strings.Contains(nav, want) {
			t.Fatalf("global nav missing %q", want)
		}
	}
}
