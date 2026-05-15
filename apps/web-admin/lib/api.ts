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
