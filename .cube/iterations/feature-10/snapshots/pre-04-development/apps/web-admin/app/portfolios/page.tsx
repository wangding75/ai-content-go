'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { fetchPortfolios, pageErrorFromEnvelope, type PageError, type PortfolioDetailResponse } from '../../lib/api';

export default function PortfoliosPage() {
  const [items, setItems] = useState<PortfolioDetailResponse[]>([]);
  const [error, setError] = useState<PageError | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const envelope = await fetchPortfolios();
        if (cancelled) return;
        if (!envelope.success || !envelope.data) {
          setError(pageErrorFromEnvelope(envelope, '加载 Portfolio 列表失败'));
          return;
        }
        setItems(envelope.data.items);
      } catch {
        if (!cancelled) setError({ message: '加载 Portfolio 列表失败' });
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    load();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <main className="page-shell" data-testid="styled-page-shell">
      <header className="page-header">
        <div>
          <p className="eyebrow">Iteration 10</p>
          <h1>Portfolio 管理</h1>
        </div>
        <button className="primary-button" type="button">新建组合</button>
      </header>
      {loading ? <p role="status">加载中...</p> : null}
      {error ? <p role="alert">{error.message}{error.request_id ? ` (${error.request_id})` : ''}</p> : null}
      {!loading && !error && items.length === 0 ? <p>暂无 Portfolio。</p> : null}
      <section className="card-grid">
        {items.map((item) => (
          <article className="summary-card" key={item.id}>
            <h2>{item.name}</h2>
            <p>{item.description}</p>
            <p>项目数：{item.project_count}</p>
            <p>健康度：{item.latest_health_score}</p>
            <p>成本：{item.estimated_monthly_cost} {item.currency}</p>
            <Link href={`/portfolios/${item.id}`}>查看详情</Link>
          </article>
        ))}
      </section>
    </main>
  );
}
