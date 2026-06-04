'use client';

import Link from 'next/link';
import { use, useEffect, useState } from 'react';
import { fetchPublishJob, pageErrorFromEnvelope, type PageError, type PublishJobDetailResponse } from '../../../lib/api';

export default function PublishJobDetailPage({ params, searchParams }: { params: Promise<{ jobId: string }>; searchParams: Promise<{ project_id?: string }> }) {
  const { jobId } = use(params);
  const { project_id: projectId = '' } = use(searchParams);
  const [job, setJob] = useState<PublishJobDetailResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);

  useEffect(() => {
    async function load() {
      if (!projectId) {
        setError({ message: '缺少 project_id' });
        return;
      }
      const envelope = await fetchPublishJob(projectId, jobId);
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '加载发布详情失败'));
        return;
      }
      setJob(envelope.data);
      setError(null);
    }
    void load();
  }, [projectId, jobId]);

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>发布详情</h1>
        <p>从详情接口刷新任务、目标、内容版本和发布日志。</p>
      </section>
      {error && <p role="alert">{error.code} {error.message} request_id={error.request_id}</p>}
      {job && (
        <>
          <section className="card">
            <div className="card__header">
              <div>
                <h2>{job.title}</h2>
                <p>{job.id} · {job.status} · {job.payload_hash}</p>
              </div>
              <div className="action-row">
                <Link href={`/publish-jobs/${job.id}/copy?project_id=${encodeURIComponent(projectId)}`}>复制载荷</Link>
                <Link href={`/publish-jobs/${job.id}/backfill?project_id=${encodeURIComponent(projectId)}`}>发布回填</Link>
              </div>
            </div>
            <p>目标：{job.target.platform} · {job.target.display_name}</p>
            <p>版本：{job.content_version.id} · {job.content_version.source}</p>
            <p>外部链接：{job.external_url || '未回填'}</p>
          </section>
          <section className="card">
            <h2>行为摘要</h2>
            {job.logs.length === 0 ? <p className="muted">暂无发布日志</p> : (
              <ul>
                {job.logs.map((log) => <li key={log.id}>{log.created_at} · {log.event_type} · {log.from_status || '-'} → {log.to_status || '-'}</li>)}
              </ul>
            )}
          </section>
        </>
      )}
    </main>
  );
}
