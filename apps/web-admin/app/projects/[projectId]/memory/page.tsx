'use client';

import { use, useEffect, useState } from 'react';
import { correctDynamicState, fetchKnowledgeMemory, fetchMemorySnapshots, pageErrorFromEnvelope, updateRecentWindowPolicy, updateStaticContext, updateStyleGuide, type KnowledgeMemoryResponse, type MemorySnapshotResponse, type PageError } from '../../../../lib/api';

export default function MemoryPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [memory, setMemory] = useState<KnowledgeMemoryResponse | null>(null);
  const [snapshots, setSnapshots] = useState<MemorySnapshotResponse[]>([]);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');

  async function load() {
    const [memoryEnvelope, snapshotEnvelope] = await Promise.all([fetchKnowledgeMemory(projectId), fetchMemorySnapshots(projectId)]);
    if (!memoryEnvelope.success || !memoryEnvelope.data) {
      setError(pageErrorFromEnvelope(memoryEnvelope, '加载记忆上下文失败'));
      return;
    }
    setMemory(memoryEnvelope.data);
    if (snapshotEnvelope.success && snapshotEnvelope.data) setSnapshots(snapshotEnvelope.data.items);
    setError(null);
  }

  useEffect(() => { void load(); }, [projectId]);

  async function submitStaticContext() {
    const envelope = await updateStaticContext(projectId, { static_context: { summary: 'updated static context' }, note: 'manual update' });
    if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '修正 StaticContext 失败'));
    setNotice(`StaticContext 已更新 version=${envelope.data.version} operation_log_id=${envelope.data.operation_log_id}`);
    await load();
  }

  async function submitStyleGuide() {
    const envelope = await updateStyleGuide(projectId, { style_guide: { tone: 'consistent' }, note: 'manual update' });
    if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '修正 StyleGuide 失败'));
    setNotice(`StyleGuide 已更新 version=${envelope.data.version} operation_log_id=${envelope.data.operation_log_id}`);
    await load();
  }

  async function submitDynamicCorrection() {
    const envelope = await correctDynamicState(projectId, { reason: '人工纠偏', changes: { status: 'corrected' }, source_refs: ['item_001'] }, `memory-correction-${Date.now()}`);
    if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '纠偏 DynamicState 失败'));
    setNotice(`DynamicState 已纠偏 snapshot=${envelope.data.memory_snapshot_id} operation_log_id=${envelope.data.operation_log_id}`);
    await load();
  }

  async function submitPolicy() {
    const envelope = await updateRecentWindowPolicy(projectId, { item_count: 5, token_limit: 2000, truncation_policy: 'time' });
    if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '更新 RecentContentWindow 失败'));
    setNotice(`RecentContentWindow 已更新 version=${envelope.data.version} operation_log_id=${envelope.data.operation_log_id}`);
    await load();
  }

  return (
    <main className="page-shell">
      <section className="page-hero"><h1>记忆上下文</h1><p>维护项目级 StaticContext、DynamicState、RecentContentWindow 和 StyleGuide。</p></section>
      {notice && <section className="card" role="status">{notice}</section>}
      {error && <section className="card" role="alert">{error.code} {error.message} request_id={error.request_id}</section>}
      {!memory ? <section className="card">加载中...</section> : <>
        <section className="card"><h2>上下文概览</h2><p>version={memory.version} updated_at={memory.updated_at}</p><pre>{JSON.stringify({ static_context: memory.static_context, dynamic_state: memory.dynamic_state, style_guide: memory.style_guide }, null, 2)}</pre></section>
        <section className="card"><h2>RecentContentWindow</h2><p>{memory.recent_window_policy.item_count} items · {memory.recent_window_policy.token_limit} tokens · {memory.recent_window_policy.truncation_policy}</p><button type="button" onClick={() => void submitPolicy()}>保存最小策略</button></section>
        <section className="card"><h2>人工修正</h2><button type="button" onClick={() => void submitStaticContext()}>修正 StaticContext</button> <button type="button" onClick={() => void submitStyleGuide()}>修正 StyleGuide</button> <button type="button" onClick={() => void submitDynamicCorrection()}>纠偏 DynamicState</button></section>
        <section className="card"><h2>记忆快照</h2>{snapshots.length === 0 ? <p>暂无记忆快照</p> : <ul>{snapshots.map(snapshot => <li key={snapshot.id}>{snapshot.source_type} · budget={snapshot.token_budget} · estimated={snapshot.estimated_tokens} · {snapshot.truncation_policy} · {snapshot.created_at}</li>)}</ul>}</section>
      </>}
    </main>
  );
}
