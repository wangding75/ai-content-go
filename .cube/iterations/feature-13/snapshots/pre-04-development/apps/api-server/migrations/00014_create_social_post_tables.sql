-- +goose Up
CREATE TABLE IF NOT EXISTS social_post_extension (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    target_platforms JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_variant_count INTEGER NOT NULL DEFAULT 3 CHECK (default_variant_count > 0 AND default_variant_count <= 10),
    caption_length_policy TEXT NOT NULL DEFAULT 'short',
    hashtag_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
    cover_copy_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
    tone_style TEXT NOT NULL DEFAULT '',
    forbidden_terms JSONB NOT NULL DEFAULT '[]'::jsonb,
    config_version INTEGER NOT NULL DEFAULT 1 CHECK (config_version > 0),
    operation_log_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id)
);

CREATE INDEX IF NOT EXISTS idx_social_post_extension_project ON social_post_extension(project_id);

CREATE TABLE IF NOT EXISTS social_post_variant (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    content_item_id TEXT NOT NULL,
    generation_run_id TEXT NOT NULL,
    workflow_run_id TEXT NOT NULL,
    variant_index INTEGER NOT NULL CHECK (variant_index > 0),
    platform TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    hashtags JSONB NOT NULL DEFAULT '[]'::jsonb,
    cover_copy TEXT NOT NULL DEFAULT '',
    tone_style TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('generated', 'selected', 'rejected', 'archived')) DEFAULT 'generated',
    content_version_id TEXT NOT NULL DEFAULT '',
    selected_at TIMESTAMPTZ,
    operation_log_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(content_item_id, generation_run_id, variant_index)
);

CREATE INDEX IF NOT EXISTS idx_social_post_variant_project_created ON social_post_variant(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_social_post_variant_content_item_status ON social_post_variant(content_item_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_social_post_variant_platform_status ON social_post_variant(platform, status, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_social_post_variant_selected_unique
    ON social_post_variant(content_item_id)
    WHERE status = 'selected';

CREATE TABLE IF NOT EXISTS social_post_asset (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    content_item_id TEXT NOT NULL,
    source_variant_id TEXT NOT NULL DEFAULT '',
    asset_type TEXT NOT NULL CHECK (asset_type IN ('tags', 'cover_copy')),
    platform TEXT NOT NULL,
    generation_run_id TEXT NOT NULL,
    workflow_run_id TEXT NOT NULL,
    result JSONB NOT NULL DEFAULT '{}'::jsonb,
    asset_suggestions JSONB NOT NULL DEFAULT '[]'::jsonb,
    operation_log_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_social_post_asset_project_type_created ON social_post_asset(project_id, asset_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_social_post_asset_content_item_platform ON social_post_asset(content_item_id, platform, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_social_post_asset_variant_type ON social_post_asset(source_variant_id, asset_type, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS social_post_asset;
DROP TABLE IF EXISTS social_post_variant;
DROP TABLE IF EXISTS social_post_extension;