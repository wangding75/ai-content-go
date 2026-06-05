'use client';

import Link from 'next/link';
import { use, useEffect, useState } from 'react';
import { createContentReview, fetchContentItems, fetchContentReviews, pageErrorFromEnvelope, type ContentItemResponse, type ContentReviewResponse, type PageError } from '../../../../lib/api';

export default function ReviewsPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = use(params);
  const [reviews, setReviews] = useState<ContentReviewResponse[]>([]);
  const [items, setItems] = useState<ContentItemResponse[]>([]);
  const [status, setStatus] = useState('pending');
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');

  async function loadReviews(nextStatus = status) {
    const envelope = await fetchContentReviews(projectId, { status: nextStatus, page: 1, page_size: 20 });
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '加载审稿列表失败'));
      return;
    }
    setReviews(envelope.data.items);
    setError(null);
  }

  useEffect(() => {
    async function load() {
      await loadReviews(status);
      const itemEnvelope = await fetchContentItems(projectId, { status: 'pending_review', page: 1, page_size: 20 });
      if (itemEnvelope.success && itemEnvelope.data) setItems(itemEnvelope.data.items);
    }
    void load();
  }, [projectId]);

  async function createReview(itemID: string) {
    const envelope = await createContentReview(itemID, { review_type: 'combined' }, `review-${Date.now()}`);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '创建审稿失败'));
      return;
    }
    setNotice(`审稿已创建：${envelope.data.review_id} request_id=${envelope.request_id}`);
    await loadReviews(status);
  }

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>审稿中心</h1>
        <p>查看待审稿内容、筛选审稿状态，并为 pending_review 内容单元创建审稿。</p>
      </section>
      <section className="card">
        <label>状态筛选 <select value={status} onChange={(event) => { setStatus(event.target.value); void loadReviews(event.target.value); }}><option value="pending">pending</option><option value="in_review">in_review</option><option value="approved">approved</option><option value="rejected">rejected</option></select></label>
        {notice && <p role="status">{notice}</p>}
        {error && <p role="alert">{error.code} {error.message} request_id={error.request_id}</p>}
      </section>
      <section className="card">
        <h2>审稿列表</h2>
        {reviews.length === 0 ? <p>暂无审稿记录</p> : <ul>{reviews.map((review) => <li key={review.id}><Link href={`/content-reviews/${review.id}`}>{review.title || review.id}</Link> · {review.status}</li>)}</ul>}
      </section>
      <section className="card">
        <h2>可创建审稿的内容单元</h2>
        {items.length === 0 ? <p>暂无 pending_review 内容单元</p> : <ul>{items.map((item) => <li key={item.id}>{item.title} <button type="button" onClick={() => void createReview(item.id)}>创建审稿</button></li>)}</ul>}
      </section>
    </main>
  );
}
