'use client';

import Link from 'next/link';
import { use, useEffect, useState } from 'react';
import { approveReview, fetchContentReview, pageErrorFromEnvelope, rejectReview, type ContentReviewDetailResponse, type PageError } from '../../../lib/api';

export default function ContentReviewDetailPage({ params }: { params: Promise<{ reviewId: string }> }) {
  const { reviewId } = use(params);
  const [review, setReview] = useState<ContentReviewDetailResponse | null>(null);
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');

  async function loadReview() {
    const envelope = await fetchContentReview(reviewId);
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '加载审稿详情失败'));
      return;
    }
    setReview(envelope.data);
    setError(null);
  }

  useEffect(() => { void loadReview(); }, [reviewId]);

  async function approve() {
    const envelope = await approveReview(reviewId, { note: 'approved from review detail' });
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '通过审稿失败'));
      return;
    }
    setNotice(`已通过：${envelope.data.operation_log_id} request_id=${envelope.request_id}`);
    await loadReview();
  }

  async function reject(trigger: boolean) {
    const envelope = await rejectReview(reviewId, { reason: 'needs improvement', regenerate_instruction: 'regenerate with stronger structure', trigger_regeneration: trigger });
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '打回审稿失败'));
      return;
    }
    setNotice(`已打回：${envelope.data.operation_log_id} ${envelope.data.regeneration_run_id ?? ''} request_id=${envelope.request_id}`);
    await loadReview();
  }

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>审稿详情</h1>
        <p>查看正文、审稿报告、版本历史，并执行通过或打回。</p>
      </section>
      {notice && <section className="card" role="status">{notice}</section>}
      {error && <section className="card" role="alert">{error.code} {error.message} request_id={error.request_id}</section>}
      {review && <section className="card">
        <h2>{review.title || review.id}</h2>
        <p>{review.status} · {review.review_type}</p>
        <div className="action-row">
          <button type="button" onClick={() => void approve()}>通过</button>
          <button type="button" onClick={() => void reject(false)}>仅打回</button>
          <button type="button" onClick={() => void reject(true)}>打回并重生成</button>
          <Link href={`/content-reviews/${review.id}/ai-report`}>AI 质检报告</Link>
          <Link href={`/content-reviews/${review.id}/edit-approve`}>编辑后通过</Link>
        </div>
        <h3>正文</h3>
        <article>{review.body}</article>
        <h3>版本</h3>
        <ul>{review.versions.map((version) => <li key={version.id}>v{version.version_no} · {version.source}</li>)}</ul>
      </section>}
    </main>
  );
}
