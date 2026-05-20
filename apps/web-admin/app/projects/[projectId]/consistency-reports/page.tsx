'use client';

import Link from 'next/link';
import { use, useEffect, useRef, useState } from 'react';
import ProjectWorkspaceNav from '../workspace-nav';
import { createConsistencyReport, fetchConsistencyReports, pageErrorFromEnvelope, type ConsistencyReportResponse, type PageError } from '../../../../lib/api';

export default function ConsistencyReportsPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [reports, setReports] = useState<ConsistencyReportResponse[]>([]);
  const [notice, setNotice] = useState('');
  const [error, setError] = useState<PageError | null>(null);
  const [statusFilterInput, setStatusFilterInput] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [sortField, setSortField] = useState('created_at');
  const [sortOrder, setSortOrder] = useState('desc');
  const [page, setPage] = useState(1);
  const [hasNext, setHasNext] = useState(false);
  const [reportRangeInput, setReportRangeInput] = useState('{"latest": true}');
  const [reportScope, setReportScope] = useState('project');
  const [severityThreshold, setSeverityThreshold] = useState('low');
  const [reportScopeInput, setReportScopeInput] = useState('project');
  const loadSequence = useRef(0);

  async function load(currentPage = page) {
    const sequence = ++loadSequence.current;
    setReports([]);
    setHasNext(false);
    try {
      const envelope = await fetchConsistencyReports(projectId, { status: statusFilter || undefined, page: currentPage, sort: sortField, order: sortOrder });
      if (sequence !== loadSequence.current) return;
      if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '加载一致性报告失败'));
      setReports(envelope.data.items);
      setHasNext(envelope.data.pagination.has_next);
      setError(null);
    } catch {
      if (sequence !== loadSequence.current) return;
      setError({ code: 'NETWORK_ERROR', message: '加载一致性报告失败' });
    }
  }

  useEffect(() => { void load(); }, [projectId, statusFilter, page, sortField, sortOrder]);

  function applyReportFilter() {
    setPage(1);
    setStatusFilter(statusFilterInput);
  }

  async function submitReport() {
    let range: Record<string, unknown>;
    try {
      const parsed = JSON.parse(reportRangeInput);
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) throw new Error('invalid object');
      range = parsed as Record<string, unknown>;
    } catch {
      setError({ code: 'VALIDATION_ERROR', message: '报告范围必须是合法 JSON 对象' });
      return;
    }
    try {
      const envelope = await createConsistencyReport(projectId, { range, scope: reportScope, severity_threshold: severityThreshold }, `consistency-report-${Date.now()}`);
      if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '创建一致性报告失败'));
      setNotice(`报告已创建 report=${envelope.data.report_id} status=${envelope.data.status}`);
      await load();
    } catch {
      setError({ code: 'NETWORK_ERROR', message: '创建一致性报告失败' });
    }
  }

  return (
    <main className="page-shell">
      <ProjectWorkspaceNav projectId={projectId} />
      <section className="page-hero"><h1>一致性报告</h1><p>创建并查看项目级一致性检查报告。</p></section>
      {notice && <section className="card" role="status">{notice}</section>}
      {error && <section className="card" role="alert">错误码: {error.code} 错误信息: {error.message} request_id={error.request_id}</section>}
      <section className="card"><h2>操作</h2>
        <div><label>状态筛选 <select value={statusFilterInput} onChange={e => setStatusFilterInput(e.target.value)}><option value="">全部</option><option value="pending">pending</option><option value="running">running</option><option value="completed">completed</option><option value="failed">failed</option></select></label><label>排序字段 <select value={sortField} onChange={e => { setSortField(e.target.value); setPage(1); }}><option value="created_at">created_at</option></select></label><label>排序方向 <select value={sortOrder} onChange={e => { setSortOrder(e.target.value); setPage(1); }}><option value="desc">desc</option><option value="asc">asc</option></select></label><button type="button" onClick={applyReportFilter}>筛选报告</button></div>
        <button type="button" onClick={() => void submitReport()}>创建一致性报告</button>
        <div><label>reportRange <textarea value={reportRangeInput} onChange={e => setReportRangeInput(e.target.value)} rows={2} cols={40} /></label><label>scope <select value={reportScopeInput} onChange={e => { setReportScopeInput(e.target.value); setReportScope(e.target.value); }}><option value="project">project</option><option value="content_unit">content_unit</option></select></label><label>severity_threshold <select value={severityThreshold} onChange={e => setSeverityThreshold(e.target.value)}><option value="low">low</option><option value="medium">medium</option><option value="high">high</option></select></label></div>
      </section>
      <section className="card"><h2>报告列表</h2>
        {reports.length === 0 ? <p>暂无一致性报告</p> : <ul>{reports.map(report => <li key={report.id}><Link href={`/projects/${encodeURIComponent(projectId)}/consistency-reports/${encodeURIComponent(report.id)}`}>查看详情</Link> · 报告状态: {report.status} · 问题数量: {report.issue_count} · 严重度: {JSON.stringify(report.severity_summary)} · {report.created_at}</li>)}</ul>}
        <div><button type="button" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>上一页</button> <span>页码: {page}</span> <button type="button" onClick={() => setPage(p => p + 1)}>下一页</button> {!hasNext && <span>已到最后一页</span>}</div>
      </section>
    </main>
  );
}
