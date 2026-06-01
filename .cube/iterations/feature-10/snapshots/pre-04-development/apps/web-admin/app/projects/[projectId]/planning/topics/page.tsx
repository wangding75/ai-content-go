'use client';

import { use, useEffect, useState } from 'react';
import ProjectWorkspaceNav from '../../workspace-nav';
import { PageError, TopicCandidateResponse, confirmTopic, fetchPlanningRun, fetchPlanningRuns, pageErrorFromEnvelope } from '@/lib/api';

export default function TopicsPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [topics, setTopics] = useState<TopicCandidateResponse[]>([]);
  const [note, setNote] = useState('直达确认');
  const [error, setError] = useState<PageError | null>(null);
  const [toast, setToast] = useState('');
  const [loading, setLoading] = useState(true);

  async function load() {
    setLoading(true);
    const runs = await fetchPlanningRuns(projectId, { sort: 'created_at', order: 'desc' });
    if (!runs.success || !runs.data) {
      setError(pageErrorFromEnvelope(runs, '加载规划运行失败'));
      setLoading(false);
      return;
    }
    const run = runs.data.items[0];
    if (!run) {
      setTopics([]);
      setLoading(false);
      return;
    }
    const detail = await fetchPlanningRun(projectId, run.id);
    if (detail.success && detail.data) {
      setTopics(detail.data.topics);
      setError(null);
    } else {
      setError(pageErrorFromEnvelope(detail, '加载候选选题失败'));
    }
    setLoading(false);
  }

  useEffect(() => {
    void load();
  }, [projectId]);

  async function submit(topicID: string) {
    const result = await confirmTopic(projectId, topicID, { note }, `topic-${topicID}-${Date.now()}`);
    if (result.success) {
      setToast(`确认成功：${result.data?.operation_log_id}`);
      await load();
    } else {
      setError(pageErrorFromEnvelope(result, '确认选题失败'));
    }
  }

  return (
    <main className="page-shell">
      <ProjectWorkspaceNav projectId={projectId} />
      <section className="page-hero"><h1>候选选题确认</h1><p>作为弹窗流程的直达与刷新恢复入口。</p></section>
      {error ? <div role="alert">{error.code ? `${error.code}：` : ''}{error.message}（request_id: {error.request_id ?? 'client'}）</div> : null}
      {toast ? <div role="status">{toast}</div> : null}
      <section className="card"><label>确认备注<input value={note} onChange={(event) => setNote(event.target.value)} /></label></section>
      <section className="card-grid">{loading ? <p>加载态</p> : topics.length === 0 ? <p>空状态：暂无候选选题</p> : topics.map((topic) => <article className="card" key={topic.candidate_id}><h2>{topic.title}</h2><p>{topic.logline}</p><p>{topic.reason}</p><button type="button" onClick={() => submit(topic.candidate_id)}>确认</button></article>)}</section>
    </main>
  );
}
