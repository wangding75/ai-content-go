'use client';

import { use, useEffect, useState } from 'react';
import ProjectWorkspaceNav from '../workspace-nav';
import { PageError, PlanningRunResponse, TopicCandidateResponse, confirmTopic, createPlanningRun, fetchPlanningRun, fetchPlanningRuns, pageErrorFromEnvelope } from '@/lib/api';

export default function PlanningPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [runs, setRuns] = useState<PlanningRunResponse[]>([]);
  const [topics, setTopics] = useState<TopicCandidateResponse[]>([]);
  const [selectedTopic, setSelectedTopic] = useState<TopicCandidateResponse | null>(null);
  const [genre, setGenre] = useState('fantasy');
  const [audience, setAudience] = useState('young-adult');
  const [count, setCount] = useState('3');
  const [templateVersionID, setTemplateVersionID] = useState('wftv-novel-planning');
  const [note, setNote] = useState('确认作为新书选题');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<PageError | null>(null);
  const [toast, setToast] = useState('');

  async function load() {
    setLoading(true);
    const result = await fetchPlanningRuns(projectId);
    if (result.success && result.data) {
      setRuns(result.data.items);
      setError(null);
      const firstRun = result.data.items[0];
      if (firstRun) {
        const detail = await fetchPlanningRun(projectId, firstRun.id);
        if (detail.success && detail.data) {
          setTopics(detail.data.topics);
        }
      }
    } else {
      setError(pageErrorFromEnvelope(result, '加载规划运行失败'));
    }
    setLoading(false);
  }

  useEffect(() => {
    void load();
  }, [projectId]);

  async function submitPlanning() {
    const result = await createPlanningRun(projectId, { genre, audience, count: Number(count), template_version_id: templateVersionID, input_override: { content_pack: 'novel' } }, `planning-${projectId}-${Date.now()}`);
    if (result.success) {
      setToast(`规划已启动：${result.data?.workflow_run_id}`);
      await load();
    } else {
      setError(pageErrorFromEnvelope(result, '启动规划失败'));
    }
  }

  async function submitConfirm() {
    if (!selectedTopic) {
      return;
    }
    const result = await confirmTopic(projectId, selectedTopic.candidate_id, { note }, `topic-${selectedTopic.candidate_id}-${Date.now()}`);
    if (result.success) {
      setToast(`选题已确认：${result.data?.operation_log_id}`);
      setSelectedTopic(null);
      await load();
    } else {
      setError(pageErrorFromEnvelope(result, '确认选题失败'));
    }
  }

  return (
    <main className="page-shell">
      <ProjectWorkspaceNav projectId={projectId} />
      <section className="page-hero">
        <div className="page-hero__header"><div><h1>内容规划</h1><p>启动 Novel Pack 新书规划，查看运行状态并确认候选选题。</p></div><button type="button" onClick={submitPlanning}>启动规划</button></div>
      </section>
      {error ? <div role="alert">{error.code ? `${error.code}：` : ''}{error.message}（request_id: {error.request_id ?? 'client'}）</div> : null}
      {toast ? <div role="status">{toast}</div> : null}
      <section className="card">
        <div className="form-grid">
          <label>类型<input value={genre} onChange={(event) => setGenre(event.target.value)} /></label>
          <label>目标读者<input value={audience} onChange={(event) => setAudience(event.target.value)} /></label>
          <label>候选数量<input value={count} onChange={(event) => setCount(event.target.value)} /></label>
          <label>Template Version ID<input value={templateVersionID} onChange={(event) => setTemplateVersionID(event.target.value)} /></label>
        </div>
      </section>
      <section className="card table-card">
        {loading ? <p>加载态</p> : runs.length === 0 ? <p>空状态：暂无规划运行</p> : <table><thead><tr><th>Run</th><th>Workflow</th><th>状态</th><th>候选数</th></tr></thead><tbody>{runs.map((run) => <tr key={run.id}><td>{run.id}</td><td>{run.workflow_run_id}</td><td><span className="badge">{run.status}</span></td><td>{run.candidate_count}</td></tr>)}</tbody></table>}
      </section>
      <section className="card">
        <div className="card__header"><h2>候选选题</h2><a href={`/projects/${projectId}/planning/topics`}>打开直达确认页</a></div>
        {topics.length === 0 ? <p>空状态：暂无候选选题</p> : <div className="card-grid">{topics.map((topic) => <article className="card" key={topic.candidate_id}><h3>{topic.title}</h3><p>{topic.logline}</p><p>{topic.reason}</p><button type="button" onClick={() => setSelectedTopic(topic)}>确认选题</button></article>)}</div>}
      </section>
      {selectedTopic ? <div className="dialog-backdrop" role="presentation"><section className="dialog-panel" role="dialog" aria-modal="true" aria-label="候选选题确认"><div className="card__header"><h2>{selectedTopic.title}</h2><button type="button" onClick={() => setSelectedTopic(null)}>关闭弹窗</button></div><p>{selectedTopic.logline}</p><label>确认备注<input value={note} onChange={(event) => setNote(event.target.value)} /></label><button type="button" onClick={submitConfirm}>提交确认</button></section></div> : null}
    </main>
  );
}
