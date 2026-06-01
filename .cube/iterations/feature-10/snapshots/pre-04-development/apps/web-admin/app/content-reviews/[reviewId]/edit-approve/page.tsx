'use client';

import { use, useEffect, useState } from 'react';
import { approveWithEdit, fetchContentReview, pageErrorFromEnvelope, type ContentReviewDetailResponse, type PageError } from '../../../../lib/api';

export default function EditApprovePage({ params }: { params: Promise<{ reviewId: string }> }) {
  const { reviewId } = use(params);
  const [review, setReview] = useState<ContentReviewDetailResponse | null>(null);
  const [editableFields, setEditableFields] = useState<Record<string, unknown>>({});
  const [note, setNote] = useState('edited and approved');
  const [error, setError] = useState<PageError | null>(null);
  const [notice, setNotice] = useState('');

  useEffect(() => {
    async function loadReview() {
      const envelope = await fetchContentReview(reviewId);
      if (!envelope.success || !envelope.data) {
        setError(pageErrorFromEnvelope(envelope, '加载审稿详情失败'));
        return;
      }
      setReview(envelope.data);
      setEditableFields({ title: envelope.data.title, body: envelope.data.body });
      setError(null);
    }
    void loadReview();
  }, [reviewId]);

  async function submit() {
    const envelope = await approveWithEdit(reviewId, { editable_fields: editableFields, note });
    if (!envelope.success || !envelope.data) {
      setError(pageErrorFromEnvelope(envelope, '编辑后通过失败'));
      return;
    }
    setNotice(`已创建版本：${envelope.data.content_version_id} operation_log_id=${envelope.data.operation_log_id} request_id=${envelope.request_id}`);
  }

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>编辑后通过</h1>
        <p>按内容类型 editable fields 编辑内容，并创建新版本后通过审稿。</p>
      </section>
      {error && <section className="card" role="alert">{error.code} {error.message} request_id={error.request_id}</section>}
      {notice && <section className="card" role="status">{notice}</section>}
      <section className="card">
        <h2>{review?.title ?? reviewId}</h2>
        <label>标题 <input value={String(editableFields.title ?? '')} onChange={(event) => setEditableFields({ ...editableFields, title: event.target.value })} /></label>
        <label>正文 <textarea value={String(editableFields.body ?? '')} onChange={(event) => setEditableFields({ ...editableFields, body: event.target.value })} /></label>
        <label>备注 <input value={note} onChange={(event) => setNote(event.target.value)} /></label>
        <button type="button" onClick={() => void submit()}>保存并通过</button>
      </section>
    </main>
  );
}
