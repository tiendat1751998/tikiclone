import { Suspense } from 'react';
import Link from 'next/link';
import { ErrorBoundary, DataTable } from '@/components/admin';
import type { ColumnDef } from '@/components/admin/DataTable';
import { formatRelativeTime } from '@shopee/shared-utils';

interface User {
  id: string;
  name: string;
  email: string;
  phone: string;
  role: string;
  status: 'active' | 'banned' | 'pending';
  order_count: number;
  total_spent: number;
  created_at: string;
  last_login: string;
}

async function getUsers(searchParams: Record<string, string>) {
  const params = new URLSearchParams();
  Object.entries(searchParams).forEach(([key, value]) => {
    if (value && value !== 'all') params.set(key, value);
  });

  const res = await fetch(
    `${process.env.API_GATEWAY_URL || 'http://localhost:8080'}/api/admin/users?${params.toString()}`,
    {
      next: { revalidate: 30 },
      headers: { 'Content-Type': 'application/json' },
    }
  );

  if (!res.ok) {
    throw new Error('Failed to fetch users');
  }

  return res.json();
}

function UsersLoading() {
  return <div className="h-96 bg-muted rounded-lg animate-pulse" />;
}

function UsersTable({
  users,
  total,
  page,
  perPage,
}: {
  users: User[];
  total: number;
  page: number;
  perPage: number;
}) {
  const columns: ColumnDef<User>[] = [
    {
      key: 'name',
      header: 'User',
      sortable: true,
      render: (row) => (
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-primary-500 flex items-center justify-center text-white font-medium">
            {row.name.charAt(0).toUpperCase()}
          </div>
          <div>
            <p className="font-medium text-foreground">{row.name}</p>
            <p className="text-xs text-muted-foreground">{row.email}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'phone',
      header: 'Phone',
      render: (row) => (
        <span className="text-muted-foreground">{row.phone || '-'}</span>
      ),
    },
    {
      key: 'role',
      header: 'Role',
      render: (row) => (
        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400 capitalize">
          {row.role}
        </span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (row) => {
        const styles = {
          active: 'bg-success-100 text-success-700 dark:bg-success-900/30 dark:text-success-400',
          banned: 'bg-danger-100 text-danger-700 dark:bg-danger-900/30 dark:text-danger-400',
          pending: 'bg-warning-100 text-warning-700 dark:bg-warning-900/30 dark:text-warning-400',
        };
        return (
          <span
            className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize ${styles[row.status]}`}
          >
            {row.status}
          </span>
        );
      },
    },
    {
      key: 'orders',
      header: 'Orders',
      sortable: true,
      render: (row) => (
        <span className="text-muted-foreground">{row.order_count}</span>
      ),
    },
    {
      key: 'last_login',
      header: 'Last Login',
      sortable: true,
      render: (row) => formatRelativeTime(row.last_login),
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (row) => (
        <div className="flex items-center gap-2">
          <button className="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400">
            View
          </button>
          {row.status === 'active' ? (
            <button className="text-sm text-danger-600 hover:text-danger-700 dark:text-danger-400">
              Ban
            </button>
          ) : (
            <button className="text-sm text-success-600 hover:text-success-700 dark:text-success-400">
              Unban
            </button>
          )}
        </div>
      ),
    },
  ];

  return (
    <DataTable
      data={users}
      columns={columns}
      emptyMessage="No users found."
      pagination={{
        page,
        perPage,
        total,
        onPageChange: () => {},
      }}
      sorting={{
        sortBy: 'created_at',
        sortOrder: 'desc',
        onSort: () => {},
      }}
    />
  );
}

export default async function UsersPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string>>;
}) {
  const params = await searchParams;

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Users</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Manage customer accounts and permissions
          </p>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-card p-4">
        <UserFilters />
      </div>

      <ErrorBoundary>
        <Suspense fallback={<UsersLoading />}>
          <UsersContent searchParams={params} />
        </Suspense>
      </ErrorBoundary>
    </div>
  );
}

async function UsersContent({ searchParams }: { searchParams: Record<string, string> }) {
  try {
    const data = await getUsers(searchParams);
    return (
      <UsersTable
        users={data.users || []}
        total={data.total || 0}
        page={data.page || 1}
        perPage={data.per_page || 20}
      />
    );
  } catch {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <p>Unable to load users. Please try again later.</p>
      </div>
    );
  }
}

function UserFilters() {
  return (
    <div className="flex flex-col sm:flex-row gap-3">
      <div className="flex-1">
        <div className="relative">
          <svg
            className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
          <input
            type="text"
            placeholder="Search by name, email, phone..."
            className="w-full pl-10 pr-4 py-2 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
        </div>
      </div>

      <select className="px-3 py-2 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500">
        <option value="all">All Status</option>
        <option value="active">Active</option>
        <option value="banned">Banned</option>
        <option value="pending">Pending</option>
      </select>

      <select className="px-3 py-2 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500">
        <option value="">All Roles</option>
        <option value="customer">Customer</option>
        <option value="vip">VIP</option>
        <option value="wholesale">Wholesale</option>
      </select>
    </div>
  );
}
