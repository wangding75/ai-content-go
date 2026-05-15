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
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/system"
)

func NewRouter(systemService system.Service, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	systemHandler := handlers.NewSystemHandler(systemService, logger)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(bearerAuth)
		r.Get("/health", systemHandler.Health)
		r.Get("/system/info", systemHandler.Info)
		r.Get("/system/config-check", systemHandler.ConfigCheck)
		r.Get("/system/db-check", systemHandler.DBCheck)
		r.Get("/system/migration-status", systemHandler.MigrationStatus)
	})
	r.Get("/openapi.yaml", serveOpenAPI)

	return r
}

func bearerAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			api.WriteError(w, r, http.StatusUnauthorized, api.ErrorUnauthorized, "missing bearer token", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
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
