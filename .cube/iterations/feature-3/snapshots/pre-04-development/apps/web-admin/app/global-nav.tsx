'use client';

import Link from 'next/link';
import { usePathname, useSearchParams } from 'next/navigation';

const navItems = [
  { href: '/', label: '首页 / 系统大盘', match: (pathname: string, search: string) => pathname === '/' && search === '' },
  { href: '/?view=projects', label: '项目管理', forceLoad: true, match: (pathname: string, search: string) => pathname === '/' && search === '?view=projects' },
  { href: '/?view=content-types', label: '项目模板管理', forceLoad: true, match: (pathname: string, search: string) => pathname === '/' && search === '?view=content-types' },
  { href: '/workflow/templates', label: '工作流模板', match: (pathname: string) => pathname === '/workflow/templates' || pathname.startsWith('/workflow/templates/') },
  { href: '/workflow/runs', label: '运行记录', match: (pathname: string) => pathname === '/workflow/runs' || pathname.startsWith('/workflow/runs/') },
  { href: '/workflow/schedules', label: '生产计划 / 调度管理', match: (pathname: string) => pathname === '/workflow/schedules' || pathname.startsWith('/workflow/schedules/') },
  { href: '/external-automation/n8n', label: '外部自动化 / n8n', match: (pathname: string) => pathname === '/external-automation/n8n' },
  { href: '/llm/cost-summary', label: '成本汇总', match: (pathname: string) => pathname === '/llm/cost-summary' },
  { href: '/agent/tasks', label: 'Agent 管理', match: (pathname: string) => pathname === '/agent/tasks' || pathname.startsWith('/agent/tasks/') },
  { href: '/llm/logs', label: 'LLM Logs', match: (pathname: string) => pathname === '/llm/logs' || pathname.startsWith('/llm/logs/') },
];

export default function GlobalNav() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const search = searchParams.toString() ? `?${searchParams.toString()}` : '';

  return (
    <nav aria-label="Iteration 2 navigation" className="app-nav">
      {navItems.map((item) => {
        const active = item.match(pathname, search);
        if ('forceLoad' in item && item.forceLoad) {
          return (
            <a key={item.href} href={item.href} aria-current={active ? 'page' : undefined} className="app-nav__link">
              {item.label}
            </a>
          );
        }
        return (
          <Link key={item.href} href={item.href} aria-current={active ? 'page' : undefined} className="app-nav__link">
            {item.label}
          </Link>
        );
      })}
    </nav>
  );
}
