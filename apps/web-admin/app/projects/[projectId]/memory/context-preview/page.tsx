'use client';

import { use, useState } from 'react';
import ProjectWorkspaceNav from '../../workspace-nav';
import { assembleContext, pageErrorFromEnvelope, previewContext, type ContextPreviewResponse, type PageError } from '../../../../../lib/api';

export default function ContextPreviewPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [preview, setPreview] = useState<ContextPreviewResponse | null>(null);
  const [notice, setNotice] = useState('');
  const [error, setError] = useState<PageError | null>(null);
  const [purpose, setPurpose] = useState('draft_generation');
  const [budget, setBudget] = useState(2000);
  const [contentItemID, setContentItemID] = useState('');

  async function submitPreview() {
    const envelope = await previewContext(projectId, { purpose, budget, content_item_id: contentItemID || undefined });
    if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '预览上下文失败'));
    setPreview(envelope.data);
    setNotice('上下文预览已生成，未落库');
    setError(null);
  }

  async function submitSnapshot() {
    const envelope = await assembleContext(projectId, { purpose, budget, content_item_id: contentItemID || undefined }, `context-snapshot-${Date.now()}`);
    if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '生成上下文快照失败'));
    setNotice(`已生成上下文快照 快照 ID=${envelope.data.context_snapshot_id} estimated=${envelope.data.estimated_tokens}`);
    setPreview(null);
    setError(null);
  }

  return (
    <main className="page-shell">
      <ProjectWorkspaceNav projectId={projectId} />
      <section className="page-hero"><h1>预览上下文</h1><p>预览组装上下文或生成可追踪的记忆快照。</p></section>
      {notice && <section className="card" role="status">{notice}</section>}
      {error && <section className="card" role="alert">错误码: {error.code} 错误信息: {error.message} request_id={error.request_id}</section>}
      <section className="card">
        <h2>参数</h2>
        <label>用途 <input value={purpose} onChange={e => setPurpose(e.target.value)} /></label>
        <label>Token 预算 <input type="number" value={budget} onChange={e => setBudget(Number(e.target.value))} /></label>
        <label>内容单元 ID <input value={contentItemID} onChange={e => setContentItemID(e.target.value)} /></label>
        <button type="button" onClick={() => void submitPreview()}>预览上下文</button> <button type="button" onClick={() => void submitSnapshot()}>生成上下文快照</button>
      </section>
      {preview && <section className="card">
        <h2>预览结果</h2>
        <p>上下文来源: {preview.sources.join(', ')}</p>
        <p>Token 预算: {preview.token_budget} · 预估 Token: {preview.estimated_tokens} · 截断策略: {preview.truncation_policy}</p>
        <h3>预览内容</h3>
        <pre>{preview.preview_text}</pre>
      </section>}
    </main>
  );
}
