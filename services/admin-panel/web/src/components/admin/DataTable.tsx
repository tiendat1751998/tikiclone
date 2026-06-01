'use client';

import { useState, useCallback, useMemo, type ReactNode } from 'react';
import { cn } from '@shopee/ui-system';

interface DataTableProps<T> {
  data: T[];
  columns: ColumnDef<T>[];
  isLoading?: boolean;
  emptyMessage?: string;
  onRowClick?: (row: T) => void;
  pagination?: {
    page: number;
    perPage: number;
    total: number;
    onPageChange: (page: number) => void;
  };
  sorting?: {
    sortBy: string;
    sortOrder: 'asc' | 'desc';
    onSort: (column: string) => void;
  };
  className?: string;
}

export interface ColumnDef<T> {
  key: string;
  header: string;
  width?: string;
  sortable?: boolean;
  render: (row: T) => ReactNode;
}

export function DataTable<T extends { id: string }>({
  data,
  columns,
  isLoading = false,
  emptyMessage = 'No data available',
  onRowClick,
  pagination,
  sorting,
  className,
}: DataTableProps<T>) {
  const renderSortIcon = useCallback((column: ColumnDef<T>) => {
    if (!column.sortable || !sorting) return null;
    const isActive = sorting.sortBy === column.key;
    return (
      <span className="ml-1 inline-flex flex-col">
        <svg
          className={cn(
            'w-3 h-3 -mb-1',
            isActive && sorting.sortOrder === 'asc'
              ? 'text-primary-500'
              : 'text-muted-foreground/40'
          )}
          fill="currentColor"
          viewBox="0 0 20 20"
        >
          <path d="M5 12l5-5 5 5H5z" />
        </svg>
        <svg
          className={cn(
            'w-3 h-3',
            isActive && sorting.sortOrder === 'desc'
              ? 'text-primary-500'
              : 'text-muted-foreground/40'
          )}
          fill="currentColor"
          viewBox="0 0 20 20"
        >
          <path d="M5 8l5 5 5-5H5z" />
        </svg>
      </span>
    );
  }, [sorting]);

  const totalPages = pagination ? Math.ceil(pagination.total / pagination.perPage) : 0;
  const startItem = pagination ? (pagination.page - 1) * pagination.perPage + 1 : 0;
  const endItem = pagination ? Math.min(pagination.page * pagination.perPage, pagination.total) : 0;

  if (isLoading) {
    return (
      <div className={cn('w-full overflow-hidden rounded-lg border border-border', className)}>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-muted/50">
              <tr>
                {columns.map((col) => (
                  <th
                    key={col.key}
                    className="px-4 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider"
                    style={{ width: col.width }}
                  >
                    {col.header}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {Array.from({ length: 5 }).map((_, i) => (
                <tr key={i} className="border-t border-border">
                  {columns.map((col) => (
                    <td key={col.key} className="px-4 py-3">
                      <div className="h-4 bg-muted rounded animate-pulse" style={{ width: '80%' }} />
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    );
  }

  return (
    <div className={cn('w-full overflow-hidden rounded-lg border border-border bg-card', className)}>
      <div className="overflow-x-auto scrollbar-thin">
        <table className="w-full">
          <thead className="bg-muted/50">
            <tr>
              {columns.map((col) => (
                <th
                  key={col.key}
                  className={cn(
                    'px-4 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider',
                    col.sortable && sorting && 'cursor-pointer hover:text-foreground select-none'
                  )}
                  style={{ width: col.width }}
                  onClick={() => {
                    if (col.sortable && sorting) {
                      sorting.onSort(col.key);
                    }
                  }}
                >
                  <span className="inline-flex items-center">
                    {col.header}
                    {renderSortIcon(col)}
                  </span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.length === 0 ? (
              <tr>
                <td colSpan={columns.length} className="px-4 py-12 text-center">
                  <div className="flex flex-col items-center">
                    <svg className="w-12 h-12 text-muted-foreground/40 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                    </svg>
                    <p className="text-sm text-muted-foreground">{emptyMessage}</p>
                  </div>
                </td>
              </tr>
            ) : (
              data.map((row, index) => (
                <tr
                  key={row.id}
                  className={cn(
                    'border-t border-border transition-colors',
                    onRowClick && 'cursor-pointer hover:bg-muted/50',
                    index % 2 === 1 && 'bg-muted/20'
                  )}
                  onClick={() => onRowClick?.(row)}
                >
                  {columns.map((col) => (
                    <td key={col.key} className="px-4 py-3 text-sm text-foreground">
                      {col.render(row)}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {pagination && data.length > 0 && (
        <div className="flex flex-col sm:flex-row items-center justify-between px-4 py-3 border-t border-border bg-muted/20 gap-3">
          <p className="text-sm text-muted-foreground">
            Showing {startItem} to {endItem} of {pagination.total} results
          </p>
          <div className="flex items-center gap-2">
            <button
              onClick={() => pagination.onPageChange(pagination.page - 1)}
              disabled={pagination.page <= 1}
              className="px-3 py-1.5 text-sm font-medium rounded-md border border-border bg-card hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Previous
            </button>
            
            <div className="flex items-center gap-1">
              {generatePageNumbers(pagination.page, totalPages).map((page, idx) => (
                page === '...' ? (
                  <span key={`ellipsis-${idx}`} className="px-2 text-muted-foreground">...</span>
                ) : (
                  <button
                    key={page}
                    onClick={() => pagination.onPageChange(page as number)}
                    className={cn(
                      'w-8 h-8 text-sm font-medium rounded-md transition-colors',
                      pagination.page === page
                        ? 'bg-primary-500 text-white'
                        : 'border border-border bg-card hover:bg-muted'
                    )}
                  >
                    {page}
                  </button>
                )
              ))}
            </div>

            <button
              onClick={() => pagination.onPageChange(pagination.page + 1)}
              disabled={pagination.page >= totalPages}
              className="px-3 py-1.5 text-sm font-medium rounded-md border border-border bg-card hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

function generatePageNumbers(currentPage: number, totalPages: number): (number | string)[] {
  if (totalPages <= 7) {
    return Array.from({ length: totalPages }, (_, i) => i + 1);
  }

  const pages: (number | string)[] = [];
  
  if (currentPage <= 3) {
    pages.push(1, 2, 3, 4, '...', totalPages);
  } else if (currentPage >= totalPages - 2) {
    pages.push(1, '...', totalPages - 3, totalPages - 2, totalPages - 1, totalPages);
  } else {
    pages.push(1, '...', currentPage - 1, currentPage, currentPage + 1, '...', totalPages);
  }
  
  return pages;
}
