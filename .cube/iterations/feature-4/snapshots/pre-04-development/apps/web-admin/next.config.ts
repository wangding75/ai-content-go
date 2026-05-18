import type { NextConfig } from 'next';

const allowedApiProxyTargets = new Set(['127.0.0.1', 'localhost', '[::1]']);

function apiProxyTarget() {
  const target = process.env.API_PROXY_TARGET ?? 'http://127.0.0.1:18080';
  const url = new URL(target);
  if (url.protocol !== 'http:' || !allowedApiProxyTargets.has(url.hostname)) {
    throw new Error('API_PROXY_TARGET must be an HTTP localhost URL');
  }
  return url.origin;
}

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: '/api/v1/:path*',
        destination: `${apiProxyTarget()}/api/v1/:path*`,
      },
    ];
  },
};

export default nextConfig;
