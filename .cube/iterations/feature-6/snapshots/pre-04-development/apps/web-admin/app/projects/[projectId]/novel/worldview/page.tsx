'use client';

import { use, useEffect, useState } from 'react';
import ProjectWorkspaceNav from '../../workspace-nav';
import { PageError, fetchWorldview, pageErrorFromEnvelope, updateWorldview } from '@/lib/api';

export default function WorldviewPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [worldview, setWorldview] = useState('{}');
  const [forbiddenRules, setForbiddenRules] = useState('');
  const [note, setNote] = useState('更新世界观');
  const [version, setVersion] = useState('');
  const [error, setError] = useState<PageError | null>(null);
  const [toast, setToast] = useState('');

  async function load() {
    const result = await fetchWorldview(projectId);
    if (result.success && result.data) {
      setWorldview(JSON.stringify(result.data.worldview, null, 2));
      setForbiddenRules(result.data.forbidden_rules.join('\n'));
      setVersion(`${result.data.version_id} / v${result.data.version}`);
      setError(null);
    } else {
      setError(pageErrorFromEnvelope(result, '加载世界观失败'));
    }
  }

  useEffect(() => {
    void load();
  }, [projectId]);

  async function submit() {
    let parsed: Record<string, unknown>;
    try {
      parsed = JSON.parse(worldview) as Record<string, unknown>;
    } catch {
      setError({ message: '世界观 JSON 格式错误', request_id: 'client' });
      return;
    }
    const result = await updateWorldview(projectId, { worldview: parsed, forbidden_rules: forbiddenRules.split('\n').filter(Boolean), note });
    if (result.success) {
      setToast(`保存成功：${result.data?.operation_log_id}`);
      await load();
    } else {
      setError(pageErrorFromEnvelope(result, '保存世界观失败'));
    }
  }

  return (
    <main className="page-shell">
      <ProjectWorkspaceNav projectId={projectId} />
      <section className="page-hero"><h1>世界观</h1><p>查看并编辑 Novel Pack 世界观与禁用规则。</p></section>
      {error ? <div role="alert">{error.code ? `${error.code}：` : ''}{error.message}（request_id: {error.request_id ?? 'client'}）</div> : null}
      {toast ? <div role="status">{toast}</div> : null}
      <section className="card form-grid"><label>版本<input readOnly value={version} /></label><label>备注<input value={note} onChange={(event) => setNote(event.target.value)} /></label><label>世界观 JSON<textarea value={worldview} onChange={(event) => setWorldview(event.target.value)} /></label><label>Forbidden Rules<textarea value={forbiddenRules} onChange={(event) => setForbiddenRules(event.target.value)} /></label><button type="button" onClick={submit}>保存世界观</button></section>
    </main>
  );
}
