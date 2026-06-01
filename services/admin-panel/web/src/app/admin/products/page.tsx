import { Suspense } from 'react';
import Link from 'next/link';
import { ErrorBoundary, DataTable, StatCard } from '@/components/admin';
import { ProductFilters } from './ProductFilters';
import type { ColumnDef } from '@/components/admin/DataTable';
import { formatVND, formatRelativeTime } from '@shopee/shared-utils';

interface Product {
  id: string;
  name: string;
  slug: string;
  price: number;
  sale_price?: number;
  quantity: number;
  category: string;
  brand: string;
  status: 'draft' | 'published' | 'archived';
  created_at: string;
  updated_at: string;
}

async function getProducts(searchParams: Record<string, string>) {
  const params = new URLSearchParams();
  Object.entries(searchParams).forEach(([key, value]) => {
    if (value && value !== 'all') params.set(key, value);
  });

  const res = await fetch(
    `${process.env.API_GATEWAY_URL || 'http://localhost:8080'}/api/admin/products?${params.toString()}`,
    {
      next: { revalidate: 30 },
      headers: { 'Content-Type': 'application/json' },
    }
  );

  if (!res.ok) {
    throw new Error('Failed to fetch products');
  }

  return res.json();
}

function ProductsLoading() {
  return <div className="h-96 bg-muted rounded-lg animate-pulse" />;
}

function ProductsTable({ products, total, page, perPage }: { products: Product[]; total: number; page: number; perPage: number }) {
  const columns: ColumnDef<Product>[] = [
    {
      key: 'name',
      header: 'Product',
      sortable: true,
      render: (row) => (
        <div>
          <p className="font-medium text-foreground">{row.name}</p>
          <p className="text-xs text-muted-foreground">{row.slug}</p>
        </div>
      ),
    },
    {
      key: 'price',
      header: 'Price',
      sortable: true,
      render: (row) => (
        <div>
          <span className="font-medium">{formatVND(row.price)}</span>
          {row.sale_price && row.sale_price < row.price && (
            <span className="ml-2 text-xs text-muted-foreground line-through">
              {formatVND(row.sale_price)}
            </span>
          )}
        </div>
      ),
    },
    {
      key: 'quantity',
      header: 'Stock',
      sortable: true,
      render: (row) => (
        <span
          className={
            row.quantity <= 10
              ? 'text-danger-600 font-medium'
              : row.quantity <= 50
              ? 'text-warning-600'
              : 'text-success-600'
          }
        >
          {row.quantity}
        </span>
      ),
    },
    {
      key: 'category',
      header: 'Category',
      render: (row) => (
        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400">
          {row.category}
        </span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (row) => {
        const styles = {
          published: 'bg-success-100 text-success-700 dark:bg-success-900/30 dark:text-success-400',
          draft: 'bg-muted text-muted-foreground',
          archived: 'bg-danger-100 text-danger-700 dark:bg-danger-900/30 dark:text-danger-400',
        };
        return (
          <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize ${styles[row.status]}`}>
            {row.status}
          </span>
        );
      },
    },
    {
      key: 'updated_at',
      header: 'Updated',
      sortable: true,
      render: (row) => formatRelativeTime(row.updated_at),
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (row) => (
        <div className="flex items-center gap-2">
          <Link
            href={`/admin/products/${row.id}`}
            className="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400"
          >
            Edit
          </Link>
          <span className="text-muted-foreground">|</span>
          <button className="text-sm text-danger-600 hover:text-danger-700 dark:text-danger-400">
            Delete
          </button>
        </div>
      ),
    },
  ];

  return (
    <DataTable
      data={products}
      columns={columns}
      emptyMessage="No products found. Create your first product to get started."
      pagination={{
        page,
        perPage,
        total,
        onPageChange: () => {},
      }}
      sorting={{
        sortBy: 'updated_at',
        sortOrder: 'desc',
        onSort: () => {},
      }}
    />
  );
}

async function ProductStats() {
  try {
    const res = await fetch(
      `${process.env.API_GATEWAY_URL || 'http://localhost:8080'}/api/admin/products/stats`,
      { next: { revalidate: 60 } }
    );

    if (!res.ok) throw new Error();

    const data = await res.json();

    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <StatCard title="Total Products" value={data.total || 0} />
        <StatCard
          title="Published"
          value={data.published || 0}
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          }
        />
        <StatCard
          title="Low Stock"
          value={data.low_stock || 0}
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
            </svg>
          }
        />
        <StatCard
          title="Drafts"
          value={data.drafts || 0}
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
          }
        />
      </div>
    );
  } catch {
    return null;
  }
}

export default async function ProductsPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string>>;
}) {
  const params = await searchParams;

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Products</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Manage your product catalog, inventory, and pricing
          </p>
        </div>
        <Link
          href="/admin/products/new"
          className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary-500 text-white font-medium text-sm hover:bg-primary-600 transition-colors"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Product
        </Link>
      </div>

      <ErrorBoundary>
        <Suspense fallback={null}>
          <ProductStats />
        </Suspense>
      </ErrorBoundary>

      <div className="rounded-xl border border-border bg-card p-4">
        <ProductFilters />
      </div>

      <ErrorBoundary>
        <Suspense fallback={<ProductsLoading />}>
          <ProductsContent searchParams={params} />
        </Suspense>
      </ErrorBoundary>
    </div>
  );
}

async function ProductsContent({ searchParams }: { searchParams: Record<string, string> }) {
  try {
    const data = await getProducts(searchParams);
    return (
      <ProductsTable
        products={data.products || []}
        total={data.total || 0}
        page={data.page || 1}
        perPage={data.per_page || 20}
      />
    );
  } catch {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <p>Unable to load products. Please try again later.</p>
      </div>
    );
  }
}
