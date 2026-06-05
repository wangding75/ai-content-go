'use client';

import { useEffect, useState } from 'react';
import { fetchArticlePackStatus, pageErrorFromEnvelope, registerArticlePack, type ArticlePackStatusResponse, type PageError } from '../../lib/api';

export default function ArticlePackPage() {
  const [status, setStatus] = useState<ArticlePackStatusResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');
  const [busy, setBusy] = useState(false);
  const defaultMetrics = Array.isArray(status?.default_metrics) ? status.default_metrics : [];

  async function load() {
    try {
      const envelope = await fetchArticlePackStatus();
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '加载 Article Pack 状态失败'));
        return;
      }
      setStatus({
        ...envelope.data,
        default_metrics: Array.isArray(envelope.data.default_metrics) ? envelope.data.default_metrics : [],
      });
      setError(null);
      setNotice(`状态已刷新 request_id=${envelope.request_id}`);
    } catch {
      setError({ message: '加载 Article Pack 状态失败' });
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function handleRegister() {
    setBusy(true);
    try {
      const envelope = await registerArticlePack(`article-pack-${Date.now()}`);
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '注册 Article Pack 失败'));
        return;
      }
      setNotice(`注册成功：${envelope.data.content_pack_id} request_id=${envelope.request_id}`);
      await load();
    } catch {
      setError({ message: '注册 Article Pack 失败' });
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="page-shell">
      <section className="page-hero">
        <div className="page-hero__header">
          <div>
            <h1>Article Pack 管理</h1>
            <p>查看 Article Pack 注册状态、默认工作流和默认指标，并支持重新注册。</p>
          </div>
          <div className="action-row">
            <button type="button" onClick={handleRegister} disabled={busy}>{busy ? '注册中…' : '注册 / 重新注册'}</button>
            <button type="button" onClick={load}>刷新</button>
          </div>
        </div>
      </section>
      {notice && <p role="status">{notice}</p>}
      {error && <section className="card" role="alert">{error.code} {error.message} request_id={error.request_id}</section>}
      <section className="card-grid">
        <article className="card">
          <h2>注册状态</h2>
          <p>{status?.registered ? '已注册' : '未注册'}</p>
          <span className="badge badge--muted">content_pack_id={status?.content_pack_id ?? '未生成'}</span>
        </article>
        <article className="card">
          <h2>内容类型</h2>
          <p>{status?.content_type?.name ?? 'Article Pack'}</p>
          <span className="badge badge--muted">code={status?.content_type?.code ?? 'article'}</span>
        </article>
        <article className="card">
          <h2>默认工作流</h2>
          <p>{status?.default_workflow_template?.name ?? '未注册'}</p>
          <span className="badge badge--muted">status={status?.default_workflow_template?.status ?? 'draft'}</span>
        </article>
      </section>
      <section className="card table-card">
        <div className="card__header">
          <h2>默认指标</h2>
          <span className="badge badge--muted">{defaultMetrics.length} 项</span>
        </div>
        {defaultMetrics.length > 0 ? (
          <table>
            <thead><tr><th>指标码</th><th>名称</th><th>单位</th></tr></thead>
            <tbody>{defaultMetrics.map((metric) => <tr key={metric.metric_code}><td>{metric.metric_code}</td><td>{metric.name}</td><td>{metric.unit}</td></tr>)}</tbody>
          </table>
        ) : <p className="muted">默认指标将在注册完成后显示。</p>}
      </section>
    </main>
  );
}
