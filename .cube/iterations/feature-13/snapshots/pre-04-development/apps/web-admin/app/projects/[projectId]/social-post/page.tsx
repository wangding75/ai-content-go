'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { fetchSocialPostConfig, pageErrorFromEnvelope, updateSocialPostConfig, createSocialPostGenerationRun, fetchSocialPostGenerationRun, type SocialPostConfigResponse, type SocialPostGenerationRunDetailResponse, type PageError } from '../../../../lib/api';

export default function SocialPostProjectPage() {
  const params = useParams<{ projectId: string }>();
  const projectId = params.projectId;
  const [config, setConfig] = useState<SocialPostConfigResponse | null>(null);
  const [runDetail, setRunDetail] = useState<SocialPostGenerationRunDetailResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');
  const [busy, setBusy] = useState(false);

  async function loadConfig() {
    try {
      const envelope = await fetchSocialPostConfig(projectId);
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '加载 Social Post 配置失败'));
        return;
      }
      setConfig(envelope.data);
      setError(null);
    } catch {
      setError({ message: '加载 Social Post 配置失败' });
    }
  }

  useEffect(() => {
    void loadConfig();
  }, [projectId]);

  async function handleSaveConfig() {
    if (!config) return;
    setBusy(true);
    try {
      const envelope = await updateSocialPostConfig(projectId, {
        target_platforms: config.target_platforms,
        default_variant_count: config.default_variant_count,
        caption_length_policy: config.caption_length_policy,
        hashtag_policy: config.hashtag_policy,
        cover_copy_policy: config.cover_copy_policy,
        tone_style: config.tone_style,
        forbidden_terms: config.forbidden_terms,
      }, `social-post-config-${Date.now()}`);
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '保存配置失败'));
        return;
      }
      setNotice(`配置已保存 version_id=${envelope.data.version_id} request_id=${envelope.request_id}`);
      await loadConfig();
    } catch {
      setError({ message: '保存配置失败' });
    } finally {
      setBusy(false);
    }
  }

  async function handleCreateRun() {
    setBusy(true);
    try {
      const envelope = await createSocialPostGenerationRun(projectId, {
        topic: 'test topic',
        source_content_item_id: '',
        platform: 'xiaohongshu',
        version_count: 3,
        tone_style: 'friendly',
        asset_options: { generate_tags: true, generate_cover_copy: false },
      }, `social-post-gen-${Date.now()}`);
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '触发生成失败'));
        return;
      }
      setNotice(`生成已触发 generation_run_id=${envelope.data.generation_run_id} request_id=${envelope.request_id}`);
    } catch {
      setError({ message: '触发生成失败' });
    } finally {
      setBusy(false);
    }
  }

  async function handleLoadRun(runId: string) {
    try {
      const envelope = await fetchSocialPostGenerationRun(projectId, runId);
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '加载生成详情失败'));
        return;
      }
      setRunDetail(envelope.data);
      setError(null);
    } catch {
      setError({ message: '加载生成详情失败' });
    }
  }

  if (error) {
    return (
      <main role="alert">
        <p>{error.message}{error.code ? ` (${error.code})` : ''}{error.request_id ? ` request_id=${error.request_id}` : ''}</p>
        <button onClick={() => { setError(null); void loadConfig(); }}>重试</button>
      </main>
    );
  }

  if (!config) {
    return <main role="status"><p>加载中...</p></main>;
  }

  return (
    <main>
      <h1>Social Post · 项目 {projectId}</h1>
      {notice && <p role="status">{notice}</p>}

      <section>
        <h2>配置</h2>
        <p>目标平台: {config.target_platforms.join(', ') || '未设置'}</p>
        <p>默认版本数: {config.default_variant_count}</p>
        <p>文案长度策略: {config.caption_length_policy}</p>
        <p>风格: {config.tone_style}</p>
        <p>配置版本: {config.config_version}</p>
        <button onClick={handleSaveConfig} disabled={busy}>保存配置</button>
      </section>

      <section>
        <h2>触发生成</h2>
        <button onClick={handleCreateRun} disabled={busy}>创建生成运行</button>
      </section>

      {runDetail && (
        <section>
          <h2>生成详情</h2>
          <p>generation_run_id: {runDetail.generation_run_id}</p>
          <p>status: {runDetail.status}</p>
          <p>content_item_id: {runDetail.content_item_id}</p>
          {runDetail.variants?.length ? (
            <ul>
              {runDetail.variants.map((v) => (
                <li key={v.id}>{v.variant_index}. {v.title} · {v.platform} · {v.status}</li>
              ))}
            </ul>
          ) : (
            <p>暂无候选文案</p>
          )}
        </section>
      )}
    </main>
  );
}