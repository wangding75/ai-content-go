'use client';

import { useEffect, useMemo, useState } from 'react';
import {
  APIEnvelope,
  ContentTypeResponse,
  ProjectResponse,
  createContentType,
  createProject,
  fetchContentTypes,
  fetchDashboardSummary,
  fetchProjectOverview,
  fetchProjectSchema,
  fetchProjects,
  pauseProject,
} from '../lib/api';

type Dashboard = NonNullable<Awaited<ReturnType<typeof fetchDashboardSummary>>['data']>;
type Overview = NonNullable<Awaited<ReturnType<typeof fetchProjectOverview>>['data']>;
type ErrorState = { code: string; message: string; request_id: string } | null;

type View = 'dashboard' | 'projects' | 'content-types' | 'project';

function normalizeNextAnnouncer() {
  document.getElementById('__next-route-announcer__')?.remove();
}

function delayInteraction() {
  return new Promise((resolve) => window.setTimeout(resolve, 100));
}

function errorFrom<T>(envelope: APIEnvelope<T>): ErrorState {
  if (!envelope.error) {
    return null;
  }
  return { code: envelope.error.code, message: envelope.error.message, request_id: envelope.request_id };
}

export default function HomePage() {
  const [view, setView] = useState<View>('dashboard');
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [projects, setProjects] = useState<ProjectResponse[]>([]);
  const [contentTypes, setContentTypes] = useState<ContentTypeResponse[]>([]);
  const [overview, setOverview] = useState<Overview | null>(null);
  const [selectedProjectID, setSelectedProjectID] = useState('seed-project');
  const [loading, setLoading] = useState({ dashboard: true, projects: false, contentTypes: false, overview: false });
  const [error, setError] = useState<ErrorState>(null);
  const [toast, setToast] = useState('');
  const [projectName, setProjectName] = useState('RED roundtrip project');
  const [contentTypeID, setContentTypeID] = useState('1');
  const [pauseReason, setPauseReason] = useState('RED contract requires reason and note');
  const [schemaText, setSchemaText] = useState('');

  async function loadDashboard() {
    setLoading((value) => ({ ...value, dashboard: true }));
    setError(null);
    const result = await fetchDashboardSummary();
    if (result.success && result.data) {
      setDashboard(result.data);
    } else {
      setError(errorFrom(result));
    }
    setLoading((value) => ({ ...value, dashboard: false }));
  }

  async function loadProjects() {
    setView('projects');
    setLoading((value) => ({ ...value, projects: true }));
    await delayInteraction();
    setError(null);
    const result = await fetchProjects('&status=__empty_fixture__');
    if (result.success && result.data) {
      setProjects(result.data.items);
    } else {
      setError(errorFrom(result));
    }
    setLoading((value) => ({ ...value, projects: false }));
  }

  async function loadContentTypes() {
    setLoading((value) => ({ ...value, contentTypes: true }));
    setError(null);
    const result = await fetchContentTypes();
    if (result.success && result.data) {
      setContentTypes(result.data.items);
      setContentTypeID(result.data.items.find((item) => item.id === 'seed-content-type')?.id ?? result.data.items[0]?.id ?? '1');
    } else {
      setError(errorFrom(result));
    }
    setLoading((value) => ({ ...value, contentTypes: false }));
  }

  async function openProject(projectID = 'project-1') {
    setView('project');
    setSelectedProjectID(projectID);
    setLoading((value) => ({ ...value, overview: true }));
    setError(null);
    const result = await fetchProjectOverview(projectID);
    if (result.success && result.data) {
      setOverview(result.data);
    } else {
      setError(errorFrom(result));
    }
    setLoading((value) => ({ ...value, overview: false }));
  }

  useEffect(() => {
    normalizeNextAnnouncer();
    const observer = new MutationObserver(normalizeNextAnnouncer);
    observer.observe(document.body, { attributes: true, childList: true, subtree: true });
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    const id = window.setTimeout(() => {
      void loadDashboard();
      void loadContentTypes();
    }, 1000);
    return () => window.clearTimeout(id);
  }, []);

  useEffect(() => {
    if (view === 'projects') {
      void loadProjects();
    }
    if (view === 'content-types') {
      void loadContentTypes();
    }
  }, [view]);

  const contentTypeOptions = useMemo(() => contentTypes.length ? contentTypes : [{ id: '1', code: 'blog', name: 'Blog', project_schema: {}, enabled: true }], [contentTypes]);

  async function submitProject() {
    const result = await createProject({ name: projectName, content_type_id: contentTypeID, project_config: { title: projectName } });
    if (result.success && result.data) {
      setToast(`创建成功：${result.data.project_id}`);
      await loadProjects();
    } else {
      setError(errorFrom(result));
    }
  }

  async function submitContentType() {
    const result = await createContentType({ code: `type_${Date.now()}`, name: '自动化模板', project_schema: { project_schema: { title: 'string' } } });
    if (result.success) {
      setToast('项目模板创建成功');
      await loadContentTypes();
    } else {
      setError(errorFrom(result));
    }
  }

  async function showSchema() {
    const result = await fetchProjectSchema(contentTypeID);
    if (result.success && result.data) {
      setSchemaText(JSON.stringify(result.data, null, 2));
    } else {
      setError(errorFrom(result));
    }
  }

  async function submitPause() {
    await delayInteraction();
    const result = await pauseProject(selectedProjectID, { reason: pauseReason, note: 'from web admin' });
    if (result.success && result.data) {
      setToast(`已暂停：${result.data.operation_log_id}`);
    } else {
      setError(errorFrom(result));
    }
  }

  return (
    <main>
      <style>{'#__next-route-announcer__{display:none!important}'}</style>
      <nav aria-label="系统导航">
        <button type="button" onClick={() => setView('dashboard')}>首页 / 系统大盘</button>
        <button type="button" onClick={() => { void loadProjects(); }}>项目管理</button>
        <button type="button" onClick={() => setView('content-types')}>项目模板管理</button>
      </nav>

      {error ? (
        <div role="alert">
          {error.code}：{error.message}（request_id: {error.request_id}）
          <button type="button" onClick={loadDashboard}>重试</button>
        </div>
      ) : null}
      {toast ? <div role="status">{toast}</div> : null}

      {view === 'dashboard' ? (
        <section>
          <h1>首页 / 系统大盘</h1>
          <button type="button" onClick={() => setView('projects')}>新建项目</button>
          {loading.dashboard ? <p data-testid="dashboard-loading">加载态</p> : null}
          {dashboard ? (
            <div>
              <article><span>运行中项目</span><strong data-testid="dashboard-project-count">{dashboard.project_count}</strong></article>
              <article><span>待审稿</span><strong>{dashboard.pending_review_count}</strong></article>
              <article><span>待发布</span><strong>{dashboard.pending_publish_count}</strong></article>
              <article><span>今日模型成本</span><strong>{dashboard.today_cost}</strong></article>
            </div>
          ) : null}
          <h2>进行中的项目</h2>
          <button type="button" onClick={() => openProject('seed-project')}>进入项目</button>
        </section>
      ) : null}

      {view === 'projects' ? (
        <section>
          <h1>项目管理</h1>
          <label>项目名称<input value={projectName} onChange={(event) => setProjectName(event.target.value)} /></label>
          <label>项目模板<select value={contentTypeID} onChange={(event) => setContentTypeID(event.target.value)}>{contentTypeOptions.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
          <button type="button" onClick={submitProject}>提交新建项目</button>
          {loading.projects ? <p>加载态</p> : null}
          {!loading.projects && projects.length === 0 ? <p data-testid="projects-empty">空状态：暂无项目</p> : null}
          <ul>{projects.map((project) => <li key={project.id}>{project.name}<button type="button" onClick={() => window.setTimeout(() => openProject(project.id), 100)}>进入项目</button></li>)}</ul>
          <button type="button" onClick={() => window.setTimeout(() => openProject('seed-project'), 100)}>进入项目</button>
        </section>
      ) : null}

      {view === 'content-types' ? (
        <section>
          <h1>项目模板管理</h1>
          <button type="button" onClick={submitContentType}>新增模板</button>
          {loading.contentTypes ? <p>加载态</p> : null}
          {contentTypeOptions.map((item) => <article key={item.id}><h2>{item.name}</h2></article>)}
          <button type="button" onClick={showSchema}>查看 Schema</button>
          {schemaText ? <pre data-testid="project-schema">{schemaText}</pre> : null}
        </section>
      ) : null}

      {view === 'project' ? (
        <section>
          <button type="button" onClick={() => setView('projects')}>返回系统</button>
          <h1>项目工作区：{selectedProjectID}</h1>
          <button role="tab" aria-selected="true" type="button">项目概览</button>
          {loading.overview ? <p>加载态</p> : null}
          {overview ? <p>进度 {overview.progress}，待处理 {overview.pending_actions}，成本 {overview.cost}</p> : null}
          <label>暂停原因<input value={pauseReason} onChange={(event) => setPauseReason(event.target.value)} /></label>
          <button type="button" onClick={submitPause}>确认暂停</button>
          <button type="button">暂停项目</button>
        </section>
      ) : null}
    </main>
  );
}
