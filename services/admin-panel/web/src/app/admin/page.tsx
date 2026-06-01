import { Suspense } from 'react';
import { ErrorBoundary, StatCard, ChartWrapper } from '@/components/admin';
import { DashboardCharts } from './DashboardCharts';
import { formatVND, formatNumber } from '@shopee/shared-utils';

async function getDashboardStats() {
  const res = await fetch(`${process.env.API_GATEWAY_URL || 'http://localhost:8080'}/api/admin/dashboard/stats`, {
    next: { revalidate: 60 },
    headers: { 'Content-Type': 'application/json' },
  });

  if (!res.ok) {
    throw new Error('Failed to fetch dashboard stats');
  }

  return res.json();
}

async function getRevenueData() {
  const res = await fetch(`${process.env.API_GATEWAY_URL || 'http://localhost:8080'}/api/admin/dashboard/revenue?days=30`, {
    next: { revalidate: 300 },
    headers: { 'Content-Type': 'application/json' },
  });

  if (!res.ok) {
    throw new Error('Failed to fetch revenue data');
  }

  return res.json();
}

function StatsGrid() {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      <Suspense fallback={<StatCard title="Total Revenue" value="--" isLoading />}>
        <RevenueStat />
      </Suspense>
      <Suspense fallback={<StatCard title="Total Orders" value="--" isLoading />}>
        <OrdersStat />
      </Suspense>
      <Suspense fallback={<StatCard title="Total Users" value="--" isLoading />}>
        <UsersStat />
      </Suspense>
      <Suspense fallback={<StatCard title="Conversion Rate" value="--" isLoading />}>
        <ConversionStat />
      </Suspense>
    </div>
  );
}

async function RevenueStat() {
  try {
    const data = await getDashboardStats();
    return (
      <StatCard
        title="Total Revenue"
        value={formatVND(data.total_revenue || 0)}
        change={data.revenue_change}
        changeLabel="vs last month"
        icon={
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        }
      />
    );
  } catch {
    return <StatCard title="Total Revenue" value="N/A" />;
  }
}

async function OrdersStat() {
  try {
    const data = await getDashboardStats();
    return (
      <StatCard
        title="Total Orders"
        value={formatNumber(data.total_orders || 0)}
        change={data.orders_change}
        changeLabel="vs last month"
        icon={
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
          </svg>
        }
      />
    );
  } catch {
    return <StatCard title="Total Orders" value="N/A" />;
  }
}

async function UsersStat() {
  try {
    const data = await getDashboardStats();
    return (
      <StatCard
        title="Total Users"
        value={formatNumber(data.total_users || 0)}
        change={data.users_change}
        changeLabel="vs last month"
        icon={
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
        }
      />
    );
  } catch {
    return <StatCard title="Total Users" value="N/A" />;
  }
}

async function ConversionStat() {
  try {
    const data = await getDashboardStats();
    return (
      <StatCard
        title="Conversion Rate"
        value={`${(data.conversion_rate || 0).toFixed(2)}%`}
        change={data.conversion_change}
        changeLabel="vs last month"
        icon={
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
          </svg>
        }
      />
    );
  } catch {
    return <StatCard title="Conversion Rate" value="N/A" />;
  }
}

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">Dashboard</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Overview of your e-commerce platform performance
        </p>
      </div>

      <ErrorBoundary>
        <StatsGrid />
      </ErrorBoundary>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ErrorBoundary>
          <Suspense fallback={<ChartWrapper title="Revenue Overview" subtitle="Last 30 days" isLoading />}>
            <RevenueChart />
          </Suspense>
        </ErrorBoundary>

        <ErrorBoundary>
          <Suspense fallback={<ChartWrapper title="Order Status" subtitle="Current distribution" isLoading />}>
            <OrderStatusChart />
          </Suspense>
        </ErrorBoundary>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <ErrorBoundary>
          <Suspense fallback={<ChartWrapper title="Top Products" subtitle="Best sellers" isLoading />}>
            <TopProductsChart />
          </Suspense>
        </ErrorBoundary>

        <ErrorBoundary>
          <Suspense fallback={<ChartWrapper title="New Users" subtitle="Daily registrations" isLoading />}>
            <NewUsersChart />
          </Suspense>
        </ErrorBoundary>

        <ErrorBoundary>
          <Suspense fallback={<ChartWrapper title="Recent Orders" subtitle="Latest transactions" isLoading />}>
            <RecentOrdersList />
          </Suspense>
        </ErrorBoundary>
      </div>
    </div>
  );
}

async function RevenueChart() {
  try {
    const data = await getRevenueData();
    return (
      <ChartWrapper title="Revenue Overview" subtitle="Last 30 days">
        <DashboardCharts.RevenueChart data={data.revenue || []} />
      </ChartWrapper>
    );
  } catch (err) {
    return (
      <ChartWrapper
        title="Revenue Overview"
        subtitle="Last 30 days"
        error="Unable to load revenue data"
      />
    );
  }
}

async function OrderStatusChart() {
  try {
    const data = await getDashboardStats();
    return (
      <ChartWrapper title="Order Status" subtitle="Current distribution">
        <DashboardCharts.OrderStatusChart data={data.order_status || []} />
      </ChartWrapper>
    );
  } catch (err) {
    return (
      <ChartWrapper
        title="Order Status"
        subtitle="Current distribution"
        error="Unable to load order status data"
      />
    );
  }
}

async function TopProductsChart() {
  try {
    const data = await getDashboardStats();
    return (
      <ChartWrapper title="Top Products" subtitle="Best sellers this month">
        <DashboardCharts.TopProductsChart data={data.top_products || []} />
      </ChartWrapper>
    );
  } catch (err) {
    return (
      <ChartWrapper
        title="Top Products"
        subtitle="Best sellers this month"
        error="Unable to load product data"
      />
    );
  }
}

async function NewUsersChart() {
  try {
    const data = await getRevenueData();
    return (
      <ChartWrapper title="New Users" subtitle="Daily registrations">
        <DashboardCharts.NewUsersChart data={data.new_users || []} />
      </ChartWrapper>
    );
  } catch (err) {
    return (
      <ChartWrapper
        title="New Users"
        subtitle="Daily registrations"
        error="Unable to load user data"
      />
    );
  }
}

async function RecentOrdersList() {
  try {
    const data = await getDashboardStats();
    return (
      <ChartWrapper title="Recent Orders" subtitle="Latest 5 transactions">
        <DashboardCharts.RecentOrdersList orders={data.recent_orders || []} />
      </ChartWrapper>
    );
  } catch (err) {
    return (
      <ChartWrapper
        title="Recent Orders"
        subtitle="Latest transactions"
        error="Unable to load recent orders"
      />
    );
  }
}
