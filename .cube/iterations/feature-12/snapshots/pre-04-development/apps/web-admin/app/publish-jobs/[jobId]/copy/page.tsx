'use client';

import { use, useEffect, useState } from 'react';
import { copyPublishPayload, fetchPublishCopyPayload, pageErrorFromEnvelope, type PageError, type PublishCopyPayloadResponse } from '../../../../lib/api';

export default function PublishJobCopyPage({ params, searchParams }: { params: Promise<{ jobId: string }>; searchParams: Promise<{ project_id?: string }> }) {
  const { jobId } = use(params);
  const { project_id: projectId = '' } = use(searchParams);
  const [payload, setPayload] = useState<PublishCopyPayloadResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');

  useEffect(() => {
    async function load() {
      if (!projectId) {
        setError({ message: '缺少 project_id' });
        return;
      }
      const envelope = await fetchPublishCopyPayload(projectId, jobId);
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '加载复制载荷失败'));
        return;
      }
      setPayload(envelope.data);
      setError(null);
    }
    void load();
  }, [projectId, jobId]);

  async function copy(scope: string) {
    const envelope = await copyPublishPayload(projectId, jobId, { copy_scope: scope, note: 'manual copy' }, `copy-${scope}-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '记录复制行为失败'));
      return;
    }
    setNotice(`复制已记录：${envelope.data.current_status} request_id=${envelope.request_id}`);
  }

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>复制发布内容</h1>
        <p>预览不会改变任务状态，点击复制按钮后才记录发布复制行为。</p>
      </section>
      {error && <p role="alert">{error.code} {error.message} request_id={error.request_id}</p>}
      {notice && <p role="status">{notice}</p>}
      {payload && (
        <section className="card">
          <h2>{payload.title}</h2>
          <p>format={payload.format} platform={payload.platform} target_id={payload.target_id}</p>
          <p>content_version_id={payload.content_version_id}</p>
          <p>payload_hash={payload.payload_hash}</p>
          <pre>{payload.body}</pre>
          <div className="action-row">
            <button type="button" onClick={() => copy('title')}>复制标题</button>
            <button type="button" onClick={() => copy('body')}>复制正文</button>
            <button type="button" onClick={() => copy('full')}>复制完整载荷</button>
          </div>
        </section>
      )}
    </main>
  );
}
