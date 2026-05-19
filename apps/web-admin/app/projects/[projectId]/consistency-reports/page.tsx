'use client';

import Link from 'next/link';
import { use, useEffect, useState } from 'react';
import { createConsistencyReport, fetchConsistencyReports, pageErrorFromEnvelope, type ConsistencyReportResponse, type PageError } from '../../../../lib/api';

export default function ConsistencyReportsPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [reports, setReports] = useState<ConsistencyReportResponse[]>([]);
  const [notice, setNotice] = useState('');
  const [error, setError] = useState<PageError | null>(null);

  async function load() {
    const envelope = await fetchConsistencyReports(projectId);
    if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '加载一致性报告失败'));
    setReports(envelope.data.items);
    setError(null);
  }

  useEffect(() => { void load(); }, [projectId]);

  async function submitReport() {
    const envelope = await createConsistencyReport(projectId, { range: { latest: true }, scope: 'project', severity_threshold: 'low' }, `consistency-report-${Date.now()}`);
    if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '创建一致性报告失败'));
    setNotice(`一致性报告已创建 report=${envelope.data.report_id} status=${envelope.data.status}`);
    await load();
  }

  return (
    <main className="page-shell">
      <section className="page-hero"><h1>一致性报告</h1><p>创建并查看项目级一致性检查报告。</p></section>
      {notice && <section className="card" role="status">{notice}</section>}
      {error && <section className="card" role="alert">{error.code} {error.message} request_id={error.request_id}</section>}
      <section className="card"><h2>操作</h2><button type="button" onClick={() => void submitReport()}>创建一致性报告</button></section>
      <section className="card"><h2>报告列表</h2>{reports.length === 0 ? <p>暂无一致性报告</p> : <ul>{reports.map(report => <li key={report.id}><Link href={`/projects/${projectId}/consistency-reports/${report.id}`}>{report.id}</Link> · {report.status} · issues={report.issue_count} · {report.created_at}</li>)}</ul>}</section>
    </main>
  );
}
