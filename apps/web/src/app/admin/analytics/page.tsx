"use client";

import { useQuery } from "@tanstack/react-query";
import { dashboardApi } from "@/lib/api/client";
import MetricCard from "@/components/admin/MetricCard";
import RevenueChart from "@/components/admin/RevenueChart";

export default function AdminAnalyticsPage() {
  const { data: metrics } = useQuery({
    queryKey: ["dashboard", "metrics", "30d"],
    queryFn: () => dashboardApi.getMetrics("30d"),
  });

  const { data: realtime } = useQuery({
    queryKey: ["dashboard", "realtime"],
    queryFn: () => dashboardApi.getRealtimeStats(),
    refetchInterval: 30000,
  });

  return (
    <div className="space-y-6">
      <h2 className="text-lg font-semibold text-tiki-text">Phân tích</h2>

      {/* Realtime stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard title="Đang online" value={realtime?.active_users?.toLocaleString() || "—"} icon="🟢" color="green" />
        <MetricCard title="Đơn hôm nay" value={realtime?.orders_today?.toLocaleString() || "—"} icon="📦" color="blue" />
        <MetricCard title="Doanh thu hôm nay" value={realtime?.revenue_today ? `${(realtime.revenue_today / 1000000).toFixed(1)}M ₫` : "—"} icon="💰" color="orange" />
        <MetricCard title="Tỷ lệ chuyển đổi" value={realtime?.conversion_rate ? `${realtime.conversion_rate.toFixed(1)}%` : "—"} icon="📈" color="purple" />
      </div>

      {/* 30-day metrics */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard title="GMV (30 ngày)" value={metrics?.gmv ? `${(metrics.gmv / 1000000).toFixed(0)}M ₫` : "—"} change={metrics?.gmv_change} icon="📊" color="blue" />
        <MetricCard title="Doanh thu" value={metrics?.revenue ? `${(metrics.revenue / 1000000).toFixed(0)}M ₫` : "—"} change={metrics?.revenue_change} icon="💰" color="green" />
        <MetricCard title="Đơn hàng" value={metrics?.orders?.toLocaleString() || "—"} change={metrics?.orders_change} icon="📦" color="purple" />
        <MetricCard title="Khách hàng mới" value={metrics?.customers?.toLocaleString() || "—"} change={metrics?.customers_change} icon="👥" color="orange" />
      </div>

      {/* Charts */}
      <div className="bg-white rounded-lg border border-tiki-border p-4">
        <h3 className="text-sm font-semibold text-tiki-text mb-4">Doanh thu 30 ngày qua</h3>
        <RevenueChart />
      </div>

      {/* AOV */}
      <div className="bg-white rounded-lg border border-tiki-border p-4">
        <h3 className="text-sm font-semibold text-tiki-text mb-3">Giá trị đơn hàng trung bình (AOV)</h3>
        <div className="text-3xl font-bold text-tiki-text">
          {metrics?.aov ? `${(metrics.aov / 1000).toFixed(0)}K ₫` : "—"}
        </div>
      </div>
    </div>
  );
}
