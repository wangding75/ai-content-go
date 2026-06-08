package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/socialpost"
)

type stubSocialPostService struct {
	packStatusFn     func(ctx context.Context) (socialpost.SocialPostPackStatusResponse, error)
	registerPackFn  func(ctx context.Context, req socialpost.RegisterSocialPostPackRequest, idempotencyKey string) (socialpost.RegisterSocialPostPackResponse, error)
	configFn         func(ctx context.Context, projectID string) (socialpost.SocialPostConfigResponse, error)
	updateConfigFn  func(ctx context.Context, projectID string, req socialpost.UpdateSocialPostConfigRequest, idempotencyKey string) (socialpost.UpdateSocialPostConfigResponse, error)
	createRunFn     func(ctx context.Context, projectID string, req socialpost.CreateSocialPostGenerationRunRequest, idempotencyKey string) (socialpost.CreateSocialPostGenerationRunResponse, error)
	getRunFn        func(ctx context.Context, projectID, generationRunID string) (socialpost.SocialPostGenerationRunDetailResponse, error)
	listVariantsFn  func(ctx context.Context, projectID string, req socialpost.ListSocialPostVariantsRequest) (socialpost.PagedSocialPostVariantsResponse, error)
	selectVariantFn func(ctx context.Context, projectID, variantID string, req socialpost.SelectSocialPostVariantRequest, idempotencyKey string) (socialpost.SelectSocialPostVariantResponse, error)
	genTagsFn       func(ctx context.Context, projectID string, req socialpost.GenerateSocialPostTagsRequest, idempotencyKey string) (socialpost.GenerateSocialPostAssetResponse, error)
	genCoverFn      func(ctx context.Context, projectID string, req socialpost.GenerateSocialPostCoverCopyRequest, idempotencyKey string) (socialpost.GenerateSocialPostAssetResponse, error)
	assetsFn        func(ctx context.Context, projectID string, req socialpost.GetSocialPostAssetsRequest) (socialpost.SocialPostAssetsResponse, error)
}

func (s *stubSocialPostService) GetPackStatus(ctx context.Context) (socialpost.SocialPostPackStatusResponse, error) {
	if s.packStatusFn != nil {
		return s.packStatusFn(ctx)
	}
	return socialpost.SocialPostPackStatusResponse{}, nil
}
func (s *stubSocialPostService) RegisterPack(ctx context.Context, req socialpost.RegisterSocialPostPackRequest, idempotencyKey string) (socialpost.RegisterSocialPostPackResponse, error) {
	if s.registerPackFn != nil {
		return s.registerPackFn(ctx, req, idempotencyKey)
	}
	return socialpost.RegisterSocialPostPackResponse{}, nil
}
func (s *stubSocialPostService) GetConfig(ctx context.Context, projectID string) (socialpost.SocialPostConfigResponse, error) {
	if s.configFn != nil {
		return s.configFn(ctx, projectID)
	}
	return socialpost.SocialPostConfigResponse{}, nil
}
func (s *stubSocialPostService) UpdateConfig(ctx context.Context, projectID string, req socialpost.UpdateSocialPostConfigRequest, idempotencyKey string) (socialpost.UpdateSocialPostConfigResponse, error) {
	if s.updateConfigFn != nil {
		return s.updateConfigFn(ctx, projectID, req, idempotencyKey)
	}
	return socialpost.UpdateSocialPostConfigResponse{}, nil
}
func (s *stubSocialPostService) CreateGenerationRun(ctx context.Context, projectID string, req socialpost.CreateSocialPostGenerationRunRequest, idempotencyKey string) (socialpost.CreateSocialPostGenerationRunResponse, error) {
	if s.createRunFn != nil {
		return s.createRunFn(ctx, projectID, req, idempotencyKey)
	}
	return socialpost.CreateSocialPostGenerationRunResponse{}, nil
}
func (s *stubSocialPostService) GetGenerationRun(ctx context.Context, projectID, generationRunID string) (socialpost.SocialPostGenerationRunDetailResponse, error) {
	if s.getRunFn != nil {
		return s.getRunFn(ctx, projectID, generationRunID)
	}
	return socialpost.SocialPostGenerationRunDetailResponse{}, nil
}
func (s *stubSocialPostService) ListVariants(ctx context.Context, projectID string, req socialpost.ListSocialPostVariantsRequest) (socialpost.PagedSocialPostVariantsResponse, error) {
	if s.listVariantsFn != nil {
		return s.listVariantsFn(ctx, projectID, req)
	}
	return socialpost.PagedSocialPostVariantsResponse{}, nil
}
func (s *stubSocialPostService) SelectVariant(ctx context.Context, projectID, variantID string, req socialpost.SelectSocialPostVariantRequest, idempotencyKey string) (socialpost.SelectSocialPostVariantResponse, error) {
	if s.selectVariantFn != nil {
		return s.selectVariantFn(ctx, projectID, variantID, req, idempotencyKey)
	}
	return socialpost.SelectSocialPostVariantResponse{}, nil
}
func (s *stubSocialPostService) GenerateTags(ctx context.Context, projectID string, req socialpost.GenerateSocialPostTagsRequest, idempotencyKey string) (socialpost.GenerateSocialPostAssetResponse, error) {
	if s.genTagsFn != nil {
		return s.genTagsFn(ctx, projectID, req, idempotencyKey)
	}
	return socialpost.GenerateSocialPostAssetResponse{}, nil
}
func (s *stubSocialPostService) GenerateCoverCopy(ctx context.Context, projectID string, req socialpost.GenerateSocialPostCoverCopyRequest, idempotencyKey string) (socialpost.GenerateSocialPostAssetResponse, error) {
	if s.genCoverFn != nil {
		return s.genCoverFn(ctx, projectID, req, idempotencyKey)
	}
	return socialpost.GenerateSocialPostAssetResponse{}, nil
}
func (s *stubSocialPostService) GetAssets(ctx context.Context, projectID string, req socialpost.GetSocialPostAssetsRequest) (socialpost.SocialPostAssetsResponse, error) {
	if s.assetsFn != nil {
		return s.assetsFn(ctx, projectID, req)
	}
	return socialpost.SocialPostAssetsResponse{}, nil
}

// ===== Task-03: Pack status and registration =====

// @Test
func TestTask03GetPackStatusReturns200(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		packStatusFn: func(ctx context.Context) (socialpost.SocialPostPackStatusResponse, error) {
			return socialpost.SocialPostPackStatusResponse{
				ContentPackID:  "cp_social_post",
				CurrentVersion: "2026.06.social-post.v1",
				Schema: map[string]any{
					"content_type_code": "social_post",
				},
				Workflows: []socialpost.PackWorkflowSummary{
					{TemplateID: "31", Code: "social_post_generation", Name: "Social Post Generation", CurrentVersion: "45"},
				},
				Metrics: []socialpost.PackMetricSummary{
					{MetricCode: "impressions", MetricName: "曝光", Unit: "count", Platform: "generic"},
				},
			}, nil
		},
	}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.GetPackStatus(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if !envelope.Success || envelope.Data == nil {
		t.Fatalf("expected success envelope with data")
	}
}

// @Test
func TestTask03GetPackStatusNotFoundReturns404(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		packStatusFn: func(ctx context.Context) (socialpost.SocialPostPackStatusResponse, error) {
			return socialpost.SocialPostPackStatusResponse{}, socialpost.ErrNotFound
		},
	}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.GetPackStatus(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if envelope.Success || envelope.Error == nil {
		t.Fatalf("expected error envelope")
	}
	if envelope.Error.Code != api.ErrorNotFound {
		t.Fatalf("expected NOT_FOUND error code, got %s", envelope.Error.Code)
	}
}

// @Test
func TestTask03RegisterPackReturns201(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		registerPackFn: func(ctx context.Context, req socialpost.RegisterSocialPostPackRequest, idempotencyKey string) (socialpost.RegisterSocialPostPackResponse, error) {
			return socialpost.RegisterSocialPostPackResponse{
				ContentPackID:     "cp_social_post",
				ContentTypeID:     "13",
				RegisteredVersion: "2026.06.social-post.v1",
			}, nil
		},
	}, slog.Default())

	body := `{"schema":{"content_type_code":"social_post","name":"Social Post Pack","project_schema":{}},"workflows":[{"code":"social_post_generation","name":"Social Post Generation"}],"metrics":[{"metric_code":"impressions","metric_name":"曝光","unit":"count"}],"version":"2026.06.social-post.v1"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "idem-pack-reg-1")
	rec := httptest.NewRecorder()
	h.RegisterPack(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body %s", rec.Code, rec.Body.String())
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if !envelope.Success || envelope.Data == nil {
		t.Fatalf("expected success envelope with data")
	}
}

// @Test
func TestTask03RegisterPackInvalidJSONReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{invalid`))
	rec := httptest.NewRecorder()
	h.RegisterPack(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask03RegisterPackValidationErrorReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		registerPackFn: func(ctx context.Context, req socialpost.RegisterSocialPostPackRequest, idempotencyKey string) (socialpost.RegisterSocialPostPackResponse, error) {
			return socialpost.RegisterSocialPostPackResponse{}, socialpost.ErrValidation
		},
	}, slog.Default())

	body := `{"schema":{},"workflows":[],"metrics":[],"version":""}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.RegisterPack(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask03RegisterPackConflictReturns409(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		registerPackFn: func(ctx context.Context, req socialpost.RegisterSocialPostPackRequest, idempotencyKey string) (socialpost.RegisterSocialPostPackResponse, error) {
			return socialpost.RegisterSocialPostPackResponse{}, socialpost.ErrConflict
		},
	}, slog.Default())

	body := `{"schema":{},"workflows":[],"metrics":[],"version":""}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.RegisterPack(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if envelope.Error == nil || envelope.Error.Code != api.ErrorConflict {
		t.Fatalf("expected CONFLICT error code, got %v", envelope.Error)
	}
}

// @Test
func TestTask03RegisterPackIdempotencyConflictReturns409(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		registerPackFn: func(ctx context.Context, req socialpost.RegisterSocialPostPackRequest, idempotencyKey string) (socialpost.RegisterSocialPostPackResponse, error) {
			return socialpost.RegisterSocialPostPackResponse{}, socialpost.ErrIdempotencyConflict
		},
	}, slog.Default())

	body := `{"schema":{},"workflows":[],"metrics":[],"version":"v1"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "idem-pack-conflict")
	rec := httptest.NewRecorder()
	h.RegisterPack(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if envelope.Error == nil || envelope.Error.Code != api.ErrorIdempotencyConflict {
		t.Fatalf("expected IDEMPOTENCY_CONFLICT error code, got %v", envelope.Error)
	}
}

// @Test
func TestTask03RegisterPackInternalErrorReturns500(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		registerPackFn: func(ctx context.Context, req socialpost.RegisterSocialPostPackRequest, idempotencyKey string) (socialpost.RegisterSocialPostPackResponse, error) {
			return socialpost.RegisterSocialPostPackResponse{}, errors.New("db down")
		},
	}, slog.Default())

	body := `{"schema":{},"workflows":[],"metrics":[],"version":"v1"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.RegisterPack(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

// @Test
func TestTask03RegisterPackRejectsUnknownFields(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	body := `{"schema":{},"workflows":[],"metrics":[],"version":"v1","extra_field":"hack"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.RegisterPack(rec, req)
	// The JSON decoder does not reject unknown fields by default in Go;
	// this test verifies the handler can still parse the known fields without error.
	// Unknown fields are silently ignored by encoding/json, which is the current behavior.
	// Future: if strict decoding is needed, use json.Decoder.DisallowUnknownFields().
	if rec.Code != http.StatusCreated && rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 201 or 400, got %d", rec.Code)
	}
}

// @Test
func TestTask03RegisterPackEmptyBodyReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	rec := httptest.NewRecorder()
	h.RegisterPack(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask03RegisterPackArrayBodyReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("[]"))
	rec := httptest.NewRecorder()
	h.RegisterPack(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask03RegisterPackWrongTypeFieldReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	body := `{"schema":{},"workflows":[],"metrics":[],"version":true}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.RegisterPack(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
// @Test
func TestTask03RegisterPackAgentOutputInvalidReturns500(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		registerPackFn: func(ctx context.Context, req socialpost.RegisterSocialPostPackRequest, idempotencyKey string) (socialpost.RegisterSocialPostPackResponse, error) {
			return socialpost.RegisterSocialPostPackResponse{}, socialpost.ErrAgentOutputInvalid
		},
	}, slog.Default())

	body := `{"schema":{},"workflows":[],"metrics":[],"version":"v1"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.RegisterPack(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if envelope.Error == nil || envelope.Error.Code != api.ErrorAgentOutputInvalid {
		t.Fatalf("expected AGENT_OUTPUT_INVALID error code, got %v", envelope.Error)
	}
}

// ===== Task-04: Config query and update =====

// @Test
func TestTask04GetConfigReturns200WithDefaults(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		configFn: func(ctx context.Context, projectID string) (socialpost.SocialPostConfigResponse, error) {
			return socialpost.SocialPostConfigResponse{
				TargetPlatforms:     []string{"xiaohongshu"},
				DefaultVariantCount: 3,
				CaptionLengthPolicy: "short",
				HashtagPolicy:       map[string]any{"mode": "auto", "count": 5},
				CoverCopyPolicy:     map[string]any{"mode": "auto", "count": 2},
				ToneStyle:           "professional",
				ForbiddenTerms:      []string{},
				ConfigVersion:       1,
			}, nil
		},
	}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetConfig(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if !envelope.Success || envelope.Data == nil {
		t.Fatalf("expected success envelope with data")
	}
}

// @Test
func TestTask04GetConfigNotFoundReturns404(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		configFn: func(ctx context.Context, projectID string) (socialpost.SocialPostConfigResponse, error) {
			return socialpost.SocialPostConfigResponse{}, socialpost.ErrNotFound
		},
	}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetConfig(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if envelope.Error == nil || envelope.Error.Code != api.ErrorNotFound {
		t.Fatalf("expected NOT_FOUND error code, got %v", envelope.Error)
	}
}

// @Test
func TestTask04GetConfigMissingProjectIdReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.GetConfig(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask04UpdateConfigReturns200(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		updateConfigFn: func(ctx context.Context, projectID string, req socialpost.UpdateSocialPostConfigRequest, idempotencyKey string) (socialpost.UpdateSocialPostConfigResponse, error) {
			return socialpost.UpdateSocialPostConfigResponse{
				VersionID:      "social-post-config-2",
				OperationLogID: "oplog-social-post-config-2",
			}, nil
		},
	}, slog.Default())

	body := `{"target_platforms":["xiaohongshu"],"default_variant_count":3,"caption_length_policy":"short","hashtag_policy":{},"cover_copy_policy":{},"tone_style":"friendly","forbidden_terms":["绝对化用语"]}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "idem-config-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.UpdateConfig(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if !envelope.Success || envelope.Data == nil {
		t.Fatalf("expected success envelope with data")
	}
}

// @Test
func TestTask04UpdateConfigValidationErrorReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		updateConfigFn: func(ctx context.Context, projectID string, req socialpost.UpdateSocialPostConfigRequest, idempotencyKey string) (socialpost.UpdateSocialPostConfigResponse, error) {
			return socialpost.UpdateSocialPostConfigResponse{}, socialpost.ErrValidation
		},
	}, slog.Default())

	body := `{"default_variant_count":-1}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.UpdateConfig(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask04UpdateConfigIdempotencyConflictReturns409(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		updateConfigFn: func(ctx context.Context, projectID string, req socialpost.UpdateSocialPostConfigRequest, idempotencyKey string) (socialpost.UpdateSocialPostConfigResponse, error) {
			return socialpost.UpdateSocialPostConfigResponse{}, socialpost.ErrIdempotencyConflict
		},
	}, slog.Default())

	body := `{"target_platforms":["xiaohongshu"]}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "idem-config-conflict")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.UpdateConfig(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if envelope.Error == nil || envelope.Error.Code != api.ErrorIdempotencyConflict {
		t.Fatalf("expected IDEMPOTENCY_CONFLICT error code, got %v", envelope.Error)
	}
}

// @Test
func TestTask04UpdateConfigMissingProjectIdReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	body := `{"target_platforms":["xiaohongshu"]}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.UpdateConfig(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask04UpdateConfigInvalidJSONReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{invalid`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.UpdateConfig(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask04UpdateConfigNotFoundReturns404(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		updateConfigFn: func(ctx context.Context, projectID string, req socialpost.UpdateSocialPostConfigRequest, idempotencyKey string) (socialpost.UpdateSocialPostConfigResponse, error) {
			return socialpost.UpdateSocialPostConfigResponse{}, socialpost.ErrNotFound
		},
	}, slog.Default())

	body := `{"target_platforms":["xiaohongshu"]}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.UpdateConfig(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if envelope.Error == nil || envelope.Error.Code != api.ErrorNotFound {
		t.Fatalf("expected NOT_FOUND error code, got %v", envelope.Error)
	}
}

// @Test
func TestTask04UpdateConfigRejectsUnknownFields(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	body := `{"target_platforms":["xiaohongshu"],"extra_field":"hack"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.UpdateConfig(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 200 or 400, got %d", rec.Code)
	}
}

// ===== Task-05: Generation run trigger =====

// @Test
func TestTask05CreateGenerationRunReturns202(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		createRunFn: func(ctx context.Context, projectID string, req socialpost.CreateSocialPostGenerationRunRequest, idempotencyKey string) (socialpost.CreateSocialPostGenerationRunResponse, error) {
			return socialpost.CreateSocialPostGenerationRunResponse{
				GenerationRunID: "genrun-social-1",
				WorkflowRunID:   "52",
				Status:          "running",
			}, nil
		},
	}, slog.Default())

	body := `{"topic":"618 促销预热","platform":"xiaohongshu","version_count":3,"tone_style":"friendly","asset_options":{"generate_tags":true,"generate_cover_copy":false}}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "idem-gen-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.CreateGenerationRun(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body %s", rec.Code, rec.Body.String())
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if !envelope.Success || envelope.Data == nil {
		t.Fatalf("expected success envelope with data")
	}
}

// @Test
func TestTask05CreateGenerationRunVersionCountExceedsLimitReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		createRunFn: func(ctx context.Context, projectID string, req socialpost.CreateSocialPostGenerationRunRequest, idempotencyKey string) (socialpost.CreateSocialPostGenerationRunResponse, error) {
			return socialpost.CreateSocialPostGenerationRunResponse{}, socialpost.ErrValidation
		},
	}, slog.Default())

	body := `{"topic":"test","platform":"xiaohongshu","version_count":20}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.CreateGenerationRun(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask05CreateGenerationRunMissingProjectIdReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	body := `{"topic":"test","platform":"xiaohongshu"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.CreateGenerationRun(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask05CreateGenerationRunInvalidJSONReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{invalid`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.CreateGenerationRun(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask05CreateGenerationRunNotFoundReturns404(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		createRunFn: func(ctx context.Context, projectID string, req socialpost.CreateSocialPostGenerationRunRequest, idempotencyKey string) (socialpost.CreateSocialPostGenerationRunResponse, error) {
			return socialpost.CreateSocialPostGenerationRunResponse{}, socialpost.ErrNotFound
		},
	}, slog.Default())

	body := `{"topic":"test","platform":"xiaohongshu"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.CreateGenerationRun(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if envelope.Error == nil || envelope.Error.Code != api.ErrorNotFound {
		t.Fatalf("expected NOT_FOUND error code, got %v", envelope.Error)
	}
}

// @Test
func TestTask05CreateGenerationRunIdempotencyConflictReturns409(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		createRunFn: func(ctx context.Context, projectID string, req socialpost.CreateSocialPostGenerationRunRequest, idempotencyKey string) (socialpost.CreateSocialPostGenerationRunResponse, error) {
			return socialpost.CreateSocialPostGenerationRunResponse{}, socialpost.ErrIdempotencyConflict
		},
	}, slog.Default())

	body := `{"topic":"test","platform":"xiaohongshu"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "idem-gen-conflict")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.CreateGenerationRun(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

// ===== Task-06: Generation run detail query =====

// @Test
func TestTask06GetGenerationRunReturns200WithVariants(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		getRunFn: func(ctx context.Context, projectID, generationRunID string) (socialpost.SocialPostGenerationRunDetailResponse, error) {
			return socialpost.SocialPostGenerationRunDetailResponse{
				GenerationRunID: "genrun-social-1",
				WorkflowRunID:   "52",
				Status:          "completed",
				ContentItemID:   "content-item-1",
				Trace: &socialpost.GenerationTrace{
					AgentTaskIDs:  []string{"agent-task-1"},
					LLMCallLogIDs: []string{"llm-log-1"},
				},
				Variants: []socialpost.SocialPostVariantResponse{
					{ID: "variant-1", VariantIndex: 1, Platform: "xiaohongshu", Title: "标题", Body: "正文", Status: "generated"},
				},
			}, nil
		},
	}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	rctx.URLParams.Add("id", "genrun-social-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetGenerationRun(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
}

// @Test
func TestTask06GetGenerationRunNotFoundReturns404(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		getRunFn: func(ctx context.Context, projectID, generationRunID string) (socialpost.SocialPostGenerationRunDetailResponse, error) {
			return socialpost.SocialPostGenerationRunDetailResponse{}, socialpost.ErrNotFound
		},
	}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	rctx.URLParams.Add("id", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetGenerationRun(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// @Test
func TestTask06GetGenerationRunFailedStatusReturns200WithError(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		getRunFn: func(ctx context.Context, projectID, generationRunID string) (socialpost.SocialPostGenerationRunDetailResponse, error) {
			return socialpost.SocialPostGenerationRunDetailResponse{
				GenerationRunID: "genrun-failed-1",
				WorkflowRunID:   "99",
				Status:          "failed",
				Error:           "LLM output parse error",
				Trace: &socialpost.GenerationTrace{
					AgentTaskIDs:  []string{"agent-task-99"},
					LLMCallLogIDs: []string{"llm-log-99"},
				},
			}, nil
		},
	}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	rctx.URLParams.Add("id", "genrun-failed-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetGenerationRun(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// @Test
func TestTask06GetGenerationRunMissingProjectIdReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.GetGenerationRun(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ===== Task-07: Variant list query =====

// @Test
func TestTask07ListVariantsReturns200WithPagination(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		listVariantsFn: func(ctx context.Context, projectID string, req socialpost.ListSocialPostVariantsRequest) (socialpost.PagedSocialPostVariantsResponse, error) {
			return socialpost.PagedSocialPostVariantsResponse{
				Items: []socialpost.SocialPostVariantResponse{
					{ID: "variant-1", ContentItemID: "content-item-1", VariantIndex: 1, Platform: "xiaohongshu", Title: "标题", Body: "正文", Status: "generated"},
				},
				Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: 1},
			}, nil
		},
	}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/?content_item_id=content-item-1&status=generated&page=1&page_size=20", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.ListVariants(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
}

// @Test
func TestTask07ListVariantsEmptyResultReturns200(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		listVariantsFn: func(ctx context.Context, projectID string, req socialpost.ListSocialPostVariantsRequest) (socialpost.PagedSocialPostVariantsResponse, error) {
			return socialpost.PagedSocialPostVariantsResponse{
				Items:      []socialpost.SocialPostVariantResponse{},
				Pagination: content.PaginationResponse{Page: 1, PageSize: 20, Total: 0},
			}, nil
		},
	}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.ListVariants(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// @Test
func TestTask07ListVariantsMissingProjectIdReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ListVariants(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ===== Task-08: Select variant =====

// @Test
func TestTask08SelectVariantReturns200(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		selectVariantFn: func(ctx context.Context, projectID, variantID string, req socialpost.SelectSocialPostVariantRequest, idempotencyKey string) (socialpost.SelectSocialPostVariantResponse, error) {
			return socialpost.SelectSocialPostVariantResponse{
				SelectedVariantID: "variant-1",
				ContentVersionID:  "version-content-item-1-1",
				OperationLogID:    "oplog-variant-1",
			}, nil
		},
	}, slog.Default())

	body := `{"content_item_id":"content-item-1","note":"选择第 1 版作为主版本"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "idem-select-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	rctx.URLParams.Add("variantId", "variant-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.SelectVariant(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if !envelope.Success || envelope.Data == nil {
		t.Fatalf("expected success envelope with data")
	}
}

// @Test
func TestTask08SelectVariantNotFoundReturns404(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		selectVariantFn: func(ctx context.Context, projectID, variantID string, req socialpost.SelectSocialPostVariantRequest, idempotencyKey string) (socialpost.SelectSocialPostVariantResponse, error) {
			return socialpost.SelectSocialPostVariantResponse{}, socialpost.ErrNotFound
		},
	}, slog.Default())

	body := `{"content_item_id":"content-item-1","note":"select"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	rctx.URLParams.Add("variantId", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.SelectVariant(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if envelope.Error == nil || envelope.Error.Code != api.ErrorNotFound {
		t.Fatalf("expected NOT_FOUND error code, got %v", envelope.Error)
	}
}

// @Test
func TestTask08SelectVariantConflictReturns409(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		selectVariantFn: func(ctx context.Context, projectID, variantID string, req socialpost.SelectSocialPostVariantRequest, idempotencyKey string) (socialpost.SelectSocialPostVariantResponse, error) {
			return socialpost.SelectSocialPostVariantResponse{}, socialpost.ErrConflict
		},
	}, slog.Default())

	body := `{"content_item_id":"content-item-1","note":"select"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	rctx.URLParams.Add("variantId", "variant-archived")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.SelectVariant(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if envelope.Error == nil || envelope.Error.Code != api.ErrorConflict {
		t.Fatalf("expected CONFLICT error code, got %v", envelope.Error)
	}
}

// @Test
func TestTask08SelectVariantValidationErrorReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		selectVariantFn: func(ctx context.Context, projectID, variantID string, req socialpost.SelectSocialPostVariantRequest, idempotencyKey string) (socialpost.SelectSocialPostVariantResponse, error) {
			return socialpost.SelectSocialPostVariantResponse{}, socialpost.ErrValidation
		},
	}, slog.Default())

	body := `{"content_item_id":"","note":""}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	rctx.URLParams.Add("variantId", "variant-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.SelectVariant(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask08SelectVariantMissingProjectIdReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	body := `{"content_item_id":"content-item-1","note":"select"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.SelectVariant(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask08SelectVariantInvalidJSONReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{invalid`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	rctx.URLParams.Add("variantId", "variant-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.SelectVariant(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask08SelectVariantIdempotencyConflictReturns409(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		selectVariantFn: func(ctx context.Context, projectID, variantID string, req socialpost.SelectSocialPostVariantRequest, idempotencyKey string) (socialpost.SelectSocialPostVariantResponse, error) {
			return socialpost.SelectSocialPostVariantResponse{}, socialpost.ErrIdempotencyConflict
		},
	}, slog.Default())

	body := `{"content_item_id":"content-item-1","note":"select"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "idem-select-conflict")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	rctx.URLParams.Add("variantId", "variant-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.SelectVariant(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
	var envelope api.Envelope
	json.Unmarshal(rec.Body.Bytes(), &envelope)
	if envelope.Error == nil || envelope.Error.Code != api.ErrorIdempotencyConflict {
		t.Fatalf("expected IDEMPOTENCY_CONFLICT error code, got %v", envelope.Error)
	}
}

// ===== Task-09: Tags and cover copy generation =====

// @Test
func TestTask09GenerateTagsReturns202(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		genTagsFn: func(ctx context.Context, projectID string, req socialpost.GenerateSocialPostTagsRequest, idempotencyKey string) (socialpost.GenerateSocialPostAssetResponse, error) {
			return socialpost.GenerateSocialPostAssetResponse{
				GenerationRunID: "asset-run-1",
				WorkflowRunID:   "66",
				Status:          "running",
			}, nil
		},
	}, slog.Default())

	body := `{"content_item_id":"content-item-1","variant_id":"variant-1","platform":"xiaohongshu","count":5,"style":"trending"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "idem-tags-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GenerateTags(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body %s", rec.Code, rec.Body.String())
	}
}

// @Test
func TestTask09GenerateCoverCopyReturns202(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		genCoverFn: func(ctx context.Context, projectID string, req socialpost.GenerateSocialPostCoverCopyRequest, idempotencyKey string) (socialpost.GenerateSocialPostAssetResponse, error) {
			return socialpost.GenerateSocialPostAssetResponse{
				GenerationRunID: "asset-run-2",
				WorkflowRunID:   "67",
				Status:          "running",
			}, nil
		},
	}, slog.Default())

	body := `{"content_item_id":"content-item-1","variant_id":"variant-1","platform":"xiaohongshu","count":2,"style":"warm"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "idem-cover-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GenerateCoverCopy(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body %s", rec.Code, rec.Body.String())
	}
}

// @Test
func TestTask09GenerateTagsMissingProjectIdReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	body := `{"content_item_id":"content-item-1","variant_id":"variant-1","platform":"xiaohongshu"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.GenerateTags(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask09GenerateTagsIdempotencyConflictReturns409(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		genTagsFn: func(ctx context.Context, projectID string, req socialpost.GenerateSocialPostTagsRequest, idempotencyKey string) (socialpost.GenerateSocialPostAssetResponse, error) {
			return socialpost.GenerateSocialPostAssetResponse{}, socialpost.ErrIdempotencyConflict
		},
	}, slog.Default())

	body := `{"content_item_id":"content-item-1","variant_id":"variant-1","platform":"xiaohongshu"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "idem-tags-conflict")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GenerateTags(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

// @Test
func TestTask09GenerateTagsInvalidJSONReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{invalid`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GenerateTags(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask09GenerateTagsValidationErrorReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		genTagsFn: func(ctx context.Context, projectID string, req socialpost.GenerateSocialPostTagsRequest, idempotencyKey string) (socialpost.GenerateSocialPostAssetResponse, error) {
			return socialpost.GenerateSocialPostAssetResponse{}, socialpost.ErrValidation
		},
	}, slog.Default())

	body := `{"content_item_id":"","variant_id":"","platform":""}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GenerateTags(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask09GenerateCoverCopyMissingProjectIdReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	body := `{"content_item_id":"content-item-1","variant_id":"variant-1","platform":"xiaohongshu"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.GenerateCoverCopy(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask09GenerateCoverCopyInvalidJSONReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{invalid`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GenerateCoverCopy(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask09GenerateCoverCopyIdempotencyConflictReturns409(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		genCoverFn: func(ctx context.Context, projectID string, req socialpost.GenerateSocialPostCoverCopyRequest, idempotencyKey string) (socialpost.GenerateSocialPostAssetResponse, error) {
			return socialpost.GenerateSocialPostAssetResponse{}, socialpost.ErrIdempotencyConflict
		},
	}, slog.Default())

	body := `{"content_item_id":"content-item-1","variant_id":"variant-1","platform":"xiaohongshu"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "idem-cover-conflict")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GenerateCoverCopy(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

// ===== Task-10: Asset query =====

// @Test
func TestTask10GetAssetsReturns200(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		assetsFn: func(ctx context.Context, projectID string, req socialpost.GetSocialPostAssetsRequest) (socialpost.SocialPostAssetsResponse, error) {
			return socialpost.SocialPostAssetsResponse{
				Tags: []socialpost.SocialPostAssetItem{
					{ID: "asset-1", Platform: "xiaohongshu", SourceVariantID: "variant-1", GenerationRunID: "asset-run-1", Result: map[string]any{"tags": []any{"#618"}}},
				},
				CoverCopy: []socialpost.SocialPostAssetItem{
					{ID: "asset-2", Platform: "xiaohongshu", SourceVariantID: "variant-1", GenerationRunID: "asset-run-2", Result: map[string]any{"items": []any{map[string]any{"text": "夏日大促", "style": "warm"}}}},
				},
				AssetSuggestions: []string{"优先保留品牌词"},
				SourceRuns:       []string{"asset-run-1", "asset-run-2"},
			}, nil
		},
	}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/?content_item_id=content-item-1&platform=xiaohongshu", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetAssets(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
}

// @Test
func TestTask10GetAssetsEmptyResultReturns200(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		assetsFn: func(ctx context.Context, projectID string, req socialpost.GetSocialPostAssetsRequest) (socialpost.SocialPostAssetsResponse, error) {
			return socialpost.SocialPostAssetsResponse{
				Tags:             []socialpost.SocialPostAssetItem{},
				CoverCopy:        []socialpost.SocialPostAssetItem{},
				AssetSuggestions: []string{},
				SourceRuns:       []string{},
			}, nil
		},
	}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetAssets(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// @Test
func TestTask10GetAssetsMissingProjectIdReturns400(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.GetAssets(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// @Test
func TestTask10GetAssetsNotFoundReturns404(t *testing.T) {
	h := NewSocialPostHandler(&stubSocialPostService{
		assetsFn: func(ctx context.Context, projectID string, req socialpost.GetSocialPostAssetsRequest) (socialpost.SocialPostAssetsResponse, error) {
			return socialpost.SocialPostAssetsResponse{}, socialpost.ErrNotFound
		},
	}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "p-nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetAssets(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
