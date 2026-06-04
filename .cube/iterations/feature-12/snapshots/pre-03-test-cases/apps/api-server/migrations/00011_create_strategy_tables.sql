-- +goose Up
CREATE TABLE IF NOT EXISTS strategy_suggestion_run (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    date_from DATE NOT NULL,
    date_to DATE NOT NULL,
    rule_codes JSONB NOT NULL DEFAULT '[]'::jsonb,
    metric_codes JSONB NOT NULL DEFAULT '[]'::jsonb,
    force_regenerate BOOLEAN NOT NULL DEFAULT FALSE,
    status TEXT NOT NULL CHECK (status IN ('generating', 'completed', 'failed')) DEFAULT 'generating',
    failure_reason TEXT NOT NULL DEFAULT '',
    suggestion_count INTEGER NOT NULL DEFAULT 0 CHECK (suggestion_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_strategy_suggestion_run_project ON strategy_suggestion_run(project_id, created_at DESC);

CREATE TABLE IF NOT EXISTS strategy_suggestion (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    suggestion_run_id TEXT NOT NULL,
    suggestion_type TEXT NOT NULL CHECK (suggestion_type IN ('keep', 'optimize', 'suspend', 'promote', 'cost_control')),
    title TEXT NOT NULL,
    trigger_reason TEXT NOT NULL,
    evidence_metrics JSONB NOT NULL DEFAULT '[]'::jsonb,
    impact_scope TEXT NOT NULL DEFAULT '',
    risk_level TEXT NOT NULL CHECK (risk_level IN ('low', 'medium', 'high')),
    confidence TEXT NOT NULL CHECK (confidence IN ('low', 'medium', 'high')),
    suggested_action TEXT NOT NULL DEFAULT '',
    expected_benefit TEXT NOT NULL DEFAULT '',
    metrics_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL CHECK (status IN ('pending', 'confirmed', 'ignored', 'executed', 'execution_failed')) DEFAULT 'pending',
    ignored_reason TEXT NOT NULL DEFAULT '',
    ignored_note TEXT NOT NULL DEFAULT '',
    confirmed_at TIMESTAMPTZ,
    ignored_at TIMESTAMPTZ,
    executed_at TIMESTAMPTZ,
    date_from DATE NOT NULL,
    date_to DATE NOT NULL,
    triggered_rules JSONB NOT NULL DEFAULT '[]'::jsonb,
    generation_method TEXT NOT NULL DEFAULT 'rule',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_strategy_suggestion_project ON strategy_suggestion(project_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_strategy_suggestion_run ON strategy_suggestion(suggestion_run_id);
CREATE INDEX IF NOT EXISTS idx_strategy_suggestion_type_status ON strategy_suggestion(project_id, suggestion_type, status);

CREATE TABLE IF NOT EXISTS strategy_execution_log (
    id TEXT PRIMARY KEY,
    suggestion_id TEXT NOT NULL REFERENCES strategy_suggestion(id),
    action_type TEXT NOT NULL,
    target_type TEXT NOT NULL DEFAULT '',
    target_id TEXT NOT NULL DEFAULT '',
    operator_note TEXT NOT NULL DEFAULT '',
    previous_status TEXT NOT NULL,
    current_status TEXT NOT NULL,
    result TEXT NOT NULL CHECK (result IN ('success', 'failed')),
    failure_reason TEXT NOT NULL DEFAULT '',
    target_interface TEXT NOT NULL DEFAULT '',
    target_resource TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_strategy_execution_log_suggestion ON strategy_execution_log(suggestion_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS strategy_execution_log;
DROP TABLE IF EXISTS strategy_suggestion;
DROP TABLE IF EXISTS strategy_suggestion_run;
