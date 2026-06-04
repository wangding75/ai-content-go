'use client';

import { useState, useEffect, useRef } from 'react';
import { fetchPluginClients, createPluginClient, updatePluginClient, rotatePluginClientKey, type PluginClientResponse } from '@/lib/api';

export default function PluginClientsPage() {
  const [clients, setClients] = useState<PluginClientResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<PluginClientResponse | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [apiKeyOnce, setApiKeyOnce] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [rotatingClientID, setRotatingClientID] = useState<string | null>(null);
  const listRequestVersion = useRef(0);
  const loadingRequestVersion = useRef<number | null>(null);

  const loadClients = async (showLoading: boolean) => {
    const requestVersion = listRequestVersion.current + 1;
    listRequestVersion.current = requestVersion;
    if (showLoading) {
      setLoading(true);
      loadingRequestVersion.current = requestVersion;
    }
    try {
      const res = await fetchPluginClients();
      if (listRequestVersion.current !== requestVersion) {
        return;
      }
      if (res.success && res.data) {
        setClients(res.data.items ?? []);
        setError(null);
      } else {
        setError(res.error?.message ?? '请求失败');
      }
    } catch {
      if (listRequestVersion.current !== requestVersion) {
        return;
      }
      setError('请求失败');
    } finally {
      if (loadingRequestVersion.current !== null && loadingRequestVersion.current <= requestVersion && listRequestVersion.current === requestVersion) {
        setLoading(false);
        loadingRequestVersion.current = null;
      }
    }
  };

  useEffect(() => {
    void loadClients(true);
  }, []);

  const handleCreate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (isCreating) {
      return;
    }
    setIsCreating(true);
    const form = new FormData(e.currentTarget);
    try {
      const res = await createPluginClient({
        name: form.get('name') as string,
        client_type: form.get('client_type') as string,
        scopes: (form.get('scopes') as string).split(',').filter(Boolean),
      });
      if (res.success && res.data) {
        setApiKeyOnce(res.data.api_key_masked);
        setShowCreate(false);
        await loadClients(false);
      } else {
        setError(res.error?.message ?? '创建失败');
      }
    } catch {
      setError('创建失败');
    } finally {
      setIsCreating(false);
    }
  };

  const handleRotateKey = async (id: string) => {
    if (rotatingClientID === id) {
      return;
    }
    setRotatingClientID(id);
    try {
      const res = await rotatePluginClientKey(id);
      if (res.success && res.data) {
        setApiKeyOnce(res.data.api_key_masked);
      } else {
        setError(res.error?.message ?? '密钥轮换失败');
      }
    } catch {
      setError('密钥轮换失败');
    } finally {
      setRotatingClientID(null);
    }
  };

  const handleUpdate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!selected || isUpdating) return;
    setIsUpdating(true);
    const form = new FormData(e.currentTarget);
    try {
      const res = await updatePluginClient(selected.id, {
        name: form.get('name') as string,
        client_type: form.get('client_type') as string,
        scopes: (form.get('scopes') as string).split(',').filter(Boolean),
        status: form.get('status') as string,
      });
      if (res.success) {
        setSelected(null);
        await loadClients(false);
      } else {
        setError(res.error?.message ?? '更新失败');
      }
    } catch {
      setError('更新失败');
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>插件客户端管理</h1>
        <button onClick={() => {
          setSelected(null);
          setShowCreate(!showCreate);
        }}>注册客户端</button>
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
              <button onClick={() => {
                setShowCreate(false);
                setSelected(c);
              }}>编辑</button>
              <button onClick={() => handleRotateKey(c.id)} disabled={rotatingClientID === c.id}>轮换密钥</button>
            </li>
          ))}
        </ul>
      )}

      {showCreate && (
        <form onSubmit={handleCreate} className="card">
          <label htmlFor="plugin-client-create-name">名称</label>
          <input id="plugin-client-create-name" name="name" placeholder="name" required />
          <label htmlFor="plugin-client-create-type">客户端类型</label>
          <input id="plugin-client-create-type" name="client_type" placeholder="client_type" required />
          <label htmlFor="plugin-client-create-scopes">权限范围</label>
          <input id="plugin-client-create-scopes" name="scopes" placeholder="scopes (comma separated)" />
          <button type="submit" disabled={isCreating}>注册</button>
        </form>
      )}

      {selected && (
        <form onSubmit={handleUpdate} className="card">
          <p>request_id: {selected.id}</p>
          <label htmlFor="plugin-client-edit-name">名称</label>
          <input id="plugin-client-edit-name" name="name" defaultValue={selected.name} required />
          <label htmlFor="plugin-client-edit-type">客户端类型</label>
          <input id="plugin-client-edit-type" name="client_type" defaultValue={selected.client_type} required />
          <label htmlFor="plugin-client-edit-scopes">权限范围</label>
          <input id="plugin-client-edit-scopes" name="scopes" defaultValue={selected.scopes.join(',')} />
          <label htmlFor="plugin-client-edit-status">状态</label>
          <input id="plugin-client-edit-status" name="status" defaultValue={selected.status} />
          <button type="submit" disabled={isUpdating}>保存</button>
        </form>
      )}
    </main>
  );
}
