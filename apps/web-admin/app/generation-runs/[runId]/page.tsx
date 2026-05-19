'use client';

import Link from 'next/link';
import { use, useEffect, useState } from 'react';
import { fetchGenerationRun, pageErrorFromEnvelope, type GenerationRunDetailResponse, type PageError } from '../../../lib/api';

export default function GenerationRunDetailPage({ params }: { params: Promise<{ runId: string }> }) {
  const { runId } = use(params);
  const [run, setRun] = useState<GenerationRunDetailResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);

  useEffect(() => {
    async function loadRun() {
      const envelope = await fetchGenerationRun(runId);
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '加载生成运行详情失败'));
        return;
      }
      setRun(envelope.data);
      setError(null);
    }
    void loadRun();
  }, [runId]);

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>生成运行详情</h1>
        <p>查看 workflow_run_id、agent_tasks、llm_call_logs 和 content_items。</p>
      </section>
      {error && <section className="card" role="alert">{error.message} request_id={error.request_id}</section>}
      {run && (
        <section className="card">
          <h2>{run.id}</h2>
          <p>workflow_run_id: {run.workflow_run_id}</p>
          <p>status: {run.status}</p>
          {run.status === 'failed' && <Link href={`/generation-runs/${run.id}/retry`}>retry</Link>}
          <h3>agent_tasks</h3>
          <pre>{JSON.stringify(run.agent_tasks, null, 2)}</pre>
          <h3>llm_call_logs</h3>
          <pre>{JSON.stringify(run.llm_call_logs, null, 2)}</pre>
          <h3>content_items</h3>
          <pre>{JSON.stringify(run.content_items, null, 2)}</pre>
        </section>
      )}
    </main>
  );
}
