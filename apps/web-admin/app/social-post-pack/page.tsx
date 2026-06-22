'use client';

import { useEffect, useState } from 'react';
import { fetchSocialPostPackStatus, pageErrorFromEnvelope, registerSocialPostPack, type SocialPostPackStatusResponse, type PageError } from '../../lib/api';

export default function SocialPostPackPage() {
  const [status, setStatus] = useState<SocialPostPackStatusResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');
  const [busy, setBusy] = useState(false);

  async function load() {
    try {
      const envelope = await fetchSocialPostPackStatus();
      if (!envelope.success) {
        if (envelope.error?.code === 'NOT_FOUND') {
          setStatus({ content_pack_id: '', content_type: null, schema: {}, workflows: [], metrics: [], current_version: '' });
          setError(null);
          return;
        }
        setError(pageErrorFromEnvelope(envelope, '加载 Social Post Pack 状态失败'));
        return;
      }
      setStatus(envelope.data);
      setError(null);
      setNotice(`状态已刷新 request_id=${envelope.request_id}`);
    } catch {
      setError({ message: '加载 Social Post Pack 状态失败' });
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function handleRegister() {
    setBusy(true);
    try {
      const envelope = await registerSocialPostPack(`social-post-pack-${Date.now()}`);
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '注册 Social Post Pack 失败'));
        return;
      }
      setNotice(`注册成功 content_pack_id=${envelope.data.content_pack_id} request_id=${envelope.request_id}`);
      await load();
    } catch {
      setError({ message: '注册 Social Post Pack 失败' });
    } finally {
      setBusy(false);
    }
  }

  if (error) {
    return (
      <main role="alert">
        <p>{error.message}{error.code ? ` (${error.code})` : ''}{error.request_id ? ` request_id=${error.request_id}` : ''}</p>
        <button onClick={() => { setError(null); void load(); }}>重试</button>
      </main>
    );
  }

  if (!status) {
    return <main role="status"><p>加载中...</p></main>;
  }

  return (
    <main>
      <h1>Social Post Pack</h1>
      {notice && <p role="status">{notice}</p>}
      <section>
        <h2>状态</h2>
        {status.content_type ? (
          <p>已注册 · content_pack_id={status.content_pack_id} · current_version={status.current_version}</p>
        ) : (
          <p>未注册</p>
        )}
      </section>
      {!status.content_type && (
        <section>
          <button onClick={handleRegister} disabled={busy}>
            {busy ? '注册中...' : '注册 Social Post Pack'}
          </button>
        </section>
      )}
      {status.content_type && (
        <>
          <section>
            <h2>工作流</h2>
            {status.workflows?.length ? (
              <ul>
                {status.workflows.map((wf) => (
                  <li key={wf.template_id}>{wf.name} · current_version={wf.current_version}</li>
                ))}
              </ul>
            ) : (
              <p>暂无工作流</p>
            )}
          </section>
          <section>
            <h2>默认指标</h2>
            {status.metrics?.length ? (
              <ul>
                {status.metrics.map((m) => (
                  <li key={m.metric_code}>{m.metric_name} ({m.metric_code}) · {m.unit}</li>
                ))}
              </ul>
            ) : (
              <p>暂无指标模板</p>
            )}
          </section>
        </>
      )}
    </main>
  );
}