-- +goose Up
CREATE TABLE IF NOT EXISTS content_review (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    content_item_id TEXT NOT NULL REFERENCES content_item(id),
    review_type TEXT NOT NULL CHECK (review_type IN ('manual', 'ai', 'combined')),
    status TEXT NOT NULL CHECK (status IN ('pending', 'in_review', 'approved', 'rejected', 'approved_with_edit')),
    current_version_id TEXT,
    report_id TEXT,
    note TEXT NOT NULL DEFAULT '',
    reject_reason TEXT NOT NULL DEFAULT '',
    regenerate_instruction TEXT NOT NULL DEFAULT '',
    operation_log_id TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_content_review_project_status ON content_review(project_id, status);
CREATE INDEX IF NOT EXISTS idx_content_review_content_item ON content_review(content_item_id);
CREATE INDEX IF NOT EXISTS idx_content_review_updated_at ON content_review(updated_at DESC);

CREATE TABLE IF NOT EXISTS content_version (
    id TEXT PRIMARY KEY,
    content_item_id TEXT NOT NULL REFERENCES content_item(id),
    project_id TEXT NOT NULL,
    version_no INTEGER NOT NULL CHECK (version_no > 0),
    source TEXT NOT NULL CHECK (source IN ('generation', 'edit_approve', 'regeneration')),
    title TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    editable_fields JSONB NOT NULL DEFAULT '{}'::jsonb,
    summary TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE(content_item_id, version_no)
);

CREATE INDEX IF NOT EXISTS idx_content_version_item_version ON content_version(content_item_id, version_no DESC);

CREATE TABLE IF NOT EXISTS review_report (
    id TEXT PRIMARY KEY,
    review_id TEXT NOT NULL REFERENCES content_review(id),
    content_item_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'generating', 'succeeded', 'failed')),
    quality_score INTEGER CHECK (quality_score >= 0 AND quality_score <= 100),
    risk_level TEXT NOT NULL DEFAULT 'unknown' CHECK (risk_level IN ('low', 'medium', 'high', 'unknown')),
    issues JSONB NOT NULL DEFAULT '[]'::jsonb,
    suggestions JSONB NOT NULL DEFAULT '[]'::jsonb,
    job_id TEXT,
    workflow_run_id TEXT,
    error_code TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_review_report_review_id ON review_report(review_id);
CREATE INDEX IF NOT EXISTS idx_review_report_status ON review_report(status);

-- +goose Down
DROP TABLE IF EXISTS review_report;
DROP TABLE IF EXISTS content_version;
DROP TABLE IF EXISTS content_review;
