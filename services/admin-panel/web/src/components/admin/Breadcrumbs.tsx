'use client';

import { useMemo } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@shopee/ui-system';

interface BreadcrumbItem {
  label: string;
  href?: string;
}

const ROUTE_LABELS: Record<string, string> = {
  admin: 'Dashboard',
  products: 'Products',
  categories: 'Categories',
  inventory: 'Inventory',
  orders: 'Orders',
  users: 'Users',
  analytics: 'Analytics',
  settings: 'Settings',
  new: 'Add New',
  edit: 'Edit',
};

export function Breadcrumbs({ items, className }: { items?: BreadcrumbItem[]; className?: string }) {
  const pathname = usePathname();

  const breadcrumbs = useMemo(() => {
    if (items) return items;

    const segments = pathname.split('/').filter(Boolean);
    const crumbs: BreadcrumbItem[] = [];

    let currentPath = '';
    for (const segment of segments) {
      currentPath += `/${segment}`;
      
      if (/^\d+$/.test(segment) || /^[0-9a-f]{8}-[0-9a-f]{4}/.test(segment)) {
        crumbs.push({ label: 'Details', href: currentPath });
      } else {
        const label = ROUTE_LABELS[segment] || segment.charAt(0).toUpperCase() + segment.slice(1).replace(/-/g, ' ');
        crumbs.push({ label, href: currentPath });
      }
    }

    return crumbs;
  }, [items, pathname]);

  if (breadcrumbs.length === 0) return null;

  return (
    <nav className={cn('flex items-center gap-2 text-sm', className)}>
      <Link href="/admin" className="text-muted-foreground hover:text-foreground transition-colors">
        Home
      </Link>
      {breadcrumbs.map((crumb, index) => (
        <span key={index} className="flex items-center gap-2">
          <svg className="w-4 h-4 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
          </svg>
          {crumb.href && index < breadcrumbs.length - 1 ? (
            <Link href={crumb.href} className="text-muted-foreground hover:text-foreground transition-colors">
              {crumb.label}
            </Link>
          ) : (
            <span className="text-foreground font-medium">{crumb.label}</span>
          )}
        </span>
      ))}
    </nav>
  );
}
