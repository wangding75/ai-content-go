package app

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/config"
	httpserver "github.com/wangding75/ai-content-go/apps/api-server/internal/http"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/metrics"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/portfolio"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/socialpost"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/system"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/store"
)

func NewServer(cfg config.Config, logger *slog.Logger) (*http.Server, error) {
	if cfg.HTTPAddr == "" {
		return nil, errors.New("http addr is required")
	}
	service := system.NewService(cfg, store.NoopDBChecker{}, store.NoopMigrationReader{})

	var opts []httpserver.RouterOption
	var db *sql.DB
	if cfg.DatabaseURL != "" {
		var err error
		db, err = sql.Open("postgres", cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("open database: %w", err)
		}
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)
		metricsStore := metrics.NewPostgresStore(db)
		portfolioStore := portfolio.NewPostgresStore(db)
		socialPostStore := socialpost.NewPostgresStore(db)
		opts = append(
			opts,
			httpserver.WithMetricsService(metrics.NewService(metricsStore)),
			httpserver.WithPortfolioService(portfolio.NewService(portfolioStore)),
			httpserver.WithSocialPostService(socialpost.NewService(socialPostStore)),
		)
	}

	router := httpserver.NewRouter(service, logger, opts...)
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: router}
	if db != nil {
		srv.RegisterOnShutdown(func() { db.Close() })
	}
	return srv, nil
}
