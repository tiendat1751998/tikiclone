'use client';

import { Suspense, useState, useTransition, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { cn } from '@shopee/ui-system';
import { formatVND, formatRelativeTime } from '@shopee/shared-utils';
import { ErrorBoundary } from '@/components/admin';

interface UserDetail {
  id: string;
  name: string;
  email: string;
  phone: string;
  avatar_url: string;
  role: string;
  status: 'active' | 'banned' | 'pending';
  email_verified: boolean;
  phone_verified: boolean;
  two_factor_enabled: boolean;
  address: {
    street: string;
    city: string;
    district: string;
    ward: string;
  };
  stats: {
    total_orders: number;
    total_spent: number;
    avg_order_value: number;
    last_order_date: string;
    member_since: string;
  };
  recent_orders: {
    id: string;
    order_number: string;
    total: number;
    status: string;
    date: string;
  }[];
  activity_log: {
    id: string;
    action: string;
    description: string;
    ip_address: string;
    user_agent: string;
    timestamp: string;
  }[];
}

async function getUserDetail(id: string) {
  const res = await fetch(
    `${process.env.API_GATEWAY_URL || 'http://localhost:8080'}/api/admin/users/${id}`,
    { next: { revalidate: 30 }, headers: { 'Content-Type': 'application/json' } }
  );
  if (!res.ok) throw new Error('Failed to fetch user details');
  return res.json();
}

function UserDetailLoading() {
  return <div className="h-96 bg-muted rounded-lg animate-pulse" />;
}

export default async function UserDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Link href="/admin/users" className="hover:text-foreground">Users</Link>
        <span>/</span>
        <span className="text-foreground">User Details</span>
      </div>

      <ErrorBoundary>
        <Suspense fallback={<UserDetailLoading />}>
          <UserDetailContent id={id} />
        </Suspense>
      </ErrorBoundary>
    </div>
  );
}

async function UserDetailContent({ id }: { id: string }) {
  let user: UserDetail;
  try {
    const data = await getUserDetail(id);
    user = data.user || data;
  } catch {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <p>Unable to load user details. Please try again later.</p>
      </div>
    );
  }

  const statusStyles = {
    active: 'bg-success-100 text-success-700 dark:bg-success-900/30 dark:text-success-400',
    banned: 'bg-danger-100 text-danger-700 dark:bg-danger-900/30 dark:text-danger-400',
    pending: 'bg-warning-100 text-warning-700 dark:bg-warning-900/30 dark:text-warning-400',
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div className="lg:col-span-1 space-y-6">
        <div className="rounded-xl border border-border bg-card p-6">
          <div className="text-center">
            <div className="w-20 h-20 rounded-full bg-primary-500 flex items-center justify-center text-white text-2xl font-bold mx-auto mb-4">
              {user.avatar_url ? (
                <img src={user.avatar_url} alt={user.name} className="w-full h-full rounded-full object-cover" />
              ) : (
                user.name.charAt(0).toUpperCase()
              )}
            </div>
            <h2 className="text-xl font-bold text-foreground">{user.name}</h2>
            <p className="text-sm text-muted-foreground">{user.email}</p>
            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium mt-2 capitalize ${statusStyles[user.status]}`}>
              {user.status}
            </span>
          </div>

          <div className="mt-6 space-y-3">
            <div className="flex items-center justify-between py-2 border-b border-border">
              <span className="text-sm text-muted-foreground">Role</span>
              <span className="text-sm font-medium text-foreground capitalize">{user.role}</span>
            </div>
            <div className="flex items-center justify-between py-2 border-b border-border">
              <span className="text-sm text-muted-foreground">Phone</span>
              <span className="text-sm font-medium text-foreground">{user.phone || '—'}</span>
            </div>
            <div className="flex items-center justify-between py-2 border-b border-border">
              <span className="text-sm text-muted-foreground">Email Verified</span>
              <span className={cn('text-sm font-medium', user.email_verified ? 'text-success-600' : 'text-muted-foreground')}>
                {user.email_verified ? 'Yes' : 'No'}
              </span>
            </div>
            <div className="flex items-center justify-between py-2 border-b border-border">
              <span className="text-sm text-muted-foreground">2FA Enabled</span>
              <span className={cn('text-sm font-medium', user.two_factor_enabled ? 'text-success-600' : 'text-muted-foreground')}>
                {user.two_factor_enabled ? 'Yes' : 'No'}
              </span>
            </div>
            <div className="flex items-center justify-between py-2">
              <span className="text-sm text-muted-foreground">Member Since</span>
              <span className="text-sm font-medium text-foreground">{formatRelativeTime(user.stats.member_since)}</span>
            </div>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <h3 className="text-sm font-semibold text-foreground mb-3">Quick Actions</h3>
          <div className="space-y-2">
            <button className="w-full px-3 py-2 rounded-lg border border-border bg-card text-sm text-foreground hover:bg-muted transition-colors text-left">
              ✉️ Send Email
            </button>
            <button className="w-full px-3 py-2 rounded-lg border border-border bg-card text-sm text-foreground hover:bg-muted transition-colors text-left">
              📝 Add Note
            </button>
            {user.status === 'active' ? (
              <button className="w-full px-3 py-2 rounded-lg border border-danger-300 bg-danger-50 dark:bg-danger-900/20 text-sm text-danger-600 hover:bg-danger-100 dark:hover:bg-danger-900/30 transition-colors text-left">
                🚫 Ban User
              </button>
            ) : (
              <button className="w-full px-3 py-2 rounded-lg border border-success-300 bg-success-50 dark:bg-success-900/20 text-sm text-success-600 hover:bg-success-100 dark:hover:bg-success-900/30 transition-colors text-left">
                ✅ Unban User
              </button>
            )}
          </div>
        </div>
      </div>

      <div className="lg:col-span-2 space-y-6">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <div className="rounded-xl border border-border bg-card p-4">
            <p className="text-sm text-muted-foreground">Total Orders</p>
            <p className="text-2xl font-bold text-foreground mt-1">{user.stats.total_orders}</p>
          </div>
          <div className="rounded-xl border border-border bg-card p-4">
            <p className="text-sm text-muted-foreground">Total Spent</p>
            <p className="text-2xl font-bold text-foreground mt-1">{formatVND(user.stats.total_spent)}</p>
          </div>
          <div className="rounded-xl border border-border bg-card p-4">
            <p className="text-sm text-muted-foreground">Avg. Order</p>
            <p className="text-2xl font-bold text-foreground mt-1">{formatVND(user.stats.avg_order_value)}</p>
          </div>
          <div className="rounded-xl border border-border bg-card p-4">
            <p className="text-sm text-muted-foreground">Last Order</p>
            <p className="text-lg font-bold text-foreground mt-1">{formatRelativeTime(user.stats.last_order_date)}</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <h3 className="text-lg font-semibold text-foreground mb-4">Recent Orders</h3>
          {user.recent_orders.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-8">No orders yet</p>
          ) : (
            <div className="space-y-3">
              {user.recent_orders.map((order) => (
                <Link
                  key={order.id}
                  href={`/admin/orders/${order.id}`}
                  className="flex items-center justify-between p-3 rounded-lg border border-border bg-muted/20 hover:bg-muted/40 transition-colors"
                >
                  <div>
                    <p className="text-sm font-medium text-foreground">#{order.order_number}</p>
                    <p className="text-xs text-muted-foreground">{formatRelativeTime(order.date)}</p>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-sm font-medium text-foreground">{formatVND(order.total)}</span>
                    <StatusBadge status={order.status} />
                  </div>
                </Link>
              ))}
            </div>
          )}
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <h3 className="text-lg font-semibold text-foreground mb-4">Activity Log</h3>
          {user.activity_log.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-8">No activity recorded</p>
          ) : (
            <div className="space-y-4 max-h-96 overflow-y-auto scrollbar-thin">
              {user.activity_log.map((log) => (
                <div key={log.id} className="flex gap-4">
                  <div className="flex flex-col items-center">
                    <div className="w-2 h-2 rounded-full bg-primary-500 mt-2" />
                    <div className="w-0.5 h-full bg-border mt-1" />
                  </div>
                  <div className="pb-4 flex-1">
                    <p className="text-sm font-medium text-foreground">{log.action}</p>
                    <p className="text-xs text-muted-foreground">{log.description}</p>
                    <p className="text-xs text-muted-foreground mt-1">
                      {log.ip_address} • {formatRelativeTime(log.timestamp)}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    pending: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
    confirmed: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
    processing: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
    shipped: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/30 dark:text-cyan-400',
    delivered: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
    cancelled: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    refunded: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
  };
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize ${styles[status] || 'bg-muted text-muted-foreground'}`}>
      {status}
    </span>
  );
}
