'use client';

import { use, useEffect, useState } from 'react';
import { fetchConsistencyReport, pageErrorFromEnvelope, type ConsistencyReportDetailResponse, type PageError } from '../../../../../lib/api';

export default function ConsistencyReportDetailPage({ params }: { params: Promise<{ projectId: string; reportId: string }> }) {
  const { projectId, reportId } = use(params);
  const [report, setReport] = useState<ConsistencyReportDetailResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);

  useEffect(() => {
    async function load() {
      const envelope = await fetchConsistencyReport(projectId, reportId);
      if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '加载一致性报告详情失败'));
      setReport(envelope.data);
      setError(null);
    }
    void load();
  }, [projectId, reportId]);

  return (
    <main className="page-shell">
      <section className="page-hero"><h1>一致性报告详情</h1><p>查看结构化一致性问题与来源快照。</p></section>
      {error && <section className="card" role="alert">{error.code} {error.message} request_id={error.request_id}</section>}
      {!report ? <section className="card">加载中...</section> : <>
        <section className="card"><h2>报告概览</h2><p>{report.id} · {report.status} · issues={report.issue_count}</p><p>source_snapshot_id={report.source_snapshot_id}</p><pre>{JSON.stringify(report.severity_summary, null, 2)}</pre></section>
        <section className="card"><h2>结构化 Issues</h2>{report.issues.length === 0 ? <p>暂无问题</p> : <ul>{report.issues.map(issue => <li key={issue.issue_id}><strong>{issue.severity}</strong> · {issue.type} · {issue.title}<p>{issue.description}</p><p>affected={issue.affected_content_items.join(', ')}</p><p>{issue.suggestion}</p></li>)}</ul>}</section>
      </>}
    </main>
  );
}
