package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	serverhttp "github.com/wangding75/ai-content-go/apps/api-server/internal/http"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/system"
)

func iteration6Router() http.Handler {
	return serverhttp.NewRouter(iteration6SystemService{}, nil)
}

type iteration6SystemService struct{}

func (iteration6SystemService) Health(ctx context.Context) (system.HealthResponse, error) {
	return system.HealthResponse{}, nil
}

func (iteration6SystemService) Info(ctx context.Context) (system.InfoResponse, error) {
	return system.InfoResponse{}, nil
}

func (iteration6SystemService) ConfigCheck(ctx context.Context) (system.ConfigCheckResponse, error) {
	return system.ConfigCheckResponse{}, nil
}

func (iteration6SystemService) DBCheck(ctx context.Context) (system.DBCheckResponse, error) {
	return system.DBCheckResponse{}, nil
}

func (iteration6SystemService) MigrationStatus(ctx context.Context) (system.MigrationStatusResponse, error) {
	return system.MigrationStatusResponse{}, nil
}

func iteration6Request(t *testing.T, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Idempotency-Key", "idem-test")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	iteration6Router().ServeHTTP(rr, req)
	return rr
}

// @Test
func TestTask05MemoryAPIRequiresBearerTokenForReadEndpoints(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/knowledge-memory", nil)
	rr := httptest.NewRecorder()
	iteration6Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("GET knowledge-memory without bearer token = %d, want 401", rr.Code)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/consistency-reports", nil)
	rr = httptest.NewRecorder()
	iteration6Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("GET consistency-reports without bearer token = %d, want 401", rr.Code)
	}
}

// @Test
func TestTask05MemoryAPIHandlesHappyPathValidationFailureAndDomainFailureThroughHTTP(t *testing.T) {
	rr := iteration6Request(t, http.MethodGet, "/api/v1/projects/project-1/knowledge-memory", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET knowledge-memory = %d body=%s", rr.Code, rr.Body.String())
	}
	var envelope struct {
		Success   bool
		Data      map[string]any
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode success envelope: %v", err)
	}
	if !envelope.Success || envelope.Data["recent_snapshot_summary"] == nil || envelope.RequestID == "" {
		t.Fatalf("success envelope must include data.recent_snapshot_summary and request_id: %s", rr.Body.String())
	}

	rr = iteration6Request(t, http.MethodPatch, "/api/v1/projects/project-1/knowledge-memory/dynamic-state-correction", []byte(`{"reason":"人工纠偏","changes":{"status":"corrected"}}`))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("missing source_refs = %d body=%s", rr.Code, rr.Body.String())
	}

	rr = iteration6Request(t, http.MethodGet, "/api/v1/projects//knowledge-memory", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("domain failure for missing project must be 404, got %d body=%s", rr.Code, rr.Body.String())
	}
}

// @Test
func TestTask05MemoryAPIRoutesAllIteration6EndpointsWithExpectedStatusCodes(t *testing.T) {
	cases := []struct {
		method, path, body string
		want               int
	}{
		{http.MethodPatch, "/api/v1/projects/project-1/knowledge-memory/static-context", `{"static_context":{"summary":"updated"},"note":"manual"}`, http.StatusOK},
		{http.MethodPatch, "/api/v1/projects/project-1/knowledge-memory/style-guide", `{"style_guide":{"tone":"consistent"},"note":"manual"}`, http.StatusOK},
		{http.MethodPatch, "/api/v1/projects/project-1/knowledge-memory/recent-window-policy", `{"item_count":5,"token_limit":2000,"truncation_policy":"time"}`, http.StatusOK},
		{http.MethodGet, "/api/v1/projects/project-1/knowledge-memory/snapshots?page=1&page_size=20", ``, http.StatusOK},
		{http.MethodGet, "/api/v1/projects/project-1/knowledge-memory/context-preview?purpose=draft_generation&budget=2000", ``, http.StatusOK},
		{http.MethodPost, "/api/v1/projects/project-1/knowledge-memory/assemble-context", `{"purpose":"draft_generation","budget":2000}`, http.StatusOK},
		{http.MethodPost, "/api/v1/content-items/content-item-1/update-dynamic-state", `{"summary":"summary","changes":{"status":"changed"},"source_version_id":"version-1"}`, http.StatusOK},
		{http.MethodPost, "/api/v1/projects/project-1/consistency-reports", `{"range":{"latest":true},"scope":"project","severity_threshold":"low"}`, http.StatusAccepted},
		{http.MethodGet, "/api/v1/projects/project-1/consistency-reports", ``, http.StatusOK},
		{http.MethodGet, "/api/v1/projects/project-1/consistency-reports/report-1", ``, http.StatusOK},
	}
	for _, tc := range cases {
		var body []byte
		if tc.body != "" {
			body = []byte(tc.body)
		}
		rr := iteration6Request(t, tc.method, tc.path, body)
		if rr.Code != tc.want {
			t.Fatalf("%s %s = %d, want %d body=%s", tc.method, tc.path, rr.Code, tc.want, rr.Body.String())
		}
	}
}

// @Test
func TestTask05MemoryAPIRejectsForbiddenAccessWith403(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-forbidden/knowledge-memory", nil)
	req.Header.Set("Authorization", "Bearer forbidden")
	rr := httptest.NewRecorder()
	iteration6Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("GET knowledge-memory with forbidden token = %d, want 403", rr.Code)
	}
}
