'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

const items = [
  { href: 'planning', label: '内容规划' },
  { href: 'planning/topics', label: '候选选题' },
  { href: 'novel/worldview', label: '世界观' },
  { href: 'novel/characters', label: '人物' },
  { href: 'novel/arcs', label: '大纲' },
  { href: 'production', label: '内容生产' },
  { href: 'content-items', label: '内容单元' },
  { href: 'reviews', label: '审稿中心' },
  { href: 'publish-jobs', label: '发布队列' },
  { href: 'metrics', label: '指标表现' },
  { href: 'metrics/input', label: '指标录入' },
  { href: 'metrics/trends', label: '趋势图' },
  { href: 'metrics/missing', label: '缺失提醒' },
  { href: 'article/config', label: 'Article 配置' },
  { href: 'article/generation-runs', label: 'Article 生成' },
  { href: 'article/metrics', label: 'Article 指标' },
  { href: 'social-post', label: 'Social Post' },
  { href: 'social-post/variants', label: 'Social 候选文案' },
  { href: 'social-post/assets', label: 'Social 资产' },
  { href: 'strategy-suggestions', label: '策略建议' },
  { href: 'memory', label: '记忆上下文' },
  { href: 'memory/context-preview', label: '上下文预览' },
  { href: 'consistency-reports', label: '一致性报告' },
];

export default function ProjectWorkspaceNav({ projectId }: { projectId: string }) {
  const pathname = usePathname();

  return (
    <nav aria-label="项目工作区导航" className="app-nav">
      {items.map((item) => {
        const href = `/projects/${projectId}/${item.href}`;
        const active = pathname === href;
        return <Link key={href} href={href} aria-current={active ? 'page' : undefined} className="app-nav__link">{item.label}</Link>;
      })}
    </nav>
  );
}
