import { Suspense } from 'react';
import GlobalNav from './global-nav';

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <Suspense fallback={null}>
          <GlobalNav />
        </Suspense>
        <Suspense fallback={null}>
          {children}
        </Suspense>
      </body>
    </html>
  );
}
