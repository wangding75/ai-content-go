-- +goose Up

CREATE TABLE IF NOT EXISTS workflow_template (
    id          BIGSERIAL PRIMARY KEY,
    code        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT '',
    category    TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'draft',
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workflow_template_version (
    id           BIGSERIAL PRIMARY KEY,
    template_id  BIGINT NOT NULL REFERENCES workflow_template(id),
    version      INT NOT NULL,
    input_schema JSONB NOT NULL DEFAULT '{}',
    output_schema JSONB NOT NULL DEFAULT '{}',
    status       TEXT NOT NULL DEFAULT 'draft',
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (template_id, version)
);

CREATE TABLE IF NOT EXISTS workflow_step_template (
    id                   BIGSERIAL PRIMARY KEY,
    template_version_id  BIGINT NOT NULL REFERENCES workflow_template_version(id),
    step_code            TEXT NOT NULL,
    step_type            TEXT NOT NULL,
    agent_code           TEXT NOT NULL DEFAULT '',
    order_index          INT NOT NULL,
    input_mapping        JSONB NOT NULL DEFAULT '{}',
    output_mapping       JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_workflow_step_template_version ON workflow_step_template(template_version_id);

CREATE TABLE IF NOT EXISTS workflow_run (
    id                   BIGSERIAL PRIMARY KEY,
    project_id           TEXT NOT NULL DEFAULT '',
    template_version_id  BIGINT NOT NULL REFERENCES workflow_template_version(id),
    status               TEXT NOT NULL DEFAULT 'pending',
    input                JSONB NOT NULL DEFAULT '{}',
    output               JSONB NOT NULL DEFAULT '{}',
    error                TEXT NOT NULL DEFAULT '',
    source               TEXT NOT NULL DEFAULT 'manual',
    parent_run_id        BIGINT,
    idempotency_key      TEXT NOT NULL DEFAULT '',
    created_at           TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workflow_run_status ON workflow_run(status);
CREATE INDEX IF NOT EXISTS idx_workflow_run_project ON workflow_run(project_id);
CREATE INDEX IF NOT EXISTS idx_workflow_run_template_version ON workflow_run(template_version_id);

CREATE TABLE IF NOT EXISTS workflow_step_run (
    id               BIGSERIAL PRIMARY KEY,
    workflow_run_id  BIGINT NOT NULL REFERENCES workflow_run(id),
    step_template_id BIGINT NOT NULL REFERENCES workflow_step_template(id),
    status           TEXT NOT NULL DEFAULT 'pending',
    input            JSONB NOT NULL DEFAULT '{}',
    output           JSONB NOT NULL DEFAULT '{}',
    error            TEXT NOT NULL DEFAULT '',
    started_at       TIMESTAMP,
    finished_at      TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_workflow_step_run_run ON workflow_step_run(workflow_run_id);

CREATE TABLE IF NOT EXISTS agent_task (
    id                BIGSERIAL PRIMARY KEY,
    workflow_run_id   BIGINT NOT NULL REFERENCES workflow_run(id),
    step_run_id       BIGINT NOT NULL REFERENCES workflow_step_run(id),
    agent_code        TEXT NOT NULL DEFAULT '',
    prompt_template_id TEXT NOT NULL DEFAULT '',
    status            TEXT NOT NULL DEFAULT 'pending',
    input             JSONB NOT NULL DEFAULT '{}',
    output            JSONB NOT NULL DEFAULT '{}',
    error             TEXT NOT NULL DEFAULT '',
    started_at        TIMESTAMP,
    finished_at       TIMESTAMP,
    created_at        TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_task_run ON agent_task(workflow_run_id);
CREATE INDEX IF NOT EXISTS idx_agent_task_step_run ON agent_task(step_run_id);
CREATE INDEX IF NOT EXISTS idx_agent_task_status ON agent_task(status);

CREATE TABLE IF NOT EXISTS llm_call_log (
    id               BIGSERIAL PRIMARY KEY,
    workflow_run_id  BIGINT NOT NULL REFERENCES workflow_run(id),
    step_run_id      BIGINT NOT NULL REFERENCES workflow_step_run(id),
    agent_task_id    BIGINT NOT NULL REFERENCES agent_task(id),
    provider         TEXT NOT NULL DEFAULT '',
    model            TEXT NOT NULL DEFAULT '',
    input_tokens     INT NOT NULL DEFAULT 0,
    output_tokens    INT NOT NULL DEFAULT 0,
    cost             NUMERIC(12,6) NOT NULL DEFAULT 0,
    currency         TEXT NOT NULL DEFAULT 'USD',
    latency_ms       INT NOT NULL DEFAULT 0,
    status           TEXT NOT NULL DEFAULT 'success',
    error            TEXT NOT NULL DEFAULT '',
    request_id       TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_llm_call_log_run ON llm_call_log(workflow_run_id);
CREATE INDEX IF NOT EXISTS idx_llm_call_log_agent_task ON llm_call_log(agent_task_id);

CREATE TABLE IF NOT EXISTS workflow_schedule (
    id                   BIGSERIAL PRIMARY KEY,
    template_version_id  BIGINT NOT NULL REFERENCES workflow_template_version(id),
    project_id           TEXT NOT NULL DEFAULT '',
    schedule_type        TEXT NOT NULL DEFAULT '',
    status               TEXT NOT NULL DEFAULT 'not_implemented',
    created_at           TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down

DROP TABLE IF EXISTS workflow_schedule;
DROP TABLE IF EXISTS llm_call_log;
DROP TABLE IF EXISTS agent_task;
DROP TABLE IF EXISTS workflow_step_run;
DROP TABLE IF EXISTS workflow_run;
DROP TABLE IF EXISTS workflow_step_template;
DROP TABLE IF EXISTS workflow_template_version;
DROP TABLE IF EXISTS workflow_template;
