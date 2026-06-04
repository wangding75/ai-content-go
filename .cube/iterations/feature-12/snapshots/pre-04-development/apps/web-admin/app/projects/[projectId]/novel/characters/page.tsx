'use client';

import { use, useEffect, useState } from 'react';
import ProjectWorkspaceNav from '../../workspace-nav';
import { CharacterResponse, PageError, createCharacter, fetchCharacters, pageErrorFromEnvelope } from '@/lib/api';

export default function CharactersPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [characters, setCharacters] = useState<CharacterResponse[]>([]);
  const [role, setRole] = useState('protagonist');
  const [name, setName] = useState('新角色');
  const [profile, setProfile] = useState('{"goal":"成长"}');
  const [note, setNote] = useState('新增人物');
  const [error, setError] = useState<PageError | null>(null);
  const [toast, setToast] = useState('');
  const [loading, setLoading] = useState(true);

  async function load() {
    setLoading(true);
    const result = await fetchCharacters(projectId, { role });
    if (result.success && result.data) {
      setCharacters(result.data.items);
      setError(null);
    } else {
      setError(pageErrorFromEnvelope(result, '加载人物失败'));
    }
    setLoading(false);
  }

  useEffect(() => {
    void load();
  }, [projectId, role]);

  async function submit() {
    let parsed: Record<string, unknown>;
    try {
      parsed = JSON.parse(profile) as Record<string, unknown>;
    } catch {
      setError({ message: '人物 Profile JSON 格式错误', request_id: 'client' });
      return;
    }
    const result = await createCharacter(projectId, { name, role, profile: parsed, note });
    if (result.success) {
      setToast(`创建成功：${result.data?.operation_log_id}`);
      await load();
    } else {
      setError(pageErrorFromEnvelope(result, '创建人物失败'));
    }
  }

  return (
    <main className="page-shell">
      <ProjectWorkspaceNav projectId={projectId} />
      <section className="page-hero"><h1>人物管理</h1><p>按角色筛选并新增 Novel Pack 人物设定。</p></section>
      {error ? <div role="alert">{error.code ? `${error.code}：` : ''}{error.message}（request_id: {error.request_id ?? 'client'}）</div> : null}
      {toast ? <div role="status">{toast}</div> : null}
      <section className="card form-grid"><label>角色筛选<input value={role} onChange={(event) => setRole(event.target.value)} /></label><label>名称<input value={name} onChange={(event) => setName(event.target.value)} /></label><label>Profile JSON<textarea value={profile} onChange={(event) => setProfile(event.target.value)} /></label><label>备注<input value={note} onChange={(event) => setNote(event.target.value)} /></label><button type="button" onClick={submit}>新增人物</button></section>
      <section className="card table-card">{loading ? <p>加载态</p> : characters.length === 0 ? <p>空状态：暂无人物</p> : <table><thead><tr><th>名称</th><th>角色</th><th>来源</th></tr></thead><tbody>{characters.map((item) => <tr key={item.character_id}><td>{item.name}</td><td><span className="badge">{item.role}</span></td><td>{item.planning_run_id ?? 'manual'}</td></tr>)}</tbody></table>}</section>
    </main>
  );
}
