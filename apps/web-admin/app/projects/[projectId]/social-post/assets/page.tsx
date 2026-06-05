'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { fetchSocialPostAssets, generateSocialPostTags, generateSocialPostCoverCopy, pageErrorFromEnvelope, type SocialPostAssetsResponse, type SocialPostAssetItem, type PageError } from '../../../../../lib/api';

export default function SocialPostAssetsPage() {
  const params = useParams<{ projectId: string }>();
  const projectId = params.projectId;
  const [assets, setAssets] = useState<SocialPostAssetsResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');
  const [busy, setBusy] = useState(false);

  async function load() {
    try {
      const envelope = await fetchSocialPostAssets(projectId, {});
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '加载资产失败'));
        return;
      }
      setAssets(envelope.data);
      setError(null);
    } catch {
      setError({ message: '加载资产失败' });
    }
  }

  useEffect(() => {
    void load();
  }, [projectId]);

  async function handleGenerateTags() {
    setBusy(true);
    try {
      const envelope = await generateSocialPostTags(projectId, {
        content_item_id: '',
        variant_id: '',
        platform: 'xiaohongshu',
        count: 5,
        style: 'trending',
      }, `generate-tags-${Date.now()}`);
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '生成标签失败'));
        return;
      }
      setNotice(`标签生成已触发 generation_run_id=${envelope.data.generation_run_id} request_id=${envelope.request_id}`);
    } catch {
      setError({ message: '生成标签失败' });
    } finally {
      setBusy(false);
    }
  }

  async function handleGenerateCoverCopy() {
    setBusy(true);
    try {
      const envelope = await generateSocialPostCoverCopy(projectId, {
        content_item_id: '',
        variant_id: '',
        platform: 'xiaohongshu',
        count: 2,
        style: 'warm',
      }, `generate-cover-copy-${Date.now()}`);
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '生成封面文案失败'));
        return;
      }
      setNotice(`封面文案生成已触发 generation_run_id=${envelope.data.generation_run_id} request_id=${envelope.request_id}`);
    } catch {
      setError({ message: '生成封面文案失败' });
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

  if (!assets) {
    return <main role="status"><p>加载中...</p></main>;
  }

  return (
    <main>
      <h1>资产 · 项目 {projectId}</h1>
      {notice && <p role="status">{notice}</p>}

      <section>
        <h2>标签</h2>
        <button onClick={handleGenerateTags} disabled={busy}>生成标签</button>
        {assets.tags.length === 0 ? (
          <p>暂无标签资产</p>
        ) : (
          <ul>
            {assets.tags.map((t: SocialPostAssetItem) => (
              <li key={t.id}>{t.platform} · source_variant_id={t.source_variant_id} · generation_run_id={t.generation_run_id}</li>
            ))}
          </ul>
        )}
      </section>

      <section>
        <h2>封面文案</h2>
        <button onClick={handleGenerateCoverCopy} disabled={busy}>生成封面文案</button>
        {assets.cover_copy.length === 0 ? (
          <p>暂无封面文案资产</p>
        ) : (
          <ul>
            {assets.cover_copy.map((c: SocialPostAssetItem) => (
              <li key={c.id}>{c.platform} · source_variant_id={c.source_variant_id} · generation_run_id={c.generation_run_id}</li>
            ))}
          </ul>
        )}
      </section>

      {assets.asset_suggestions.length > 0 && (
        <section>
          <h2>建议</h2>
          <ul>
            {assets.asset_suggestions.map((s, i) => <li key={i}>{s}</li>)}
          </ul>
        </section>
      )}
    </main>
  );
}