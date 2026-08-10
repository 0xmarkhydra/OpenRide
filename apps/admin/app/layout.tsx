import type { ReactNode } from 'react';
import './globals.css';

export const metadata = {
  title: 'FlashX Admin',
  description: 'Operations dashboard for FlashX ride-hailing platform',
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="vi">
      <body>{children}</body>
    </html>
  );
}
