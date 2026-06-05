'use client';

import Link from 'next/link';
import { use, useEffect, useState } from 'react';
import { fetchMetricRecords, fetchMetricSummary, fetchMissingMetricDates, pageErrorFromEnvelope, type MetricRecordResponse, type MetricSummaryResponse, type MissingMetricDatesResponse, type PageError } from '../../../../lib/api';

export default function MetricsPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [summary, setSummary] = useState<MetricSummaryResponse | null>(null);
  const [records, setRecords] = useState<MetricRecordResponse[]>([]);
  const [missing, setMissing] = useState<MissingMetricDatesResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');

  async function load() {
    const [summaryEnvelope, recordsEnvelope, missingEnvelope] = await Promise.all([
      fetchMetricSummary(projectId, { date_from: '2026-05-01', date_to: '2026-05-25', metric_codes: ['views', 'likes'] }),
      fetchMetricRecords({ project_id: projectId, page: 1, page_size: 10, sort: 'metric_date', order: 'desc' }),
      fetchMissingMetricDates(projectId, { date_from: '2026-05-01', date_to: '2026-05-25' }),
    ]);
    if (!summaryEnvelope.success || !summaryEnvelope.data) {
      setError(pageErrorFromEnvelope(summaryEnvelope, '加载指标汇总失败'));
      return;
    }
    if (!recordsEnvelope.success || !recordsEnvelope.data) {
      setError(pageErrorFromEnvelope(recordsEnvelope, '加载指标记录失败'));
      return;
    }
    if (!missingEnvelope.success || !missingEnvelope.data) {
      setError(pageErrorFromEnvelope(missingEnvelope, '加载缺失提醒失败'));
      return;
    }
    setSummary(summaryEnvelope.data);
    setRecords(recordsEnvelope.data.items);
    setMissing(missingEnvelope.data);
    setError(null);
    setNotice(`指标数据已刷新 request_id=${summaryEnvelope.request_id}`);
  }

  useEffect(() => {
    void load();
  }, [projectId]);

  return (
    <main className="page-shell">
      <section className="page-hero">
        <div className="page-hero__header">
          <div>
            <h1>指标表现</h1>
            <p>查看项目指标汇总、稳定快照、趋势入口和缺失数据提醒。</p>
          </div>
          <div className="action-row">
            <Link href={`/projects/${projectId}/metrics/input`}>录入指标</Link>
            <button type="button" onClick={load}>刷新</button>
          </div>
        </div>
      </section>
      {notice && <p role="status">{notice}</p>}
      {error && <section className="card" role="alert">{error.code} {error.message} request_id={error.request_id}</section>}
      <section className="card-grid">
        {(summary?.items ?? []).map((item) => (
          <article className="card" key={item.metric_code}>
            <h2>{item.metric_code}</h2>
            <p>{item.value} {item.unit}</p>
            <span className="badge badge--muted">{item.aggregation_method} · {item.source_record_count} 条来源</span>
          </article>
        ))}
        {summary && summary.items.length === 0 && <article className="card"><h2>暂无指标</h2><p>当前筛选范围内还没有可汇总的数据。</p></article>}
      </section>
      <section className="card">
        <div className="card__header">
          <h2>稳定快照</h2>
          <span className="badge badge--muted">summary_snapshot_id={summary?.summary_snapshot_id ?? '未生成'}</span>
        </div>
        <p>source_record_count={summary?.source_record_count ?? 0}</p>
      </section>
      <section className="card table-card">
        <div className="card__header">
          <h2>最近记录</h2>
          <Link href={`/projects/${projectId}/metrics/trends`}>趋势图</Link>
        </div>
        {records.length === 0 ? <p className="muted">暂无指标记录</p> : (
          <table>
            <thead><tr><th>指标</th><th>平台</th><th>日期</th><th>原始值</th><th>来源</th></tr></thead>
            <tbody>{records.map((record) => <tr key={record.id}><td>{record.metric_code}</td><td>{record.platform}</td><td>{record.metric_date}</td><td>{record.raw_value}</td><td>{record.source_type}</td></tr>)}</tbody>
          </table>
        )}
      </section>
      <section className="card">
        <div className="card__header">
          <h2>缺失提醒</h2>
          <Link href={`/projects/${projectId}/metrics/missing`}>缺失提醒</Link>
        </div>
        <p>待补录：{missing?.items.length ?? 0}</p>
      </section>
    </main>
  );
}
