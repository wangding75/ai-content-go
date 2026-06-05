'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { fetchSocialPostVariants, selectSocialPostVariant, pageErrorFromEnvelope, type SocialPostVariantResponse, type PageError } from '../../../../../lib/api';

export default function SocialPostVariantsPage() {
  const params = useParams<{ projectId: string }>();
  const projectId = params.projectId;
  const [variants, setVariants] = useState<SocialPostVariantResponse[]>([]);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');
  const [busy, setBusy] = useState(false);

  async function load() {
    try {
      const envelope = await fetchSocialPostVariants(projectId, { page: 1, page_size: 20 });
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '加载候选文案失败'));
        return;
      }
      setVariants(envelope.data.items);
      setError(null);
    } catch {
      setError({ message: '加载候选文案失败' });
    }
  }

  useEffect(() => {
    void load();
  }, [projectId]);

  async function handleSelect(variantId: string, contentItemId: string) {
    setBusy(true);
    try {
      const envelope = await selectSocialPostVariant(projectId, variantId, {
        content_item_id: contentItemId,
        note: '选择为主版本',
      }, `select-variant-${Date.now()}`);
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '选择主版本失败'));
        return;
      }
      setNotice(`主版本已选择 content_version_id=${envelope.data.content_version_id} request_id=${envelope.request_id}`);
      await load();
    } catch {
      setError({ message: '选择主版本失败' });
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

  return (
    <main>
      <h1>候选文案 · 项目 {projectId}</h1>
      {notice && <p role="status">{notice}</p>}
      {variants.length === 0 ? (
        <p>暂无候选文案</p>
      ) : (
        <ul>
          {variants.map((v) => (
            <li key={v.id}>
              {v.variant_index}. {v.title} · {v.platform} · {v.status}
              {v.status === 'generated' && (
                <button onClick={() => handleSelect(v.id, v.content_item_id)} disabled={busy}>选择主版本</button>
              )}
              {v.content_version_id && <span> · content_version_id={v.content_version_id}</span>}
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}