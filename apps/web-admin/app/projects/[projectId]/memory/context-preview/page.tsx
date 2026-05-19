'use client';

import { use, useState } from 'react';
import { assembleContext, pageErrorFromEnvelope, previewContext, type ContextPreviewResponse, type PageError } from '../../../../../lib/api';

export default function ContextPreviewPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [preview, setPreview] = useState<ContextPreviewResponse | null>(null);
  const [notice, setNotice] = useState('');
  const [error, setError] = useState<PageError | null>(null);

  async function submitPreview() {
    const envelope = await previewContext(projectId, { purpose: 'draft_generation', budget: 2000 });
    if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '预览上下文失败'));
    setPreview(envelope.data);
    setNotice('上下文预览已生成，未落库');
    setError(null);
  }

  async function submitSnapshot() {
    const envelope = await assembleContext(projectId, { purpose: 'draft_generation', budget: 2000 }, `context-snapshot-${Date.now()}`);
    if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '生成上下文快照失败'));
    setNotice(`上下文快照已生成 snapshot=${envelope.data.context_snapshot_id} estimated=${envelope.data.estimated_tokens}`);
    setError(null);
  }

  return (
    <main className="page-shell">
      <section className="page-hero"><h1>上下文预览</h1><p>预览组装上下文或生成可追踪的记忆快照。</p></section>
      {notice && <section className="card" role="status">{notice}</section>}
      {error && <section className="card" role="alert">{error.code} {error.message} request_id={error.request_id}</section>}
      <section className="card"><h2>操作</h2><button type="button" onClick={() => void submitPreview()}>预览上下文</button> <button type="button" onClick={() => void submitSnapshot()}>生成上下文快照</button></section>
      {preview && <section className="card"><h2>预览结果</h2><p>{preview.estimated_tokens}/{preview.token_budget} tokens · {preview.truncation_policy}</p><p>sources={preview.sources.join(', ')}</p><pre>{preview.preview_text}</pre></section>}
    </main>
  );
}
