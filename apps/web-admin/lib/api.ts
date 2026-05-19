const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? '';
const IS_DEVELOPMENT = process.env.NODE_ENV === 'development';
const USE_LOCAL_DEV_TOKEN = (() => {
  if (!IS_DEVELOPMENT) {
    return false;
  }
  if (API_BASE_URL === '') {
    return true;
  }
  try {
    const url = new URL(API_BASE_URL);
    return url.protocol === 'http:' && ['127.0.0.1', 'localhost', '[::1]'].includes(url.hostname);
  } catch {
    return false;
  }
})();

export type APIEnvelope<T> = {
  success: boolean;
  data: T | null;
  error: { code: string; message: string; details: Array<{ field?: string; reason: string }> } | null;
  request_id: string;
};

export type PageError = { code?: string; message: string; request_id?: string };

export function pageErrorFromEnvelope<T>(envelope: APIEnvelope<T>, fallbackMessage: string): PageError {
  return {
    code: envelope.error?.code,
    message: envelope.error?.message ?? fallbackMessage,
    request_id: envelope.request_id,
  };
}

const sensitiveKeyPattern = /(api[_-]?key|authorization|bearer|cookie|credential|password|secret|token)/i;
const sensitiveValuePattern = /(Bearer\s+)[^\s,;]+|((?:api[_-]?key|authorization|cookie|credential|password|secret|token)=)[^\s,;&]+|\beyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b|\b(?:sk|pk|api)[_-][A-Za-z0-9]{16,}\b/gi;

export function redactSensitive(value: unknown): unknown {
  if (typeof value === 'string') {
    return value.replace(sensitiveValuePattern, (_match, bearerPrefix: string | undefined, keyPrefix: string | undefined) => `${bearerPrefix ?? keyPrefix}[REDACTED]`);
  }
  if (Array.isArray(value)) {
    return value.map(item => redactSensitive(item));
  }
  if (value && typeof value === 'object') {
    return Object.fromEntries(
      Object.entries(value as Record<string, unknown>).map(([key, item]) => [
        key,
        sensitiveKeyPattern.test(key) ? '[REDACTED]' : redactSensitive(item),
      ]),
    );
  }
  return value;
}

export type Pagination = { page: number; page_size: number; total: number; has_next: boolean };
export type PagedResponse<T> = { items: T[]; pagination: Pagination };

export type HealthResponse = {
  status: string;
  service: string;
  version: string;
  timestamp: string;
};

export type InfoResponse = {
  app_name: string;
  environment: string;
  build_commit: string;
};

export type ConfigCheckResponse = {
  items: Array<{ key: string; required: boolean; configured: boolean; status: string }>;
  summary: { total: number; configured: number; missing: number };
};

export type DashboardSummaryResponse = {
  project_count: number;
  pending_review_count: number;
  pending_publish_count: number;
  failed_task_count: number;
  today_cost: number;
};

export type ContentTypeResponse = {
  id: string;
  code: string;
  name: string;
  project_schema: Record<string, unknown>;
  enabled: boolean;
};

export type ProjectResponse = {
  id: string;
  name: string;
  content_type_id: string;
  content_type_code: string;
  status: string;
  project_config: Record<string, unknown>;
};

export type ProjectOverviewResponse = {
  project_id: string;
  progress: number;
  pending_actions: number;
  cost: number;
};

export type PromptTemplateResponse = {
  id: string;
  code: string;
  agent_code: string;
  template: string;
  variables: string[];
};

export type LLMProviderResponse = {
  id: string;
  provider_type: string;
  base_url: string;
  api_key_masked: string;
  enabled: boolean;
};

async function request<T>(path: string, init: RequestInit = {}): Promise<APIEnvelope<T>> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      ...(USE_LOCAL_DEV_TOKEN ? { Authorization: 'Bearer dev' } : {}),
      ...(init.body ? { 'Content-Type': 'application/json' } : {}),
      ...init.headers,
    },
    cache: 'no-store',
  });
  const envelope = (await response.json()) as APIEnvelope<T>;
  if (!response.ok || !envelope.success) {
    return envelope;
  }
  return envelope;
}

export async function fetchHealth(): Promise<APIEnvelope<HealthResponse>> {
  return request<HealthResponse>('/api/v1/health');
}

export async function fetchSystemInfo(): Promise<APIEnvelope<InfoResponse>> {
  return request<InfoResponse>('/api/v1/system/info');
}

export async function fetchConfigCheck(): Promise<APIEnvelope<ConfigCheckResponse>> {
  return request<ConfigCheckResponse>('/api/v1/system/config-check');
}

export async function fetchDashboardSummary(): Promise<APIEnvelope<DashboardSummaryResponse>> {
  return request<DashboardSummaryResponse>('/api/v1/dashboard/summary');
}

export async function fetchContentTypes(): Promise<APIEnvelope<PagedResponse<ContentTypeResponse>>> {
  return request<PagedResponse<ContentTypeResponse>>('/api/v1/content-types?page=1&page_size=20&sort=code&order=asc');
}

export async function createContentType(input: { code: string; name: string; project_schema: Record<string, unknown> }): Promise<APIEnvelope<{ content_type_id: string }>> {
  return request<{ content_type_id: string }>('/api/v1/content-types', { method: 'POST', body: JSON.stringify(input) });
}

export async function fetchProjectSchema(contentTypeID: string): Promise<APIEnvelope<{ content_type_id: string; project_schema: Record<string, unknown> }>> {
  return request<{ content_type_id: string; project_schema: Record<string, unknown> }>(`/api/v1/content-types/${contentTypeID}/project-schema`);
}

export async function fetchProjects(query = ''): Promise<APIEnvelope<PagedResponse<ProjectResponse>>> {
  return request<PagedResponse<ProjectResponse>>(`/api/v1/projects?page=1&page_size=20${query}`);
}

export async function createProject(input: { name: string; content_type_id: string; project_config: Record<string, unknown> }): Promise<APIEnvelope<{ project_id: string; status: string }>> {
  return request<{ project_id: string; status: string }>('/api/v1/projects', { method: 'POST', body: JSON.stringify(input) });
}

export async function fetchProjectOverview(projectID: string): Promise<APIEnvelope<ProjectOverviewResponse>> {
  return request<ProjectOverviewResponse>(`/api/v1/projects/${projectID}/overview`);
}

export async function pauseProject(projectID: string, input: { reason: string; note: string }): Promise<APIEnvelope<{ project_id: string; status: string; operation_log_id: string }>> {
  return request<{ project_id: string; status: string; operation_log_id: string }>(`/api/v1/projects/${projectID}/pause`, { method: 'POST', body: JSON.stringify(input) });
}

export async function fetchPromptTemplates(): Promise<APIEnvelope<PagedResponse<PromptTemplateResponse>>> {
  return request<PagedResponse<PromptTemplateResponse>>('/api/v1/prompt-templates?page=1&page_size=20');
}

export async function createPromptTemplate(input: { code: string; template: string; variables: string[] }): Promise<APIEnvelope<{ prompt_template_id: string }>> {
  return request<{ prompt_template_id: string }>('/api/v1/prompt-templates', { method: 'POST', body: JSON.stringify(input) });
}

export async function fetchLLMProviders(): Promise<APIEnvelope<PagedResponse<LLMProviderResponse>>> {
  return request<PagedResponse<LLMProviderResponse>>('/api/v1/llm-providers?page=1&page_size=20');
}

export async function createLLMProvider(input: { provider_type: string; base_url: string; api_key: string }): Promise<APIEnvelope<{ provider_id: string; api_key_masked: string }>> {
  return request<{ provider_id: string; api_key_masked: string }>('/api/v1/llm-providers', { method: 'POST', body: JSON.stringify(input) });
}

// ---- Iteration 2: Workflow Types ----

export type WorkflowTemplateResponse = {
  id: string;
  code: string;
  name: string;
  content_type: string;
  category: string;
  description: string;
  status: string;
  created_at: string;
};

export type WorkflowTemplateVersionResponse = {
  id: string;
  template_id: string;
  version: number;
  input_schema: Record<string, unknown>;
  output_schema: Record<string, unknown>;
  status: string;
  created_at: string;
};

export type WorkflowStepTemplateResponse = {
  id: string;
  template_version_id: string;
  step_code: string;
  step_type: string;
  agent_code: string;
  order_index: number;
};

export type WorkflowRunResponse = {
  id: string;
  project_id: string;
  template_version_id: string;
  status: string;
  source: string;
  parent_run_id?: string;
  created_at: string;
  updated_at: string;
};

export type WorkflowStepRunResponse = {
  id: string;
  workflow_run_id: string;
  step_template_id: string;
  status: string;
  input: Record<string, unknown>;
  output: Record<string, unknown>;
  error?: string;
  started_at?: string;
  finished_at?: string;
};

export type AgentTaskResponse = {
  id: string;
  workflow_run_id: string;
  step_run_id: string;
  agent_code: string;
  status: string;
  created_at: string;
};

export type AgentTaskDetailResponse = AgentTaskResponse & {
  input: Record<string, unknown>;
  output: Record<string, unknown>;
  error?: string;
  started_at?: string;
  finished_at?: string;
  llm_call_log_count: number;
  llm_call_log_ids: string[];
};

export type LLMCallLogResponse = {
  id: string;
  workflow_run_id: string;
  step_run_id: string;
  agent_task_id: string;
  provider: string;
  model: string;
  input_tokens: number;
  output_tokens: number;
  cost: number;
  currency: string;
  latency_ms: number;
  status: string;
};

// ---- Iteration 2: Workflow API Functions ----

export async function fetchWorkflowTemplates(params?: { content_type?: string; category?: string; status?: string; page?: number; page_size?: number }): Promise<APIEnvelope<PagedResponse<WorkflowTemplateResponse>>> {
  const q = new URLSearchParams({ page: String(params?.page ?? 1), page_size: String(params?.page_size ?? 20), ...(params?.content_type ? { content_type: params.content_type } : {}), ...(params?.category ? { category: params.category } : {}), ...(params?.status ? { status: params.status } : {}) }).toString();
  return request<PagedResponse<WorkflowTemplateResponse>>(`/api/v1/workflow-templates?${q}`);
}

export async function createWorkflowTemplate(input: { code: string; name: string; content_type: string; category?: string; description?: string }): Promise<APIEnvelope<{ workflow_template_id: string; status: string }>> {
  return request<{ workflow_template_id: string; status: string }>('/api/v1/workflow-templates', { method: 'POST', body: JSON.stringify(input) });
}

export async function fetchWorkflowTemplate(id: string): Promise<APIEnvelope<WorkflowTemplateResponse>> {
  return request<WorkflowTemplateResponse>(`/api/v1/workflow-templates/${id}`);
}

export async function fetchWorkflowVersions(templateID: string): Promise<APIEnvelope<PagedResponse<WorkflowTemplateVersionResponse>>> {
  return request<PagedResponse<WorkflowTemplateVersionResponse>>(`/api/v1/workflow-templates/${templateID}/versions?page=1&page_size=20`);
}

export async function createWorkflowVersion(templateID: string, input: { steps: Array<{ step_code: string; step_type: string; agent_code?: string; order_index: number }> }): Promise<APIEnvelope<{ template_version_id: string; step_count: number; status: string }>> {
  return request<{ template_version_id: string; step_count: number; status: string }>(`/api/v1/workflow-templates/${templateID}/versions`, { method: 'POST', body: JSON.stringify(input) });
}

export async function publishWorkflowVersion(versionID: string, input: { note?: string }, idempotencyKey?: string): Promise<APIEnvelope<{ previous_status: string; current_status: string; operation_log_id: string }>> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (idempotencyKey) headers['Idempotency-Key'] = idempotencyKey;
  return request<{ previous_status: string; current_status: string; operation_log_id: string }>(`/api/v1/workflow-template-versions/${versionID}/publish`, { method: 'POST', body: JSON.stringify(input), headers });
}

export async function fetchWorkflowRuns(params?: { project_id?: string; template_version_id?: string; status?: string; page?: number; page_size?: number }): Promise<APIEnvelope<PagedResponse<WorkflowRunResponse>>> {
  const q = new URLSearchParams({ page: String(params?.page ?? 1), page_size: String(params?.page_size ?? 20), ...(params?.project_id ? { project_id: params.project_id } : {}), ...(params?.template_version_id ? { template_version_id: params.template_version_id } : {}), ...(params?.status ? { status: params.status } : {}) }).toString();
  return request<PagedResponse<WorkflowRunResponse>>(`/api/v1/workflow-runs?${q}`);
}

export async function createWorkflowRun(input: { project_id: string; template_version_id: string; input?: Record<string, unknown> }, idempotencyKey?: string): Promise<APIEnvelope<{ workflow_run_id: string; status: string }>> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (idempotencyKey) headers['Idempotency-Key'] = idempotencyKey;
  return request<{ workflow_run_id: string; status: string }>('/api/v1/workflow-runs', { method: 'POST', body: JSON.stringify(input), headers });
}

export async function fetchWorkflowRun(id: string): Promise<APIEnvelope<WorkflowRunResponse & { input: Record<string, unknown>; output: Record<string, unknown>; error?: string; step_count: number; agent_task_count: number }>> {
  return request<WorkflowRunResponse & { input: Record<string, unknown>; output: Record<string, unknown>; error?: string; step_count: number; agent_task_count: number }>(`/api/v1/workflow-runs/${id}`);
}

export async function fetchWorkflowRunSteps(runID: string): Promise<APIEnvelope<{ items: WorkflowStepRunResponse[] }>> {
  return request<{ items: WorkflowStepRunResponse[] }>(`/api/v1/workflow-runs/${runID}/steps`);
}

export async function cancelWorkflowRun(runID: string, input: { reason: string; note?: string }, idempotencyKey?: string): Promise<APIEnvelope<{ previous_status: string; current_status: string; operation_log_id: string }>> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (idempotencyKey) headers['Idempotency-Key'] = idempotencyKey;
  return request<{ previous_status: string; current_status: string; operation_log_id: string }>(`/api/v1/workflow-runs/${runID}/cancel`, { method: 'POST', body: JSON.stringify(input), headers });
}

export async function retryWorkflowRun(runID: string, input: { reason: string; input_override?: Record<string, unknown> }, idempotencyKey?: string): Promise<APIEnvelope<{ new_workflow_run_id: string; status: string }>> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (idempotencyKey) headers['Idempotency-Key'] = idempotencyKey;
  return request<{ new_workflow_run_id: string; status: string }>(`/api/v1/workflow-runs/${runID}/retry`, { method: 'POST', body: JSON.stringify(input), headers });
}

export async function fetchAgentTasks(params?: { workflow_run_id?: string; step_run_id?: string; agent_code?: string; status?: string; page?: number; page_size?: number }): Promise<APIEnvelope<PagedResponse<AgentTaskResponse>>> {
  const q = new URLSearchParams({ page: String(params?.page ?? 1), page_size: String(params?.page_size ?? 20), ...(params?.workflow_run_id ? { workflow_run_id: params.workflow_run_id } : {}), ...(params?.step_run_id ? { step_run_id: params.step_run_id } : {}), ...(params?.agent_code ? { agent_code: params.agent_code } : {}), ...(params?.status ? { status: params.status } : {}) }).toString();
  return request<PagedResponse<AgentTaskResponse>>(`/api/v1/agent-tasks?${q}`);
}

export async function fetchAgentTask(id: string): Promise<APIEnvelope<AgentTaskDetailResponse>> {
  return request<AgentTaskDetailResponse>(`/api/v1/agent-tasks/${id}`);
}

export async function fetchLLMCallLogs(params?: { workflow_run_id?: string; agent_task_id?: string; provider?: string; model?: string; status?: string; page?: number; page_size?: number }): Promise<APIEnvelope<PagedResponse<LLMCallLogResponse>>> {
  const q = new URLSearchParams({ page: String(params?.page ?? 1), page_size: String(params?.page_size ?? 20), ...(params?.workflow_run_id ? { workflow_run_id: params.workflow_run_id } : {}), ...(params?.agent_task_id ? { agent_task_id: params.agent_task_id } : {}), ...(params?.provider ? { provider: params.provider } : {}), ...(params?.model ? { model: params.model } : {}), ...(params?.status ? { status: params.status } : {}) }).toString();
  return request<PagedResponse<LLMCallLogResponse>>(`/api/v1/llm-call-logs?${q}`);
}

export async function fetchLLMCallLog(id: string): Promise<APIEnvelope<LLMCallLogResponse & { error?: string; request_id: string }>> {
  return request<LLMCallLogResponse & { error?: string; request_id: string }>(`/api/v1/llm-call-logs/${id}`);
}

export type WorkflowScheduleResponse = {
  id: string;
  project_id: string;
  template_version_id: string;
  cron_expression: string;
  enabled: boolean;
  next_run_at: string;
  daily_content_count: number;
  created_at: string;
  updated_at: string;
};

export type ScheduleTriggerResponse = {
  id: string;
  schedule_id: string;
  trigger_type: string;
  workflow_run_id?: string;
  status: string;
  error?: string;
  triggered_at: string;
};

export type LLMCostSummaryResponse = {
  calls: number;
  input_tokens: number;
  output_tokens: number;
  tokens: number;
  cost: number;
  currency: string;
  by_model: Array<{ model: string; calls: number; tokens: number; cost: number }>;
};

export type ExternalProviderResponse = {
  id: string;
  provider_type: string;
  base_url: string;
  token_masked: string;
  enabled: boolean;
};

export type ExternalBindingResponse = {
  id: string;
  provider_id: string;
  trigger_event: string;
  webhook_url: string;
  enabled: boolean;
};

export async function fetchWorkflowSchedules(): Promise<APIEnvelope<PagedResponse<WorkflowScheduleResponse>>> {
  return request<PagedResponse<WorkflowScheduleResponse>>('/api/v1/workflow-schedules?page=1&page_size=20');
}

export async function createWorkflowSchedule(input: { project_id: string; template_version_id: string; cron_expression: string; daily_content_count?: number }): Promise<APIEnvelope<{ schedule_id: string; next_run_at: string; daily_content_count: number; status: string }>> {
  return request<{ schedule_id: string; next_run_at: string; daily_content_count: number; status: string }>('/api/v1/workflow-schedules', { method: 'POST', body: JSON.stringify(input) });
}

export async function enableWorkflowSchedule(id: string, input: { note?: string }): Promise<APIEnvelope<{ previous_enabled: boolean; current_enabled: boolean; next_run_at: string; operation_log_id: string }>> {
  return request<{ previous_enabled: boolean; current_enabled: boolean; next_run_at: string; operation_log_id: string }>(`/api/v1/workflow-schedules/${id}/enable`, { method: 'POST', body: JSON.stringify(input) });
}

export async function disableWorkflowSchedule(id: string, input: { reason: string; note?: string }): Promise<APIEnvelope<{ previous_enabled: boolean; current_enabled: boolean; next_run_at: string; operation_log_id: string }>> {
  return request<{ previous_enabled: boolean; current_enabled: boolean; next_run_at: string; operation_log_id: string }>(`/api/v1/workflow-schedules/${id}/disable`, { method: 'POST', body: JSON.stringify(input) });
}

export async function testRunWorkflowSchedule(id: string, input: { input_override?: Record<string, unknown> }): Promise<APIEnvelope<{ workflow_run_id: string; status: string; trigger_log_id: string }>> {
  return request<{ workflow_run_id: string; status: string; trigger_log_id: string }>(`/api/v1/workflow-schedules/${id}/test-run`, { method: 'POST', body: JSON.stringify(input) });
}

export async function fetchScheduleTriggers(id: string): Promise<APIEnvelope<PagedResponse<ScheduleTriggerResponse>>> {
  return request<PagedResponse<ScheduleTriggerResponse>>(`/api/v1/workflow-schedules/${id}/triggers?page=1&page_size=20`);
}

export const fetchWorkflowScheduleTriggers = fetchScheduleTriggers;

export async function fetchLLMCostSummary(): Promise<APIEnvelope<LLMCostSummaryResponse>> {
  return request<LLMCostSummaryResponse>('/api/v1/llm-call-logs/summary');
}

export async function fetchExternalProviders(): Promise<APIEnvelope<PagedResponse<ExternalProviderResponse>>> {
  return request<PagedResponse<ExternalProviderResponse>>('/api/v1/external-automation/providers?page=1&page_size=20');
}

export async function createExternalProvider(input: { provider_type: string; base_url: string; token: string }): Promise<APIEnvelope<{ provider_id: string; token_masked: string }>> {
  return request<{ provider_id: string; token_masked: string }>('/api/v1/external-automation/providers', { method: 'POST', body: JSON.stringify(input) });
}

export async function fetchExternalBindings(): Promise<APIEnvelope<PagedResponse<ExternalBindingResponse>>> {
  return request<PagedResponse<ExternalBindingResponse>>('/api/v1/external-automation/bindings?page=1&page_size=20');
}

export async function createExternalBinding(input: { provider_id: string; trigger_event: string; webhook_url: string }): Promise<APIEnvelope<{ binding_id: string }>> {
  return request<{ binding_id: string }>('/api/v1/external-automation/bindings', { method: 'POST', body: JSON.stringify(input) });
}

export type PlanningRunResponse = {
  id: string;
  project_id: string;
  workflow_run_id: string;
  template_version_id: string;
  status: string;
  genre: string;
  audience: string;
  candidate_count: number;
  created_at: string;
  updated_at: string;
};

export type TopicCandidateResponse = {
  candidate_id: string;
  planning_run_id: string;
  snapshot_id: string;
  title: string;
  logline: string;
  status: string;
  score: number;
  reason: string;
  confirmed_topic_id?: string;
};

export type PlanningRunDetailResponse = PlanningRunResponse & {
  topics: TopicCandidateResponse[];
  step_runs: Array<{ id: string; status: string; error?: string }>;
  agent_tasks: Array<{ id: string; agent_code: string; status: string; error?: string }>;
  llm_call_logs: Array<{ id: string; model: string; status: string; cost: number }>;
};

export type WorldviewResponse = {
  project_id: string;
  version_id: string;
  version: number;
  worldview: Record<string, unknown>;
  forbidden_rules: string[];
  planning_run_id?: string;
  snapshot_id?: string;
};

export type CharacterResponse = {
  character_id: string;
  project_id: string;
  name: string;
  role: string;
  profile: Record<string, unknown>;
  planning_run_id?: string;
  snapshot_id?: string;
};

export type ArcResponse = {
  arc_id: string;
  project_id: string;
  title: string;
  summary: string;
  order_index: number;
  planning_run_id?: string;
  snapshot_id?: string;
};

function idempotencyHeaders(idempotencyKey?: string): Record<string, string> {
  return idempotencyKey ? { 'Idempotency-Key': idempotencyKey } : {};
}

export async function fetchPlanningRuns(projectID: string, params?: { status?: string; page?: number; page_size?: number; sort?: string; order?: string }): Promise<APIEnvelope<PagedResponse<PlanningRunResponse>>> {
  const q = new URLSearchParams({ page: String(params?.page ?? 1), page_size: String(params?.page_size ?? 20), ...(params?.status ? { status: params.status } : {}), ...(params?.sort ? { sort: params.sort } : {}), ...(params?.order ? { order: params.order } : {}) }).toString();
  return request<PagedResponse<PlanningRunResponse>>(`/api/v1/projects/${projectID}/novel/planning-runs?${q}`);
}

export async function createPlanningRun(projectID: string, input: { genre: string; audience: string; count: number; template_version_id: string; input_override?: Record<string, unknown> }, idempotencyKey?: string): Promise<APIEnvelope<{ planning_run_id: string; workflow_run_id: string; status: string }>> {
  return request<{ planning_run_id: string; workflow_run_id: string; status: string }>(`/api/v1/projects/${projectID}/novel/planning-runs`, { method: 'POST', body: JSON.stringify(input), headers: idempotencyHeaders(idempotencyKey) });
}

export async function fetchPlanningRun(projectID: string, runID: string): Promise<APIEnvelope<PlanningRunDetailResponse>> {
  return request<PlanningRunDetailResponse>(`/api/v1/projects/${projectID}/novel/planning-runs/${runID}`);
}

export async function confirmTopic(projectID: string, topicID: string, input: { note: string }, idempotencyKey?: string): Promise<APIEnvelope<{ confirmed_topic_id: string; previous_status: string; current_status: string; operation_log_id: string }>> {
  return request<{ confirmed_topic_id: string; previous_status: string; current_status: string; operation_log_id: string }>(`/api/v1/projects/${projectID}/novel/topics/${topicID}/confirm`, { method: 'POST', body: JSON.stringify(input), headers: idempotencyHeaders(idempotencyKey) });
}

export async function fetchWorldview(projectID: string): Promise<APIEnvelope<WorldviewResponse>> {
  return request<WorldviewResponse>(`/api/v1/projects/${projectID}/novel/worldview`);
}

export async function updateWorldview(projectID: string, input: { worldview: Record<string, unknown>; forbidden_rules: string[]; note: string }): Promise<APIEnvelope<{ version_id: string; operation_log_id: string }>> {
  return request<{ version_id: string; operation_log_id: string }>(`/api/v1/projects/${projectID}/novel/worldview`, { method: 'PATCH', body: JSON.stringify(input) });
}

export async function fetchCharacters(projectID: string, params?: { role?: string; page?: number; page_size?: number }): Promise<APIEnvelope<PagedResponse<CharacterResponse>>> {
  const q = new URLSearchParams({ page: String(params?.page ?? 1), page_size: String(params?.page_size ?? 20), ...(params?.role ? { role: params.role } : {}) }).toString();
  return request<PagedResponse<CharacterResponse>>(`/api/v1/projects/${projectID}/novel/characters?${q}`);
}

export async function createCharacter(projectID: string, input: { name: string; role: string; profile: Record<string, unknown>; note: string }): Promise<APIEnvelope<{ character_id: string; operation_log_id: string }>> {
  return request<{ character_id: string; operation_log_id: string }>(`/api/v1/projects/${projectID}/novel/characters`, { method: 'POST', body: JSON.stringify(input) });
}

export async function fetchArcs(projectID: string, params?: { page?: number; page_size?: number; sort?: string; order?: string }): Promise<APIEnvelope<PagedResponse<ArcResponse>>> {
  const q = new URLSearchParams({ page: String(params?.page ?? 1), page_size: String(params?.page_size ?? 20), ...(params?.sort ? { sort: params.sort } : {}), ...(params?.order ? { order: params.order } : {}) }).toString();
  return request<PagedResponse<ArcResponse>>(`/api/v1/projects/${projectID}/novel/arcs?${q}`);
}

export type GenerationRunResponse = {
  id: string;
  project_id: string;
  workflow_run_id: string;
  template_version_id: string;
  status: string;
  trigger_type: string;
  retry_of_generation_run_id?: string;
  created_at: string;
  updated_at: string;
};

export type ContentItemResponse = {
  id: string;
  project_id: string;
  generation_run_id: string;
  content_type_code: string;
  status: string;
  sequence_no: number;
  title: string;
  version: number;
  created_at: string;
  updated_at: string;
};

export type GenerationRunDetailResponse = GenerationRunResponse & {
  step_runs: Array<{ id: string; status: string; error?: string }>;
  agent_tasks: Array<{ id: string; agent_code: string; status: string; error?: string }>;
  llm_call_logs: Array<{ id: string; model: string; status: string; cost: number }>;
  content_items: ContentItemResponse[];
  error?: string;
};

export type ContentItemDetailResponse = ContentItemResponse & {
  body: string;
  metadata: Record<string, unknown>;
  extension: Record<string, unknown>;
};

export async function createGenerationRun(projectID: string, input: Record<string, unknown>, idempotencyKey?: string): Promise<APIEnvelope<{ generation_run_id: string; workflow_run_id: string; status: string }>> {
  return request<{ generation_run_id: string; workflow_run_id: string; status: string }>(`/api/v1/projects/${projectID}/generation-runs`, { method: 'POST', body: JSON.stringify(input), headers: idempotencyHeaders(idempotencyKey) });
}

export async function createBatchGenerationRuns(projectID: string, input: Record<string, unknown>, idempotencyKey?: string): Promise<APIEnvelope<{ generation_run_ids: string[]; workflow_run_ids: string[]; accepted_count: number }>> {
  return request<{ generation_run_ids: string[]; workflow_run_ids: string[]; accepted_count: number }>(`/api/v1/projects/${projectID}/generation-runs/batch`, { method: 'POST', body: JSON.stringify(input), headers: idempotencyHeaders(idempotencyKey) });
}

export async function fetchGenerationRuns(projectID: string, params?: { status?: string; page?: number; page_size?: number; sort?: string; order?: string }): Promise<APIEnvelope<PagedResponse<GenerationRunResponse>>> {
  const q = new URLSearchParams({ page: String(params?.page ?? 1), page_size: String(params?.page_size ?? 20), ...(params?.status ? { status: params.status } : {}), ...(params?.sort ? { sort: params.sort } : {}), ...(params?.order ? { order: params.order } : {}) }).toString();
  return request<PagedResponse<GenerationRunResponse>>(`/api/v1/projects/${projectID}/generation-runs?${q}`);
}

export async function fetchGenerationRun(runID: string): Promise<APIEnvelope<GenerationRunDetailResponse>> {
  return request<GenerationRunDetailResponse>(`/api/v1/generation-runs/${runID}`);
}

export async function retryGenerationRun(runID: string, input: { reason: string; input_override?: Record<string, unknown> }, idempotencyKey?: string): Promise<APIEnvelope<{ new_generation_run_id: string; workflow_run_id: string; operation_log_id: string }>> {
  return request<{ new_generation_run_id: string; workflow_run_id: string; operation_log_id: string }>(`/api/v1/generation-runs/${runID}/retry`, { method: 'POST', body: JSON.stringify(input), headers: idempotencyHeaders(idempotencyKey) });
}

export async function fetchContentItems(projectID: string, params?: { status?: string; page?: number; page_size?: number; sort?: string; order?: string }): Promise<APIEnvelope<PagedResponse<ContentItemResponse>>> {
  const q = new URLSearchParams({ page: String(params?.page ?? 1), page_size: String(params?.page_size ?? 20), ...(params?.status ? { status: params.status } : {}), ...(params?.sort ? { sort: params.sort } : {}), ...(params?.order ? { order: params.order } : {}) }).toString();
  return request<PagedResponse<ContentItemResponse>>(`/api/v1/projects/${projectID}/content-items?${q}`);
}

export async function fetchContentItem(itemID: string): Promise<APIEnvelope<ContentItemDetailResponse>> {
  return request<ContentItemDetailResponse>(`/api/v1/content-items/${itemID}`);
}

// ---- Iteration 5: Review and Quality Control Types ----

export type ContentReviewResponse = {
  id: string;
  project_id: string;
  content_item_id: string;
  review_type: string;
  status: string;
  title: string;
  updated_at: string;
};

export type ContentVersionResponse = {
  id: string;
  content_item_id: string;
  version_no: number;
  source: string;
  title: string;
  body: string;
  editable_fields: Record<string, unknown>;
  summary: string;
  created_at: string;
};

export type ReviewReportSummaryResponse = {
  id: string;
  status: string;
  quality_score: number;
  risk_level: string;
};

export type ContentReviewDetailResponse = ContentReviewResponse & {
  body: string;
  metadata: Record<string, unknown>;
  extension: Record<string, unknown>;
  report_summary: ReviewReportSummaryResponse;
  versions: ContentVersionResponse[];
};

export type ReviewReportResponse = {
  id: string;
  review_id: string;
  content_item_id: string;
  status: string;
  quality_score: number;
  risk_level: string;
  issues: Array<{ code: string; severity: string; message: string }>;
  suggestions: Array<{ code: string; message: string }>;
  job_id?: string;
  workflow_run_id?: string;
  error_code?: string;
  error_message?: string;
};

export async function fetchContentReviews(projectID: string, params?: { status?: string; page?: number; page_size?: number }): Promise<APIEnvelope<PagedResponse<ContentReviewResponse>>> {
  const q = new URLSearchParams({ project_id: projectID, page: String(params?.page ?? 1), page_size: String(params?.page_size ?? 20), ...(params?.status ? { status: params.status } : {}) }).toString();
  return request<PagedResponse<ContentReviewResponse>>(`/api/v1/content-reviews?${q}`);
}

export async function createContentReview(contentItemID: string, input: { review_type: string }, idempotencyKey: string): Promise<APIEnvelope<{ review_id: string; status: string }>> {
  return request<{ review_id: string; status: string }>(`/api/v1/content-items/${contentItemID}/reviews`, { method: 'POST', body: JSON.stringify(input), headers: { 'Idempotency-Key': idempotencyKey } });
}

export async function fetchContentReview(reviewID: string): Promise<APIEnvelope<ContentReviewDetailResponse>> {
  return request<ContentReviewDetailResponse>(`/api/v1/content-reviews/${reviewID}`);
}

export async function triggerAIReport(reviewID: string, input: { report_type: string; config: Record<string, unknown> }, idempotencyKey: string): Promise<APIEnvelope<{ report_id: string; job_id: string; workflow_run_id: string; status: string }>> {
  return request<{ report_id: string; job_id: string; workflow_run_id: string; status: string }>(`/api/v1/content-reviews/${reviewID}/ai-report`, { method: 'POST', body: JSON.stringify(input), headers: { 'Idempotency-Key': idempotencyKey } });
}

export async function fetchAIReport(reviewID: string): Promise<APIEnvelope<ReviewReportResponse>> {
  return request<ReviewReportResponse>(`/api/v1/content-reviews/${reviewID}/ai-report`);
}

export async function approveReview(reviewID: string, input: { note: string }): Promise<APIEnvelope<{ review_id: string; status: string; operation_log_id: string }>> {
  return request<{ review_id: string; status: string; operation_log_id: string }>(`/api/v1/content-reviews/${reviewID}/approve`, { method: 'POST', body: JSON.stringify(input) });
}

export async function rejectReview(reviewID: string, input: { reason: string; regenerate_instruction: string; trigger_regeneration: boolean }): Promise<APIEnvelope<{ review_id: string; status: string; operation_log_id: string; regeneration_run_id?: string; job_id?: string }>> {
  return request<{ review_id: string; status: string; operation_log_id: string; regeneration_run_id?: string; job_id?: string }>(`/api/v1/content-reviews/${reviewID}/reject`, { method: 'POST', body: JSON.stringify(input) });
}

export async function approveWithEdit(reviewID: string, input: { editable_fields: Record<string, unknown>; note: string }): Promise<APIEnvelope<{ review_id: string; status: string; content_version_id: string; operation_log_id: string }>> {
  return request<{ review_id: string; status: string; content_version_id: string; operation_log_id: string }>(`/api/v1/content-reviews/${reviewID}/approve-with-edit`, { method: 'POST', body: JSON.stringify(input) });
}

export async function fetchContentVersions(contentItemID: string): Promise<APIEnvelope<PagedResponse<ContentVersionResponse>>> {
  return request<PagedResponse<ContentVersionResponse>>(`/api/v1/content-items/${contentItemID}/versions?page=1&page_size=20`);
}

// ---- Iteration 6: Knowledge Memory Types ----

export type RecentWindowPolicy = { item_count: number; token_limit: number; truncation_policy: string };
export type SnapshotSummaryResponse = { id: string; source_type: string; estimated_tokens: number; truncation_policy: string; created_at: string };
export type KnowledgeMemoryResponse = { id: string; project_id: string; static_context: Record<string, unknown>; dynamic_state: Record<string, unknown>; recent_window_policy: RecentWindowPolicy; style_guide: Record<string, unknown>; version: number; updated_at: string; recent_snapshot_summary: SnapshotSummaryResponse };
export type MemorySnapshotResponse = { id: string; project_id: string; content_item_id?: string; source_type: string; token_budget: number; estimated_tokens: number; truncation_policy: string; created_at: string };
export type ContextPreviewResponse = { sources: string[]; token_budget: number; estimated_tokens: number; truncation_policy: string; preview_text: string };
export type ConsistencyIssue = { issue_id: string; severity: string; type: string; title: string; description: string; affected_content_items: string[]; suggestion: string };
export type ConsistencyReportResponse = { id: string; project_id: string; status: string; issue_count: number; severity_summary: Record<string, number>; created_at: string };
export type ConsistencyReportDetailResponse = ConsistencyReportResponse & { source_snapshot_id: string; issues: ConsistencyIssue[]; error_code?: string; error_message?: string };

export async function fetchKnowledgeMemory(projectID: string): Promise<APIEnvelope<KnowledgeMemoryResponse>> {
  return request<KnowledgeMemoryResponse>(`/api/v1/projects/${projectID}/knowledge-memory`);
}

export async function updateStaticContext(projectID: string, input: { static_context: Record<string, unknown>; note: string }): Promise<APIEnvelope<{ version: number; operation_log_id: string }>> {
  return request<{ version: number; operation_log_id: string }>(`/api/v1/projects/${projectID}/knowledge-memory/static-context`, { method: 'PATCH', body: JSON.stringify(input) });
}

export async function updateStyleGuide(projectID: string, input: { style_guide: Record<string, unknown>; note: string }): Promise<APIEnvelope<{ version: number; operation_log_id: string }>> {
  return request<{ version: number; operation_log_id: string }>(`/api/v1/projects/${projectID}/knowledge-memory/style-guide`, { method: 'PATCH', body: JSON.stringify(input) });
}

export async function correctDynamicState(projectID: string, input: { reason: string; changes: Record<string, unknown>; source_refs: string[] }, idempotencyKey: string): Promise<APIEnvelope<{ memory_snapshot_id: string; dynamic_state_version: number; operation_log_id: string }>> {
  return request<{ memory_snapshot_id: string; dynamic_state_version: number; operation_log_id: string }>(`/api/v1/projects/${projectID}/knowledge-memory/dynamic-state-correction`, { method: 'PATCH', body: JSON.stringify(input), headers: { 'Idempotency-Key': idempotencyKey } });
}

export async function updateRecentWindowPolicy(projectID: string, input: RecentWindowPolicy): Promise<APIEnvelope<RecentWindowPolicy & { version: number; operation_log_id: string }>> {
  return request<RecentWindowPolicy & { version: number; operation_log_id: string }>(`/api/v1/projects/${projectID}/knowledge-memory/recent-window-policy`, { method: 'PATCH', body: JSON.stringify(input) });
}

export async function fetchMemorySnapshots(projectID: string, params?: { page?: number; page_size?: number; content_item_id?: string }): Promise<APIEnvelope<PagedResponse<MemorySnapshotResponse>>> {
  const q = new URLSearchParams({ page: String(params?.page ?? 1), page_size: String(params?.page_size ?? 20), ...(params?.content_item_id ? { content_item_id: params.content_item_id } : {}) }).toString();
  return request<PagedResponse<MemorySnapshotResponse>>(`/api/v1/projects/${projectID}/knowledge-memory/snapshots?${q}`);
}

export async function previewContext(projectID: string, params: { purpose: string; budget: number; content_item_id?: string }): Promise<APIEnvelope<ContextPreviewResponse>> {
  const q = new URLSearchParams({ purpose: params.purpose, budget: String(params.budget), ...(params.content_item_id ? { content_item_id: params.content_item_id } : {}) }).toString();
  return request<ContextPreviewResponse>(`/api/v1/projects/${projectID}/knowledge-memory/context-preview?${q}`);
}

export async function assembleContext(projectID: string, input: { purpose: string; budget: number; content_item_id?: string }, idempotencyKey: string): Promise<APIEnvelope<{ context_snapshot_id: string; estimated_tokens: number; truncation_policy: string }>> {
  return request<{ context_snapshot_id: string; estimated_tokens: number; truncation_policy: string }>(`/api/v1/projects/${projectID}/knowledge-memory/assemble-context`, { method: 'POST', body: JSON.stringify(input), headers: { 'Idempotency-Key': idempotencyKey } });
}

export async function updateDynamicState(contentItemID: string, input: { summary: string; changes: Record<string, unknown>; source_version_id: string }, idempotencyKey: string): Promise<APIEnvelope<{ memory_snapshot_id: string; dynamic_state_version: number }>> {
  return request<{ memory_snapshot_id: string; dynamic_state_version: number }>(`/api/v1/content-items/${contentItemID}/update-dynamic-state`, { method: 'POST', body: JSON.stringify(input), headers: { 'Idempotency-Key': idempotencyKey } });
}

export async function createConsistencyReport(projectID: string, input: { range: Record<string, unknown>; scope: string; severity_threshold: string }, idempotencyKey: string): Promise<APIEnvelope<{ report_id: string; status: string }>> {
  return request<{ report_id: string; status: string }>(`/api/v1/projects/${projectID}/consistency-reports`, { method: 'POST', body: JSON.stringify(input), headers: { 'Idempotency-Key': idempotencyKey } });
}

export async function fetchConsistencyReports(projectID: string, params?: { status?: string; page?: number; page_size?: number }): Promise<APIEnvelope<PagedResponse<ConsistencyReportResponse>>> {
  const q = new URLSearchParams({ page: String(params?.page ?? 1), page_size: String(params?.page_size ?? 20), ...(params?.status ? { status: params.status } : {}) }).toString();
  return request<PagedResponse<ConsistencyReportResponse>>(`/api/v1/projects/${projectID}/consistency-reports?${q}`);
}

export async function fetchConsistencyReport(projectID: string, reportID: string): Promise<APIEnvelope<ConsistencyReportDetailResponse>> {
  return request<ConsistencyReportDetailResponse>(`/api/v1/projects/${projectID}/consistency-reports/${reportID}`);
}
