import * as React from 'react';
import { Slot } from '@radix-ui/react-slot';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '../../lib/utils';

const buttonVariants = cva(
  'inline-flex min-h-11 items-center justify-center gap-2 whitespace-nowrap rounded-[14px] text-sm font-semibold transition-[background,box-shadow,transform,color,border-color] duration-200 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-primary/15 disabled:pointer-events-none disabled:opacity-45 active:scale-[.985]',
  {
    variants: {
      variant: {
        default: 'bg-primary text-white shadow-[0_8px_22px_rgba(7,148,85,.18)] hover:bg-primary-strong hover:shadow-[0_10px_26px_rgba(7,148,85,.24)]',
        accent: 'bg-primary text-white shadow-[0_8px_22px_rgba(7,148,85,.18)] hover:bg-primary-strong',
        destructive: 'bg-danger-soft text-danger hover:bg-[#fee4e2]',
        outline: 'border border-border bg-white text-foreground shadow-[0_1px_2px_rgba(16,24,40,.03)] hover:bg-surface-soft',
        ghost: 'text-muted hover:bg-surface-soft hover:text-foreground',
      },
      size: {
        default: 'px-4 py-2.5',
        sm: 'min-h-9 rounded-[12px] px-3 text-xs',
        lg: 'min-h-12 rounded-[16px] px-6 text-[15px]',
        icon: 'size-11 min-h-11 rounded-full p-0',
      },
    },
    defaultVariants: { variant: 'default', size: 'default' },
  },
);

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement>, VariantProps<typeof buttonVariants> { asChild?: boolean }
export function Button({ className, variant, size, asChild = false, ...props }: ButtonProps) {
  const Comp = asChild ? Slot : 'button';
  return <Comp className={cn(buttonVariants({ variant, size, className }))} {...props} />;
}
export { buttonVariants };
