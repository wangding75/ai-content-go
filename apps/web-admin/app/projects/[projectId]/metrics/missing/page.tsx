'use client';

import Link from 'next/link';
import { use, useEffect, useState } from 'react';
import { fetchMissingMetricDates, pageErrorFromEnvelope, type MissingMetricDatesResponse, type PageError } from '../../../../../lib/api';

export default function MissingMetricsPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [missing, setMissing] = useState<MissingMetricDatesResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);

  async function load() {
    const envelope = await fetchMissingMetricDates(projectId, { date_from: '2026-05-01', date_to: '2026-05-25' });
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '加载缺失提醒失败'));
      return;
    }
    setMissing(envelope.data);
    setError(null);
  }

  useEffect(() => {
    void load();
  }, [projectId]);

  return (
    <main className="page-shell">
      <section className="page-hero">
        <div className="page-hero__header">
          <div><h1>缺失数据提醒</h1><p>只检查已发布、目标有效、模板启用且必填的指标。</p></div>
          <button type="button" onClick={load}>刷新</button>
        </div>
      </section>
      {error && <section className="card" role="alert">{error.code} {error.message} request_id={error.request_id}</section>}
      <section className="card table-card">
        <h2>待补录日期</h2>
        {!missing || missing.items.length === 0 ? <p className="muted">暂无缺失提醒</p> : (
          <table>
            <thead><tr><th>指标</th><th>平台</th><th>周期</th><th>日期</th><th>原因</th><th>操作</th></tr></thead>
            <tbody>{missing.items.map((item) => {
              const query = new URLSearchParams({ platform: item.platform, target_id: item.target_id, metric_code: item.metric_code, period: item.period, metric_date: item.metric_date }).toString();
              return <tr key={`${item.publish_job_id}-${item.metric_code}-${item.metric_date}`}><td>{item.metric_code}</td><td>{item.platform}</td><td>{item.period}</td><td>{item.metric_date}</td><td>{item.missing_reason}</td><td><Link href={`/projects/${projectId}/metrics/input?${query}`}>补录</Link></td></tr>;
            })}</tbody>
          </table>
        )}
      </section>
    </main>
  );
}
