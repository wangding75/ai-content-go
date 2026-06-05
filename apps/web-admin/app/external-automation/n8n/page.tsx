'use client';

import { useEffect, useMemo, useState } from 'react';
import {
  ExternalBindingResponse,
  ExternalProviderResponse,
  PageError,
  ExternalCallbackLogResponse,
  createExternalBinding,
  createExternalProvider,
  fetchExternalBindings,
  fetchExternalProviders,
  fetchCallbackLogs,
  rotateCallbackToken,
  updateCallbackAuth,
  receiveExternalCallback,
  testExternalCallback,
  pageErrorFromEnvelope,
} from '../../../lib/api';

export default function ExternalAutomationN8NPage() {
  const [providers, setProviders] = useState<ExternalProviderResponse[]>([]);
  const [bindings, setBindings] = useState<ExternalBindingResponse[]>([]);
  const [token, setToken] = useState('');
  const [baseURL, setBaseURL] = useState('https://n8n.example.invalid');
  const [triggerFilter, setTriggerFilter] = useState('all');
  const [selectedProviderID, setSelectedProviderID] = useState('');
  const [selectedBindingID, setSelectedBindingID] = useState('');
  const [error, setError] = useState<PageError | null>(null);
  const [toast, setToast] = useState('');
  const [callbackTokenOnce, setCallbackTokenOnce] = useState<string | null>(null);
  const [callbackLogs, setCallbackLogs] = useState<ExternalCallbackLogResponse[]>([]);
  const [callbackLogsLoading, setCallbackLogsLoading] = useState(false);
  const [rotateReason, setRotateReason] = useState('');
  const [testEventType, setTestEventType] = useState('workflow_run.completed');
  const [testPayload, setTestPayload] = useState('{"test": true}');

  async function load() {
    const providerResult = await fetchExternalProviders();
    const bindingResult = await fetchExternalBindings();
    if (providerResult.success && providerResult.data && bindingResult.success && bindingResult.data) {
      setProviders(providerResult.data.items);
      setBindings(bindingResult.data.items);
      setSelectedProviderID((value) => value || providerResult.data.items[0]?.id || '');
      setError(null);
    } else if (!providerResult.success) {
      setError(pageErrorFromEnvelope(providerResult, '加载外部自动化失败'));
    } else {
      setError(pageErrorFromEnvelope(bindingResult, '加载外部自动化失败'));
    }
  }

  useEffect(() => {
    void load();
  }, []);

  const filteredBindings = useMemo(() => bindings.filter((binding) => triggerFilter === 'all' || binding.trigger_event === triggerFilter), [bindings, triggerFilter]);

  async function submitProvider() {
    const result = await createExternalProvider({ provider_type: 'n8n', base_url: baseURL, token });
    if (result.success) {
      setToken('');
      setToast(`Provider 创建成功：${result.data?.token_masked}`);
      await load();
    } else {
      setError(pageErrorFromEnvelope(result, 'Provider 创建失败'));
    }
  }

  async function submitBinding() {
    const providerID = selectedProviderID || providers[0]?.id;
    if (!providerID) {
      setError({ message: '请先保存 Provider', request_id: 'client' });
      return;
    }
    const result = await createExternalBinding({ provider_id: providerID, trigger_event: 'workflow_run.completed', webhook_url: 'https://n8n.example.invalid/webhook/run' });
    if (result.success) {
      setToast(`Binding 创建成功：${result.data?.binding_id}`);
      await load();
    } else {
      setError(pageErrorFromEnvelope(result, 'Binding 创建失败'));
    }
  }

  async function handleRotateToken() {
    if (!selectedBindingID || !rotateReason) {
      setError({ message: '请选择 Binding 并填写轮换原因', request_id: 'client' });
      return;
    }
    const result = await rotateCallbackToken(selectedBindingID);
    if (result.success && result.data) {
      setCallbackTokenOnce(result.data.token_masked ?? null);
      setToast(`Token 轮换成功（operation_log_id: ${result.data.operation_log_id}）`);
      setRotateReason('');
    } else {
      setError(pageErrorFromEnvelope(result, 'Token 轮换失败'));
    }
  }

  async function handleUpdateAuth() {
    if (!selectedBindingID) {
      setError({ message: '请选择 Binding', request_id: 'client' });
      return;
    }
    const result = await updateCallbackAuth(selectedBindingID, {
      callback_auth_type: 'HMAC-SHA256',
      change_reason: 'webhook 升级',
    });
    if (result.success) {
      setToast(`Auth 更新成功（operation_log_id: ${result.data?.operation_log_id}）`);
    } else {
      setError(pageErrorFromEnvelope(result, 'Auth 更新失败'));
    }
  }

  async function handleTestCallback() {
    if (!selectedBindingID) {
      setError({ message: '请选择 Binding', request_id: 'client' });
      return;
    }
    let payload: Record<string, unknown> = {};
    try {
      payload = JSON.parse(testPayload);
    } catch {
      setError({ message: 'Payload 格式无效，需为 JSON', request_id: 'client' });
      return;
    }
    const result = await testExternalCallback(selectedBindingID, {
      event_type: testEventType,
      payload,
    });
    if (result.success && result.data) {
      setToast(`测试回调 sent（callback_log_id: ${result.data.callback_log_id}）`);
    } else {
      setError(pageErrorFromEnvelope(result, '测试回调失败'));
    }
  }

  async function loadCallbackLogs() {
    setCallbackLogsLoading(true);
    const result = await fetchCallbackLogs(selectedBindingID);
    if (result.success && result.data) {
      setCallbackLogs(result.data.items);
      setError(null);
    } else {
      setError(pageErrorFromEnvelope(result, '加载回调日志失败'));
    }
    setCallbackLogsLoading(false);
  }

  return (
    <main className="page-shell" data-testid="styled-page-shell">
      <section className="page-hero">
        <div className="page-hero__header">
          <div>
            <h1>外部自动化 / n8n</h1>
            <p>管理 n8n Provider 和 workflow_run.completed webhook binding，页面只展示 masked token。</p>
          </div>
          <span className="badge badge--success">token_masked</span>
        </div>
      </section>

      {error ? <div role="alert">{error.code ? `${error.code}：` : ''}{error.message}（request_id: {error.request_id ?? 'client'}）</div> : null}
      {toast ? <div role="status">{toast}</div> : null}

      <section className="card">
        <div className="card__header"><h2>Provider 表单</h2><span className="badge badge--muted">n8n</span></div>
        <div className="form-grid">
          <label>Base URL<input value={baseURL} onChange={(event) => setBaseURL(event.target.value)} /></label>
          <label>Token<input type="password" value={token} onChange={(event) => setToken(event.target.value)} /></label>
        </div>
        <div className="action-row"><button type="button" onClick={submitProvider}>保存 Provider</button></div>
      </section>

      <section className="card">
        <div className="card__header"><h2>Binding 表单</h2><button type="button" onClick={submitBinding}>保存 Binding</button></div>
        <div className="form-grid">
          <label>Provider
            <select value={selectedProviderID} onChange={(event) => setSelectedProviderID(event.target.value)}>
              {providers.map((provider) => <option key={provider.id} value={provider.id}>{provider.provider_type} / {provider.token_masked}</option>)}
            </select>
          </label>
          <label>事件筛选
            <select value={triggerFilter} onChange={(event) => setTriggerFilter(event.target.value)}>
              <option value="all">全部</option>
              <option value="workflow_run.completed">workflow_run.completed</option>
            </select>
          </label>
        </div>
      </section>

      <section className="card-grid">
        <article className="card"><h2>Providers</h2><ul>{providers.map((provider) => <li key={provider.id}>{provider.provider_type} {provider.base_url} <span className="badge badge--success">{provider.token_masked}</span></li>)}</ul></article>
        <article className="card"><h2>Bindings</h2><ul>{filteredBindings.map((binding) => <li key={binding.id}>{binding.trigger_event} {binding.webhook_url} <span className={binding.enabled ? 'badge badge--success' : 'badge badge--muted'}>{binding.enabled ? '启用' : '停用'}</span></li>)}</ul></article>
      </section>

      <section className="card">
        <div className="card__header"><h2>Callback Token 轮换</h2></div>
        <div className="form-grid">
          <label>Binding<select value={selectedBindingID} onChange={(event) => setSelectedBindingID(event.target.value)}>{bindings.map((b) => <option key={b.id} value={b.id}>{b.trigger_event} / {b.webhook_url}</option>)}</select></label>
          <label>轮换原因<input value={rotateReason} onChange={(event) => setRotateReason(event.target.value)} placeholder="轮换原因" /></label>
        </div>
        <div className="action-row"><button type="button" onClick={handleRotateToken}>轮换 Token</button></div>
        {callbackTokenOnce && <p className="card">callback_token_once: {callbackTokenOnce}</p>}
      </section>

      <section className="card">
        <div className="card__header"><h2>Callback Auth 更新</h2></div>
        <div className="form-grid">
          <label>Binding<select value={selectedBindingID} onChange={(event) => setSelectedBindingID(event.target.value)}>{bindings.map((b) => <option key={b.id} value={b.id}>{b.trigger_event} / {b.webhook_url}</option>)}</select></label>
        </div>
        <div className="action-row"><button type="button" onClick={handleUpdateAuth}>更新 Auth</button></div>
      </section>

      <section className="card">
        <div className="card__header"><h2>测试回调</h2></div>
        <div className="form-grid">
          <label>Binding<select value={selectedBindingID} onChange={(event) => setSelectedBindingID(event.target.value)}>{bindings.map((b) => <option key={b.id} value={b.id}>{b.trigger_event} / {b.webhook_url}</option>)}</select></label>
          <label>事件类型<input value={testEventType} onChange={(event) => setTestEventType(event.target.value)} /></label>
          <label>Payload<textarea value={testPayload} onChange={(event) => setTestPayload(event.target.value)} rows={2} /></label>
        </div>
        <div className="action-row"><button type="button" onClick={handleTestCallback}>发送测试回调</button></div>
      </section>

      <section className="card">
        <div className="card__header"><h2>回调日志</h2><button type="button" onClick={loadCallbackLogs}>刷新</button></div>
        {callbackLogsLoading ? <p>loading</p> : callbackLogs.length === 0 ? <p className="card">暂无</p> : (
          <ul>{callbackLogs.map((log) => (
            <li key={log.id} className="card">
              <p>callback_log_id: {log.id}</p>
              <p>request_id: {log.binding_id}</p>
              <p>状态: {log.status}</p>
              <p>事件类型: {log.event_type}</p>
              <p>创建时间: {log.created_at || '-'}</p>
            </li>
          ))}</ul>
        )}
      </section>
    </main>
  );
}
