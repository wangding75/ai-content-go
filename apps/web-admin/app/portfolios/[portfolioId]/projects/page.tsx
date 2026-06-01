'use client';

import { useEffect, useState } from 'react';
import { fetchPortfolioProjects, pageErrorFromEnvelope, type PageError, type PortfolioProjectResponse } from '../../../../lib/api';

export default function PortfolioProjectsPage({ params }: { params: { portfolioId: string } }) {
  const [items, setItems] = useState<PortfolioProjectResponse[]>([]);
  const [error, setError] = useState<PageError | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const envelope = await fetchPortfolioProjects(params.portfolioId);
        if (cancelled) return;
        if (!envelope.success || !envelope.data) {
          setError(pageErrorFromEnvelope(envelope, '加载 Portfolio 项目失败'));
          return;
        }
        setItems(envelope.data.items);
      } catch {
        if (!cancelled) setError({ message: '加载 Portfolio 项目失败' });
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
          <p className="eyebrow">Portfolio Projects</p>
          <h1>项目优先级</h1>
        </div>
        <button className="primary-button" type="button">加入项目</button>
      </header>
      {error ? <p role="alert">{error.message}{error.request_id ? ` (${error.request_id})` : ''}</p> : null}
      {items.length === 0 && !error ? <p>暂无项目。</p> : null}
      <table>
        <thead><tr><th>项目</th><th>角色</th><th>优先级</th><th>权重</th></tr></thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.project_id}><td>{item.project_name || item.project_id}</td><td>{item.role}</td><td>{item.priority}</td><td>{item.weight}</td></tr>
          ))}
        </tbody>
      </table>
    </main>
  );
}
