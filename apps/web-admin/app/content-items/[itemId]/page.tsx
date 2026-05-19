'use client';

import { use, useEffect, useState } from 'react';
import { fetchContentItem, pageErrorFromEnvelope, type ContentItemDetailResponse, type PageError } from '../../../lib/api';

export default function ContentItemDetailPage({ params }: { params: Promise<{ itemId: string }> }) {
  const { itemId } = use(params);
  const [item, setItem] = useState<ContentItemDetailResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);

  useEffect(() => {
    async function loadItem() {
      const envelope = await fetchContentItem(itemId);
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '加载内容详情失败'));
        return;
      }
      setItem(envelope.data);
      setError(null);
    }
    void loadItem();
  }, [itemId]);

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>ContentItem 详情</h1>
        <p>查看正文、metadata、extension 与 generation_run_id。</p>
      </section>
      {error && <section className="card" role="alert">{error.message} request_id={error.request_id}</section>}
      {item && (
        <section className="card">
          <h2>{item.title}</h2>
          <p>generation_run_id: {item.generation_run_id}</p>
          <h3>body</h3>
          <article>{item.body}</article>
          <h3>metadata</h3>
          <pre>{JSON.stringify(item.metadata, null, 2)}</pre>
          <h3>extension</h3>
          <pre>{JSON.stringify(item.extension, null, 2)}</pre>
        </section>
      )}
    </main>
  );
}
