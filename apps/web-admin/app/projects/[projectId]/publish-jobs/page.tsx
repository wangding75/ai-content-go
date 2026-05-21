'use client';

import Link from 'next/link';
import { use, useEffect, useState } from 'react';
import { createPublishJob, fetchPublishJobs, fetchPublishTargets, pageErrorFromEnvelope, type PageError, type PublishJobResponse, type PublishTargetResponse } from '../../../../lib/api';

export default function PublishJobsPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [jobs, setJobs] = useState<PublishJobResponse[]>([]);
  const [targets, setTargets] = useState<PublishTargetResponse[]>([]);
  const [status, setStatus] = useState('');
  const [targetID, setTargetID] = useState('');
  const [page, setPage] = useState(1);
  const [hasNext, setHasNext] = useState(false);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');

  async function load(nextPage = page) {
    const [jobsEnvelope, targetsEnvelope] = await Promise.all([
      fetchPublishJobs(projectId, { status: status || undefined, target_id: targetID || undefined, page: nextPage, page_size: 20, sort: 'created_at', order: 'desc' }),
      fetchPublishTargets(projectId, { enabled: true, page: 1, page_size: 20 }),
    ]);
    if (!jobsEnvelope.success || !jobsEnvelope.data) {
      setError(pageErrorFromEnvelope(jobsEnvelope, '加载发布队列失败'));
      return;
    }
    if (!targetsEnvelope.success || !targetsEnvelope.data) {
      setError(pageErrorFromEnvelope(targetsEnvelope, '加载发布目标失败'));
      return;
    }
    setJobs(jobsEnvelope.data.items);
    setTargets(targetsEnvelope.data.items);
    setPage(nextPage);
    setHasNext(jobsEnvelope.data.page * jobsEnvelope.data.page_size < jobsEnvelope.data.total);
    setError(null);
  }

  useEffect(() => {
    void load(1);
  }, [projectId]);

  async function applyFilters() {
    await load(1);
  }

  async function changePage(nextPage: number) {
    await load(Math.max(1, nextPage));
  }

  async function createSampleJob() {
    const targetID = targets[0]?.id ?? 'publish-target-1';
    const envelope = await createPublishJob(projectId, { content_item_id: 'content-item-1', content_version_id: 'version-1', target_id: targetID }, `publish-job-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '创建发布任务失败'));
      return;
    }
    setNotice(`发布任务已入队：${envelope.data.publish_job_id} request_id=${envelope.request_id}`);
    await load();
  }

  return (
    <main className="page-shell">
      <section className="page-hero">
        <div className="page-hero__header">
          <div>
            <h1>发布队列</h1>
            <p>按目标和状态查看人工发布任务，复制载荷并回填发布结果。</p>
          </div>
          <button type="button" onClick={createSampleJob}>创建发布任务</button>
        </div>
      </section>
      <section className="card">
        <div className="toolbar">
          <label>状态筛选
            <select value={status} onChange={(event) => setStatus(event.target.value)}>
              <option value="">全部</option>
              <option value="queued">queued</option>
              <option value="copied">copied</option>
              <option value="published">published</option>
              <option value="failed">failed</option>
            </select>
          </label>
          <label>目标筛选
            <select value={targetID} onChange={(event) => setTargetID(event.target.value)}>
              <option value="">全部目标</option>
              {targets.map((target) => <option key={target.id} value={target.id}>{target.platform} · {target.display_name}</option>)}
            </select>
          </label>
          <button type="button" onClick={applyFilters}>筛选</button>
          <button type="button">新建发布目标</button>
          <span className="muted">目标数：{targets.length}</span>
        </div>
        {notice && <p role="status">{notice}</p>}
        {error && <p role="alert">{error.code} {error.message} request_id={error.request_id}</p>}
      </section>
      <section className="card table-card">
        <h2>任务列表</h2>
        {jobs.length === 0 ? <p className="muted">暂无发布任务</p> : (
          <>
            <table>
              <thead><tr><th>内容</th><th>目标</th><th>状态</th><th>最近错误</th><th>操作</th></tr></thead>
              <tbody>
                {jobs.map((job) => (
                  <tr key={job.id}>
                    <td>{job.title}<br /><span className="muted">{job.content_version_id}</span></td>
                    <td>{job.target_platform} · {job.target_display}</td>
                    <td><span className="badge badge--muted">{job.status}</span></td>
                    <td>{job.last_error || '-'}</td>
                    <td className="action-row">
                      <Link href={`/publish-jobs/${job.id}?project_id=${encodeURIComponent(projectId)}`}>详情</Link>
                      <Link href={`/publish-jobs/${job.id}/copy?project_id=${encodeURIComponent(projectId)}`}>复制</Link>
                      <Link href={`/publish-jobs/${job.id}/backfill?project_id=${encodeURIComponent(projectId)}`}>回填</Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            <div className="action-row">
              <button type="button" disabled={page <= 1} onClick={() => changePage(page - 1)}>上一页</button>
              <span className="muted">第 {page} 页</span>
              <button type="button" disabled={!hasNext} onClick={() => changePage(page + 1)}>下一页</button>
            </div>
          </>
        )}
      </section>
    </main>
  );
}
