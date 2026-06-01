import { Suspense } from 'react';
import Link from 'next/link';
import { ErrorBoundary, DataTable } from '@/components/admin';
import type { ColumnDef } from '@/components/admin/DataTable';
import { formatVND, formatRelativeTime } from '@shopee/shared-utils';

interface Order {
  id: string;
  order_number: string;
  customer_name: string;
  customer_email: string;
  total: number;
  status: 'pending' | 'confirmed' | 'processing' | 'shipped' | 'delivered' | 'cancelled' | 'refunded';
  item_count: number;
  created_at: string;
}

async function getOrders(searchParams: Record<string, string>) {
  const params = new URLSearchParams();
  Object.entries(searchParams).forEach(([key, value]) => {
    if (value && value !== 'all') params.set(key, value);
  });

  const res = await fetch(
    `${process.env.API_GATEWAY_URL || 'http://localhost:8080'}/api/admin/orders?${params.toString()}`,
    {
      next: { revalidate: 30 },
      headers: { 'Content-Type': 'application/json' },
    }
  );

  if (!res.ok) {
    throw new Error('Failed to fetch orders');
  }

  return res.json();
}

function OrdersLoading() {
  return <div className="h-96 bg-muted rounded-lg animate-pulse" />;
}

function OrdersTable({
  orders,
  total,
  page,
  perPage,
}: {
  orders: Order[];
  total: number;
  page: number;
  perPage: number;
}) {
  const columns: ColumnDef<Order>[] = [
    {
      key: 'order_number',
      header: 'Order ID',
      sortable: true,
      render: (row) => (
        <Link
          href={`/admin/orders/${row.id}`}
          className="text-primary-600 hover:text-primary-700 dark:text-primary-400 font-medium"
        >
          #{row.order_number}
        </Link>
      ),
    },
    {
      key: 'customer',
      header: 'Customer',
      render: (row) => (
        <div>
          <p className="font-medium text-foreground">{row.customer_name}</p>
          <p className="text-xs text-muted-foreground">{row.customer_email}</p>
        </div>
      ),
    },
    {
      key: 'total',
      header: 'Total',
      sortable: true,
      render: (row) => (
        <span className="font-medium">{formatVND(row.total)}</span>
      ),
    },
    {
      key: 'items',
      header: 'Items',
      render: (row) => (
        <span className="text-muted-foreground">{row.item_count} items</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (row) => {
        const styles = {
          pending: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
          confirmed: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
          processing: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
          shipped: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/30 dark:text-cyan-400',
          delivered: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
          cancelled: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
          refunded: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
        };
        return (
          <span
            className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium capitalize ${styles[row.status]}`}
          >
            {row.status}
          </span>
        );
      },
    },
    {
      key: 'created_at',
      header: 'Date',
      sortable: true,
      render: (row) => formatRelativeTime(row.created_at),
    },
  ];

  return (
    <DataTable
      data={orders}
      columns={columns}
      emptyMessage="No orders found. Orders will appear here when customers make purchases."
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
      onRowClick={(row) => {
        window.location.href = `/admin/orders/${row.id}`;
      }}
    />
  );
}

export default async function OrdersPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string>>;
}) {
  const params = await searchParams;

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Orders</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Manage and track customer orders
          </p>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-card p-4">
        <OrderFilters />
      </div>

      <ErrorBoundary>
        <Suspense fallback={<OrdersLoading />}>
          <OrdersContent searchParams={params} />
        </Suspense>
      </ErrorBoundary>
    </div>
  );
}

async function OrdersContent({ searchParams }: { searchParams: Record<string, string> }) {
  try {
    const data = await getOrders(searchParams);
    return (
      <OrdersTable
        orders={data.orders || []}
        total={data.total || 0}
        page={data.page || 1}
        perPage={data.per_page || 20}
      />
    );
  } catch {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <p>Unable to load orders. Please try again later.</p>
      </div>
    );
  }
}

function OrderFilters() {
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
            placeholder="Search by order ID, customer name..."
            className="w-full pl-10 pr-4 py-2 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
        </div>
      </div>

      <select className="px-3 py-2 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500">
        <option value="all">All Status</option>
        <option value="pending">Pending</option>
        <option value="confirmed">Confirmed</option>
        <option value="processing">Processing</option>
        <option value="shipped">Shipped</option>
        <option value="delivered">Delivered</option>
        <option value="cancelled">Cancelled</option>
        <option value="refunded">Refunded</option>
      </select>

      <input
        type="date"
        className="px-3 py-2 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
      />
    </div>
  );
}
