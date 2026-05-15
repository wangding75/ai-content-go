const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? 'http://127.0.0.1:18080';
const API_TOKEN = process.env.NEXT_PUBLIC_API_TOKEN ?? 'dev';

export type APIEnvelope<T> = {
  success: boolean;
  data: T | null;
  error: { code: string; message: string; details: Array<{ field?: string; reason: string }> } | null;
  request_id: string;
};

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
      Authorization: `Bearer ${API_TOKEN}`,
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
