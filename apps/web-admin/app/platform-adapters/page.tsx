'use client';

import { useState, useEffect } from 'react';
import { fetchPlatformAdapters, createPlatformAdapter, updatePlatformAdapter, fetchPlatformAdapter, type PlatformAdapterResponse, type PlatformAdapterDetailResponse } from '@/lib/api';

export default function PlatformAdaptersPage() {
  const [adapters, setAdapters] = useState<PlatformAdapterResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<PlatformAdapterDetailResponse | null>(null);
  const [showCreate, setShowCreate] = useState(false);

  useEffect(() => {
    fetchPlatformAdapters().then((res) => {
      if (res.success && res.data) {
        setAdapters(res.data.items);
      } else {
        setError(res.error?.message ?? '请求失败');
      }
      setLoading(false);
    });
  }, []);

  const handleCreate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = new FormData(e.currentTarget);
    const res = await createPlatformAdapter({
      platform: form.get('platform') as string,
      display_name: form.get('display_name') as string,
      publish_mode: form.get('publish_mode') as string,
      target_type: form.get('target_type') as string,
      config: {},
    });
    if (res.success && res.data) {
      setShowCreate(false);
      const list = await fetchPlatformAdapters();
      if (list.success && list.data) setAdapters(list.data.items);
    } else {
      setError(res.error?.message ?? '创建失败');
    }
  };

  const handleSelect = async (id: string) => {
    const res = await fetchPlatformAdapter(id);
    if (res.success && res.data) {
      setSelected(res.data);
    }
  };

  const handleUpdate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!selected) return;
    const form = new FormData(e.currentTarget);
    const res = await updatePlatformAdapter(selected.id, {
      platform: form.get('platform') as string,
      display_name: form.get('display_name') as string,
      publish_mode: form.get('publish_mode') as string,
      target_type: form.get('target_type') as string,
      config: {},
      enabled: form.get('enabled') === 'on',
    });
    if (res.success) {
      setSelected(null);
      const list = await fetchPlatformAdapters();
      if (list.success && list.data) setAdapters(list.data.items);
    } else {
      setError(res.error?.message ?? '更新失败');
    }
  };

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>平台 Adapter 管理</h1>
        <button onClick={() => setShowCreate(!showCreate)}>创建 Adapter</button>
      </section>

      {error && <p className="error">{error}</p>}

      {loading ? (
        <p>loading</p>
      ) : adapters.length === 0 ? (
        <p className="card">暂无</p>
      ) : (
        <ul>
          {adapters.map((a) => (
            <li key={a.id} className="card">
              <button onClick={() => handleSelect(a.id)}>
                {a.display_name} ({a.platform}) - {a.publish_mode}
              </button>
            </li>
          ))}
        </ul>
      )}

      {showCreate && (
        <form onSubmit={handleCreate} className="card">
          <input name="platform" placeholder="platform" required />
          <input name="display_name" placeholder="display_name" required />
          <input name="publish_mode" placeholder="publish_mode" required />
          <input name="target_type" placeholder="target_type" required />
          <button type="submit">创建</button>
        </form>
      )}

      {selected && (
        <form onSubmit={handleUpdate} className="card">
          <p>request_id: {selected.id}</p>
          <input name="platform" defaultValue={selected.platform} required />
          <input name="display_name" defaultValue={selected.display_name} required />
          <input name="publish_mode" defaultValue={selected.publish_mode} required />
          <input name="target_type" defaultValue={selected.target_type} required />
          <label><input name="enabled" type="checkbox" defaultChecked={selected.enabled} /> 启用</label>
          <button type="submit">保存</button>
        </form>
      )}
    </main>
  );
}
