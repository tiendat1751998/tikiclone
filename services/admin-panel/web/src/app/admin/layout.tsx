'use client';

import { useEffect, useState, useCallback, useMemo } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useAuthStore } from '@shopee/shared-auth';
import { cn } from '@shopee/ui-system';
import { AdminSidebar } from '@/components/admin/AdminSidebar';
import { ThemeToggle } from '@/components/admin/ThemeToggle';

interface NavItem {
  label: string;
  href: string;
  icon: string;
  roles: string[];
  badge?: string;
}

const NAV_ITEMS: NavItem[] = [
  { label: 'Dashboard', href: '/admin', icon: '📊', roles: ['super_admin', 'product_manager', 'order_manager', 'viewer'] },
  { label: 'Products', href: '/admin/products', icon: '📦', roles: ['super_admin', 'product_manager'] },
  { label: 'Categories', href: '/admin/categories', icon: '🏷️', roles: ['super_admin', 'product_manager'] },
  { label: 'Inventory', href: '/admin/inventory', icon: '📋', roles: ['super_admin', 'product_manager'], badge: '3' },
  { label: 'Orders', href: '/admin/orders', icon: '🛒', roles: ['super_admin', 'order_manager'] },
  { label: 'Users', href: '/admin/users', icon: '👥', roles: ['super_admin'] },
  { label: 'Analytics', href: '/admin/analytics', icon: '📈', roles: ['super_admin', 'viewer'] },
  { label: 'Settings', href: '/admin/settings', icon: '⚙️', roles: ['super_admin'] },
];

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();
  const { isAuthenticated, user, logout } = useAuthStore();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    const checkMobile = () => {
      const mobile = window.innerWidth < 768;
      setIsMobile(mobile);
      if (mobile) setSidebarOpen(false);
    };
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  useEffect(() => {
    if (!isAuthenticated && pathname !== '/login') {
      router.push(`/login?redirect=${encodeURIComponent(pathname)}`);
    }
  }, [isAuthenticated, pathname, router]);

  const handleLogout = useCallback(() => {
    logout();
    router.push('/login');
  }, [logout, router]);

  const filteredNavItems = useMemo(() => {
    const role = (user as any)?.role ?? 'viewer';
    return NAV_ITEMS.filter(item => item.roles.includes(role));
  }, [user?.role]);

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-muted">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <AdminSidebar
        items={filteredNavItems}
        isOpen={sidebarOpen}
        onToggle={() => setSidebarOpen(!sidebarOpen)}
        currentPath={pathname}
        user={user}
        onLogout={handleLogout}
      />

      <div
        className={cn(
          'transition-all duration-300',
          sidebarOpen && !isMobile ? 'ml-64' : 'ml-0',
          sidebarOpen && isMobile ? 'ml-0' : ''
        )}
      >
        <header className="sticky top-0 z-30 h-16 border-b border-border bg-card/80 backdrop-blur-sm">
          <div className="flex h-full items-center justify-between px-4 md:px-6">
            <div className="flex items-center gap-3">
              <button
                onClick={() => setSidebarOpen(!sidebarOpen)}
                className="p-2 rounded-md hover:bg-muted transition-colors md:hidden"
                aria-label="Toggle sidebar"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                </svg>
              </button>
              <h1 className="text-lg font-semibold text-foreground hidden sm:block">
                Tiki Admin
              </h1>
            </div>

            <div className="flex items-center gap-3">
              <ThemeToggle />
              <div className="flex items-center gap-2">
                <div className="w-8 h-8 rounded-full bg-primary-500 flex items-center justify-center text-white text-sm font-medium">
                  {user?.display_name?.charAt(0)?.toUpperCase() ?? 'A'}
                </div>
                <span className="text-sm font-medium text-foreground hidden sm:block">
                  {user?.display_name ?? 'Admin'}
                </span>
              </div>
            </div>
          </div>
        </header>

        <main className="p-4 md:p-6 min-h-[calc(100vh-4rem)]">
          {children}
        </main>
      </div>

      {sidebarOpen && isMobile && (
        <div
          className="fixed inset-0 bg-black/50 z-40 md:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}
    </div>
  );
}
