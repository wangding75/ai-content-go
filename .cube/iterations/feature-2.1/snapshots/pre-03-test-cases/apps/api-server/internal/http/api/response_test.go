package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
)

func TestWriteErrorRejectsSuccessfulHTTPStatus(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	api.WriteError(w, req, http.StatusOK, api.ErrorInternal, "boom", nil)

	if w.Code < 400 {
		t.Fatalf("expected error response to use non-2xx status, got %d", w.Code)
	}

	var body api.Envelope
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Success {
		t.Fatalf("expected success=false")
	}
}
