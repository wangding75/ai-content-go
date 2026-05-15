package store

import (
	"context"
	"time"
)

type DBChecker interface {
	Check(ctx context.Context) (DBCheckResult, error)
}

type MigrationReader interface {
	Status(ctx context.Context) (MigrationStatusResult, error)
}

type OperationLogger interface {
	Log(ctx context.Context, entry OperationLogEntry) error
}

type DBCheckResult struct {
	Database  string
	Status    string
	LatencyMS int64
}

type MigrationStatusResult struct {
	AppliedMigrations []MigrationInfo
	PendingMigrations []MigrationInfo
}

type MigrationInfo struct {
	Version   string
	Name      string
	AppliedAt *time.Time
}

type OperationLogEntry struct {
	RequestID    string
	ActorID      string
	ActorType    string
	Action       string
	ResourceType string
	ResourceID   string
	Reason       string
	Metadata     map[string]any
}

type NoopDBChecker struct{}

func (NoopDBChecker) Check(ctx context.Context) (DBCheckResult, error) {
	return DBCheckResult{Database: "postgres", Status: "not_configured", LatencyMS: 0}, nil
}

type NoopMigrationReader struct{}

func (NoopMigrationReader) Status(ctx context.Context) (MigrationStatusResult, error) {
	return MigrationStatusResult{AppliedMigrations: []MigrationInfo{}, PendingMigrations: []MigrationInfo{}}, nil
}
