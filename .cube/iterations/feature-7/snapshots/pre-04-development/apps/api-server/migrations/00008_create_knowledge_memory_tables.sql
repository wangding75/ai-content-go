CREATE TABLE IF NOT EXISTS knowledge_memory (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL UNIQUE,
  static_context JSONB NOT NULL DEFAULT '{}'::jsonb,
  dynamic_state JSONB NOT NULL DEFAULT '{}'::jsonb,
  recent_window_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
  style_guide JSONB NOT NULL DEFAULT '{}'::jsonb,
  version INTEGER NOT NULL DEFAULT 1,
  operation_log_id TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_knowledge_memory_project ON knowledge_memory(project_id);

CREATE TABLE IF NOT EXISTS memory_snapshot (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  content_item_id TEXT,
  source_type TEXT NOT NULL CHECK (source_type IN ('assemble_context', 'dynamic_state_update', 'dynamic_state_correction')),
  source_id TEXT NOT NULL DEFAULT '',
  assembled_context JSONB NOT NULL DEFAULT '{}'::jsonb,
  source_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
  token_budget INTEGER NOT NULL DEFAULT 0,
  estimated_tokens INTEGER NOT NULL DEFAULT 0,
  truncation_policy TEXT NOT NULL DEFAULT 'time',
  triggered_by TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_memory_snapshot_project_created ON memory_snapshot(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_memory_snapshot_content_item ON memory_snapshot(content_item_id);

CREATE TABLE IF NOT EXISTS consistency_report (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  range JSONB NOT NULL DEFAULT '{}'::jsonb,
  scope TEXT NOT NULL DEFAULT 'project',
  severity_threshold TEXT NOT NULL DEFAULT 'low',
  status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed')),
  issue_count INTEGER NOT NULL DEFAULT 0,
  severity_summary JSONB NOT NULL DEFAULT '{}'::jsonb,
  issues JSONB NOT NULL DEFAULT '[]'::jsonb,
  source_snapshot_id TEXT,
  error_code TEXT,
  error_message TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_consistency_report_project_created ON consistency_report(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_consistency_report_project_status_created ON consistency_report(project_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_consistency_report_project_severity ON consistency_report(project_id, severity_threshold);

CREATE TABLE IF NOT EXISTS idempotency_record (
  id TEXT PRIMARY KEY,
  scope TEXT NOT NULL,
  endpoint TEXT NOT NULL,
  idempotency_key TEXT NOT NULL,
  request_hash TEXT NOT NULL,
  response_ref_type TEXT NOT NULL,
  response_ref_id TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(scope, endpoint, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_idempotency_scope_endpoint ON idempotency_record(scope, endpoint);
