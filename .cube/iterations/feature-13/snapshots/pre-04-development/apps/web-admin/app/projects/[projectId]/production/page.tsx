'use client';

import { use, useEffect, useState } from 'react';
import { createBatchGenerationRuns, createGenerationRun, fetchGenerationRuns, pageErrorFromEnvelope, type GenerationRunResponse, type PageError } from '../../../../lib/api';

export default function ProductionPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);  const [runs, setRuns] = useState<GenerationRunResponse[]>([]);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');
  const [lastRequestID, setLastRequestID] = useState('');

  async function loadRuns() {
    const envelope = await fetchGenerationRuns(projectId, { page: 1, page_size: 20 });
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '加载生成运行失败'));
      return;
    }
    setRuns(envelope.data.items);
    setLastRequestID(envelope.request_id);
    setError(null);
  }

  useEffect(() => {
    void loadRuns();
  }, [projectId]);

  async function startManualGeneration() {
    const envelope = await createGenerationRun(projectId, {
      confirmed_topic_id: 'topic-1',
      worldview_version_id: 'worldview-v1',
      arc_id: 'arc-1',
      target_count: 1,
      start_sequence_no: 1,
      template_version_id: 'wftv-generation',
      generation_config: {},
    }, `manual-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '手动生成失败'));
      return;
    }
    setNotice(`手动生成已受理：${envelope.data.generation_run_id} request_id=${envelope.request_id}`);
    await loadRuns();
  }

  async function startBatchGeneration() {
    const envelope = await createBatchGenerationRuns(projectId, {
      range: { start_sequence_no: 1, end_sequence_no: 3 },
      batch_size: 3,
      template_version_id: 'wftv-generation',
      generation_config: {},
    }, `batch-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '批量生成失败'));
      return;
    }
    setNotice(`批量生成已受理：${envelope.data.accepted_count} request_id=${envelope.request_id}`);
    await loadRuns();
  }

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>内容生产</h1>
        <p>发起单次或批量内容生成，并跟踪 generation run 的受理状态。</p>
      </section>
      <section className="card">
        <h2>生成操作</h2>
        <div className="action-row">
          <button type="button" onClick={startManualGeneration}>手动生成</button>
          <button type="button" onClick={startBatchGeneration}>批量生成</button>
        </div>
        {lastRequestID && <p>request_id={lastRequestID}</p>}
        {notice && <p role="status">{notice}</p>}
        {error && <p role="alert">{error.message} request_id={error.request_id}</p>}
      </section>
      <section className="card">
        <h2>生成运行</h2>
        <ul>
          {runs.map((run) => <li key={run.id}>{run.id} · {run.workflow_run_id} · {run.status}</li>)}
        </ul>
      </section>
    </main>
  );
}
