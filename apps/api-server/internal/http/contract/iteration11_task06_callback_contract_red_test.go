package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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