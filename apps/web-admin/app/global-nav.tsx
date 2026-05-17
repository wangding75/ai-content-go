'use client';

import Link from 'next/link';
import { usePathname, useSearchParams } from 'next/navigation';

const navItems = [
  { href: '/', label: '首页 / 系统大盘', match: (pathname: string, search: string) => pathname === '/' && search === '' },
  { href: '/?view=projects', label: '项目管理', forceLoad: true, match: (pathname: string, search: string) => pathname === '/' && search === '?view=projects' },
  { href: '/?view=content-types', label: '项目模板管理', forceLoad: true, match: (pathname: string, search: string) => pathname === '/' && search === '?view=content-types' },
  { href: '/workflow/templates', label: '工作流模板', match: (pathname: string) => pathname === '/workflow/templates' || pathname.startsWith('/workflow/templates/') },
  { href: '/workflow/runs', label: '运行记录', match: (pathname: string) => pathname === '/workflow/runs' || pathname.startsWith('/workflow/runs/') },
  { href: '/agent/tasks', label: 'Agent 管理', match: (pathname: string) => pathname === '/agent/tasks' || pathname.startsWith('/agent/tasks/') },
  { href: '/llm/logs', label: 'LLM Logs', match: (pathname: string) => pathname === '/llm/logs' || pathname.startsWith('/llm/logs/') },
];

export default function GlobalNav() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const search = searchParams.toString() ? `?${searchParams.toString()}` : '';

  return (
    <nav
      aria-label="Iteration 2 navigation"
      style={{
        alignItems: 'center',
        borderBottom: '1px solid #e5e7eb',
        display: 'flex',
        flexWrap: 'wrap',
        gap: 8,
        padding: '12px 24px',
      }}
    >
      {navItems.map((item) => {
        const active = item.match(pathname, search);
        const linkStyle = {
          background: active ? '#2563eb' : '#ffffff',
          border: `1px solid ${active ? '#2563eb' : '#dbeafe'}`,
          borderRadius: 6,
          color: active ? '#ffffff' : '#1d4ed8',
          padding: '6px 10px',
          textDecoration: 'none',
        };
        if ('forceLoad' in item && item.forceLoad) {
          return (
            <a key={item.href} href={item.href} aria-current={active ? 'page' : undefined} style={linkStyle}>
              {item.label}
            </a>
          );
        }
        return (
          <Link key={item.href} href={item.href} aria-current={active ? 'page' : undefined} style={linkStyle}>
            {item.label}
          </Link>
        );
      })}
    </nav>
  );
}
