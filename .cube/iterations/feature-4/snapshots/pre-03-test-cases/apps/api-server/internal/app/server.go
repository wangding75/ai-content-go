package app

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/config"
	httpserver "github.com/wangding75/ai-content-go/apps/api-server/internal/http"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/system"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/store"
)

func NewServer(cfg config.Config, logger *slog.Logger) (*http.Server, error) {
	if cfg.HTTPAddr == "" {
		return nil, errors.New("http addr is required")
	}
	service := system.NewService(cfg, store.NoopDBChecker{}, store.NoopMigrationReader{})
	router := httpserver.NewRouter(service, logger)
	return &http.Server{Addr: cfg.HTTPAddr, Handler: router}, nil
}
