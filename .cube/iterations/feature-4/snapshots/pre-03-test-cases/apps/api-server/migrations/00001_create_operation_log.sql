-- +goose Up
CREATE TABLE operation_log (
    id BIGSERIAL PRIMARY KEY,
    request_id TEXT NOT NULL,
    actor_id TEXT,
    actor_type TEXT NOT NULL DEFAULT 'system',
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    reason TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (metadata ? 'secret' = FALSE),
    CHECK (metadata ? 'token' = FALSE),
    CHECK (metadata ? 'password' = FALSE)
);

CREATE INDEX idx_operation_log_resource ON operation_log(resource_type, resource_id);
CREATE INDEX idx_operation_log_created_at ON operation_log(created_at DESC);
CREATE INDEX idx_operation_log_request_id ON operation_log(request_id);

-- +goose Down
DROP TABLE IF EXISTS operation_log;
