CREATE TABLE IF NOT EXISTS workflow_schedule (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    template_version_id TEXT NOT NULL,
    cron_expression TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    next_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_workflow_schedule_project_id ON workflow_schedule(project_id);
CREATE INDEX IF NOT EXISTS idx_workflow_schedule_enabled ON workflow_schedule(enabled);

CREATE TABLE IF NOT EXISTS production_plan (
    id TEXT PRIMARY KEY,
    schedule_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    daily_content_count INTEGER NOT NULL CHECK (daily_content_count > 0),
    input_template JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_production_plan_schedule_id ON production_plan(schedule_id);

CREATE TABLE IF NOT EXISTS schedule_trigger_log (
    id TEXT PRIMARY KEY,
    schedule_id TEXT NOT NULL,
    trigger_type TEXT NOT NULL,
    workflow_run_id TEXT,
    status TEXT NOT NULL,
    error TEXT,
    triggered_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_schedule_trigger_log_schedule_id ON schedule_trigger_log(schedule_id);
CREATE INDEX IF NOT EXISTS idx_schedule_trigger_log_run_id ON schedule_trigger_log(workflow_run_id);

CREATE TABLE IF NOT EXISTS external_workflow_provider (
    id TEXT PRIMARY KEY,
    provider_type TEXT NOT NULL,
    base_url TEXT NOT NULL,
    token_masked TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (provider_type, base_url)
);

CREATE TABLE IF NOT EXISTS external_workflow_binding (
    id TEXT PRIMARY KEY,
    provider_id TEXT NOT NULL,
    trigger_event TEXT NOT NULL,
    webhook_url TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_external_binding_provider_id ON external_workflow_binding(provider_id);
CREATE INDEX IF NOT EXISTS idx_external_binding_event ON external_workflow_binding(trigger_event);

CREATE TABLE IF NOT EXISTS external_workflow_call_log (
    id TEXT PRIMARY KEY,
    provider_id TEXT NOT NULL,
    binding_id TEXT,
    status TEXT NOT NULL,
    error TEXT,
    request_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
