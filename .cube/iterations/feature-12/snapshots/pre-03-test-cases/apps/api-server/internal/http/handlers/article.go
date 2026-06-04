package handlers

import (
	"log/slog"
	"net/http"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/article"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/metrics"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type ArticleHandler struct {
	svc        article.Service
	contentSvc content.Service
	workflow   workflow.Service
	metricsSvc metrics.Service
	submitter  engine.Submitter
	logger     *slog.Logger
}

func NewArticleHandler(svc article.Service, contentSvc content.Service, wfSvc workflow.Service, metricsSvc metrics.Service, submitter engine.Submitter, logger *slog.Logger) *ArticleHandler {
	return &ArticleHandler{svc: svc, contentSvc: contentSvc, workflow: wfSvc, metricsSvc: metricsSvc, submitter: submitter, logger: logger}
}

func (h *ArticleHandler) RegisterPack(w http.ResponseWriter, r *http.Request) {
	_ = r.Header.Get("Idempotency-Key")
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "not implemented", nil)
}

func (h *ArticleHandler) GetPackStatus(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "not implemented", nil)
}

func (h *ArticleHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "not implemented", nil)
}

func (h *ArticleHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "not implemented", nil)
}

func (h *ArticleHandler) CreateGenerationRun(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "not implemented", nil)
}

func (h *ArticleHandler) ListGenerationRuns(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "not implemented", nil)
}

func (h *ArticleHandler) GetGenerationRun(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "not implemented", nil)
}

func (h *ArticleHandler) RetryGenerationRun(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "not implemented", nil)
}

func (h *ArticleHandler) GetContentSnapshot(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "not implemented", nil)
}

func (h *ArticleHandler) GetMetricsConfig(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "not implemented", nil)
}

func (h *ArticleHandler) UpdateMetricsConfig(w http.ResponseWriter, r *http.Request) {
	api.WriteError(w, r, http.StatusNotImplemented, api.ErrorInternal, "not implemented", nil)
}
