import type { ReactNode } from 'react';
import './globals.css';

export const metadata = {
  title: 'FlashX - Tài xế của bạn, khi bạn cần',
  description: 'Nền tảng dịch vụ lái xe uy tín, an toàn. Lái hộ ô tô, lái hộ xe máy, đăng kiểm hộ.',
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="vi">
      <body>{children}</body>
    </html>
  );
}
