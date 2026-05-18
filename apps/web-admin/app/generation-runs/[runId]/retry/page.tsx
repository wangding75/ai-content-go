'use client';

import { use, useState } from 'react';
import { pageErrorFromEnvelope, retryGenerationRun, type PageError } from '../../../../lib/api';

export default function GenerationRetryPage({ params }: { params: Promise<{ runId: string }> }) {
  const { runId } = use(params);
  const [reason, setReason] = useState('provider recovered');
  const [inputOverride, setInputOverride] = useState('{}');
  const [result, setResult] = useState('');
  const [error, setError] = useState<PageError | null>(null);

  async function submitRetry() {
    let input_override: Record<string, unknown> = {};
    try {
      input_override = JSON.parse(inputOverride) as Record<string, unknown>;
    } catch {
      setError({ message: 'input_override must be valid JSON' });
      return;
    }
    const envelope = await retryGenerationRun(runId, { reason, input_override }, `retry-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '重试失败'));
      return;
    }
    setResult(`new_generation_run_id=${envelope.data.new_generation_run_id} workflow_run_id=${envelope.data.workflow_run_id} operation_log_id=${envelope.data.operation_log_id} request_id=${envelope.request_id}`);
    setError(null);
  }

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>失败重试</h1>
        <p>为失败的 generation run 提交 retry。</p>
      </section>
      <section className="card">
        <label>reason<input value={reason} onChange={(event) => setReason(event.target.value)} /></label>
        <label>input_override<textarea value={inputOverride} onChange={(event) => setInputOverride(event.target.value)} /></label>
        <button type="button" onClick={submitRetry}>retry</button>
        {result && <p role="status">{result}</p>}
        {error && <p role="alert">{error.message} request_id={error.request_id}</p>}
      </section>
    </main>
  );
}
