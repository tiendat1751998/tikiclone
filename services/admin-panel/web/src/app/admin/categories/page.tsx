import { Suspense } from 'react';
import Link from 'next/link';
import { ErrorBoundary, DataTable } from '@/components/admin';
import type { ColumnDef } from '@/components/admin/DataTable';
import { formatRelativeTime } from '@shopee/shared-utils';

interface Category {
  id: string;
  name: string;
  slug: string;
  parent_id: string | null;
  parent_name: string | null;
  product_count: number;
  is_active: boolean;
  sort_order: number;
  created_at: string;
}

async function getCategories(searchParams: Record<string, string>) {
  const params = new URLSearchParams();
  Object.entries(searchParams).forEach(([key, value]) => {
    if (value && value !== 'all') params.set(key, value);
  });

  const res = await fetch(
    `${process.env.API_GATEWAY_URL || 'http://localhost:8080'}/api/admin/categories?${params.toString()}`,
    { next: { revalidate: 60 }, headers: { 'Content-Type': 'application/json' } }
  );

  if (!res.ok) throw new Error('Failed to fetch categories');
  return res.json();
}

function CategoriesLoading() {
  return <div className="h-96 bg-muted rounded-lg animate-pulse" />;
}

function CategoriesTable({
  categories,
  total,
  page,
  perPage,
}: {
  categories: Category[];
  total: number;
  page: number;
  perPage: number;
}) {
  const columns: ColumnDef<Category>[] = [
    {
      key: 'name',
      header: 'Category',
      sortable: true,
      render: (row) => (
        <div>
          <p className="font-medium text-foreground">{row.name}</p>
          <p className="text-xs text-muted-foreground">/{row.slug}</p>
        </div>
      ),
    },
    {
      key: 'parent',
      header: 'Parent',
      render: (row) => row.parent_name ? (
        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-muted text-muted-foreground">
          {row.parent_name}
        </span>
      ) : (
        <span className="text-muted-foreground">—</span>
      ),
    },
    {
      key: 'products',
      header: 'Products',
      sortable: true,
      render: (row) => (
        <span className="text-muted-foreground">{row.product_count}</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (row) => (
        <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${row.is_active ? 'bg-success-100 text-success-700 dark:bg-success-900/30 dark:text-success-400' : 'bg-muted text-muted-foreground'}`}>
          {row.is_active ? 'Active' : 'Inactive'}
        </span>
      ),
    },
    {
      key: 'sort_order',
      header: 'Order',
      sortable: true,
      render: (row) => <span className="text-muted-foreground">{row.sort_order}</span>,
    },
    {
      key: 'created_at',
      header: 'Created',
      sortable: true,
      render: (row) => formatRelativeTime(row.created_at),
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (row) => (
        <div className="flex items-center gap-2">
          <Link href={`/admin/categories/${row.id}`} className="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400">
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
      data={categories}
      columns={columns}
      emptyMessage="No categories found. Create your first category to organize products."
      pagination={{ page, perPage, total, onPageChange: () => {} }}
      sorting={{ sortBy: 'sort_order', sortOrder: 'asc', onSort: () => {} }}
    />
  );
}

export default async function CategoriesPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string>>;
}) {
  const params = await searchParams;

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Categories</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Organize your product catalog with hierarchical categories
          </p>
        </div>
        <Link
          href="/admin/categories/new"
          className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary-500 text-white font-medium text-sm hover:bg-primary-600 transition-colors"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Category
        </Link>
      </div>

      <div className="rounded-xl border border-border bg-card p-4">
        <CategoryFilters />
      </div>

      <ErrorBoundary>
        <Suspense fallback={<CategoriesLoading />}>
          <CategoriesContent searchParams={params} />
        </Suspense>
      </ErrorBoundary>
    </div>
  );
}

async function CategoriesContent({ searchParams }: { searchParams: Record<string, string> }) {
  try {
    const data = await getCategories(searchParams);
    return (
      <CategoriesTable
        categories={data.categories || []}
        total={data.total || 0}
        page={data.page || 1}
        perPage={data.per_page || 20}
      />
    );
  } catch {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <p>Unable to load categories. Please try again later.</p>
      </div>
    );
  }
}

function CategoryFilters() {
  return (
    <div className="flex flex-col sm:flex-row gap-3">
      <div className="flex-1">
        <div className="relative">
          <svg className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            placeholder="Search categories..."
            className="w-full pl-10 pr-4 py-2 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
        </div>
      </div>
      <select className="px-3 py-2 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500">
        <option value="all">All Status</option>
        <option value="active">Active</option>
        <option value="inactive">Inactive</option>
      </select>
      <select className="px-3 py-2 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500">
        <option value="">All Levels</option>
        <option value="root">Root Categories</option>
        <option value="child">Subcategories</option>
      </select>
    </div>
  );
}
