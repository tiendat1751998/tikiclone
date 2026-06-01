'use client';

import { type ReactNode, Suspense } from 'react';
import { cn } from '@shopee/ui-system';

interface ChartWrapperProps {
  title: string;
  subtitle?: string;
  children: ReactNode;
  isLoading?: boolean;
  error?: string | null;
  className?: string;
  actions?: ReactNode;
}

export function ChartWrapper({
  title,
  subtitle,
  children,
  isLoading = false,
  error = null,
  className,
  actions,
}: ChartWrapperProps) {
  return (
    <div className={cn('rounded-xl border border-border bg-card p-6', className)}>
      <div className="flex items-start justify-between mb-6">
        <div>
          <h3 className="text-lg font-semibold text-foreground">{title}</h3>
          {subtitle && (
            <p className="text-sm text-muted-foreground mt-1">{subtitle}</p>
          )}
        </div>
        {actions && <div className="flex items-center gap-2">{actions}</div>}
      </div>

      {error ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <div className="w-12 h-12 rounded-full bg-danger-100 dark:bg-danger-900/30 flex items-center justify-center mb-3">
            <svg className="w-6 h-6 text-danger-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
            </svg>
          </div>
          <p className="text-sm text-muted-foreground">{error}</p>
          <p className="text-xs text-muted-foreground mt-1">Data temporarily unavailable</p>
        </div>
      ) : isLoading ? (
        <div className="animate-pulse">
          <div className="h-48 bg-muted rounded-lg" />
        </div>
      ) : (
        <Suspense fallback={<div className="h-48 bg-muted rounded-lg animate-pulse" />}>
          {children}
        </Suspense>
      )}
    </div>
  );
}
