'use client';

import { useState, useEffect } from 'react';
import { fetchPluginClients, createPluginClient, updatePluginClient, rotatePluginClientKey, type PluginClientResponse } from '@/lib/api';

export default function PluginClientsPage() {
  const [clients, setClients] = useState<PluginClientResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<PluginClientResponse | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [apiKeyOnce, setApiKeyOnce] = useState<string | null>(null);

  useEffect(() => {
    fetchPluginClients().then((res) => {
      if (res.success && res.data) {
        setClients(res.data.items);
      } else {
        setError(res.error?.message ?? '请求失败');
      }
      setLoading(false);
    });
  }, []);

  const handleCreate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = new FormData(e.currentTarget);
    const res = await createPluginClient({
      name: form.get('name') as string,
      client_type: form.get('client_type') as string,
      scopes: (form.get('scopes') as string).split(',').filter(Boolean),
    });
    if (res.success && res.data) {
      setApiKeyOnce(res.data.api_key_masked);
      setShowCreate(false);
      const list = await fetchPluginClients();
      if (list.success && list.data) setClients(list.data.items);
    } else {
      setError(res.error?.message ?? '创建失败');
    }
  };

  const handleRotateKey = async (id: string) => {
    const res = await rotatePluginClientKey(id);
    if (res.success && res.data) {
      setApiKeyOnce(res.data.api_key_masked);
    } else {
      setError(res.error?.message ?? '密钥轮换失败');
    }
  };

  const handleUpdate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!selected) return;
    const form = new FormData(e.currentTarget);
    const res = await updatePluginClient(selected.id, {
      name: form.get('name') as string,
      client_type: form.get('client_type') as string,
      scopes: (form.get('scopes') as string).split(',').filter(Boolean),
      status: form.get('status') as string,
    });
    if (res.success) {
      setSelected(null);
      const list = await fetchPluginClients();
      if (list.success && list.data) setClients(list.data.items);
    } else {
      setError(res.error?.message ?? '更新失败');
    }
  };

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>插件客户端管理</h1>
        <button onClick={() => setShowCreate(!showCreate)}>注册客户端</button>
      </section>

      {error && <p className="error">{error}</p>}

      {apiKeyOnce && <p className="card">api_key_once: {apiKeyOnce}</p>}

      {loading ? (
        <p>loading</p>
      ) : clients.length === 0 ? (
        <p className="card">暂无</p>
      ) : (
        <ul>
          {clients.map((c) => (
            <li key={c.id} className="card">
              <span>{c.name} ({c.client_type}) - {c.status}</span>
              <button onClick={() => setSelected(c)}>编辑</button>
              <button onClick={() => handleRotateKey(c.id)}>轮换密钥</button>
            </li>
          ))}
        </ul>
      )}

      {showCreate && (
        <form onSubmit={handleCreate} className="card">
          <input name="name" placeholder="name" required />
          <input name="client_type" placeholder="client_type" required />
          <input name="scopes" placeholder="scopes (comma separated)" />
          <button type="submit">注册</button>
        </form>
      )}

      {selected && (
        <form onSubmit={handleUpdate} className="card">
          <p>request_id: {selected.id}</p>
          <input name="name" defaultValue={selected.name} required />
          <input name="client_type" defaultValue={selected.client_type} required />
          <input name="scopes" defaultValue={selected.scopes.join(',')} />
          <input name="status" defaultValue={selected.status} />
          <button type="submit">保存</button>
        </form>
      )}
    </main>
  );
}
