package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/socialpost"
)

type SocialPostHandler struct {
	svc    socialpost.Service
	logger *slog.Logger
}

func NewSocialPostHandler(svc socialpost.Service, logger *slog.Logger) *SocialPostHandler {
	return &SocialPostHandler{svc: svc, logger: logger}
}

func mapSocialPostError(err error) int {
	if errors.Is(err, socialpost.ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, socialpost.ErrForbidden) {
		return http.StatusForbidden
	}
	if errors.Is(err, socialpost.ErrConflict) || errors.Is(err, socialpost.ErrIdempotencyConflict) {
		return http.StatusConflict
	}
	if errors.Is(err, socialpost.ErrValidation) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func mapSocialPostEnvelopeCode(err error) api.ErrorCode {
	if errors.Is(err, socialpost.ErrNotFound) {
		return api.ErrorNotFound
	}
	if errors.Is(err, socialpost.ErrConflict) {
		return api.ErrorConflict
	}
	if errors.Is(err, socialpost.ErrIdempotencyConflict) {
		return api.ErrorIdempotencyConflict
	}
	if errors.Is(err, socialpost.ErrValidation) {
		return api.ErrorValidation
	}
	if errors.Is(err, socialpost.ErrForbidden) {
		return api.ErrorForbidden
	}
	if errors.Is(err, socialpost.ErrAgentOutputInvalid) {
		return api.ErrorAgentOutputInvalid
	}
	return api.ErrorInternal
}

func (h *SocialPostHandler) GetPackStatus(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetPackStatus(r.Context())
	if err != nil {
		api.WriteError(w, r, mapSocialPostError(err), mapSocialPostEnvelopeCode(err), err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *SocialPostHandler) RegisterPack(w http.ResponseWriter, r *http.Request) {
	var req socialpost.RegisterSocialPostPackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid JSON body", []api.ErrorDetail{{Reason: err.Error()}})
		return
	}
	idempotencyKey := r.Header.Get("Idempotency-Key")

	resp, err := h.svc.RegisterPack(r.Context(), req, idempotencyKey)
	if err != nil {
		api.WriteError(w, r, mapSocialPostError(err), mapSocialPostEnvelopeCode(err), err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, resp)
}

func (h *SocialPostHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId", nil)
		return
	}

	resp, err := h.svc.GetConfig(r.Context(), projectID)
	if err != nil {
		api.WriteError(w, r, mapSocialPostError(err), mapSocialPostEnvelopeCode(err), err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *SocialPostHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId", nil)
		return
	}

	var req socialpost.UpdateSocialPostConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid JSON body", []api.ErrorDetail{{Reason: err.Error()}})
		return
	}
	idempotencyKey := r.Header.Get("Idempotency-Key")

	resp, err := h.svc.UpdateConfig(r.Context(), projectID, req, idempotencyKey)
	if err != nil {
		api.WriteError(w, r, mapSocialPostError(err), mapSocialPostEnvelopeCode(err), err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *SocialPostHandler) CreateGenerationRun(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId", nil)
		return
	}

	var req socialpost.CreateSocialPostGenerationRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid JSON body", []api.ErrorDetail{{Reason: err.Error()}})
		return
	}
	idempotencyKey := r.Header.Get("Idempotency-Key")

	resp, err := h.svc.CreateGenerationRun(r.Context(), projectID, req, idempotencyKey)
	if err != nil {
		api.WriteError(w, r, mapSocialPostError(err), mapSocialPostEnvelopeCode(err), err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusAccepted, resp)
}

func (h *SocialPostHandler) GetGenerationRun(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	id := chi.URLParam(r, "id")
	if projectID == "" || id == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId or id", nil)
		return
	}

	resp, err := h.svc.GetGenerationRun(r.Context(), projectID, id)
	if err != nil {
		api.WriteError(w, r, mapSocialPostError(err), mapSocialPostEnvelopeCode(err), err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *SocialPostHandler) ListVariants(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId", nil)
		return
	}

	page, pageSize := 1, 20
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	req := socialpost.ListSocialPostVariantsRequest{
		ContentItemID: r.URL.Query().Get("content_item_id"),
		Status:        r.URL.Query().Get("status"),
		Platform:      r.URL.Query().Get("platform"),
		Page:          page,
		PageSize:      pageSize,
	}

	resp, err := h.svc.ListVariants(r.Context(), projectID, req)
	if err != nil {
		api.WriteError(w, r, mapSocialPostError(err), mapSocialPostEnvelopeCode(err), err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *SocialPostHandler) SelectVariant(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	variantID := chi.URLParam(r, "variantId")
	if projectID == "" || variantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId or variantId", nil)
		return
	}

	var req socialpost.SelectSocialPostVariantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid JSON body", []api.ErrorDetail{{Reason: err.Error()}})
		return
	}
	idempotencyKey := r.Header.Get("Idempotency-Key")

	resp, err := h.svc.SelectVariant(r.Context(), projectID, variantID, req, idempotencyKey)
	if err != nil {
		api.WriteError(w, r, mapSocialPostError(err), mapSocialPostEnvelopeCode(err), err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}

func (h *SocialPostHandler) GenerateTags(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId", nil)
		return
	}

	var req socialpost.GenerateSocialPostTagsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid JSON body", []api.ErrorDetail{{Reason: err.Error()}})
		return
	}
	idempotencyKey := r.Header.Get("Idempotency-Key")

	resp, err := h.svc.GenerateTags(r.Context(), projectID, req, idempotencyKey)
	if err != nil {
		api.WriteError(w, r, mapSocialPostError(err), mapSocialPostEnvelopeCode(err), err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusAccepted, resp)
}

func (h *SocialPostHandler) GenerateCoverCopy(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId", nil)
		return
	}

	var req socialpost.GenerateSocialPostCoverCopyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid JSON body", []api.ErrorDetail{{Reason: err.Error()}})
		return
	}
	idempotencyKey := r.Header.Get("Idempotency-Key")

	resp, err := h.svc.GenerateCoverCopy(r.Context(), projectID, req, idempotencyKey)
	if err != nil {
		api.WriteError(w, r, mapSocialPostError(err), mapSocialPostEnvelopeCode(err), err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusAccepted, resp)
}

func (h *SocialPostHandler) GetAssets(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "missing projectId", nil)
		return
	}

	req := socialpost.GetSocialPostAssetsRequest{
		ContentItemID: r.URL.Query().Get("content_item_id"),
		Platform:      r.URL.Query().Get("platform"),
		VariantID:     r.URL.Query().Get("variant_id"),
	}

	resp, err := h.svc.GetAssets(r.Context(), projectID, req)
	if err != nil {
		api.WriteError(w, r, mapSocialPostError(err), mapSocialPostEnvelopeCode(err), err.Error(), nil)
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, resp)
}