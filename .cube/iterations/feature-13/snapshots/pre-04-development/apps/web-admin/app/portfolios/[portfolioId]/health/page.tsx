'use client';

import { useEffect, useState } from 'react';
import { fetchPortfolioCostSummary, fetchPortfolioHealthSummary, fetchPortfolioStatusSnapshots, pageErrorFromEnvelope, type PageError, type PortfolioCostSummaryResponse, type PortfolioHealthSummaryResponse, type PortfolioStatusSnapshotResponse } from '../../../../lib/api';

export default function PortfolioHealthPage({ params }: { params: { portfolioId: string } }) {
  const [health, setHealth] = useState<PortfolioHealthSummaryResponse | null>(null);
  const [cost, setCost] = useState<PortfolioCostSummaryResponse | null>(null);
  const [snapshots, setSnapshots] = useState<PortfolioStatusSnapshotResponse[]>([]);
  const [error, setError] = useState<PageError | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const healthEnvelope = await fetchPortfolioHealthSummary(params.portfolioId);
        if (cancelled) return;
        if (!healthEnvelope.success || !healthEnvelope.data) {
          setError(pageErrorFromEnvelope(healthEnvelope, '加载健康汇总失败'));
          return;
        }
        setHealth(healthEnvelope.data);
        const [costEnvelope, snapshotsEnvelope] = await Promise.all([
          fetchPortfolioCostSummary(params.portfolioId),
          fetchPortfolioStatusSnapshots(params.portfolioId),
        ]);
        if (!cancelled && costEnvelope.success && costEnvelope.data) setCost(costEnvelope.data);
        if (!cancelled && snapshotsEnvelope.success && snapshotsEnvelope.data) setSnapshots(snapshotsEnvelope.data.items);
      } catch {
        if (!cancelled) setError({ message: '加载健康汇总失败' });
      }
    }
    load();
    return () => {
      cancelled = true;
    };
  }, [params.portfolioId]);

  return (
    <main className="page-shell" data-testid="styled-page-shell">
      <header className="page-header">
        <div>
          <p className="eyebrow">Portfolio Health</p>
          <h1>健康 / 成本汇总</h1>
        </div>
        <button className="primary-button" type="button">重新计算</button>
      </header>
      {error ? <p role="alert">{error.message}{error.request_id ? ` (${error.request_id})` : ''}</p> : null}
      <section className="card-grid">
        <article className="summary-card"><h2>健康状态</h2><p>{health?.health_status ?? 'pending'}</p></article>
        <article className="summary-card"><h2>健康分</h2><p>{health?.health_score ?? 0}</p></article>
        <article className="summary-card"><h2>月成本</h2><p>{cost?.estimated_monthly_cost ?? 0} {cost?.currency ?? 'CNY'}</p></article>
      </section>
      <section>
        <h2>快照历史</h2>
        {snapshots.length === 0 ? <p>暂无快照。</p> : null}
        <ul>
          {snapshots.map((snapshot) => <li key={snapshot.id}>{snapshot.health_status} / {snapshot.calculation_status}</li>)}
        </ul>
      </section>
    </main>
  );
}
