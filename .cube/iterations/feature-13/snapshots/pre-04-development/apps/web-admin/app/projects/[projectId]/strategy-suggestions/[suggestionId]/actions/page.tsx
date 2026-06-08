'use client';

import { use, useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { fetchStrategySuggestion, confirmSuggestion, ignoreSuggestion, executeSuggestion, retrySuggestion, pageErrorFromEnvelope, type StrategySuggestionDetail, type PageError } from '../../../../../../lib/api';

export default function SuggestionActionsPage({ params }: { params: Promise<{ projectId: string; suggestionId: string }> }) {
  const { projectId, suggestionId } = use(params);
  const router = useRouter();
  const [detail, setDetail] = useState<StrategySuggestionDetail | null>(null);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');
  const [ignoreReason, setIgnoreReason] = useState('');
  const [ignoreNote, setIgnoreNote] = useState('');
  const [executeActionType, setExecuteActionType] = useState('');
  const [executeTargetType, setExecuteTargetType] = useState('');
  const [executeTargetID, setExecuteTargetID] = useState('');
  const [executeNote, setExecuteNote] = useState('');
  const [retryNote, setRetryNote] = useState('');

  async function load() {
    const envelope = await fetchStrategySuggestion(suggestionId);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '加载建议失败'));
      return;
    }
    setDetail(envelope.data);
    setError(null);
  }

  useEffect(() => { load(); }, [suggestionId]);

  async function handleConfirm() {
    const idempotencyKey = `confirm-${suggestionId}-${Date.now()}`;
    const envelope = await confirmSuggestion(suggestionId, {}, idempotencyKey);
    if (!envelope.success) {
      setError(pageErrorFromEnvelope(envelope, '确认失败'));
      return;
    }
    setNotice(`已确认 operation_log_id=${envelope.data?.operation_log_id}`);
    load();
  }

  async function handleIgnore() {
    if (!ignoreReason.trim()) {
      setError({ message: '忽略原因为必填项' });
      return;
    }
    const idempotencyKey = `ignore-${suggestionId}-${Date.now()}`;
    const envelope = await ignoreSuggestion(suggestionId, { reason: ignoreReason, note: ignoreNote }, idempotencyKey);
    if (!envelope.success) {
      setError(pageErrorFromEnvelope(envelope, '忽略失败'));
      return;
    }
    setNotice(`已忽略 operation_log_id=${envelope.data?.operation_log_id}`);
    setIgnoreReason('');
    setIgnoreNote('');
    load();
  }

  async function handleExecute() {
    if (!executeActionType || !executeTargetType || !executeTargetID) {
      setError({ message: '动作类型、目标类型和目标 ID 为必填项' });
      return;
    }
    const idempotencyKey = `execute-${suggestionId}-${Date.now()}`;
    const envelope = await executeSuggestion(suggestionId, { action_type: executeActionType, target_type: executeTargetType, target_id: executeTargetID, operator_note: executeNote }, idempotencyKey);
    if (!envelope.success) {
      setError(pageErrorFromEnvelope(envelope, '执行失败'));
      return;
    }
    setNotice(`执行完成 execution_log_id=${envelope.data?.execution_log_id} 状态=${envelope.data?.current_status}`);
    setExecuteActionType('');
    setExecuteTargetType('');
    setExecuteTargetID('');
    setExecuteNote('');
    load();
  }

  async function handleRetry() {
    const idempotencyKey = `retry-${suggestionId}-${Date.now()}`;
    const envelope = await retrySuggestion(suggestionId, { operator_note: retryNote }, idempotencyKey);
    if (!envelope.success) {
      setError(pageErrorFromEnvelope(envelope, '重试失败'));
      return;
    }
    setNotice(`重试完成 execution_log_id=${envelope.data?.execution_log_id} 状态=${envelope.data?.current_status}`);
    setRetryNote('');
    load();
  }

  if (!detail) return <div className="page-loading">加载中...</div>;

  const canConfirm = detail.status === 'pending';
  const canIgnore = detail.status === 'pending';
  const canExecute = detail.status === 'confirmed';
  const canRetry = detail.status === 'execution_failed';

  return (
    <div className="page-container">
      <div className="page-header">
        <a href={`/projects/${projectId}/strategy-suggestions/${suggestionId}`} className="back-link">返回详情</a>
        <h1>策略建议操作</h1>
        <span className={`status-badge status-${detail.status}`}>{detail.status}</span>
      </div>

      {notice && <div className="notice">{notice}</div>}
      {error && (
        <div className="error-banner">
          <p>{error.message}</p>
          {error.request_id && <p className="error-request-id">request_id: {error.request_id}</p>}
        </div>
      )}

      <div className="actions-grid">
        <div className="action-card" data-disabled={!canConfirm}>
          <h3>确认建议</h3>
          <p>确认表示人工认可此建议，仅改变建议状态</p>
          <button onClick={handleConfirm} disabled={!canConfirm} className="btn btn-primary">确认</button>
        </div>

        <div className="action-card" data-disabled={!canIgnore}>
          <h3>忽略建议</h3>
          <div className="form-group">
            <label>忽略原因（必填）</label>
            <input value={ignoreReason} onChange={(e) => setIgnoreReason(e.target.value)} disabled={!canIgnore} className="form-input" />
          </div>
          <div className="form-group">
            <label>备注</label>
            <input value={ignoreNote} onChange={(e) => setIgnoreNote(e.target.value)} disabled={!canIgnore} className="form-input" />
          </div>
          <button onClick={handleIgnore} disabled={!canIgnore} className="btn btn-secondary">忽略</button>
        </div>

        <div className="action-card" data-disabled={!canExecute}>
          <h3>执行建议</h3>
          <div className="form-group">
            <label>动作类型（必填）</label>
            <input value={executeActionType} onChange={(e) => setExecuteActionType(e.target.value)} disabled={!canExecute} className="form-input" />
          </div>
          <div className="form-group">
            <label>目标类型（必填）</label>
            <input value={executeTargetType} onChange={(e) => setExecuteTargetType(e.target.value)} disabled={!canExecute} className="form-input" />
          </div>
          <div className="form-group">
            <label>目标 ID（必填）</label>
            <input value={executeTargetID} onChange={(e) => setExecuteTargetID(e.target.value)} disabled={!canExecute} className="form-input" />
          </div>
          <div className="form-group">
            <label>操作备注</label>
            <input value={executeNote} onChange={(e) => setExecuteNote(e.target.value)} disabled={!canExecute} className="form-input" />
          </div>
          <button onClick={handleExecute} disabled={!canExecute} className="btn btn-primary">执行</button>
        </div>

        <div className="action-card" data-disabled={!canRetry}>
          <h3>重试执行</h3>
          <div className="form-group">
            <label>操作备注</label>
            <input value={retryNote} onChange={(e) => setRetryNote(e.target.value)} disabled={!canRetry} className="form-input" />
          </div>
          <button onClick={handleRetry} disabled={!canRetry} className="btn btn-primary">重试</button>
        </div>
      </div>
    </div>
  );
}
