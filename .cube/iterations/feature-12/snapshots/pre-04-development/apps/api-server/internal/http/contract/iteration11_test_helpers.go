package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	serverhttp "github.com/wangding75/ai-content-go/apps/api-server/internal/http"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/external"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/metrics"
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

type iteration11PublishJobLookup struct{}

func (iteration11PublishJobLookup) FindPublishJobContext(_ context.Context, publishJobID string) (*metrics.PublishJobContext, error) {
	if publishJobID == "job-001" {
		return &metrics.PublishJobContext{
			ProjectID:        "project-1",
			ContentItemID:    "content-item-1",
			ContentVersionID: "version-1",
			TargetID:         "publish-target-1",
			ContentType:      "article",
			Platform:         "manual",
		}, nil
	}
	return nil, nil
}

var (
	iteration11MetricsOnce  sync.Once
	iteration11MetricsSvc   metrics.Service
	iteration11ExternalOnce sync.Once
	iteration11ExternalSvc  external.Service
)

func iteration11MetricsService() metrics.Service {
	iteration11MetricsOnce.Do(func() {
		store := metrics.NewMemoryStore()
		iteration11MetricsSvc = metrics.NewServiceWithPublishLookup(store, iteration11PublishJobLookup{})
	})
	return iteration11MetricsSvc
}

func iteration11ExternalService() external.Service {
	iteration11ExternalOnce.Do(func() {
		iteration11ExternalSvc = external.NewService()
	})
	return iteration11ExternalSvc
}

func iteration11Router() http.Handler {
	return serverhttp.NewRouter(iteration11SystemService{}, nil,
		serverhttp.WithMetricsService(iteration11MetricsService()),
		serverhttp.WithExternalService(iteration11ExternalService()),
	)
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
