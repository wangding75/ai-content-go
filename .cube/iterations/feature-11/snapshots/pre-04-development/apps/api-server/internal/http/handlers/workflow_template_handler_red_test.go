package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httpserver "github.com/wangding75/ai-content-go/apps/api-server/internal/http"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/system"
)

type fakeSystemSvc struct{}

func (fakeSystemSvc) Health(ctx context.Context) (system.HealthResponse, error) {
	return system.HealthResponse{}, nil
}
func (fakeSystemSvc) Info(ctx context.Context) (system.InfoResponse, error) {
	return system.InfoResponse{}, nil
}
func (fakeSystemSvc) ConfigCheck(ctx context.Context) (system.ConfigCheckResponse, error) {
	return system.ConfigCheckResponse{}, nil
}
func (fakeSystemSvc) DBCheck(ctx context.Context) (system.DBCheckResponse, error) {
	return system.DBCheckResponse{}, nil
}
func (fakeSystemSvc) MigrationStatus(ctx context.Context) (system.MigrationStatusResponse, error) {
	return system.MigrationStatusResponse{}, nil
}

func newTestRouter() http.Handler {
	return httpserver.NewRouter(fakeSystemSvc{}, nil)
}

func authedRequest(method, path string, body []byte) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Authorization", "Bearer test-token")
	return req
}

// @Test
func TestListWorkflowTemplatesEndpointReturns200WithPaginatedResponse(t *testing.T) {
	router := newTestRouter()
	req := authedRequest(http.MethodGet, "/api/v1/workflow-templates?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["success"] != true {
		t.Fatalf("expected success=true, got: %v", body)
	}
	data, ok := body["data"].(map[string]any)
	if !ok || data["items"] == nil || data["pagination"] == nil {
		t.Fatalf("expected data with items and pagination, got: %v", body)
	}
}

// @Test
func TestCreateWorkflowTemplateEndpointReturns201WithTemplateID(t *testing.T) {
	router := newTestRouter()
	payload, _ := json.Marshal(map[string]any{
		"code": "article-wf", "name": "Article Workflow", "content_type": "article", "category": "editorial",
	})
	req := authedRequest(http.MethodPost, "/api/v1/workflow-templates", payload)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	data, _ := body["data"].(map[string]any)
	if data["workflow_template_id"] == nil || data["workflow_template_id"] == "" {
		t.Fatalf("expected workflow_template_id in response, got: %v", body)
	}
}

// @Test
func TestGetWorkflowTemplateEndpointReturns404ForMissingID(t *testing.T) {
	router := newTestRouter()
	req := authedRequest(http.MethodGet, "/api/v1/workflow-templates/nonexistent-id", nil)
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
func TestCreateWorkflowVersionEndpointReturns201WithVersionID(t *testing.T) {
	router := newTestRouter()

	// create template first
	templatePayload, _ := json.Marshal(map[string]any{
		"code": "ver-test-wf", "name": "Version Test", "content_type": "article",
	})
	createReq := authedRequest(http.MethodPost, "/api/v1/workflow-templates", templatePayload)
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("setup: create template failed %d: %s", createW.Code, createW.Body.String())
	}
	var createBody map[string]any
	json.NewDecoder(createW.Body).Decode(&createBody)
	data, _ := createBody["data"].(map[string]any)
	templateID, _ := data["workflow_template_id"].(string)

	// create version
	versionPayload, _ := json.Marshal(map[string]any{
		"input_schema":  map[string]any{},
		"output_schema": map[string]any{},
		"steps": []map[string]any{
			{"step_code": "s1", "step_type": "agent", "agent_code": "writer", "order_index": 1},
		},
	})
	versionReq := authedRequest(http.MethodPost, "/api/v1/workflow-templates/"+templateID+"/versions", versionPayload)
	versionW := httptest.NewRecorder()
	router.ServeHTTP(versionW, versionReq)

	if versionW.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", versionW.Code, versionW.Body.String())
	}
	var vBody map[string]any
	json.NewDecoder(versionW.Body).Decode(&vBody)
	vData, _ := vBody["data"].(map[string]any)
	if vData["template_version_id"] == nil || vData["template_version_id"] == "" {
		t.Fatalf("expected template_version_id, got: %v", vBody)
	}
}

// @Test
func TestPublishVersionEndpointReturns200WithStatusTransition(t *testing.T) {
	router := newTestRouter()

	// create template
	tp, _ := json.Marshal(map[string]any{"code": "pub-wf", "name": "Pub", "content_type": "article"})
	cr := authedRequest(http.MethodPost, "/api/v1/workflow-templates", tp)
	cw := httptest.NewRecorder()
	router.ServeHTTP(cw, cr)
	if cw.Code != http.StatusCreated {
		t.Fatalf("setup failed: %d", cw.Code)
	}
	var cb map[string]any
	json.NewDecoder(cw.Body).Decode(&cb)
	cd, _ := cb["data"].(map[string]any)
	tid := cd["workflow_template_id"].(string)

	// create version
	vp, _ := json.Marshal(map[string]any{
		"steps": []map[string]any{{"step_code": "s1", "step_type": "agent", "agent_code": "a", "order_index": 1}},
	})
	vr := authedRequest(http.MethodPost, "/api/v1/workflow-templates/"+tid+"/versions", vp)
	vw := httptest.NewRecorder()
	router.ServeHTTP(vw, vr)
	if vw.Code != http.StatusCreated {
		t.Fatalf("setup version failed: %d", vw.Code)
	}
	var vb map[string]any
	json.NewDecoder(vw.Body).Decode(&vb)
	vd, _ := vb["data"].(map[string]any)
	vid := vd["template_version_id"].(string)

	// publish
	pubPayload, _ := json.Marshal(map[string]any{"note": "first release"})
	pubReq := authedRequest(http.MethodPost, "/api/v1/workflow-template-versions/"+vid+"/publish", pubPayload)
	pubReq.Header.Set("Idempotency-Key", "pub-key-1")
	pubW := httptest.NewRecorder()
	router.ServeHTTP(pubW, pubReq)

	if pubW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", pubW.Code, pubW.Body.String())
	}
	var pb map[string]any
	json.NewDecoder(pubW.Body).Decode(&pb)
	pd, _ := pb["data"].(map[string]any)
	if pd["previous_status"] != "draft" || pd["current_status"] != "published" {
		t.Fatalf("unexpected publish data: %v", pd)
	}
}
