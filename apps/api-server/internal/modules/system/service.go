package system

import (
	"context"
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/config"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/store"
)

type Service interface {
	Health(ctx context.Context) (HealthResponse, error)
	Info(ctx context.Context) (InfoResponse, error)
	ConfigCheck(ctx context.Context) (ConfigCheckResponse, error)
	DBCheck(ctx context.Context) (DBCheckResponse, error)
	MigrationStatus(ctx context.Context) (MigrationStatusResponse, error)
}

type service struct {
	cfg             config.Config
	dbChecker       store.DBChecker
	migrationReader store.MigrationReader
}

func NewService(cfg config.Config, dbChecker store.DBChecker, migrationReader store.MigrationReader) Service {
	return &service{cfg: cfg, dbChecker: dbChecker, migrationReader: migrationReader}
}

func (s *service) Health(ctx context.Context) (HealthResponse, error) {
	version := s.cfg.BuildCommit
	if version == "" {
		version = "dev"
	}
	return HealthResponse{Status: "ok", Service: s.cfg.AppName, Version: version, Timestamp: time.Now().UTC()}, nil
}

func (s *service) Info(ctx context.Context) (InfoResponse, error) {
	return InfoResponse{AppName: s.cfg.AppName, Environment: s.cfg.Environment, BuildCommit: s.cfg.BuildCommit}, nil
}

func (s *service) ConfigCheck(ctx context.Context) (ConfigCheckResponse, error) {
	items := []ConfigCheckItem{
		{Key: "APP_NAME", Required: true, Configured: s.cfg.AppName != "", Status: statusFor(s.cfg.AppName != "")},
		{Key: "APP_ENV", Required: true, Configured: s.cfg.Environment != "", Status: statusFor(s.cfg.Environment != "")},
		{Key: "DATABASE_URL", Required: true, Configured: s.cfg.DatabaseURL != "", Status: statusFor(s.cfg.DatabaseURL != "")},
	}
	missing := 0
	for _, item := range items {
		if item.Required && !item.Configured {
			missing++
		}
	}
	return ConfigCheckResponse{Items: items, Summary: ConfigCheckSummary{Total: len(items), Configured: len(items) - missing, Missing: missing}}, nil
}

func (s *service) DBCheck(ctx context.Context) (DBCheckResponse, error) {
	result, err := s.dbChecker.Check(ctx)
	if err != nil {
		return DBCheckResponse{}, err
	}
	return DBCheckResponse{Database: result.Database, Status: result.Status, LatencyMS: result.LatencyMS}, nil
}

func (s *service) MigrationStatus(ctx context.Context) (MigrationStatusResponse, error) {
	result, err := s.migrationReader.Status(ctx)
	if err != nil {
		return MigrationStatusResponse{}, err
	}
	return MigrationStatusResponse{AppliedMigrations: convertMigrations(result.AppliedMigrations), PendingMigrations: convertMigrations(result.PendingMigrations)}, nil
}

func statusFor(configured bool) string {
	if configured {
		return "configured"
	}
	return "missing"
}

func convertMigrations(items []store.MigrationInfo) []MigrationInfo {
	out := make([]MigrationInfo, 0, len(items))
	for _, item := range items {
		out = append(out, MigrationInfo{Version: item.Version, Name: item.Name, AppliedAt: item.AppliedAt})
	}
	return out
}
