import { Suspense } from 'react';
import { ErrorBoundary, ChartWrapper, StatCard } from '@/components/admin';
import { formatVND, formatNumber } from '@shopee/shared-utils';

async function getAnalyticsData(period: string) {
  const res = await fetch(
    `${process.env.API_GATEWAY_URL || 'http://localhost:8080'}/api/admin/analytics?period=${period}`,
    {
      next: { revalidate: 300 },
      headers: { 'Content-Type': 'application/json' },
    }
  );

  if (!res.ok) {
    throw new Error('Failed to fetch analytics data');
  }

  return res.json();
}

export default async function AnalyticsPage({
  searchParams,
}: {
  searchParams: Promise<{ period?: string }>;
}) {
  const params = await searchParams;
  const period = params.period || '30d';

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Analytics</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Detailed insights into your business performance
          </p>
        </div>
        <div className="flex items-center gap-2">
          {['7d', '30d', '90d', '1y'].map((p) => (
            <a
              key={p}
              href={`/admin/analytics?period=${p}`}
              className={`px-3 py-1.5 text-sm font-medium rounded-lg transition-colors ${
                period === p
                  ? 'bg-primary-500 text-white'
                  : 'bg-muted text-muted-foreground hover:text-foreground'
              }`}
            >
              {p === '7d' ? '7 Days' : p === '30d' ? '30 Days' : p === '90d' ? '90 Days' : '1 Year'}
            </a>
          ))}
        </div>
      </div>

      <ErrorBoundary>
        <Suspense fallback={<AnalyticsLoading />}>
          <AnalyticsContent period={period} />
        </Suspense>
      </ErrorBoundary>
    </div>
  );
}

function AnalyticsLoading() {
  return (
    <>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="h-32 bg-muted rounded-xl animate-pulse" />
        ))}
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {Array.from({ length: 2 }).map((_, i) => (
          <div key={i} className="h-80 bg-muted rounded-xl animate-pulse" />
        ))}
      </div>
    </>
  );
}

async function AnalyticsContent({ period }: { period: string }) {
  try {
    const data = await getAnalyticsData(period);

    return (
      <>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <StatCard
            title="Total Revenue"
            value={formatVND(data.revenue?.total || 0)}
            change={data.revenue?.growth}
            changeLabel="vs previous period"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
          />
          <StatCard
            title="Orders"
            value={formatNumber(data.orders?.total || 0)}
            change={data.orders?.growth}
            changeLabel="vs previous period"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
              </svg>
            }
          />
          <StatCard
            title="New Customers"
            value={formatNumber(data.customers?.new || 0)}
            change={data.customers?.growth}
            changeLabel="vs previous period"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
              </svg>
            }
          />
          <StatCard
            title="Avg. Order Value"
            value={formatVND(data.aov?.value || 0)}
            change={data.aov?.growth}
            changeLabel="vs previous period"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
              </svg>
            }
          />
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <ChartWrapper title="Revenue Trend" subtitle={`Revenue over the last ${period}`}>
            <RevenueTrendChart data={data.revenue?.trend || []} />
          </ChartWrapper>

          <ChartWrapper title="Orders by Status" subtitle="Current order distribution">
            <OrdersByStatusChart data={data.orders?.by_status || []} />
          </ChartWrapper>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <ChartWrapper title="Top Categories" subtitle="Sales by category">
            <TopCategoriesChart data={data.categories?.top || []} />
          </ChartWrapper>

          <ChartWrapper title="Customer Acquisition" subtitle="New vs returning">
            <CustomerAcquisitionChart data={data.customers?.acquisition || []} />
          </ChartWrapper>

          <ChartWrapper title="Conversion Funnel" subtitle="Visitor to customer">
            <ConversionFunnelChart data={data.funnel || []} />
          </ChartWrapper>
        </div>
      </>
    );
  } catch {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <p>Unable to load analytics data. Please try again later.</p>
      </div>
    );
  }
}

function RevenueTrendChart({ data }: { data: { date: string; value: number }[] }) {
  const { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } = require('recharts');

  return (
    <ResponsiveContainer width="100%" height={300}>
      <LineChart data={data}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
        <XAxis dataKey="date" tick={{ fontSize: 12 }} className="fill-muted-foreground" />
        <YAxis tick={{ fontSize: 12 }} className="fill-muted-foreground" tickFormatter={(v: number) => formatVND(v)} />
        <Tooltip
          contentStyle={{ backgroundColor: 'hsl(var(--card))', border: '1px solid hsl(var(--border))', borderRadius: '8px' }}
          formatter={(value: number) => [formatVND(value), 'Revenue']}
        />
        <Line type="monotone" dataKey="value" stroke="#f97316" strokeWidth={2} dot={false} />
      </LineChart>
    </ResponsiveContainer>
  );
}

function OrdersByStatusChart({ data }: { data: { status: string; count: number }[] }) {
  const { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Legend } = require('recharts');

  const COLORS = ['#f97316', '#3b82f6', '#22c55e', '#eab308', '#ef4444', '#8b5cf6'];

  return (
    <ResponsiveContainer width="100%" height={300}>
      <PieChart>
        <Pie data={data} cx="50%" cy="50%" outerRadius={100} dataKey="count" nameKey="status" label>
          {data.map((_: unknown, index: number) => (
            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
          ))}
        </Pie>
        <Tooltip contentStyle={{ backgroundColor: 'hsl(var(--card))', border: '1px solid hsl(var(--border))', borderRadius: '8px' }} />
        <Legend />
      </PieChart>
    </ResponsiveContainer>
  );
}

function TopCategoriesChart({ data }: { data: { name: string; sales: number }[] }) {
  const { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } = require('recharts');

  return (
    <ResponsiveContainer width="100%" height={300}>
      <BarChart data={data} layout="vertical">
        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
        <XAxis type="number" tick={{ fontSize: 12 }} className="fill-muted-foreground" />
        <YAxis type="category" dataKey="name" tick={{ fontSize: 11 }} className="fill-muted-foreground" width={80} />
        <Tooltip contentStyle={{ backgroundColor: 'hsl(var(--card))', border: '1px solid hsl(var(--border))', borderRadius: '8px' }} />
        <Bar dataKey="sales" fill="#f97316" radius={[0, 4, 4, 0]} />
      </BarChart>
    </ResponsiveContainer>
  );
}

function CustomerAcquisitionChart({ data }: { data: { type: string; count: number }[] }) {
  const { PieChart, Pie, Cell, Tooltip, ResponsiveContainer } = require('recharts');

  const COLORS = ['#22c55e', '#3b82f6'];

  return (
    <ResponsiveContainer width="100%" height={300}>
      <PieChart>
        <Pie data={data} cx="50%" cy="50%" innerRadius={60} outerRadius={100} dataKey="count" nameKey="type" label>
          {data.map((_: unknown, index: number) => (
            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
          ))}
        </Pie>
        <Tooltip contentStyle={{ backgroundColor: 'hsl(var(--card))', border: '1px solid hsl(var(--border))', borderRadius: '8px' }} />
      </PieChart>
    </ResponsiveContainer>
  );
}

function ConversionFunnelChart({ data }: { data: { stage: string; count: number }[] }) {
  const maxCount = Math.max(...data.map((d: { count: number }) => d.count));

  return (
    <div className="space-y-3 py-4">
      {data.map((item: { stage: string; count: number }, index: number) => (
        <div key={index} className="space-y-1">
          <div className="flex justify-between text-sm">
            <span className="text-foreground capitalize">{item.stage}</span>
            <span className="text-muted-foreground">{formatNumber(item.count)}</span>
          </div>
          <div className="h-6 bg-muted rounded-full overflow-hidden">
            <div
              className="h-full bg-primary-500 rounded-full transition-all duration-500"
              style={{ width: `${(item.count / maxCount) * 100}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  );
}
