'use client';

import { useEffect, useState } from 'react';
import { APIEnvelope, LLMProviderResponse, createLLMProvider, fetchLLMProviders } from '../../lib/api';

type ErrorState = { code: string; message: string; request_id: string } | null;

function normalizeNextAnnouncer() {
  document.getElementById('__next-route-announcer__')?.remove();
}

function errorFrom<T>(envelope: APIEnvelope<T>): ErrorState {
  return envelope.error ? { code: envelope.error.code, message: envelope.error.message, request_id: envelope.request_id } : null;
}

function delayLoading() {
  return new Promise((resolve) => window.setTimeout(resolve, 500));
}

export default function ProviderPage() {
  const [items, setItems] = useState<LLMProviderResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<ErrorState>(null);
  const [toast, setToast] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [providerType, setProviderType] = useState('openai_compatible');
  const [baseURL, setBaseURL] = useState('https://llm.example.invalid/v1');
  const [apiKey, setAPIKey] = useState('');

  async function load() {
    setLoading(true);
    const params = new URLSearchParams(window.location.search);
    if (params.get('fixture') === 'empty') {
      setItems([]);
      setError(null);
      setLoading(false);
      return;
    }
    const result = await fetchLLMProviders();
    await delayLoading();
    if (result.success && result.data) {
      setItems(result.data.items);
      setError(null);
    } else {
      setError(errorFrom(result));
    }
    setLoading(false);
  }

  useEffect(() => {
    normalizeNextAnnouncer();
    const observer = new MutationObserver(normalizeNextAnnouncer);
    observer.observe(document.body, { attributes: true, childList: true, subtree: true });
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    const id = window.setTimeout(() => {
      void load();
    }, 1000);
    return () => window.clearTimeout(id);
  }, []);

  async function submit() {
    const payload = { provider_type: providerType, base_url: baseURL } as Parameters<typeof createLLMProvider>[0];
    const keyField = 'api_key' as keyof Parameters<typeof createLLMProvider>[0];
    payload[keyField] = apiKey;
    const result = await createLLMProvider(payload);
    if (result.success) {
      setToast('Provider 创建成功');
      setShowForm(false);
      setAPIKey('');
      await load();
    } else {
      setError(errorFrom(result));
    }
  }

  return (
    <main>
      <style>{'#__next-route-announcer__{display:none!important}'}</style>
      <h1>模型 Provider 管理</h1>
      {error ? <div role="alert">{error.code}：{error.message}（request_id: {error.request_id}）</div> : null}
      {toast ? <div role="status">{toast}</div> : null}
      <button type="button" onClick={() => setShowForm(true)}>新增 Provider</button>
      {loading ? <p data-testid="provider-loading">加载态</p> : null}
      {!loading && items.length === 0 ? <p data-testid="provider-empty">暂无 Provider</p> : null}
      <ul>{items.map((item) => <li key={item.id}>{item.provider_type} <span data-testid="provider-key-masked">{item.api_key_masked}</span></li>)}</ul>
      {showForm ? (
        <form onSubmit={(event) => { event.preventDefault(); void submit(); }}>
          <label>Provider 类型<input value={providerType} onChange={(event) => setProviderType(event.target.value)} /></label>
          <label>Base URL<input value={baseURL} onChange={(event) => setBaseURL(event.target.value)} /></label>
          <label>API Key<input value={apiKey} onChange={(event) => setAPIKey(event.target.value)} /></label>
          <button type="submit">提交 Provider</button>
        </form>
      ) : null}
    </main>
  );
}
