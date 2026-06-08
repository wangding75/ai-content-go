'use client';

import { use, useEffect, useState } from 'react';
import { fetchProjectArticleMetrics, pageErrorFromEnvelope, updateProjectArticleMetrics, type ArticleProjectMetricResponse, type PageError } from '../../../../../lib/api';

interface ArticleMetricsPageProps {
  params: Promise<{ projectId: string }>;
}

const metricOptions = ['views', 'likes', 'shares', 'comments', 'avg_read_time'];

export default function ArticleMetricsPage({ params }: ArticleMetricsPageProps) {
  const { projectId } = use(params);
  const [items, setItems] = useState<ArticleProjectMetricResponse[]>([]);
  const [enabledCodes, setEnabledCodes] = useState<string[]>([]);
  const [wechatOverrides, setWechatOverrides] = useState('');
  const [note, setNote] = useState('');
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');

  async function load() {
    const envelope = await fetchProjectArticleMetrics(projectId);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '加载 Article 指标配置失败'));
      return;
    }
    setItems(envelope.data.items);
    setEnabledCodes(envelope.data.items.filter((item) => item.enabled).map((item) => item.metric_code));
    setError(null);
    setNotice(`指标配置已刷新 request_id=${envelope.request_id}`);
  }

  useEffect(() => {
    void load();
  }, [projectId]);

  async function handleSave() {
    const wechatCodes = wechatOverrides.split(',').map((item) => item.trim()).filter(Boolean);
    const envelope = await updateProjectArticleMetrics(projectId, {
      enabled_metric_codes: enabledCodes,
      platform_overrides: wechatCodes.length > 0 ? { wechat: wechatCodes } : undefined,
      note,
    }, `article-metrics-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '保存 Article 指标配置失败'));
      return;
    }
    setNotice(`指标配置已保存 version_id=${envelope.data.version_id} request_id=${envelope.request_id}`);
    await load();
  }

  function toggleMetric(metricCode: string) {
    setEnabledCodes((current) => current.includes(metricCode) ? current.filter((code) => code !== metricCode) : [...current, metricCode]);
  }

  return (
    <main className="page-shell">
      <section className="page-hero">
        <div className="page-hero__header">
          <div>
            <h1>Article 指标配置</h1>
            <p>查看指标模板、平台差异和项目启用状态，并支持启用或停用指标。</p>
          </div>
          <div className="action-row">
            <button type="button" onClick={load}>刷新</button>
            <button type="button" onClick={handleSave}>保存配置</button>
          </div>
        </div>
      </section>
      {notice && <p role="status">{notice}</p>}
      {error && <section className="card" role="alert">{error.code} {error.message} request_id={error.request_id}</section>}
      <section className="card">
        <div className="card__header">
          <h2>启用指标</h2>
          <span className="badge badge--muted">project_id={projectId}</span>
        </div>
        <div className="card-grid">
          {metricOptions.map((metricCode) => (
            <label key={metricCode}>
              <input type="checkbox" checked={enabledCodes.includes(metricCode)} onChange={() => toggleMetric(metricCode)} />
              {metricCode}
            </label>
          ))}
        </div>
      </section>
      <section className="card">
        <div className="card__header">
          <h2>平台差异</h2>
          <span className="badge badge--muted">wechat override</span>
        </div>
        <label>
          微信覆盖指标
          <input value={wechatOverrides} onChange={(event) => setWechatOverrides(event.target.value)} placeholder="views, likes" />
        </label>
        <label>
          备注
          <input value={note} onChange={(event) => setNote(event.target.value)} placeholder="记录平台差异说明" />
        </label>
      </section>
      <section className="card table-card">
        <div className="card__header">
          <h2>当前模板</h2>
          <span className="badge badge--muted">{items.length} 项</span>
        </div>
        {items.length === 0 ? <p className="muted">当前项目尚未启用指标。</p> : (
          <table>
            <thead><tr><th>指标码</th><th>名称</th><th>平台</th><th>单位</th><th>状态</th></tr></thead>
            <tbody>{items.map((item) => <tr key={item.metric_code}><td>{item.metric_code}</td><td>{item.name}</td><td>{item.platform}</td><td>{item.unit}</td><td>{item.enabled ? '启用' : '停用'}</td></tr>)}</tbody>
          </table>
        )}
      </section>
    </main>
  );
}
