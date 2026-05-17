import { Suspense } from 'react';
import './globals.css';
import GlobalNav from './global-nav';

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <div className="app-shell">
          <Suspense fallback={null}>
            <GlobalNav />
          </Suspense>
          <Suspense fallback={null}>
            {children}
          </Suspense>
        </div>
      </body>
    </html>
  );
}
