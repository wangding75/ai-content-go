'use client';

import { useEffect } from 'react';

function hideRouteAnnouncer() {
  const announcer = document.getElementById('__next-route-announcer__');
  if (!announcer) {
    return;
  }
  announcer.setAttribute('aria-hidden', 'true');
  announcer.setAttribute('role', 'presentation');
  announcer.style.display = 'none';
}

export default function RouteAnnouncerGuard() {
  useEffect(() => {
    hideRouteAnnouncer();
    const observer = new MutationObserver(hideRouteAnnouncer);
    observer.observe(document.documentElement, { attributes: true, childList: true, subtree: true });
    const id = window.setInterval(hideRouteAnnouncer, 20);
    return () => {
      observer.disconnect();
      window.clearInterval(id);
    };
  }, []);

  return null;
}
