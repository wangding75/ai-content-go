package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// @Test
func TestCreateWorkflowRunEndpointReturns202WithPendingStatus(t *testing.T) {
	router := newTestRouter()

	// setup: template + published version
	tid, vid := setupPublishedVersion(t, router)
	_ = tid

	payload, _ := json.Marshal(map[string]any{
		"project_id":          "proj-run-1",
		"template_version_id": vid,
		"input":               map[string]any{"title": "Test"},
	})
	req := authedRequest(http.MethodPost, "/api/v1/workflow-runs", payload)
	req.Header.Set("Idempotency-Key", "run-key-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	data, _ := body["data"].(map[string]any)
	if data["workflow_run_id"] == nil || data["workflow_run_id"] == "" {
		t.Fatalf("expected workflow_run_id, got: %v", body)
	}
	if data["status"] != "pending" {
		t.Fatalf("expected status=pending, got: %v", data["status"])
	}
}

// @Test
func TestGetWorkflowRunEndpointReturns200WithDetailResponse(t *testing.T) {
	router := newTestRouter()
	_, vid := setupPublishedVersion(t, router)

	createPayload, _ := json.Marshal(map[string]any{
		"project_id": "proj-get-run", "template_version_id": vid, "input": map[string]any{},
	})
	createReq := authedRequest(http.MethodPost, "/api/v1/workflow-runs", createPayload)
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusAccepted {
		t.Fatalf("setup: create run failed %d", createW.Code)
	}
	var cb map[string]any
	json.NewDecoder(createW.Body).Decode(&cb)
	cd, _ := cb["data"].(map[string]any)
	runID := cd["workflow_run_id"].(string)

	req := authedRequest(http.MethodGet, "/api/v1/workflow-runs/"+runID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	data, _ := body["data"].(map[string]any)
	if data["id"] == nil || data["step_count"] == nil {
		t.Fatalf("expected detail with id and step_count, got: %v", body)
	}
}

// @Test
func TestListWorkflowRunsEndpointSupportsStatusFilter(t *testing.T) {
	router := newTestRouter()
	req := authedRequest(http.MethodGet, "/api/v1/workflow-runs?status=pending&page=1&page_size=10", nil)
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

// @Test
func TestCancelWorkflowRunEndpointReturns200WithCancelledStatus(t *testing.T) {
	router := newTestRouter()
	_, vid := setupPublishedVersion(t, router)

	cp, _ := json.Marshal(map[string]any{
		"project_id": "proj-cancel-run", "template_version_id": vid, "input": map[string]any{},
	})
	cr := authedRequest(http.MethodPost, "/api/v1/workflow-runs", cp)
	cw := httptest.NewRecorder()
	router.ServeHTTP(cw, cr)
	if cw.Code != http.StatusAccepted {
		t.Fatalf("setup: %d", cw.Code)
	}
	var cb map[string]any
	json.NewDecoder(cw.Body).Decode(&cb)
	cd := cb["data"].(map[string]any)
	runID := cd["workflow_run_id"].(string)

	cancelPayload, _ := json.Marshal(map[string]any{"reason": "test cancel"})
	cancelReq := authedRequest(http.MethodPost, "/api/v1/workflow-runs/"+runID+"/cancel", cancelPayload)
	cancelW := httptest.NewRecorder()
	router.ServeHTTP(cancelW, cancelReq)

	if cancelW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", cancelW.Code, cancelW.Body.String())
	}
	var body map[string]any
	json.NewDecoder(cancelW.Body).Decode(&body)
	data, _ := body["data"].(map[string]any)
	if data["current_status"] != "cancelled" {
		t.Fatalf("expected current_status=cancelled, got: %v", data)
	}
}

// @Test
func TestGetRunStepsEndpointReturnsStepList(t *testing.T) {
	router := newTestRouter()
	_, vid := setupPublishedVersion(t, router)

	cp, _ := json.Marshal(map[string]any{
		"project_id": "proj-steps-run", "template_version_id": vid, "input": map[string]any{},
	})
	cr := authedRequest(http.MethodPost, "/api/v1/workflow-runs", cp)
	cw := httptest.NewRecorder()
	router.ServeHTTP(cw, cr)
	if cw.Code != http.StatusAccepted {
		t.Fatalf("setup: %d", cw.Code)
	}
	var cb map[string]any
	json.NewDecoder(cw.Body).Decode(&cb)
	cd := cb["data"].(map[string]any)
	runID := cd["workflow_run_id"].(string)

	req := authedRequest(http.MethodGet, "/api/v1/workflow-runs/"+runID+"/steps", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	data, _ := body["data"].(map[string]any)
	if data["items"] == nil {
		t.Fatalf("expected items in steps response, got: %v", body)
	}
}

// setupPublishedVersion creates a template + published version and returns (templateID, versionID).
func setupPublishedVersion(t *testing.T, router http.Handler) (string, string) {
	t.Helper()

	tp, _ := json.Marshal(map[string]any{"code": "setup-wf-" + t.Name(), "name": "Setup", "content_type": "article"})
	cr := authedRequest(http.MethodPost, "/api/v1/workflow-templates", tp)
	cw := httptest.NewRecorder()
	router.ServeHTTP(cw, cr)
	if cw.Code != http.StatusCreated {
		t.Fatalf("setup: create template failed %d: %s", cw.Code, cw.Body.String())
	}
	var cb map[string]any
	json.NewDecoder(cw.Body).Decode(&cb)
	cd := cb["data"].(map[string]any)
	tid := cd["workflow_template_id"].(string)

	vp, _ := json.Marshal(map[string]any{
		"input_schema": map[string]any{}, "output_schema": map[string]any{},
		"steps": []map[string]any{{"step_code": "s1", "step_type": "agent", "agent_code": "writer", "order_index": 1}},
	})
	vr := authedRequest(http.MethodPost, "/api/v1/workflow-templates/"+tid+"/versions", vp)
	vw := httptest.NewRecorder()
	router.ServeHTTP(vw, vr)
	if vw.Code != http.StatusCreated {
		t.Fatalf("setup: create version failed %d: %s", vw.Code, vw.Body.String())
	}
	var vb map[string]any
	json.NewDecoder(vw.Body).Decode(&vb)
	vd := vb["data"].(map[string]any)
	vid := vd["template_version_id"].(string)

	pubP, _ := json.Marshal(map[string]any{"note": "setup publish"})
	pubR := authedRequest(http.MethodPost, "/api/v1/workflow-template-versions/"+vid+"/publish", pubP)
	pubW := httptest.NewRecorder()
	router.ServeHTTP(pubW, pubR)
	if pubW.Code != http.StatusOK {
		t.Fatalf("setup: publish failed %d: %s", pubW.Code, pubW.Body.String())
	}

	return tid, vid
}
