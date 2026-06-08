'use client';

import { use, useEffect, useState } from 'react';
import { createArticleGenerationRun, fetchArticleConfig, fetchArticleGenerationRuns, pageErrorFromEnvelope, retryArticleGenerationRun, updateArticleConfig, type ArticleConfigResponse, type ArticleGenerationRunSummaryResponse, type PageError } from '../../../../lib/api';

interface ArticlePageProps {
  params: Promise<{ projectId: string }>;
}

const emptyConfig: ArticleConfigResponse = {
  topic_style: '',
  audience_profile: '',
  seo_config: { keywords: [] },
  source_policy: '',
  structure_policy: '',
  default_workflow_template_version_id: '',
  enabled_metric_codes: [],
  version: '',
};

export default function ArticleProjectPage({ params }: ArticlePageProps) {
  const { projectId } = use(params);
  const [config, setConfig] = useState<ArticleConfigResponse>(emptyConfig);
  const [runs, setRuns] = useState<ArticleGenerationRunSummaryResponse[]>([]);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');
  const [generationForm, setGenerationForm] = useState({ topic: '', audience: '', sourceRefs: '', seoKeywords: '', targetPlatform: 'blog', outlineRequired: true });

  async function load() {
    const [configEnvelope, runsEnvelope] = await Promise.all([
      fetchArticleConfig(projectId),
      fetchArticleGenerationRuns(projectId),
    ]);

    if (configEnvelope.success && configEnvelope.data) {
      setConfig(configEnvelope.data);
      setError(null);
    } else if (configEnvelope.error?.code !== 'NOT_FOUND') {
      setError(pageErrorFromEnvelope(configEnvelope, '加载 Article 配置失败'));
      return;
    }

    if (!runsEnvelope.success || !runsEnvelope.data) {
      setError(pageErrorFromEnvelope(runsEnvelope, '加载生成运行失败'));
      return;
    }

    setRuns(runsEnvelope.data.items);
    setNotice(`页面已刷新 request_id=${runsEnvelope.request_id}`);
  }

  useEffect(() => {
    void load();
  }, [projectId]);

  async function handleSaveConfig() {
    const envelope = await updateArticleConfig(projectId, {
      topic_style: config.topic_style,
      audience_profile: config.audience_profile,
      seo_config: { keywords: config.seo_config?.keywords ?? [] },
      source_policy: config.source_policy,
      structure_policy: config.structure_policy,
      default_workflow_template_version_id: config.default_workflow_template_version_id,
    }, `article-config-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '保存 Article 配置失败'));
      return;
    }
    setNotice(`配置已保存 version_id=${envelope.data.version_id} request_id=${envelope.request_id}`);
    await load();
  }

  async function handleCreateRun() {
    const envelope = await createArticleGenerationRun(projectId, {
      topic: generationForm.topic,
      audience: generationForm.audience,
      source_refs: generationForm.sourceRefs.split(',').map((item) => item.trim()).filter(Boolean),
      seo_keywords: generationForm.seoKeywords.split(',').map((item) => item.trim()).filter(Boolean),
      outline_required: generationForm.outlineRequired,
      target_platform: generationForm.targetPlatform,
      generation_config: {},
    }, `article-run-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '创建 Article 生成失败'));
      return;
    }
    setNotice(`生成已受理 run_id=${envelope.data.generation_run_id} request_id=${envelope.request_id}`);
    await load();
  }

  async function handleRetry(runID: string) {
    const envelope = await retryArticleGenerationRun(projectId, runID, { reason: 'manual retry from web admin' }, `article-retry-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '重试生成失败'));
      return;
    }
    setNotice(`重试已受理 new_run_id=${envelope.data.new_generation_run_id} request_id=${envelope.request_id}`);
    await load();
  }

  return (
    <main className="page-shell">
      <section className="page-hero">
        <div className="page-hero__header">
          <div>
            <h1>Article 内容规划与生产</h1>
            <p>管理 Article 扩展配置、发起生成运行，并查看运行状态与失败重试入口。</p>
          </div>
          <div className="action-row">
            <button type="button" onClick={load}>刷新</button>
          </div>
        </div>
      </section>
      {notice && <p role="status">{notice}</p>}
      {error && <section className="card" role="alert">{error.code} {error.message} request_id={error.request_id}</section>}
      <section className="card">
        <div className="card__header">
          <h2>扩展配置</h2>
          <span className="badge badge--muted">version={config.version || '未保存'}</span>
        </div>
        <div className="card-grid">
          <label>
            主题风格
            <input value={config.topic_style} onChange={(event) => setConfig({ ...config, topic_style: event.target.value })} />
          </label>
          <label>
            受众画像
            <input value={config.audience_profile} onChange={(event) => setConfig({ ...config, audience_profile: event.target.value })} />
          </label>
          <label>
            来源策略
            <input value={config.source_policy} onChange={(event) => setConfig({ ...config, source_policy: event.target.value })} />
          </label>
          <label>
            结构策略
            <input value={config.structure_policy} onChange={(event) => setConfig({ ...config, structure_policy: event.target.value })} />
          </label>
          <label>
            默认工作流版本
            <input value={config.default_workflow_template_version_id} onChange={(event) => setConfig({ ...config, default_workflow_template_version_id: event.target.value })} />
          </label>
          <label>
            SEO 关键词
            <input value={(config.seo_config?.keywords ?? []).join(', ')} onChange={(event) => setConfig({ ...config, seo_config: { keywords: event.target.value.split(',').map((item) => item.trim()).filter(Boolean) } })} />
          </label>
        </div>
        <div className="action-row">
          <button type="button" onClick={handleSaveConfig}>保存配置</button>
        </div>
      </section>
      <section className="card">
        <div className="card__header">
          <h2>生成输入</h2>
          <span className="badge badge--muted">project_id={projectId}</span>
        </div>
        <div className="card-grid">
          <label>
            Topic
            <input value={generationForm.topic} onChange={(event) => setGenerationForm({ ...generationForm, topic: event.target.value })} />
          </label>
          <label>
            Audience
            <input value={generationForm.audience} onChange={(event) => setGenerationForm({ ...generationForm, audience: event.target.value })} />
          </label>
          <label>
            Target Platform
            <input value={generationForm.targetPlatform} onChange={(event) => setGenerationForm({ ...generationForm, targetPlatform: event.target.value })} />
          </label>
          <label>
            Source Refs
            <input value={generationForm.sourceRefs} onChange={(event) => setGenerationForm({ ...generationForm, sourceRefs: event.target.value })} />
          </label>
          <label>
            SEO Keywords
            <input value={generationForm.seoKeywords} onChange={(event) => setGenerationForm({ ...generationForm, seoKeywords: event.target.value })} />
          </label>
        </div>
        <div className="action-row">
          <button type="button" onClick={handleCreateRun}>发起生成</button>
        </div>
      </section>
      <section className="card table-card">
        <div className="card__header">
          <h2>生成运行</h2>
          <span className="badge badge--muted">{runs.length} 条</span>
        </div>
        {runs.length === 0 ? <p className="muted">暂无生成运行。</p> : (
          <table>
            <thead><tr><th>Run ID</th><th>Workflow Run</th><th>Topic</th><th>Status</th><th>操作</th></tr></thead>
            <tbody>
              {runs.map((run) => (
                <tr key={run.generation_run_id}>
                  <td>{run.generation_run_id}</td>
                  <td>{run.workflow_run_id}</td>
                  <td>{run.topic}</td>
                  <td>{run.status}</td>
                  <td><button type="button" onClick={() => handleRetry(run.generation_run_id)}>失败重试</button></td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
    </main>
  );
}
