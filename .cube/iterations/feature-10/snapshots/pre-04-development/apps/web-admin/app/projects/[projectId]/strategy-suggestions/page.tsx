'use client';

import Link from 'next/link';
import { use, useEffect, useState } from 'react';
import { fetchStrategySuggestions, generateStrategySuggestions, pageErrorFromEnvelope, type StrategySuggestionItem, type PageError } from '../../../../lib/api';

export default function StrategySuggestionsPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [items, setItems] = useState<StrategySuggestionItem[]>([]);
  const [error, setError] = useState<PageError | null>(null);
  const [loading, setLoading] = useState(true);
  const [notice, setNotice] = useState('');

  async function load() {
    setLoading(true);
    const envelope = await fetchStrategySuggestions(projectId);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '加载策略建议失败'));
      setLoading(false);
      return;
    }
    setItems(envelope.data.items);
    setError(null);
    setLoading(false);
  }

  async function handleGenerate() {
    setNotice('正在生成策略建议...');
    const envelope = await generateStrategySuggestions(projectId, {
      date_from: '2026-05-01',
      date_to: '2026-05-25',
    });
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '生成策略建议失败'));
      return;
    }
    setNotice(`建议生成任务已提交 run_id=${envelope.data.suggestion_run_id}`);
    setTimeout(load, 2000);
  }

  useEffect(() => { load(); }, [projectId]);

  if (loading) return <div className="page-loading">加载中...</div>;
  if (error) return (
    <div className="page-error">
      <p className="error-message">{error.message}</p>
      {error.code && <p className="error-code">错误码: {error.code}</p>}
      {error.request_id && <p className="error-request-id">request_id: {error.request_id}</p>}
      <button onClick={load} className="btn btn-secondary">重试</button>
    </div>
  );

  return (
    <div className="page-container">
      <div className="page-header">
        <h1>策略建议</h1>
        <button onClick={handleGenerate} className="btn btn-primary">生成建议</button>
      </div>
      {notice && <div className="notice">{notice}</div>}
      {items.length === 0 ? (
        <div className="empty-state">
          <p>暂无策略建议，点击"生成建议"开始分析</p>
        </div>
      ) : (
        <div className="card-list">
          {items.map((item) => (
            <Link key={item.id} href={`/projects/${projectId}/strategy-suggestions/${item.id}`} className="card card-link">
              <div className="card-header">
                <span className={`status-badge status-${item.status}`}>{item.status}</span>
                <span className={`type-badge type-${item.suggestion_type}`}>{item.suggestion_type}</span>
              </div>
              <h3 className="card-title">{item.title}</h3>
              <p className="card-text">{item.trigger_reason}</p>
              <div className="card-meta">
                <span>风险: {item.risk_level}</span>
                <span>置信度: {item.confidence}</span>
                <span>{item.date_from} ~ {item.date_to}</span>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
