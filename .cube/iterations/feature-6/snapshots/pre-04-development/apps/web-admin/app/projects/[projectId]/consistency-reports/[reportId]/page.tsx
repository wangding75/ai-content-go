'use client';

import { use, useEffect, useRef, useState } from 'react';
import { fetchConsistencyReport, pageErrorFromEnvelope, type ConsistencyReportDetailResponse, type PageError } from '../../../../../lib/api';

export default function ConsistencyReportDetailPage({ params }: { params: Promise<{ projectId: string; reportId: string }> }) {
  const { projectId, reportId } = use(params);
  const [report, setReport] = useState<ConsistencyReportDetailResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);
  const [loading, setLoading] = useState(true);
  const loadSequence = useRef(0);

  useEffect(() => {
    async function load() {
      const sequence = ++loadSequence.current;
      setLoading(true);
      setReport(null);
      try {
        const envelope = await fetchConsistencyReport(projectId, reportId);
        if (sequence !== loadSequence.current) return;
        if (!envelope.success || !envelope.data) {
          setError(pageErrorFromEnvelope(envelope, '加载一致性报告详情失败'));
          return;
        }
        setReport(envelope.data);
        setError(null);
      } catch {
        if (sequence !== loadSequence.current) return;
        setError({ code: 'NETWORK_ERROR', message: '加载一致性报告详情失败' });
      } finally {
        if (sequence === loadSequence.current) setLoading(false);
      }
    }
    void load();
  }, [projectId, reportId]);

  return (
    <main className="page-shell">
      <section className="page-hero"><h1>一致性报告详情</h1><p>查看结构化一致性问题和关联快照信息。</p></section>
      {error && <section className="card" role="alert">错误码: {error.code} 错误信息: {error.message} request_id={error.request_id}</section>}
      {loading ? <section className="card" role="status">加载中</section> : !report ? null : <>
        <section className="card"><h2>报告概览</h2><p>来源快照: {report.source_snapshot_id}</p><p>报告状态: {report.status} · 问题数量: {report.issue_count}</p><div>严重度摘要: <pre>{JSON.stringify(report.severity_summary, null, 2)}</pre></div>{report.error_code && <p>失败原因: 错误码={report.error_code} 错误信息={report.error_message}</p>}</section>
        <section className="card"><h2>结构化 Issues</h2>
          {report.issues.length === 0 ? <p>暂无问题</p> : <ul>{report.issues.map(issue => <li key={issue.issue_id}><p>问题编号: {issue.issue_id} · 严重度: {issue.severity} · 问题类型: {issue.type}</p><p>问题标题: {issue.title}</p><p>问题描述: {issue.description}</p><p>受影响内容: {issue.affected_content_items.join(', ')}</p><p>修复建议: {issue.suggestion}</p></li>)}</ul>}
        </section>
      </>}
    </main>
  );
}
