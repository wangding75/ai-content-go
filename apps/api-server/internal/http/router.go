package http

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/handlers"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/dashboard"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/llm"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/prompt"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/system"
)

func NewRouter(systemService system.Service, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(cors)

	systemHandler := handlers.NewSystemHandler(systemService, logger)
	dashboardHandler := handlers.NewDashboardHandler(dashboard.NewService(), logger)
	contentHandler := handlers.NewContentHandler(content.NewService(), logger)
	promptHandler := handlers.NewPromptHandler(prompt.NewService(), logger)
	llmHandler := handlers.NewLLMHandler(llm.NewService(), logger)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(bearerAuth)
		r.Get("/health", systemHandler.Health)
		r.Get("/system/info", systemHandler.Info)
		r.Get("/system/config-check", systemHandler.ConfigCheck)
		r.Get("/system/db-check", systemHandler.DBCheck)
		r.Get("/system/migration-status", systemHandler.MigrationStatus)
		r.Get("/dashboard/summary", dashboardHandler.Summary)
		r.Get("/content-types", contentHandler.ListContentTypes)
		r.Post("/content-types", contentHandler.CreateContentType)
		r.Get("/content-types/{id}/project-schema", contentHandler.ProjectSchema)
		r.Get("/projects", contentHandler.ListProjects)
		r.Post("/projects", contentHandler.CreateProject)
		r.Get("/projects/{id}/overview", contentHandler.ProjectOverview)
		r.Post("/projects/{id}/pause", contentHandler.PauseProject)
		r.Get("/prompt-templates", promptHandler.ListTemplates)
		r.Post("/prompt-templates", promptHandler.CreateTemplate)
		r.Get("/llm-providers", llmHandler.ListProviders)
		r.Post("/llm-providers", llmHandler.CreateProvider)
	})
	r.Get("/openapi.yaml", serveOpenAPI)

	return r
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func bearerAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") {
			api.WriteError(w, r, http.StatusUnauthorized, api.ErrorUnauthorized, "missing bearer token", nil)
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
		if token == "" || !validBearerToken(token) {
			api.WriteError(w, r, http.StatusUnauthorized, api.ErrorUnauthorized, "invalid bearer token", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func validBearerToken(token string) bool {
	expected := strings.TrimSpace(os.Getenv("API_BEARER_TOKEN"))
	return expected == "" || token == expected
}

func serveOpenAPI(w http.ResponseWriter, r *http.Request) {
	path, err := openAPIPath()
	if err != nil {
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, "openapi document unavailable", nil)
		return
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, "openapi document not found", nil)
			return
		}
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, "openapi document unavailable", nil)
		return
	}
	http.ServeFile(w, r, path)
}

func openAPIPath() (string, error) {
	if path := os.Getenv("OPENAPI_FILE"); path != "" {
		if !filepath.IsAbs(path) || filepath.Base(path) != "openapi.yaml" {
			return "", errOpenAPIPathUnavailable
		}
		return path, nil
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errOpenAPIPathUnavailable
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../openapi/openapi.yaml")), nil
}

var errOpenAPIPathUnavailable = errors.New("openapi path unavailable")
