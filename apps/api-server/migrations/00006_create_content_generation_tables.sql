CREATE TABLE IF NOT EXISTS generation_run (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    workflow_run_id TEXT NOT NULL UNIQUE,
    template_version_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'retrying')),
    trigger_type TEXT NOT NULL CHECK (trigger_type IN ('manual', 'batch', 'retry')),
    confirmed_topic_id TEXT,
    worldview_version_id TEXT,
    arc_id TEXT,
    target_count INTEGER NOT NULL CHECK (target_count > 0),
    start_sequence_no INTEGER NOT NULL CHECK (start_sequence_no > 0),
    generation_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    retry_of_generation_run_id TEXT REFERENCES generation_run(id),
    idempotency_key TEXT,
    error_code TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_generation_run_project_id ON generation_run(project_id);
CREATE INDEX IF NOT EXISTS idx_generation_run_workflow_run_id ON generation_run(workflow_run_id);
CREATE INDEX IF NOT EXISTS idx_generation_run_status ON generation_run(status);
CREATE INDEX IF NOT EXISTS idx_generation_run_retry_of ON generation_run(retry_of_generation_run_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_generation_run_project_idempotency_key ON generation_run(project_id, idempotency_key) WHERE idempotency_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS content_item (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    generation_run_id TEXT NOT NULL REFERENCES generation_run(id),
    content_type_code TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('planned', 'generating', 'generated', 'generation_failed', 'pending_review')),
    sequence_no INTEGER NOT NULL CHECK (sequence_no > 0),
    title TEXT NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    version INTEGER NOT NULL DEFAULT 1,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    error_code TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_content_item_project_id ON content_item(project_id);
CREATE INDEX IF NOT EXISTS idx_content_item_generation_run_id ON content_item(generation_run_id);
CREATE INDEX IF NOT EXISTS idx_content_item_status ON content_item(status);
CREATE INDEX IF NOT EXISTS idx_content_item_project_sequence ON content_item(project_id, sequence_no);
CREATE UNIQUE INDEX IF NOT EXISTS uq_content_item_project_type_sequence_version ON content_item(project_id, content_type_code, sequence_no, version);

CREATE TABLE IF NOT EXISTS novel_chapter_extension (
    content_item_id TEXT PRIMARY KEY REFERENCES content_item(id),
    project_id TEXT NOT NULL,
    confirmed_topic_id TEXT NOT NULL,
    worldview_version_id TEXT NOT NULL,
    arc_id TEXT NOT NULL,
    chapter_no INTEGER NOT NULL CHECK (chapter_no > 0),
    script JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_novel_chapter_extension_project_id ON novel_chapter_extension(project_id);
CREATE INDEX IF NOT EXISTS idx_novel_chapter_extension_arc_id ON novel_chapter_extension(arc_id);
