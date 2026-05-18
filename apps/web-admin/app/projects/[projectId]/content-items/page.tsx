'use client';

import Link from 'next/link';
import { use, useEffect, useState } from 'react';
import { fetchContentItems, pageErrorFromEnvelope, type ContentItemResponse, type PageError } from '../../../../lib/api';

export default function ContentItemsPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [items, setItems] = useState<ContentItemResponse[]>([]);
  const [error, setError] = useState<PageError | null>(null);

  useEffect(() => {
    async function loadItems() {
      const envelope = await fetchContentItems(projectId, { status: 'pending_review', page: 1, page_size: 20 });
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '加载内容单元失败'));
        return;
      }
      setItems(envelope.data.items);
      setError(null);
    }
    void loadItems();
  }, [projectId]);

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>内容单元</h1>
        <p>查看 pending_review 内容单元，page_size=20。</p>
      </section>
      {error && <section className="card" role="alert">{error.message} request_id={error.request_id}</section>}
      <section className="card">
        <ul>
          {items.map((item) => (
            <li key={item.id}>
              <Link href={`/content-items/${item.id}`}>{item.title}</Link> · {item.status} · /content-items/{item.id}
            </li>
          ))}
        </ul>
      </section>
    </main>
  );
}
