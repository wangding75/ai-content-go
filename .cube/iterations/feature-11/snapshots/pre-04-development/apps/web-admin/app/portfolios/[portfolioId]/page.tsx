'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { fetchPortfolio, fetchPortfolioStrategySummary, pageErrorFromEnvelope, type PageError, type PortfolioDetailResponse, type PortfolioStrategySummaryResponse } from '../../../lib/api';

export default function PortfolioDetailPage({ params }: { params: { portfolioId: string } }) {
  const [portfolio, setPortfolio] = useState<PortfolioDetailResponse | null>(null);
  const [strategy, setStrategy] = useState<PortfolioStrategySummaryResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const portfolioEnvelope = await fetchPortfolio(params.portfolioId);
        if (cancelled) return;
        if (!portfolioEnvelope.success || !portfolioEnvelope.data) {
          setError(pageErrorFromEnvelope(portfolioEnvelope, '加载 Portfolio 详情失败'));
          return;
        }
        setPortfolio(portfolioEnvelope.data);
        const strategyEnvelope = await fetchPortfolioStrategySummary(params.portfolioId);
        if (!cancelled && strategyEnvelope.success && strategyEnvelope.data) {
          setStrategy(strategyEnvelope.data);
        }
      } catch {
        if (!cancelled) setError({ message: '加载 Portfolio 详情失败' });
      }
    }
    load();
    return () => {
      cancelled = true;
    };
  }, [params.portfolioId]);

  return (
    <main className="page-shell" data-testid="styled-page-shell">
      {error ? <p role="alert">{error.message}{error.request_id ? ` (${error.request_id})` : ''}</p> : null}
      {!portfolio && !error ? <p role="status">加载中...</p> : null}
      {portfolio ? (
        <>
          <header className="page-header">
            <div>
              <p className="eyebrow">Portfolio</p>
              <h1>{portfolio.name}</h1>
              <p>{portfolio.description}</p>
            </div>
          </header>
          <section className="card-grid">
            <article className="summary-card"><h2>健康度</h2><p>{portfolio.latest_health_score}</p></article>
            <article className="summary-card"><h2>项目数</h2><p>{portfolio.project_count}</p></article>
            <article className="summary-card"><h2>策略建议</h2><p>{strategy?.pending ?? 0} 待处理</p></article>
          </section>
          <nav className="button-row">
            <Link href={`/portfolios/${portfolio.id}/projects`}>管理项目</Link>
            <Link href={`/portfolios/${portfolio.id}/health`}>健康 / 成本</Link>
          </nav>
        </>
      ) : null}
    </main>
  );
}
