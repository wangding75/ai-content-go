-- Iteration 11 platform adapter extension skeleton
ALTER TABLE publish_target ADD COLUMN IF NOT EXISTS target_type TEXT NOT NULL DEFAULT 'default';

CREATE TABLE IF NOT EXISTS platform_adapter_config (
    id TEXT PRIMARY KEY,
    platform TEXT NOT NULL,
    display_name TEXT NOT NULL,
    publish_mode TEXT NOT NULL CHECK (publish_mode IN ('manual_plugin', 'external_callback', 'manual_only')),
    target_type TEXT NOT NULL,
    field_mapping JSONB NOT NULL DEFAULT '{}'::jsonb,
    fill_rules JSONB NOT NULL DEFAULT '{}'::jsonb,
    collect_rules JSONB NOT NULL DEFAULT '{}'::jsonb,
    credential_ref TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    operation_log_id TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE(platform, target_type)
);

CREATE TABLE IF NOT EXISTS platform_adapter_revision (
    id TEXT PRIMARY KEY,
    adapter_id TEXT NOT NULL,
    version INTEGER NOT NULL,
    change_reason TEXT NOT NULL,
    snapshot JSONB NOT NULL,
    operation_log_id TEXT,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS plugin_client (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    client_type TEXT NOT NULL CHECK (client_type IN ('chrome_extension')),
    version TEXT NOT NULL,
    scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
    status TEXT NOT NULL CHECK (status IN ('enabled', 'disabled')),
    api_key_hash TEXT NOT NULL UNIQUE,
    api_key_masked TEXT NOT NULL,
    last_active_at TIMESTAMPTZ,
    operation_log_id TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS plugin_access_token (
    id TEXT PRIMARY KEY,
    client_id TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    scopes JSONB NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL
);

ALTER TABLE publish_job ADD COLUMN IF NOT EXISTS plugin_lock_id TEXT;
ALTER TABLE publish_job ADD COLUMN IF NOT EXISTS plugin_client_id TEXT;
ALTER TABLE publish_job ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ;
ALTER TABLE publish_job ADD COLUMN IF NOT EXISTS adapter_config_id TEXT;
ALTER TABLE publish_job ADD COLUMN IF NOT EXISTS adapter_version INTEGER;
ALTER TABLE publish_job ADD COLUMN IF NOT EXISTS filled_at TIMESTAMPTZ;
ALTER TABLE publish_job ADD COLUMN IF NOT EXISTS platform_error_summary TEXT NOT NULL DEFAULT '';

ALTER TABLE publish_log DROP CONSTRAINT IF EXISTS publish_log_event_type_check;
ALTER TABLE publish_log ADD CONSTRAINT publish_log_event_type_check CHECK (event_type IN ('job_created', 'payload_copied', 'marked_published', 'marked_failed', 'requeued', 'plugin_locked', 'plugin_filled', 'plugin_published', 'plugin_failed'));

CREATE TABLE IF NOT EXISTS platform_collect_log (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    platform TEXT NOT NULL,
    target_account TEXT NOT NULL DEFAULT '',
    publish_job_id TEXT NOT NULL,
    content_item_id TEXT NOT NULL,
    content_version_id TEXT NOT NULL,
    target_id TEXT NOT NULL,
    content_type TEXT NOT NULL,
    external_url TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL CHECK (source_type IN ('extension','external_callback')),
    raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    parsed_metrics JSONB NOT NULL DEFAULT '[]'::jsonb,
    status TEXT NOT NULL CHECK (status IN ('received','parse_failed','ready','confirmed','rejected')),
    error_summary TEXT NOT NULL DEFAULT '',
    collected_at TIMESTAMPTZ NOT NULL,
    operation_log_id TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

ALTER TABLE external_workflow_binding ADD COLUMN IF NOT EXISTS callback_auth_type TEXT NOT NULL DEFAULT 'token' CHECK (callback_auth_type IN ('token','signature'));
ALTER TABLE external_workflow_binding ADD COLUMN IF NOT EXISTS callback_token_hash TEXT NOT NULL DEFAULT '';
ALTER TABLE external_workflow_binding ADD COLUMN IF NOT EXISTS signing_secret_ref TEXT NOT NULL DEFAULT '';
ALTER TABLE external_workflow_binding ADD COLUMN IF NOT EXISTS callback_token_masked TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS external_callback_log (
    id TEXT PRIMARY KEY,
    provider_id TEXT,
    binding_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    payload JSONB NOT NULL,
    accepted BOOLEAN NOT NULL,
    rejected_reason TEXT NOT NULL DEFAULT '',
    boundary_violation BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE(binding_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_publish_target_platform_type ON publish_target(platform, target_type);
CREATE INDEX IF NOT EXISTS idx_publish_job_plugin_lock ON publish_job(plugin_lock_id);
CREATE INDEX IF NOT EXISTS idx_publish_job_adapter_status ON publish_job(adapter_config_id, status);
CREATE INDEX IF NOT EXISTS idx_publish_job_locked_until ON publish_job(locked_until);
CREATE INDEX IF NOT EXISTS idx_platform_collect_log_project_status ON platform_collect_log(project_id, status);
CREATE INDEX IF NOT EXISTS idx_external_callback_binding_created ON external_callback_log(binding_id, created_at);

-- Atomic lock contract: UPDATE publish_job SET plugin_lock_id = $2, plugin_client_id = $3, locked_until = $4 WHERE id = $1 AND status IN ('queued','copied','failed') AND (locked_until IS NULL OR locked_until < $5) RETURNING id;
