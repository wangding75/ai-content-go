'use client';

import { useState, useEffect } from 'react';
import { fetchPlatformCollectLogs, fetchPlatformCollectLog, confirmPlatformCollectLogMetrics, type PlatformCollectLogResponse, type PlatformCollectLogDetailResponse } from '@/lib/api';

export default function PlatformCollectLogsPage() {
  const [logs, setLogs] = useState<PlatformCollectLogResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<PlatformCollectLogDetailResponse | null>(null);

  useEffect(() => {
    fetchPlatformCollectLogs().then((res) => {
      if (res.success && res.data) {
        setLogs(res.data.items);
      } else {
        setError(res.error?.message ?? '请求失败');
      }
      setLoading(false);
    });
  }, []);

  const handleSelect = async (id: string) => {
    const res = await fetchPlatformCollectLog(id);
    if (res.success && res.data) {
      setSelected(res.data);
    }
  };

  const handleConfirm = async (id: string) => {
    const res = await confirmPlatformCollectLogMetrics(id, { metric_values: {} }, 'confirm-' + id);
    if (res.success) {
      setSelected(null);
      const list = await fetchPlatformCollectLogs();
      if (list.success && list.data) setLogs(list.data.items);
    } else {
      setError(res.error?.message ?? '确认失败');
    }
  };

  return (
    <main className="page-shell">
      <section className="page-hero">
        <h1>平台采集日志</h1>
      </section>

      {error && <p className="error">{error}</p>}

      {loading ? (
        <p>loading</p>
      ) : logs.length === 0 ? (
        <p className="card">暂无</p>
      ) : (
        <ul>
          {logs.map((log) => (
            <li key={log.id} className="card">
              <button onClick={() => handleSelect(log.id)}>
                {log.platform} - {log.status} - {log.error_summary || '错误摘要：无'}
              </button>
            </li>
          ))}
        </ul>
      )}

      {selected && (
        <div className="card">
          <p>request_id: {selected.id}</p>
          <p>项目: {selected.project_id}</p>
          <p>平台: {selected.platform}</p>
          <p>状态: {selected.status}</p>
          <p>错误摘要: {selected.error_summary || '错误摘要：无'}</p>
          {selected.status === 'ready' && (
            <button onClick={() => handleConfirm(selected.id)}>确认指标</button>
          )}
        </div>
      )}
    </main>
  );
}
