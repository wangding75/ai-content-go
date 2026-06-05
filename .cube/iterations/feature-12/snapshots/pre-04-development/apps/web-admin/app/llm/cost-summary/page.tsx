'use client';

import { useEffect, useMemo, useState } from 'react';
import { LLMCostSummaryResponse, PageError, fetchLLMCostSummary, pageErrorFromEnvelope } from '../../../lib/api';

export default function LLMCostSummaryPage() {
  const [summary, setSummary] = useState<LLMCostSummaryResponse | null>(null);
  const [modelFilter, setModelFilter] = useState('all');
  const [selectedModel, setSelectedModel] = useState<string | null>(null);
  const [error, setError] = useState<PageError | null>(null);
  const [toast, setToast] = useState('');

  async function load() {
    const result = await fetchLLMCostSummary();
    if (result.success && result.data) {
      setSummary(result.data);
      setToast(`刷新完成：${result.data.calls} calls`);
      setError(null);
    } else {
      setError(pageErrorFromEnvelope(result, '加载成本汇总失败'));
    }
  }

  useEffect(() => {
    void load();
  }, []);

  const models = summary?.by_model.length ? summary.by_model : [{ model: 'no-model-data', calls: 0, tokens: 0, cost: 0 }];
  const filteredModels = useMemo(() => models.filter((item) => modelFilter === 'all' || item.model === modelFilter), [models, modelFilter]);
  const selected = models.find((item) => item.model === selectedModel) ?? null;

  return (
    <main className="page-shell" data-testid="styled-page-shell">
      <section className="page-hero">
        <div className="page-hero__header">
          <div>
            <h1>成本汇总</h1>
            <p>查看 LLM 调用、tokens、成本和模型维度聚合，支持刷新、筛选和详情跳转。</p>
          </div>
          <button type="button" onClick={load}>刷新汇总</button>
        </div>
      </section>

      {error ? <div role="alert">{error.code ? `${error.code}：` : ''}{error.message}（request_id: {error.request_id ?? 'client'}）</div> : null}
      {toast ? <div role="status">{toast}</div> : null}

      <section className="card-grid">
        <article className="card metric"><span className="muted">calls</span><strong>{summary?.calls ?? 0}</strong></article>
        <article className="card metric"><span className="muted">tokens</span><strong>{summary?.tokens ?? 0}</strong></article>
        <article className="card metric"><span className="muted">cost</span><strong>{summary?.cost ?? 0} {summary?.currency ?? 'USD'}</strong></article>
      </section>

      <section className="card">
        <div className="toolbar">
          <label>模型筛选
            <select value={modelFilter} onChange={(event) => setModelFilter(event.target.value)}>
              <option value="all">全部模型</option>
              {models.map((item) => <option key={item.model} value={item.model}>{item.model}</option>)}
            </select>
          </label>
          <span className="badge badge--success">by_model</span>
        </div>
      </section>

      <section className="card table-card">
        <table>
          <thead><tr><th>Model</th><th>calls</th><th>tokens</th><th>cost</th><th>操作</th></tr></thead>
          <tbody>{filteredModels.map((item) => <tr key={item.model}><td>{item.model}</td><td>{item.calls}</td><td>{item.tokens}</td><td>{item.cost}</td><td><button type="button" onClick={() => setSelectedModel(item.model)}>查看详情</button></td></tr>)}</tbody>
        </table>
      </section>

      {selected ? (
        <section className="card" data-testid="cost-detail">
          <div className="card__header"><h2>模型详情：{selected.model}</h2><button type="button" onClick={() => setSelectedModel(null)}>关闭详情</button></div>
          <p>调用次数：{selected.calls}</p>
          <p>Token：{selected.tokens}</p>
          <p>成本：{selected.cost} {summary?.currency ?? 'USD'}</p>
        </section>
      ) : null}
    </main>
  );
}
