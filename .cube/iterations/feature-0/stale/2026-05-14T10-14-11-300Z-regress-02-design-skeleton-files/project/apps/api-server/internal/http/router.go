package http

import (
	"log/slog"
	"net/http"
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
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi/openapi.yaml")
	})

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
