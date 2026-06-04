'use client';

import { use, useState } from 'react';
import { markPublishJobFailed, markPublishJobPublished, pageErrorFromEnvelope, requeuePublishJob, type PageError } from '../../../../lib/api';

export default function PublishJobBackfillPage({ params, searchParams }: { params: Promise<{ jobId: string }>; searchParams: Promise<{ project_id?: string }> }) {
  const { jobId } = use(params);
  const { project_id: projectId = '' } = use(searchParams);
  const [externalUrl, setExternalUrl] = useState('');
  const [publishedAt, setPublishedAt] = useState('');
  const [reason, setReason] = useState('');
  const [note, setNote] = useState('');
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');

  function requireProject() {
    if (projectId) {
      return true;
    }
    setError({ message: '缺少 project_id' });
    return false;
  }

  async function markPublished() {
    if (!requireProject()) {
      return;
    }
    const envelope = await markPublishJobPublished(projectId, jobId, { external_url: externalUrl, published_at: publishedAt || undefined, reason, note }, `published-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '标记已发布失败'));
      return;
    }
    setNotice(`已发布：${envelope.data.current_status} request_id=${envelope.request_id}`);
    setError(null);
  }

  async function markFailed() {
    if (!requireProject()) {
      return;
    }
    const envelope = await markPublishJobFailed(projectId, jobId, { reason, retryable: true, note }, `failed-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '标记失败失败'));
      return;
    }
    setNotice(`已标记失败：${envelope.data.current_status} request_id=${envelope.request_id}`);
    setError(null);
  }

  async function requeue() {
    if (!requireProject()) {
      return;
    }
    const envelope = await requeuePublishJob(projectId, jobId, { reason, note }, `requeue-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '重新入队失败'));
      return;
    }
    setNotice(`已重新入队：retry_count=${envelope.data.retry_count} request_id=${envelope.request_id}`);
    setError(null);
  }

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>发布回填</h1>
        <p>人工完成平台发布后回填结果，或记录失败原因并重新入队。</p>
      </section>
      <section className="card">
        <div className="form-grid">
          <label>外部链接
            <input value={externalUrl} onChange={(event) => setExternalUrl(event.target.value)} placeholder="https://example.com/published" />
          </label>
          <label>published_at
            <input value={publishedAt} onChange={(event) => setPublishedAt(event.target.value)} placeholder="2026-05-21T09:00:00Z" />
          </label>
          <label>原因
            <input value={reason} onChange={(event) => setReason(event.target.value)} placeholder="必填失败原因或无链接原因" />
          </label>
          <label>备注
            <input value={note} onChange={(event) => setNote(event.target.value)} placeholder="人工发布备注" />
          </label>
        </div>
        <div className="action-row">
          <button type="button" onClick={markPublished}>标记已发布</button>
          <button type="button" onClick={markFailed}>标记失败</button>
          <button type="button" onClick={requeue}>重新入队</button>
        </div>
        {notice && <p role="status">{notice}</p>}
        {error && <p role="alert">{error.code} {error.message} request_id={error.request_id}</p>}
      </section>
    </main>
  );
}
