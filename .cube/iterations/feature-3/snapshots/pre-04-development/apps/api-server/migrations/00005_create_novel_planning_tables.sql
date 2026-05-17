CREATE TABLE IF NOT EXISTS planning_run (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    workflow_run_id TEXT NOT NULL UNIQUE,
    template_version_id TEXT NOT NULL,
    status TEXT NOT NULL,
    genre TEXT NOT NULL,
    audience TEXT NOT NULL,
    candidate_count INTEGER NOT NULL CHECK (candidate_count > 0),
    input_override JSONB NOT NULL DEFAULT '{}'::jsonb,
    idempotency_key TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (project_id, id)
);

CREATE INDEX IF NOT EXISTS idx_planning_run_project_id ON planning_run(project_id);
CREATE INDEX IF NOT EXISTS idx_planning_run_workflow_run_id ON planning_run(workflow_run_id);
CREATE INDEX IF NOT EXISTS idx_planning_run_status ON planning_run(status);
CREATE INDEX IF NOT EXISTS idx_planning_run_idempotency_key ON planning_run(idempotency_key);

CREATE TABLE IF NOT EXISTS planning_snapshot (
    id TEXT PRIMARY KEY,
    planning_run_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    snapshot_type TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (project_id, id),
    FOREIGN KEY (project_id, planning_run_id) REFERENCES planning_run(project_id, id)
);

CREATE INDEX IF NOT EXISTS idx_planning_snapshot_run_id ON planning_snapshot(planning_run_id);
CREATE INDEX IF NOT EXISTS idx_planning_snapshot_project_id ON planning_snapshot(project_id);

CREATE TABLE IF NOT EXISTS novel_topic_candidate (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    planning_run_id TEXT NOT NULL,
    snapshot_id TEXT NOT NULL,
    title TEXT NOT NULL,
    logline TEXT NOT NULL,
    status TEXT NOT NULL,
    score NUMERIC(5,2) NOT NULL,
    reason TEXT NOT NULL,
    confirmed_topic_id TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (project_id, id),
    FOREIGN KEY (project_id, planning_run_id) REFERENCES planning_run(project_id, id),
    FOREIGN KEY (project_id, snapshot_id) REFERENCES planning_snapshot(project_id, id)
);

CREATE INDEX IF NOT EXISTS idx_novel_topic_candidate_project_id ON novel_topic_candidate(project_id);
CREATE INDEX IF NOT EXISTS idx_novel_topic_candidate_run_id ON novel_topic_candidate(planning_run_id);
CREATE INDEX IF NOT EXISTS idx_novel_topic_candidate_status ON novel_topic_candidate(status);

CREATE TABLE IF NOT EXISTS novel_worldview (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    planning_run_id TEXT,
    snapshot_id TEXT,
    worldview JSONB NOT NULL DEFAULT '{}'::jsonb,
    forbidden_rules JSONB NOT NULL DEFAULT '[]'::jsonb,
    version INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (project_id, planning_run_id) REFERENCES planning_run(project_id, id),
    FOREIGN KEY (project_id, snapshot_id) REFERENCES planning_snapshot(project_id, id)
);

CREATE INDEX IF NOT EXISTS idx_novel_worldview_project_id ON novel_worldview(project_id);

CREATE TABLE IF NOT EXISTS novel_character (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    planning_run_id TEXT,
    snapshot_id TEXT,
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    profile JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (project_id, planning_run_id) REFERENCES planning_run(project_id, id),
    FOREIGN KEY (project_id, snapshot_id) REFERENCES planning_snapshot(project_id, id)
);

CREATE INDEX IF NOT EXISTS idx_novel_character_project_id ON novel_character(project_id);
CREATE INDEX IF NOT EXISTS idx_novel_character_role ON novel_character(role);

CREATE TABLE IF NOT EXISTS novel_arc (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    planning_run_id TEXT,
    snapshot_id TEXT,
    title TEXT NOT NULL,
    summary TEXT NOT NULL,
    order_index INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (project_id, planning_run_id) REFERENCES planning_run(project_id, id),
    FOREIGN KEY (project_id, snapshot_id) REFERENCES planning_snapshot(project_id, id)
);

CREATE INDEX IF NOT EXISTS idx_novel_arc_project_id ON novel_arc(project_id);
CREATE INDEX IF NOT EXISTS idx_novel_arc_order ON novel_arc(project_id, order_index);
