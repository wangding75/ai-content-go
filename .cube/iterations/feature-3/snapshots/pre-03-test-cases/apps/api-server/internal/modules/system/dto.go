package system

import "time"

type HealthResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

type InfoResponse struct {
	AppName     string `json:"app_name"`
	Environment string `json:"environment"`
	BuildCommit string `json:"build_commit"`
}

type ConfigCheckResponse struct {
	Items   []ConfigCheckItem  `json:"items"`
	Summary ConfigCheckSummary `json:"summary"`
}

type ConfigCheckItem struct {
	Key        string `json:"key"`
	Required   bool   `json:"required"`
	Configured bool   `json:"configured"`
	Status     string `json:"status"`
}

type ConfigCheckSummary struct {
	Total      int `json:"total"`
	Configured int `json:"configured"`
	Missing    int `json:"missing"`
}

type DBCheckResponse struct {
	Database  string `json:"database"`
	Status    string `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
}

type MigrationStatusResponse struct {
	AppliedMigrations []MigrationInfo `json:"applied_migrations"`
	PendingMigrations []MigrationInfo `json:"pending_migrations"`
}

type MigrationInfo struct {
	Version   string     `json:"version"`
	Name      string     `json:"name"`
	AppliedAt *time.Time `json:"applied_at"`
}
