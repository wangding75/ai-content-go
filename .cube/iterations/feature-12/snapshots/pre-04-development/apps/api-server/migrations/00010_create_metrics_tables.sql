-- +goose Up
CREATE TABLE IF NOT EXISTS metric_template (
    id TEXT PRIMARY KEY,
    content_type TEXT NOT NULL,
    platform TEXT NOT NULL,
    metric_code TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    unit TEXT NOT NULL,
    value_type TEXT NOT NULL CHECK (value_type IN ('integer', 'decimal', 'percentage', 'currency', 'duration')),
    aggregation_method TEXT NOT NULL CHECK (aggregation_method IN ('sum', 'avg', 'max', 'min', 'latest')),
    period TEXT NOT NULL CHECK (period IN ('day', 'week', 'month')),
    required BOOLEAN NOT NULL DEFAULT FALSE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(content_type, platform, metric_code)
);

CREATE INDEX IF NOT EXISTS idx_metric_template_lookup ON metric_template(content_type, platform, enabled);

CREATE TABLE IF NOT EXISTS metric_record (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    content_item_id TEXT NOT NULL,
    content_version_id TEXT NOT NULL,
    publish_job_id TEXT NOT NULL,
    target_id TEXT NOT NULL,
    content_type TEXT NOT NULL,
    metric_template_id TEXT NOT NULL REFERENCES metric_template(id),
    platform TEXT NOT NULL,
    external_url TEXT NOT NULL DEFAULT '',
    metric_code TEXT NOT NULL,
    metric_date DATE NOT NULL,
    period TEXT NOT NULL CHECK (period IN ('day', 'week', 'month')),
    raw_value TEXT NOT NULL,
    normalized_value NUMERIC(20, 6) NOT NULL,
    source_type TEXT NOT NULL CHECK (source_type IN ('manual', 'import', 'extension', 'external_callback')),
    source_ref TEXT NOT NULL DEFAULT '',
    collected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    operation_log_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, platform, target_id, content_version_id, metric_code, metric_date, period)
);

CREATE INDEX IF NOT EXISTS idx_metric_record_project_metric_date ON metric_record(project_id, metric_code, metric_date DESC);
CREATE INDEX IF NOT EXISTS idx_metric_record_template_date ON metric_record(metric_template_id, metric_date DESC);
CREATE INDEX IF NOT EXISTS idx_metric_record_target_date ON metric_record(project_id, platform, target_id, metric_date DESC);
CREATE INDEX IF NOT EXISTS idx_metric_record_content_item ON metric_record(content_item_id, metric_date DESC);

CREATE TABLE IF NOT EXISTS metric_summary_snapshot (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    date_from DATE NOT NULL,
    date_to DATE NOT NULL,
    platform TEXT NOT NULL DEFAULT '',
    target_id TEXT NOT NULL DEFAULT '',
    metric_codes JSONB NOT NULL DEFAULT '[]'::jsonb,
    aggregation_method TEXT NOT NULL DEFAULT 'mixed',
    summary JSONB NOT NULL DEFAULT '{}'::jsonb,
    source_record_count INTEGER NOT NULL DEFAULT 0 CHECK (source_record_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_metric_summary_project_created ON metric_summary_snapshot(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_metric_summary_project_range ON metric_summary_snapshot(project_id, date_from, date_to);

-- +goose Down
DROP TABLE IF EXISTS metric_summary_snapshot;
DROP TABLE IF EXISTS metric_record;
DROP TABLE IF EXISTS metric_template;
