package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/llm"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/prompt"
)

const maxJSONBodyBytes = 1 << 20

type ContentHandler struct {
	service content.Service
	logger  *slog.Logger
}

func NewContentHandler(service content.Service, logger *slog.Logger) *ContentHandler {
	return &ContentHandler{service: service, logger: logger}
}

func (h *ContentHandler) ListContentTypes(w http.ResponseWriter, r *http.Request) {
	req, err := parseListContentTypesRequest(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid query parameters", nil)
		return
	}
	data, err := h.service.ListContentTypes(r.Context(), req)
	if err != nil {
		writeContentError(w, r, err, "content types failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *ContentHandler) CreateContentType(w http.ResponseWriter, r *http.Request) {
	var req content.CreateContentTypeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	data, err := h.service.CreateContentType(r.Context(), req)
	if err != nil {
		writeContentError(w, r, err, "create content type failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, data)
}

func (h *ContentHandler) ProjectSchema(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.ProjectSchema(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeContentError(w, r, err, "project schema failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *ContentHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	req, err := parseListProjectsRequest(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid query parameters", nil)
		return
	}
	data, err := h.service.ListProjects(r.Context(), req)
	if err != nil {
		writeContentError(w, r, err, "projects failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *ContentHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req content.CreateProjectRequest
	if err := decodeJSON(w, r, &req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	data, err := h.service.CreateProject(r.Context(), req)
	if err != nil {
		writeContentError(w, r, err, "create project failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusCreated, data)
}

func (h *ContentHandler) ProjectOverview(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.ProjectOverview(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeContentError(w, r, err, "project overview failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func (h *ContentHandler) PauseProject(w http.ResponseWriter, r *http.Request) {
	var req content.PauseProjectRequest
	if err := decodeJSON(w, r, &req); err != nil {
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, "invalid request body", nil)
		return
	}
	data, err := h.service.PauseProject(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		writeContentError(w, r, err, "pause project failed")
		return
	}
	api.WriteSuccess(w, r, http.StatusOK, data)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func parseListContentTypesRequest(r *http.Request) (content.ListContentTypesRequest, error) {
	pagination, err := parsePagination(r)
	if err != nil {
		return content.ListContentTypesRequest{}, err
	}
	req := content.ListContentTypesRequest{PaginationRequest: pagination}
	if value := r.URL.Query().Get("enabled"); value != "" {
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return content.ListContentTypesRequest{}, err
		}
		req.Enabled = &enabled
	}
	return req, nil
}

func parseListProjectsRequest(r *http.Request) (content.ListProjectsRequest, error) {
	pagination, err := parsePagination(r)
	if err != nil {
		return content.ListProjectsRequest{}, err
	}
	return content.ListProjectsRequest{
		PaginationRequest: pagination,
		Status:            r.URL.Query().Get("status"),
		ContentType:       r.URL.Query().Get("content_type"),
	}, nil
}

func parsePagination(r *http.Request) (content.PaginationRequest, error) {
	query := r.URL.Query()
	page, err := parseOptionalInt(query.Get("page"))
	if err != nil {
		return content.PaginationRequest{}, err
	}
	pageSize, err := parseOptionalInt(query.Get("page_size"))
	if err != nil {
		return content.PaginationRequest{}, err
	}
	return content.PaginationRequest{Page: page, PageSize: pageSize, Sort: query.Get("sort"), Order: query.Get("order")}, nil
}

func parseOptionalInt(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

func writeContentError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, content.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, content.ErrNotFound):
		api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, message, nil)
	case errors.Is(err, content.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}

func writePromptError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, prompt.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, prompt.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}

func writeLLMError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, llm.ErrValidation):
		api.WriteError(w, r, http.StatusBadRequest, api.ErrorValidation, message, nil)
	case errors.Is(err, llm.ErrConflict):
		api.WriteError(w, r, http.StatusConflict, api.ErrorConflict, message, nil)
	default:
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, message, nil)
	}
}
