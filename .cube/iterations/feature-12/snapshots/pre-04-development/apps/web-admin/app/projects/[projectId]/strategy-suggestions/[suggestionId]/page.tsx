'use client';

import Link from 'next/link';
import { use, useEffect, useState } from 'react';
import { fetchStrategySuggestion, fetchExecutionLogs, pageErrorFromEnvelope, type StrategySuggestionDetail, type ExecutionLogItem, type PageError } from '../../../../../lib/api';

export default function SuggestionDetailPage({ params }: { params: Promise<{ projectId: string; suggestionId: string }> }) {
  const { projectId, suggestionId } = use(params);
  const [detail, setDetail] = useState<StrategySuggestionDetail | null>(null);
  const [logs, setLogs] = useState<ExecutionLogItem[]>([]);
  const [error, setError] = useState<PageError | null>(null);
  const [loading, setLoading] = useState(true);

  async function load() {
    setLoading(true);
    const [detailEnvelope, logsEnvelope] = await Promise.all([
      fetchStrategySuggestion(suggestionId),
      fetchExecutionLogs(suggestionId),
    ]);
    if (!detailEnvelope.success || !detailEnvelope.data) {
      setError(pageErrorFromEnvelope(detailEnvelope, '加载建议详情失败'));
      setLoading(false);
      return;
    }
    setDetail(detailEnvelope.data);
    if (logsEnvelope.success && logsEnvelope.data) {
      setLogs(logsEnvelope.data.items);
    }
    setError(null);
    setLoading(false);
  }

  useEffect(() => { load(); }, [suggestionId]);

  if (loading) return <div className="page-loading">加载中...</div>;
  if (error) return (
    <div className="page-error">
      <p className="error-message">{error.message}</p>
      {error.code && <p className="error-code">错误码: {error.code}</p>}
      {error.request_id && <p className="error-request-id">request_id: {error.request_id}</p>}
      <button onClick={load} className="btn btn-secondary">重试</button>
    </div>
  );
  if (!detail) return null;

  return (
    <div className="page-container">
      <div className="page-header">
        <Link href={`/projects/${projectId}/strategy-suggestions`} className="back-link">返回列表</Link>
        <h1>{detail.title}</h1>
        <div className="header-badges">
          <span className={`status-badge status-${detail.status}`}>{detail.status}</span>
          <span className={`type-badge type-${detail.suggestion_type}`}>{detail.suggestion_type}</span>
        </div>
        {(detail.status === 'pending' || detail.status === 'confirmed' || detail.status === 'execution_failed') && (
          <Link href={`/projects/${projectId}/strategy-suggestions/${suggestionId}/actions`} className="btn btn-primary">操作</Link>
        )}
      </div>

      <div className="detail-grid">
        <div className="detail-card">
          <h3>触发原因</h3>
          <p>{detail.trigger_reason}</p>
        </div>
        <div className="detail-card">
          <h3>影响范围</h3>
          <p>{detail.impact_scope}</p>
        </div>
        <div className="detail-card">
          <h3>风险等级</h3>
          <p>{detail.risk_level}</p>
        </div>
        <div className="detail-card">
          <h3>置信度</h3>
          <p>{detail.confidence}</p>
        </div>
        <div className="detail-card">
          <h3>建议动作</h3>
          <p>{detail.suggested_action}</p>
        </div>
        <div className="detail-card">
          <h3>预期收益</h3>
          <p>{detail.expected_benefit}</p>
        </div>
      </div>

      <div className="detail-card">
        <h3>证据指标</h3>
        <table className="data-table">
          <thead><tr><th>指标</th><th>数值</th><th>趋势</th></tr></thead>
          <tbody>
            {detail.evidence_metrics.map((m, i) => (
              <tr key={i}><td>{m.metric_code}</td><td>{m.value}</td><td>{m.trend}</td></tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="detail-card">
        <h3>指标快照</h3>
        <p>snapshot_id: {detail.metrics_snapshot.summary_snapshot_id || '无'}</p>
      </div>

      {detail.ignored_reason && (
        <div className="detail-card">
          <h3>忽略原因</h3>
          <p>{detail.ignored_reason}</p>
          {detail.ignored_note && <p className="note">{detail.ignored_note}</p>}
        </div>
      )}

      {logs.length > 0 && (
        <div className="detail-card">
          <h3>执行记录</h3>
          <table className="data-table">
            <thead><tr><th>动作</th><th>目标</th><th>结果</th><th>时间</th></tr></thead>
            <tbody>
              {logs.map((log) => (
                <tr key={log.id}>
                  <td>{log.action_type}</td>
                  <td>{log.target_type}/{log.target_id}</td>
                  <td>{log.result}{log.failure_reason ? `: ${log.failure_reason}` : ''}</td>
                  <td>{log.created_at}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
