'use client';

import Link from 'next/link';
import { cn } from '@shopee/ui-system';
import type { User } from '@shopee/shared-auth';

interface NavItem {
  label: string;
  href: string;
  icon: string;
}

interface AdminSidebarProps {
  items: NavItem[];
  isOpen: boolean;
  onToggle: () => void;
  currentPath: string;
  user: User | null;
  onLogout: () => void;
}

export function AdminSidebar({
  items,
  isOpen,
  currentPath,
  user,
  onLogout,
}: AdminSidebarProps) {
  return (
    <aside
      className={cn(
        'fixed top-0 left-0 z-50 h-full w-64 bg-card border-r border-border shadow-sidebar transition-transform duration-300',
        isOpen ? 'translate-x-0' : '-translate-x-full'
      )}
    >
      <div className="flex flex-col h-full">
        <div className="h-16 flex items-center justify-center border-b border-border">
          <Link href="/admin" className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-primary-500 flex items-center justify-center">
              <span className="text-white font-bold text-sm">T</span>
            </div>
            <span className="text-xl font-bold text-foreground">Tiki Admin</span>
          </Link>
        </div>

        <nav className="flex-1 overflow-y-auto scrollbar-thin p-4">
          <ul className="space-y-1">
            {items.map((item) => {
              const isActive = currentPath === item.href ||
                (item.href !== '/admin' && currentPath.startsWith(item.href));

              return (
                <li key={item.href}>
                  <Link
                    href={item.href}
                    className={cn(
                      'flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors',
                      isActive
                        ? 'bg-primary-500/10 text-primary-600 dark:text-primary-400'
                        : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                    )}
                  >
                    <span className="text-lg">{item.icon}</span>
                    <span>{item.label}</span>
                  </Link>
                </li>
              );
            })}
          </ul>
        </nav>

        <div className="border-t border-border p-4">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-10 h-10 rounded-full bg-primary-500 flex items-center justify-center text-white font-medium">
              {user?.display_name?.charAt(0)?.toUpperCase() ?? 'A'}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-foreground truncate">
                {user?.display_name ?? 'Admin User'}
              </p>
              <p className="text-xs text-muted-foreground truncate">
                {user?.email ?? 'admin@tiki.vn'}
              </p>
            </div>
          </div>
          <button
            onClick={onLogout}
            className="w-full flex items-center justify-center gap-2 px-3 py-2 text-sm font-medium text-muted-foreground hover:text-danger-600 hover:bg-danger-50 dark:hover:bg-danger-900/20 rounded-lg transition-colors"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
            Logout
          </button>
        </div>
      </div>
    </aside>
  );
}
