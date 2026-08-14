'use client';

import { useEffect, type ReactNode } from 'react';
import { X } from 'lucide-react';
import { Button } from './ui/button';

export function DetailPanel({
  open,
  title,
  description,
  onClose,
  children,
}: {
  open: boolean;
  title: string;
  description?: string;
  onClose: () => void;
  children: ReactNode;
}) {
  useEffect(() => {
    if (!open) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [open, onClose]);

  if (!open) return null;
  return (
    <div className='fixed inset-0 z-[70] bg-foreground/20 p-2 backdrop-blur-sm md:p-4' onMouseDown={onClose}>
      <section
        role='dialog'
        aria-modal='true'
        aria-labelledby='detail-panel-title'
        className='ml-auto flex h-full w-full max-w-2xl flex-col overflow-hidden rounded-[28px] border border-border bg-surface shadow-lg'
        onMouseDown={event => event.stopPropagation()}
      >
        <header className='flex items-start justify-between gap-4 border-b border-border px-5 py-4 md:px-6'>
          <div>
            <h2 id='detail-panel-title' className='text-xl font-bold tracking-tight text-foreground'>{title}</h2>
            {description ? <p className='mt-1 text-sm leading-6 text-muted'>{description}</p> : null}
          </div>
          <Button size='icon' variant='ghost' onClick={onClose} aria-label='Đóng chi tiết'><X aria-hidden='true' className='size-5' /></Button>
        </header>
        <div className='min-h-0 flex-1 overflow-y-auto p-5 md:p-6'>{children}</div>
      </section>
    </div>
  );
}
