'use client';

import { use, useEffect, useState } from 'react';
import { fetchAIReport, pageErrorFromEnvelope, triggerAIReport, type PageError, type ReviewReportResponse } from '../../../../lib/api';

export default function AIReportPage({ params }: { params: Promise<{ reviewId: string }> }) {
  const { reviewId } = use(params);
  const [report, setReport] = useState<ReviewReportResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');

  async function loadReport() {
    const envelope = await fetchAIReport(reviewId);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '加载 AI 质检报告失败'));
      return;
    }
    setReport(envelope.data);
    setError(null);
  }

  useEffect(() => { void loadReport(); }, [reviewId]);

  async function generateReport() {
    const envelope = await triggerAIReport(reviewId, { report_type: 'default', config: {} }, `ai-report-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '触发 AI 质检失败'));
      return;
    }
    setNotice(`质检已受理：${envelope.data.job_id || envelope.data.workflow_run_id} request_id=${envelope.request_id}`);
    await loadReport();
  }

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>AI 质检报告</h1>
        <p>查看风险等级、问题项和建议；报告缺失时可触发异步生成。</p>
      </section>
      <section className="card">
        <button type="button" onClick={() => void generateReport()}>生成质检报告</button>
        {notice && <p role="status">{notice}</p>}
        {error && <p role="alert">{error.code} {error.message} request_id={error.request_id}</p>}
      </section>
      {report ? <section className="card">
        <h2>{report.status} · {report.risk_level} · {report.quality_score}</h2>
        <h3>问题项</h3>
        <ul>{report.issues.map((issue) => <li key={issue.code}>{issue.severity} · {issue.message}</li>)}</ul>
        <h3>建议</h3>
        <ul>{report.suggestions.map((suggestion) => <li key={suggestion.code}>{suggestion.message}</li>)}</ul>
      </section> : <section className="card"><p>暂无质检报告</p></section>}
    </main>
  );
}
