'use client';

import { useEffect, useMemo, useState } from 'react';
import {
  PageError,
  ScheduleTriggerResponse,
  WorkflowScheduleResponse,
  createWorkflowSchedule,
  disableWorkflowSchedule,
  enableWorkflowSchedule,
  fetchWorkflowScheduleTriggers,
  fetchWorkflowSchedules,
  pageErrorFromEnvelope,
  testRunWorkflowSchedule,
} from '../../../lib/api';

const pageSize = 5;

export default function WorkflowSchedulesPage() {
  const [items, setItems] = useState<WorkflowScheduleResponse[]>([]);
  const [triggers, setTriggers] = useState<ScheduleTriggerResponse[]>([]);
  const [selectedSchedule, setSelectedSchedule] = useState<WorkflowScheduleResponse | null>(null);
  const [statusFilter, setStatusFilter] = useState('all');
  const [page, setPage] = useState(1);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [projectID, setProjectID] = useState('seed-project');
  const [templateVersionID, setTemplateVersionID] = useState('wftv-seed');
  const [cronExpression, setCronExpression] = useState('0 9 * * *');
  const [dailyContentCount, setDailyContentCount] = useState('5');
  const [lastCreatedScheduleID, setLastCreatedScheduleID] = useState('');
  const [error, setError] = useState<PageError | null>(null);
  const [toast, setToast] = useState('');

  async function loadSchedules() {
    const result = await fetchWorkflowSchedules();
    if (result.success && result.data) {
      setItems(result.data.items);
      setError(null);
    } else {
      setError(pageErrorFromEnvelope(result, '加载调度失败'));
    }
  }

  useEffect(() => {
    void loadSchedules();
  }, []);

  const filteredItems = useMemo(() => {
    return items.filter((item) => statusFilter === 'all' || (statusFilter === 'enabled' ? item.enabled : !item.enabled));
  }, [items, statusFilter]);
  const totalPages = Math.max(1, Math.ceil(filteredItems.length / pageSize));
  const pagedItems = filteredItems.slice((page - 1) * pageSize, page * pageSize);

  async function submitSchedule() {
    const result = await createWorkflowSchedule({ project_id: projectID, template_version_id: templateVersionID, cron_expression: cronExpression, daily_content_count: Number(dailyContentCount) });
    if (result.success) {
      setLastCreatedScheduleID(result.data?.schedule_id ?? '');
      setStatusFilter('all');
      setPage(1);
      setToast(`创建成功：${result.data?.schedule_id}`);
      setShowCreateDialog(false);
      await loadSchedules();
    } else {
      setError(pageErrorFromEnvelope(result, '创建调度失败'));
    }
  }

  async function toggle(id: string, enabled: boolean) {
    const result = enabled ? await disableWorkflowSchedule(id, { reason: 'manual pause' }) : await enableWorkflowSchedule(id, { note: 'manual start' });
    if (result.success) {
      setToast(`状态已更新：${result.data?.operation_log_id}`);
      await loadSchedules();
    } else {
      setError(pageErrorFromEnvelope(result, '更新调度失败'));
    }
  }

  async function run(id: string) {
    const result = await testRunWorkflowSchedule(id, { input_override: { topic: 'manual' } });
    if (result.success) {
      setToast(`试跑已提交：${result.data?.workflow_run_id}`);
    } else {
      setError(pageErrorFromEnvelope(result, '试跑失败'));
    }
  }

  async function openDetail(item: WorkflowScheduleResponse) {
    setSelectedSchedule(item);
    const result = await fetchWorkflowScheduleTriggers(item.id);
    if (result.success && result.data) {
      setTriggers(result.data.items);
      setError(null);
    } else {
      setError(pageErrorFromEnvelope(result, '加载触发记录失败'));
    }
  }

  return (
    <main className="page-shell" data-testid="styled-page-shell">
      <section className="page-hero">
        <div className="page-hero__header">
          <div>
            <h1>生产计划 / 调度管理</h1>
            <p>配置可手动试跑的内容生产计划，跟踪 daily_content_count、启停状态和触发记录。</p>
          </div>
          <button type="button" onClick={() => setShowCreateDialog(true)}>新建调度</button>
        </div>
      </section>

      {error ? <div role="alert">{error.code ? `${error.code}：` : ''}{error.message}（request_id: {error.request_id ?? 'client'}）</div> : null}
      {toast ? <div role="status">{toast}</div> : null}

      <section className="card">
        <div className="toolbar">
          <label>状态筛选
            <select value={statusFilter} onChange={(event) => { setStatusFilter(event.target.value); setPage(1); }}>
              <option value="all">全部</option>
              <option value="enabled">启用</option>
              <option value="disabled">停用</option>
            </select>
          </label>
          <button type="button" onClick={() => {
            const target = items.find((item) => item.id === lastCreatedScheduleID) ?? pagedItems[0] ?? items[0];
            if (target) {
              void run(target.id);
            } else {
              setError({ message: '暂无可试跑调度', request_id: 'client' });
            }
          }}>试跑</button>
        </div>
      </section>

      <section className="card table-card">
        <table>
          <thead><tr><th>Project</th><th>daily_content_count</th><th>状态</th><th>下次运行</th><th>操作</th></tr></thead>
          <tbody>{pagedItems.map((item) => <tr key={item.id}><td>{item.project_id}</td><td>{item.daily_content_count}</td><td><span className={item.enabled ? 'badge badge--success' : 'badge badge--muted'}>{item.enabled ? '启用' : '停用'}</span></td><td>{item.next_run_at}</td><td className="action-row"><button type="button" onClick={() => toggle(item.id, item.enabled)}>{item.enabled ? '停用' : '启用'}</button><button type="button" onClick={() => openDetail(item)}>查看详情</button></td></tr>)}</tbody>
        </table>
        <div className="pagination">
          <span className="muted">第 {page} / {totalPages} 页，共 {filteredItems.length} 条</span>
          <div className="action-row">
            <button type="button" disabled={page <= 1} onClick={() => setPage((value) => value - 1)}>上一页</button>
            <button type="button" disabled={page >= totalPages} onClick={() => setPage((value) => value + 1)}>下一页</button>
          </div>
        </div>
      </section>

      {selectedSchedule ? (
        <section className="card" data-testid="schedule-detail">
          <div className="card__header"><h2>调度详情</h2><button type="button" onClick={() => setSelectedSchedule(null)}>关闭详情</button></div>
          <p>Schedule ID: {selectedSchedule.id}</p>
          <p>Cron: {selectedSchedule.cron_expression}</p>
          <h3>触发记录</h3>
          <ul>{triggers.map((trigger) => <li key={trigger.id}>{trigger.trigger_type} / {trigger.status} / {trigger.workflow_run_id ?? 'pending'}</li>)}</ul>
        </section>
      ) : null}

      {showCreateDialog ? (
        <div className="dialog-backdrop" role="presentation">
          <section className="dialog-panel" role="dialog" aria-modal="true" aria-label="新建调度弹窗">
            <div className="card__header"><h2>新建调度</h2><button type="button" onClick={() => setShowCreateDialog(false)}>关闭弹窗</button></div>
            <div className="form-grid">
              <label>Project ID<input value={projectID} onChange={(event) => setProjectID(event.target.value)} /></label>
              <label>Template Version ID<input value={templateVersionID} onChange={(event) => setTemplateVersionID(event.target.value)} /></label>
              <label>Cron<input value={cronExpression} onChange={(event) => setCronExpression(event.target.value)} /></label>
              <label>daily_content_count<input value={dailyContentCount} onChange={(event) => setDailyContentCount(event.target.value)} /></label>
            </div>
            <div className="action-row"><button type="button" onClick={submitSchedule}>提交调度</button></div>
          </section>
        </div>
      ) : null}
    </main>
  );
}
