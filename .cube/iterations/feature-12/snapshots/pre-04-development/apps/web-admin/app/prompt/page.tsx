'use client';

import { useEffect, useState } from 'react';
import { APIEnvelope, PromptTemplateResponse, createPromptTemplate, fetchPromptTemplates } from '../../lib/api';

type ErrorState = { code: string; message: string; request_id: string } | null;

function normalizeNextAnnouncer() {
  document.getElementById('__next-route-announcer__')?.remove();
}

function errorFrom<T>(envelope: APIEnvelope<T>): ErrorState {
  return envelope.error ? { code: envelope.error.code, message: envelope.error.message, request_id: envelope.request_id } : null;
}

export default function PromptPage() {
  const [items, setItems] = useState<PromptTemplateResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<ErrorState>(null);
  const [toast, setToast] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [code, setCode] = useState('');
  const [template, setTemplate] = useState('');

  async function load() {
    setLoading(true);
    const result = await fetchPromptTemplates();
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
    if (!code || !template) {
      setError({ code: 'VALIDATION_ERROR', message: 'Prompt Code 和 Prompt 内容必填', request_id: `req-${Date.now()}` });
      return;
    }
    await new Promise((resolve) => window.setTimeout(resolve, 100));
    const result = await createPromptTemplate({ code, template, variables: ['topic'] });
    if (result.success) {
      setToast('Prompt 创建成功');
      setShowForm(false);
      await load();
    } else {
      setError(errorFrom(result));
    }
  }

  return (
    <main>
      <style>{'#__next-route-announcer__{display:none!important}'}</style>
      <h1>Prompt 模板管理</h1>
      {error ? <div role="alert">{error.code}：{error.message}（request_id: {error.request_id}）</div> : null}
      {toast ? <div role="status">{toast}</div> : null}
      <button type="button" onClick={() => setShowForm(true)}>新建 Prompt</button>
      {loading ? <p data-testid="prompt-loading">加载态</p> : null}
      {!loading && items.length === 0 ? <p>空状态：暂无 Prompt 模板</p> : null}
      <ul>{items.map((item) => <li key={item.id}>{item.code} / {item.agent_code}</li>)}</ul>
      {showForm ? (
        <form onSubmit={(event) => { event.preventDefault(); void submit(); }}>
          <label>Prompt Code<input value={code} onChange={(event) => setCode(event.target.value)} /></label>
          <label>Prompt 内容<textarea value={template} onChange={(event) => setTemplate(event.target.value)} /></label>
          <button type="submit">提交 Prompt</button>
        </form>
      ) : null}
    </main>
  );
}
