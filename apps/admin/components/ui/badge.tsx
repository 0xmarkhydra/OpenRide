import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '../../lib/utils';

const badgeVariants = cva('inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold', {
  variants: {
    variant: {
      default: 'bg-primary text-white',
      secondary: 'bg-surface-soft text-muted',
      success: 'bg-success-soft text-success',
      warning: 'bg-warning-soft text-warning',
      destructive: 'bg-danger-soft text-danger',
      outline: 'border border-border bg-white text-muted',
    },
  },
  defaultVariants: { variant: 'default' },
});

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof badgeVariants> {}
export function Badge({ className, variant, ...props }: BadgeProps) {
  return <div className={cn(badgeVariants({ variant }), className)} {...props} />;
}
