export type APIEnvelope<T> = {
  success: boolean;
  data: T | null;
  error: { code: string; message: string; details: Array<{ field?: string; reason: string }> } | null;
  request_id: string;
};

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

export async function fetchHealth(): Promise<APIEnvelope<HealthResponse>> {
  throw new Error('not implemented');
}

export async function fetchSystemInfo(): Promise<APIEnvelope<InfoResponse>> {
  throw new Error('not implemented');
}

export async function fetchConfigCheck(): Promise<APIEnvelope<ConfigCheckResponse>> {
  throw new Error('not implemented');
}
