'use client';

import { type ReactNode } from 'react';
import { cn } from '@shopee/ui-system';

interface StatCardProps {
  title: string;
  value: string | number;
  change?: number;
  changeLabel?: string;
  icon?: ReactNode;
  isLoading?: boolean;
  className?: string;
}

export function StatCard({
  title,
  value,
  change,
  changeLabel = 'vs last period',
  icon,
  isLoading = false,
  className,
}: StatCardProps) {
  if (isLoading) {
    return (
      <div className={cn('rounded-xl border border-border bg-card p-6', className)}>
        <div className="animate-pulse">
          <div className="h-4 bg-muted rounded w-1/2 mb-4" />
          <div className="h-8 bg-muted rounded w-2/3 mb-2" />
          <div className="h-3 bg-muted rounded w-1/3" />
        </div>
      </div>
    );
  }

  const isPositiveChange = change !== undefined && change > 0;
  const isNegativeChange = change !== undefined && change < 0;

  return (
    <div className={cn('rounded-xl border border-border bg-card p-6', className)}>
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <p className="text-sm font-medium text-muted-foreground">{title}</p>
          <p className="mt-2 text-3xl font-bold text-foreground">{value}</p>
          {change !== undefined && (
            <div className="mt-2 flex items-center gap-1">
              <span
                className={cn(
                  'text-sm font-medium',
                  isPositiveChange && 'text-success-600 dark:text-success-400',
                  isNegativeChange && 'text-danger-600 dark:text-danger-400',
                  !isPositiveChange && !isNegativeChange && 'text-muted-foreground'
                )}
              >
                {isPositiveChange && '+'}
                {change}%
              </span>
              <span className="text-xs text-muted-foreground">{changeLabel}</span>
            </div>
          )}
        </div>
        {icon && (
          <div className="p-3 rounded-lg bg-primary-500/10 text-primary-600 dark:text-primary-400">
            {icon}
          </div>
        )}
      </div>
    </div>
  );
}
