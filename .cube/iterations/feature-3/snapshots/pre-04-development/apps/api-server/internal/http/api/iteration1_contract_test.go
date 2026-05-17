package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
)

// @Test
func TestIteration1ErrorEnvelopeCarriesConflictAndFieldDetails(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/content-types", nil)
	req.Header.Set("X-Request-Id", "req-content-type-conflict")
	w := httptest.NewRecorder()

	api.WriteError(w, req, http.StatusConflict, api.ErrorConflict, "content type code already exists", []api.ErrorDetail{{Field: "code", Reason: "must be unique"}})

	if w.Code != http.StatusConflict {
		t.Fatalf("expected conflict status, got %d", w.Code)
	}
	var body api.Envelope
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Success || body.Data != nil || body.Error == nil {
		t.Fatalf("expected error envelope without data, got %#v", body)
	}
	if body.Error.Code != api.ErrorConflict {
		t.Fatalf("expected CONFLICT, got %s", body.Error.Code)
	}
	if len(body.Error.Details) != 1 || body.Error.Details[0].Field != "code" || body.Error.Details[0].Reason != "must be unique" {
		t.Fatalf("expected field-level error details, got %#v", body.Error.Details)
	}
	if body.RequestID != "req-content-type-conflict" {
		t.Fatalf("expected request id to be preserved, got %q", body.RequestID)
	}
}

// @Test
func TestPaginationResponseContractUsesPageSizeTotalAndHasNext(t *testing.T) {
	payload := struct {
		Items      []string `json:"items"`
		Pagination struct {
			Page     int  `json:"page"`
			PageSize int  `json:"page_size"`
			Total    int  `json:"total"`
			HasNext  bool `json:"has_next"`
		} `json:"pagination"`
	}{Items: []string{"article"}}
	payload.Pagination.Page = 1
	payload.Pagination.PageSize = 20
	payload.Pagination.Total = 21
	payload.Pagination.HasNext = true

	req := httptest.NewRequest(http.MethodGet, "/api/v1/content-types", nil)
	w := httptest.NewRecorder()

	api.WriteSuccess(w, req, http.StatusOK, payload)

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Pagination struct {
				Page     int  `json:"page"`
				PageSize int  `json:"page_size"`
				Total    int  `json:"total"`
				HasNext  bool `json:"has_next"`
			} `json:"pagination"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success || body.Data.Pagination.Page != 1 || body.Data.Pagination.PageSize != 20 || body.Data.Pagination.Total != 21 || !body.Data.Pagination.HasNext {
		t.Fatalf("pagination contract mismatch: %#v", body.Data.Pagination)
	}
}
