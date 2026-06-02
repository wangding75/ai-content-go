-- +goose Up
CREATE TABLE IF NOT EXISTS project_portfolio (
    id text PRIMARY KEY,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    scope_type text NOT NULL,
    owner_id text NOT NULL DEFAULT '',
    health_policy jsonb NOT NULL DEFAULT '{}',
    status text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_project_portfolio_status_created_at ON project_portfolio (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_project_portfolio_scope_type ON project_portfolio (scope_type);
CREATE INDEX IF NOT EXISTS idx_project_portfolio_owner_id ON project_portfolio (owner_id);
CREATE INDEX IF NOT EXISTS idx_project_portfolio_name ON project_portfolio (name);

CREATE TABLE IF NOT EXISTS portfolio_project (
    portfolio_id text NOT NULL REFERENCES project_portfolio(id),
    project_id text NOT NULL,
    role text NOT NULL DEFAULT 'member',
    priority integer NOT NULL CHECK (priority >= 1),
    weight numeric(10, 2) NOT NULL DEFAULT 1 CHECK (weight >= 0),
    note text NOT NULL DEFAULT '',
    added_by text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (portfolio_id, project_id)
);

CREATE INDEX IF NOT EXISTS idx_portfolio_project_project_id ON portfolio_project (project_id);
CREATE INDEX IF NOT EXISTS idx_portfolio_project_role ON portfolio_project (role);
CREATE INDEX IF NOT EXISTS idx_portfolio_project_priority ON portfolio_project (portfolio_id, priority ASC);

CREATE TABLE IF NOT EXISTS portfolio_status_snapshot (
    id text PRIMARY KEY,
    portfolio_id text NOT NULL REFERENCES project_portfolio(id),
    date_range_start date NOT NULL,
    date_range_end date NOT NULL,
    health_score numeric(5, 2) NOT NULL,
    health_status text NOT NULL,
    total_projects integer NOT NULL,
    active_projects integer NOT NULL,
    warning_projects integer NOT NULL,
    estimated_monthly_cost numeric(12, 2) NOT NULL,
    currency text NOT NULL DEFAULT 'CNY',
    risk_summary jsonb NOT NULL DEFAULT '{}',
    cost_summary jsonb NOT NULL DEFAULT '{}',
    strategy_summary jsonb NOT NULL DEFAULT '{}',
    source_refs jsonb NOT NULL DEFAULT '[]',
    calculation_status text NOT NULL,
    calculated_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_portfolio_status_snapshot_portfolio_created ON portfolio_status_snapshot (portfolio_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_portfolio_status_snapshot_status ON portfolio_status_snapshot (calculation_status);
CREATE INDEX IF NOT EXISTS idx_portfolio_status_snapshot_date_range ON portfolio_status_snapshot (portfolio_id, date_range_start, date_range_end);

-- +goose Down
DROP TABLE IF EXISTS portfolio_status_snapshot;
DROP TABLE IF EXISTS portfolio_project;
DROP TABLE IF EXISTS project_portfolio;
