import type { ReactNode } from 'react';
import './globals.css';

export const metadata = {
  title: 'OpenRide Operator',
  description: 'Operations dashboard for the OpenRide open mobility marketplace',
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="vi">
      <body>{children}</body>
    </html>
  );
}
