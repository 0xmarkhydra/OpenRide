import type { ReactNode } from 'react';
import './tailwind.css';
import './globals.css';

export const metadata = {
  title: 'FlashX Admin',
  description: 'Trung tâm vận hành ba dịch vụ FlashX',
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return <html lang='vi'><body>{children}</body></html>;
}
