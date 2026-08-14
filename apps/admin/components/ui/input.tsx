import * as React from 'react';
import { cn } from '../../lib/utils';

export const Input = React.forwardRef<HTMLInputElement, React.ComponentProps<'input'>>(function Input({ className, type, ...props }, ref) {
  return <input ref={ref} type={type} className={cn('flex min-h-11 w-full rounded-[14px] border border-border bg-surface px-3.5 py-2.5 text-sm text-foreground shadow-sm outline-none placeholder:text-muted-soft focus:border-primary focus:ring-4 focus:ring-primary/10 disabled:cursor-not-allowed disabled:opacity-50', className)} {...props} />;
});
