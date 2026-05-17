-- +goose Up
CREATE TABLE IF NOT EXISTS content_type (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    project_schema JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_content_type_enabled ON content_type(enabled);

CREATE TABLE IF NOT EXISTS content_project (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    content_type_id BIGINT NOT NULL REFERENCES content_type(id),
    status TEXT NOT NULL,
    project_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    target_platform TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_content_project_status ON content_project(status);
CREATE INDEX IF NOT EXISTS idx_content_project_content_type ON content_project(content_type_id);

CREATE TABLE IF NOT EXISTS prompt_template (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    agent_code TEXT NOT NULL DEFAULT '',
    template TEXT NOT NULL,
    variables JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prompt_template_agent_code ON prompt_template(agent_code);

CREATE TABLE IF NOT EXISTS llm_provider_config (
    id BIGSERIAL PRIMARY KEY,
    provider_type TEXT NOT NULL,
    base_url TEXT NOT NULL,
    api_key TEXT NOT NULL,
    api_key_masked TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_llm_provider_type_base_url UNIQUE(provider_type, base_url)
);

-- +goose Down
DROP TABLE IF EXISTS llm_provider_config;
DROP TABLE IF EXISTS prompt_template;
DROP TABLE IF EXISTS content_project;
DROP TABLE IF EXISTS content_type;
