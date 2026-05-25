'use client';

import { use, useEffect, useState } from 'react';
import { batchCreateMetricRecords, createMetricRecord, createMetricTemplate, fetchMetricTemplates, pageErrorFromEnvelope, type MetricTemplateResponse, type PageError } from '../../../../../lib/api';

export default function MetricInputPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [templates, setTemplates] = useState<MetricTemplateResponse[]>([]);
  const [metricCode, setMetricCode] = useState('views');
  const [metricName, setMetricName] = useState('阅读量');
  const [rawValue, setRawValue] = useState('100');
  const [batchText, setBatchText] = useState('views,2026-05-25,100');
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');

  async function loadTemplates() {
    const envelope = await fetchMetricTemplates({ enabled: true, page: 1, page_size: 20 });
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '加载指标模板失败'));
      return;
    }
    setTemplates(envelope.data.items);
    setError(null);
  }

  useEffect(() => {
    void loadTemplates();
  }, []);

  async function createTemplate() {
    const envelope = await createMetricTemplate({ content_type: 'article', platform: 'manual', metric_code: metricCode, metric_name: metricName, unit: 'count', value_type: 'integer', aggregation_method: 'sum', period: 'day', required: true, enabled: true });
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '创建指标模板失败'));
      return;
    }
    setNotice(`指标模板已创建：${envelope.data.metric_template_id}`);
    await loadTemplates();
  }

  async function submitRecord() {
    const envelope = await createMetricRecord({ project_id: projectId, content_item_id: 'content-item-1', content_version_id: 'version-1', publish_job_id: 'publish-job-1', target_id: 'publish-target-1', platform: 'manual', external_url: '', metric_code: metricCode, metric_date: '2026-05-25', period: 'day', raw_value: rawValue, source_type: 'manual', source_ref: 'web-admin' }, `metric-record-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '录入指标失败'));
      return;
    }
    setNotice(`指标已保存：${envelope.data.metric_record_id} 标准化值 ${envelope.data.normalized_value}`);
    setError(null);
  }

  async function submitBatch() {
    const records = batchText.split('\n').filter(Boolean).map((line, index) => {
      const [code, date, value] = line.split(',');
      return { project_id: projectId, content_item_id: index === 0 ? 'content-item-1' : '', content_version_id: 'version-1', publish_job_id: 'publish-job-1', target_id: 'publish-target-1', platform: 'manual', external_url: '', metric_code: code ?? metricCode, metric_date: date ?? '2026-05-25', period: 'day', raw_value: value ?? rawValue, source_type: 'import', source_ref: `row-${index + 1}` };
    });
    const envelope = await batchCreateMetricRecords({ records, import_source: 'web-admin' }, `metric-batch-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '批量导入失败'));
      return;
    }
    setNotice(`批量导入完成：成功 ${envelope.data.created_count}，失败 ${envelope.data.failed_count}，errors=${envelope.data.errors.length}`);
    setError(null);
  }

  return (
    <main className="page-shell">
      <section className="page-hero">
        <div className="page-hero__header">
          <div>
            <h1>指标录入</h1>
            <p>维护指标模板，录入单条指标或批量导入平台表现数据。</p>
          </div>
          <button type="button" onClick={loadTemplates}>刷新模板</button>
        </div>
      </section>
      {notice && <p role="status">{notice}</p>}
      {error && <section className="card" role="alert">{error.code} {error.message} request_id={error.request_id}</section>}
      <section className="card">
        <div className="card__header"><h2>模板创建</h2><button type="button" onClick={createTemplate}>创建模板</button></div>
        <div className="form-grid">
          <label>指标编码<input value={metricCode} onChange={(event) => setMetricCode(event.target.value)} /></label>
          <label>指标名称<input value={metricName} onChange={(event) => setMetricName(event.target.value)} /></label>
          <label>原始值<input value={rawValue} onChange={(event) => setRawValue(event.target.value)} /></label>
        </div>
      </section>
      <section className="card">
        <h2>可用模板</h2>
        {templates.length === 0 ? <p className="muted">暂无指标模板</p> : <p>{templates.map((template) => template.metric_code).join(' / ')}</p>}
      </section>
      <section className="card">
        <div className="card__header"><h2>录入动作</h2><div className="action-row"><button type="button" onClick={submitRecord}>保存指标</button><button type="button" onClick={submitBatch}>批量导入</button></div></div>
        <label>批量导入<textarea value={batchText} onChange={(event) => setBatchText(event.target.value)} /></label>
        <p>逐条错误：批量导入后会展示 errors 明细。</p>
      </section>
    </main>
  );
}
