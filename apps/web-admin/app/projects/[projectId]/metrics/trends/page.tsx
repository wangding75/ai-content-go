'use client';

import { use, useEffect, useState } from 'react';
import { fetchMetricTrends, pageErrorFromEnvelope, type MetricTrendResponse, type PageError } from '../../../../../lib/api';

export default function MetricTrendsPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [trend, setTrend] = useState<MetricTrendResponse | null>(null);
  const [bucket, setBucket] = useState('day');
  const [error, setError] = useState<PageError | null>(null);

  async function load() {
    const envelope = await fetchMetricTrends(projectId, { metric_code: 'views', date_from: '2026-05-01', date_to: '2026-05-25', bucket });
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '加载趋势失败'));
      return;
    }
    setTrend(envelope.data);
    setError(null);
  }

  useEffect(() => {
    void load();
  }, [projectId]);

  return (
    <main className="page-shell">
      <section className="page-hero">
        <div className="page-hero__header">
          <div><h1>趋势图</h1><p>按日、周、月查看指标趋势，缺失点不会按 0 展示。</p></div>
          <div className="action-row"><select value={bucket} onChange={(event) => setBucket(event.target.value)}><option value="day">day</option><option value="week">week</option><option value="month">month</option></select><button type="button" onClick={load}>查询</button></div>
        </div>
      </section>
      {error && <section className="card" role="alert">{error.code} {error.message} request_id={error.request_id}</section>}
      <section className="card">
        <h2>{trend?.metric_code ?? 'views'}</h2>
        <p>聚合：{trend?.aggregation_method ?? '-'} · 来源记录：{trend?.source_record_count ?? 0}</p>
        <p className="muted">query_signature={trend?.query_signature ?? '-'}</p>
      </section>
      <section className="card table-card">
        <h2>序列</h2>
        {!trend || trend.series.length === 0 ? <p className="muted">暂无趋势数据</p> : <table><thead><tr><th>桶</th><th>值</th><th>来源</th><th>缺失</th></tr></thead><tbody>{trend.series.map((point) => <tr key={point.bucket_start}><td>{point.bucket_start}</td><td>{point.value}</td><td>{point.source_record_count}</td><td>{point.missing ? '是' : '否'}</td></tr>)}</tbody></table>}
      </section>
      <section className="card">
        <h2>缺失点</h2>
        <p>{trend?.missing_points.map((point) => `${point.metric_date}:${point.reason}`).join(' / ') || '无缺失点'}</p>
      </section>
    </main>
  );
}
