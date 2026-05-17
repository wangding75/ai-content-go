'use client';

import { use, useEffect, useState } from 'react';
import ProjectWorkspaceNav from '../../workspace-nav';
import { ArcResponse, PageError, fetchArcs, pageErrorFromEnvelope } from '@/lib/api';

export default function ArcsPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [arcs, setArcs] = useState<ArcResponse[]>([]);
  const [order, setOrder] = useState('asc');
  const [error, setError] = useState<PageError | null>(null);
  const [loading, setLoading] = useState(true);

  async function load() {
    setLoading(true);
    const result = await fetchArcs(projectId, { sort: 'order_index', order });
    if (result.success && result.data) {
      setArcs(result.data.items);
      setError(null);
    } else {
      setError(pageErrorFromEnvelope(result, '加载大纲失败'));
    }
    setLoading(false);
  }

  useEffect(() => {
    void load();
  }, [projectId, order]);

  return (
    <main className="page-shell">
      <ProjectWorkspaceNav projectId={projectId} />
      <section className="page-hero"><div className="page-hero__header"><div><h1>大纲管理</h1><p>查看 Novel Pack 弧线大纲、排序与规划来源。</p></div><label>排序<select value={order} onChange={(event) => setOrder(event.target.value)}><option value="asc">升序</option><option value="desc">降序</option></select></label></div></section>
      {error ? <div role="alert">{error.code ? `${error.code}：` : ''}{error.message}（request_id: {error.request_id ?? 'client'}）</div> : null}
      <section className="card table-card">{loading ? <p>加载态</p> : arcs.length === 0 ? <p>空状态：暂无大纲</p> : <table><thead><tr><th>顺序</th><th>标题</th><th>摘要</th><th>来源</th></tr></thead><tbody>{arcs.map((arc) => <tr key={arc.arc_id}><td>{arc.order_index}</td><td>{arc.title}</td><td>{arc.summary}</td><td>{arc.planning_run_id ?? 'manual'}</td></tr>)}</tbody></table>}</section>
    </main>
  );
}
