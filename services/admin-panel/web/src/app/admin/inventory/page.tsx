import { Suspense } from 'react';
import { ErrorBoundary, DataTable, StatCard } from '@/components/admin';
import type { ColumnDef } from '@/components/admin/DataTable';
import { formatVND, formatRelativeTime } from '@shopee/shared-utils';

interface InventoryItem {
  id: string;
  product_id: string;
  product_name: string;
  product_sku: string;
  quantity: number;
  reserved_quantity: number;
  available_quantity: number;
  low_stock_threshold: number;
  warehouse: string;
  last_restocked: string;
  status: 'in_stock' | 'low_stock' | 'out_of_stock' | 'discontinued';
}

async function getInventory(searchParams: Record<string, string>) {
  const params = new URLSearchParams();
  Object.entries(searchParams).forEach(([key, value]) => {
    if (value && value !== 'all') params.set(key, value);
  });

  const res = await fetch(
    `${process.env.API_GATEWAY_URL || 'http://localhost:8080'}/api/admin/inventory?${params.toString()}`,
    {
      next: { revalidate: 30 },
      headers: { 'Content-Type': 'application/json' },
    }
  );

  if (!res.ok) throw new Error('Failed to fetch inventory');
  return res.json();
}

async function getInventoryStats() {
  const res = await fetch(
    `${process.env.API_GATEWAY_URL || 'http://localhost:8080'}/api/admin/inventory/stats`,
    { next: { revalidate: 60 } }
  );
  if (!res.ok) throw new Error();
  return res.json();
}

function InventoryLoading() {
  return <div className="h-96 bg-muted rounded-lg animate-pulse" />;
}

async function InventoryStatsSection() {
  try {
    const data = await getInventoryStats();
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <StatCard title="Total SKUs" value={data.total_sku || 0} />
        <StatCard
          title="In Stock"
          value={data.in_stock || 0}
          icon={<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg>}
        />
        <StatCard
          title="Low Stock"
          value={data.low_stock || 0}
          icon={<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" /></svg>}
        />
        <StatCard
          title="Out of Stock"
          value={data.out_of_stock || 0}
          icon={<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>}
        />
      </div>
    );
  } catch {
    return null;
  }
}

function InventoryTable({
  items,
  total,
  page,
  perPage,
}: {
  items: InventoryItem[];
  total: number;
  page: number;
  perPage: number;
}) {
  const columns: ColumnDef<InventoryItem>[] = [
    {
      key: 'product',
      header: 'Product',
      sortable: true,
      render: (row) => (
        <div>
          <p className="font-medium text-foreground">{row.product_name}</p>
          <p className="text-xs text-muted-foreground">SKU: {row.product_sku}</p>
        </div>
      ),
    },
    {
      key: 'warehouse',
      header: 'Warehouse',
      render: (row) => (
        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-muted text-muted-foreground">
          {row.warehouse}
        </span>
      ),
    },
    {
      key: 'quantity',
      header: 'Total Qty',
      sortable: true,
      render: (row) => <span className="font-medium">{row.quantity}</span>,
    },
    {
      key: 'available',
      header: 'Available',
      sortable: true,
      render: (row) => (
        <span className={row.available_quantity <= row.low_stock_threshold ? 'text-danger-600 font-medium' : ''}>
          {row.available_quantity}
        </span>
      ),
    },
    {
      key: 'reserved',
      header: 'Reserved',
      render: (row) => <span className="text-muted-foreground">{row.reserved_quantity}</span>,
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (row) => {
        const styles = {
          in_stock: 'bg-success-100 text-success-700 dark:bg-success-900/30 dark:text-success-400',
          low_stock: 'bg-warning-100 text-warning-700 dark:bg-warning-900/30 dark:text-warning-400',
          out_of_stock: 'bg-danger-100 text-danger-700 dark:bg-danger-900/30 dark:text-danger-400',
          discontinued: 'bg-muted text-muted-foreground',
        };
        return (
          <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize ${styles[row.status]}`}>
            {row.status.replace('_', ' ')}
          </span>
        );
      },
    },
    {
      key: 'last_restocked',
      header: 'Last Restocked',
      sortable: true,
      render: (row) => formatRelativeTime(row.last_restocked),
    },
  ];

  return (
    <DataTable
      data={items}
      columns={columns}
      emptyMessage="No inventory records found."
      pagination={{ page, perPage, total, onPageChange: () => {} }}
      sorting={{ sortBy: 'product_name', sortOrder: 'asc', onSort: () => {} }}
    />
  );
}

export default async function InventoryPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string>>;
}) {
  const params = await searchParams;

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Inventory</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Track stock levels, warehouses, and restock alerts
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button className="inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-border bg-card text-foreground font-medium text-sm hover:bg-muted transition-colors">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
            </svg>
            Export CSV
          </button>
          <button className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary-500 text-white font-medium text-sm hover:bg-primary-600 transition-colors">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Adjust Stock
          </button>
        </div>
      </div>

      <ErrorBoundary>
        <Suspense fallback={null}>
          <InventoryStatsSection />
        </Suspense>
      </ErrorBoundary>

      <div className="rounded-xl border border-border bg-card p-4">
        <InventoryFilters />
      </div>

      <ErrorBoundary>
        <Suspense fallback={<InventoryLoading />}>
          <InventoryContent searchParams={params} />
        </Suspense>
      </ErrorBoundary>
    </div>
  );
}

async function InventoryContent({ searchParams }: { searchParams: Record<string, string> }) {
  try {
    const data = await getInventory(searchParams);
    return (
      <InventoryTable
        items={data.items || []}
        total={data.total || 0}
        page={data.page || 1}
        perPage={data.per_page || 20}
      />
    );
  } catch {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <p>Unable to load inventory data. Please try again later.</p>
      </div>
    );
  }
}

function InventoryFilters() {
  return (
    <div className="flex flex-col sm:flex-row gap-3">
      <div className="flex-1">
        <div className="relative">
          <svg className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            placeholder="Search by product name or SKU..."
            className="w-full pl-10 pr-4 py-2 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
        </div>
      </div>
      <select className="px-3 py-2 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500">
        <option value="all">All Status</option>
        <option value="in_stock">In Stock</option>
        <option value="low_stock">Low Stock</option>
        <option value="out_of_stock">Out of Stock</option>
      </select>
      <select className="px-3 py-2 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500">
        <option value="">All Warehouses</option>
        <option value="hanoi">Hanoi Warehouse</option>
        <option value="hcmc">HCMC Warehouse</option>
        <option value="danang">Danang Warehouse</option>
      </select>
      <label className="flex items-center gap-2 cursor-pointer px-3 py-2">
        <input type="checkbox" className="w-4 h-4 rounded border-border text-primary-500 focus:ring-primary-500" />
        <span className="text-sm text-foreground">Low stock only</span>
      </label>
    </div>
  );
}
