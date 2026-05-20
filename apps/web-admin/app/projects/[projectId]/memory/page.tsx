'use client';

import { use, useEffect, useRef, useState } from 'react';
import ProjectWorkspaceNav from '../workspace-nav';
import { correctDynamicState, fetchKnowledgeMemory, fetchMemorySnapshots, pageErrorFromEnvelope, updateRecentWindowPolicy, updateStaticContext, updateStyleGuide, type KnowledgeMemoryResponse, type MemorySnapshotResponse, type PageError } from '../../../../lib/api';

export default function MemoryPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [memory, setMemory] = useState<KnowledgeMemoryResponse | null>(null);
  const [snapshots, setSnapshots] = useState<MemorySnapshotResponse[]>([]);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');
  const [loading, setLoading] = useState(true);
  const [filterContentItemIDInput, setFilterContentItemIDInput] = useState('');
  const [filterContentItemID, setFilterContentItemID] = useState('');
  const [sortField, setSortField] = useState('created_at');
  const [sortOrder, setSortOrder] = useState('desc');
  const [snapshotPage, setSnapshotPage] = useState(1);
  const [snapshotHasNext, setSnapshotHasNext] = useState(false);
  const [correctionReason, setCorrectionReason] = useState('');
  const [correctionChanges, setCorrectionChanges] = useState('');
  const [correctionSourceRefs, setCorrectionSourceRefs] = useState('');
  const [staticContextInput, setStaticContextInput] = useState('');
  const [staticContextNote, setStaticContextNote] = useState('');
  const [styleGuideInput, setStyleGuideInput] = useState('');
  const [styleGuideNote, setStyleGuideNote] = useState('');
  const [policyItemCount, setPolicyItemCount] = useState('5');
  const [policyTokenLimit, setPolicyTokenLimit] = useState('2000');
  const [policyTruncationPolicy, setPolicyTruncationPolicy] = useState('time');
  const [policyNote, setPolicyNote] = useState('');
  const loadSequence = useRef(0);

  async function load(page = snapshotPage) {
    const sequence = ++loadSequence.current;
    setLoading(true);
    setMemory(null);
    setSnapshots([]);
    setSnapshotHasNext(false);
    const [memoryResult, snapshotResult] = await Promise.allSettled([
      fetchKnowledgeMemory(projectId),
      fetchMemorySnapshots(projectId, { page, content_item_id: filterContentItemID || undefined, sort: sortField, order: sortOrder }),
    ]);
    if (sequence !== loadSequence.current) return;
    if (memoryResult.status === 'rejected') {
      setError({ code: 'NETWORK_ERROR', message: '加载记忆上下文失败' });
      setLoading(false);
      return;
    }
    const memoryEnvelope = memoryResult.value;
    if (!memoryEnvelope.success || !memoryEnvelope.data) {
      setError(pageErrorFromEnvelope(memoryEnvelope, '加载记忆上下文失败'));
      setLoading(false);
      return;
    }
    setMemory(memoryEnvelope.data);
    setStaticContextInput(JSON.stringify(memoryEnvelope.data.static_context, null, 2));
    setStyleGuideInput(JSON.stringify(memoryEnvelope.data.style_guide, null, 2));
    setPolicyItemCount(String(memoryEnvelope.data.recent_window_policy.item_count));
    setPolicyTokenLimit(String(memoryEnvelope.data.recent_window_policy.token_limit));
    setPolicyTruncationPolicy(memoryEnvelope.data.recent_window_policy.truncation_policy);
    if (snapshotResult.status === 'rejected') {
      setError({ code: 'NETWORK_ERROR', message: '加载记忆快照失败' });
      setLoading(false);
      return;
    }
    const snapshotEnvelope = snapshotResult.value;
    if (!snapshotEnvelope.success || !snapshotEnvelope.data) {
      setError(pageErrorFromEnvelope(snapshotEnvelope, '加载记忆快照失败'));
      setLoading(false);
      return;
    }
    setSnapshots(snapshotEnvelope.data.items);
    setSnapshotHasNext(snapshotEnvelope.data.pagination.has_next);
    setError(null);
    setLoading(false);
  }


  useEffect(() => { void load(); }, [projectId, snapshotPage, filterContentItemID, sortField, sortOrder]);

  function applySnapshotFilter() {
    setSnapshotPage(1);
    setFilterContentItemID(filterContentItemIDInput);
  }

  async function submitStaticContext() {
    let staticContext: Record<string, unknown>;
    try {
      const parsed = JSON.parse(staticContextInput);
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) throw new Error('invalid object');
      staticContext = parsed as Record<string, unknown>;
    } catch {
      setError({ code: 'VALIDATION_ERROR', message: 'StaticContext 必须是合法 JSON 对象' });
      return;
    }
    try {
      const envelope = await updateStaticContext(projectId, { static_context: staticContext, note: staticContextNote || 'manual update' });
      if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '修正 StaticContext 失败'));
      setNotice(`StaticContext 已更新 version=${envelope.data.version} operation_log_id=${envelope.data.operation_log_id}`);
      await load();
    } catch {
      setError({ code: 'NETWORK_ERROR', message: '修正 StaticContext 失败' });
    }
  }

  async function submitStyleGuide() {
    let styleGuide: Record<string, unknown>;
    try {
      const parsed = JSON.parse(styleGuideInput);
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) throw new Error('invalid object');
      styleGuide = parsed as Record<string, unknown>;
    } catch {
      setError({ code: 'VALIDATION_ERROR', message: 'StyleGuide 必须是合法 JSON 对象' });
      return;
    }
    try {
      const envelope = await updateStyleGuide(projectId, { style_guide: styleGuide, note: styleGuideNote || 'manual update' });
      if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '修正 StyleGuide 失败'));
      setNotice(`StyleGuide 已更新 version=${envelope.data.version} operation_log_id=${envelope.data.operation_log_id}`);
      await load();
    } catch {
      setError({ code: 'NETWORK_ERROR', message: '修正 StyleGuide 失败' });
    }
  }

  async function submitDynamicCorrection() {
    if (!correctionReason || !correctionChanges || !correctionSourceRefs) return;
    let changes: Record<string, unknown>;
    try {
      const parsed = JSON.parse(correctionChanges);
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) throw new Error('invalid object');
      changes = parsed as Record<string, unknown>;
    } catch {
      setError({ code: 'VALIDATION_ERROR', message: '纠偏内容必须是合法 JSON 对象' });
      return;
    }
    const sourceRefs = correctionSourceRefs.split(',').map(ref => ref.trim()).filter(Boolean);
    if (sourceRefs.length === 0) return;
    try {
      const envelope = await correctDynamicState(projectId, { reason: correctionReason, changes, source_refs: sourceRefs }, `memory-correction-${Date.now()}`);
      if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '纠偏 DynamicState 失败'));
      setNotice(`DynamicState 已纠偏 snapshot=${envelope.data.memory_snapshot_id} operation_log_id=${envelope.data.operation_log_id}`);
      setCorrectionReason(''); setCorrectionChanges(''); setCorrectionSourceRefs('');
      await load();
    } catch {
      setError({ code: 'NETWORK_ERROR', message: '纠偏 DynamicState 失败' });
    }
  }

  async function submitPolicy() {
    const itemCount = Number(policyItemCount);
    const tokenLimit = Number(policyTokenLimit);
    if (!Number.isFinite(itemCount) || !Number.isFinite(tokenLimit)) {
      setError({ code: 'VALIDATION_ERROR', message: 'item_count 和 token_limit 必须是数字' });
      return;
    }
    try {
      const envelope = await updateRecentWindowPolicy(projectId, { item_count: itemCount, token_limit: tokenLimit, truncation_policy: policyTruncationPolicy, note: policyNote || undefined });
      if (!envelope.success || !envelope.data) return setError(pageErrorFromEnvelope(envelope, '更新 RecentContentWindow 失败'));
      setNotice(`RecentContentWindow 已更新 version=${envelope.data.version} operation_log_id=${envelope.data.operation_log_id}`);
      await load();
    } catch {
      setError({ code: 'NETWORK_ERROR', message: '更新 RecentContentWindow 失败' });
    }
  }

  return (
    <main className="page-shell">
      <ProjectWorkspaceNav projectId={projectId} />
      <section className="page-hero"><h1>记忆上下文</h1><p>维护项目级上下文、动态状态、窗口策略和风格指南。</p></section>
      {notice && <section className="card" role="status">{notice}</section>}
      {error && <section className="card" role="alert">错误码: {error.code} 错误信息: {error.message} request_id={error.request_id}</section>}
      {loading ? <section className="card" role="status">加载中</section> : !memory ? null : <>
        <section className="card"><h2>上下文概览</h2><p>version={memory.version} updated_at={memory.updated_at}</p><pre>{JSON.stringify({ StaticContext: memory.static_context, DynamicState: memory.dynamic_state, StyleGuide: memory.style_guide }, null, 2)}</pre>{memory.recent_snapshot_summary && <p>最近快照摘要: id={memory.recent_snapshot_summary.id} source={memory.recent_snapshot_summary.source_type} tokens={memory.recent_snapshot_summary.estimated_tokens} policy={memory.recent_snapshot_summary.truncation_policy}</p>}</section>
        <section className="card"><h2>RecentContentWindow</h2><p>{memory.recent_window_policy.item_count} items · {memory.recent_window_policy.token_limit} tokens · {memory.recent_window_policy.truncation_policy}</p>
          <div><label>item_count <input type="number" value={policyItemCount} onChange={e => setPolicyItemCount(e.target.value)} /></label><label>token_limit <input type="number" value={policyTokenLimit} onChange={e => setPolicyTokenLimit(e.target.value)} /></label><label>truncation_policy <select value={policyTruncationPolicy} onChange={e => setPolicyTruncationPolicy(e.target.value)}><option value="time">time</option><option value="token">token</option></select></label><label>note <input value={policyNote} onChange={e => setPolicyNote(e.target.value)} /></label><button type="button" onClick={() => void submitPolicy()}>保存最小策略</button></div></section>
        <section className="card"><h2>人工修正</h2>
          <div><h3>修正 StaticContext</h3><textarea value={staticContextInput} onChange={e => setStaticContextInput(e.target.value)} rows={4} cols={60} /><label>note <input value={staticContextNote} onChange={e => setStaticContextNote(e.target.value)} /></label><button type="button" onClick={() => void submitStaticContext()}>修正 StaticContext</button></div>
          <div><h3>修正 StyleGuide</h3><textarea value={styleGuideInput} onChange={e => setStyleGuideInput(e.target.value)} rows={4} cols={60} /><label>note <input value={styleGuideNote} onChange={e => setStyleGuideNote(e.target.value)} /></label><button type="button" onClick={() => void submitStyleGuide()}>修正 StyleGuide</button></div>
          <div><h3>纠偏 DynamicState</h3><label>纠偏原因 <input value={correctionReason} onChange={e => setCorrectionReason(e.target.value)} /></label><label>纠偏内容 <textarea value={correctionChanges} onChange={e => setCorrectionChanges(e.target.value)} rows={2} cols={40} /></label><label>来源引用 <input value={correctionSourceRefs} onChange={e => setCorrectionSourceRefs(e.target.value)} /></label><button type="button" onClick={() => void submitDynamicCorrection()}>纠偏 DynamicState</button></div>
        </section>
        <section className="card"><h2>记忆快照</h2>
          <div><label>内容单元筛选 <input value={filterContentItemIDInput} onChange={e => setFilterContentItemIDInput(e.target.value)} /></label><label>排序字段 <select value={sortField} onChange={e => { setSortField(e.target.value); setSnapshotPage(1); }}><option value="created_at">created_at</option></select></label><label>排序方向 <select value={sortOrder} onChange={e => { setSortOrder(e.target.value); setSnapshotPage(1); }}><option value="desc">desc</option><option value="asc">asc</option></select></label><button type="button" onClick={applySnapshotFilter}>筛选快照</button></div>
          {snapshots.length === 0 ? <p>暂无记忆快照</p> : <ul>{snapshots.map(snapshot => <li key={snapshot.id}>来源: {snapshot.source_type} · Token 预算: {snapshot.token_budget} · 预估 Token: {snapshot.estimated_tokens} · 截断策略: {snapshot.truncation_policy} · 创建时间: {snapshot.created_at}</li>)}</ul>}
          <div><button type="button" disabled={snapshotPage <= 1} onClick={() => setSnapshotPage(p => p - 1)}>上一页</button> <span>页码: {snapshotPage}</span> <button type="button" onClick={() => setSnapshotPage(p => p + 1)}>下一页</button> {!snapshotHasNext && <span>已到最后一页</span>}</div>
        </section>
      </>}
    </main>
  );
}
