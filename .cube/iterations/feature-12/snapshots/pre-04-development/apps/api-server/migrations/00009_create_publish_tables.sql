-- +goose Up
CREATE TABLE IF NOT EXISTS publish_target (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    platform TEXT NOT NULL,
    account_name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    config_summary TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    operation_log_id TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE(project_id, platform, account_name, display_name)
);

CREATE INDEX IF NOT EXISTS idx_publish_target_project_enabled ON publish_target(project_id, enabled);

CREATE TABLE IF NOT EXISTS publish_job (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    content_item_id TEXT NOT NULL REFERENCES content_item(id),
    content_version_id TEXT NOT NULL REFERENCES content_version(id),
    target_id TEXT NOT NULL REFERENCES publish_target(id),
    status TEXT NOT NULL CHECK (status IN ('queued', 'copied', 'published', 'failed', 'canceled')),
    payload_hash TEXT NOT NULL,
    scheduled_at TIMESTAMPTZ,
    copied_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    external_url TEXT NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    failed_at TIMESTAMPTZ,
    retryable BOOLEAN NOT NULL DEFAULT FALSE,
    retry_count INTEGER NOT NULL DEFAULT 0 CHECK (retry_count >= 0),
    operation_log_id TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_publish_job_project_status_created ON publish_job(project_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_publish_job_target_status ON publish_job(target_id, status);
CREATE INDEX IF NOT EXISTS idx_publish_job_scheduled ON publish_job(project_id, scheduled_at);
CREATE UNIQUE INDEX IF NOT EXISTS uq_publish_job_version_target_active ON publish_job(content_version_id, target_id) WHERE status IN ('queued', 'copied', 'failed');

CREATE TABLE IF NOT EXISTS publish_log (
    id TEXT PRIMARY KEY,
    publish_job_id TEXT NOT NULL REFERENCES publish_job(id),
    event_type TEXT NOT NULL CHECK (event_type IN ('job_created', 'payload_copied', 'marked_published', 'marked_failed', 'requeued')),
    from_status TEXT NOT NULL DEFAULT '',
    to_status TEXT NOT NULL DEFAULT '',
    actor_id TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    payload_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    operation_log_id TEXT,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_publish_log_job_created ON publish_log(publish_job_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS publish_log;
DROP TABLE IF EXISTS publish_job;
DROP TABLE IF EXISTS publish_target;
