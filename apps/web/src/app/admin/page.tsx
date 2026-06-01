"use client";

import { useQuery } from "@tanstack/react-query";
import { dashboardApi } from "@/lib/api/client";
import MetricCard from "@/components/admin/MetricCard";
import RevenueChart from "@/components/admin/RevenueChart";

export default function AdminDashboardPage() {
  const { data: metrics } = useQuery({
    queryKey: ["dashboard", "metrics"],
    queryFn: () => dashboardApi.getMetrics("7d"),
  });

  const { data: alerts } = useQuery({
    queryKey: ["dashboard", "alerts"],
    queryFn: () => dashboardApi.getAlerts(),
  });

  return (
    <div className="space-y-6">
      <h2 className="text-lg font-semibold text-tiki-text">Tổng quan</h2>

      {/* Metrics grid */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard
          title="Doanh thu"
          value={metrics?.revenue ? `${(metrics.revenue / 1000000).toFixed(1)}M ₫` : "—"}
          change={metrics?.revenue_change}
          icon="💰"
          color="blue"
        />
        <MetricCard
          title="Đơn hàng"
          value={metrics?.orders?.toLocaleString("vi-VN") || "—"}
          change={metrics?.orders_change}
          icon="📦"
          color="purple"
        />
        <MetricCard
          title="Khách hàng"
          value={metrics?.customers?.toLocaleString("vi-VN") || "—"}
          change={metrics?.customers_change}
          icon="👥"
          color="green"
        />
        <MetricCard
          title="GMV"
          value={metrics?.gmv ? `${(metrics.gmv / 1000000).toFixed(1)}M ₫` : "—"}
          change={metrics?.gmv_change}
          icon="📊"
          color="orange"
        />
      </div>

      {/* Revenue chart */}
      <div className="bg-white rounded-lg border border-tiki-border p-4">
        <h3 className="text-sm font-semibold text-tiki-text mb-4">Doanh thu 7 ngày qua</h3>
        <RevenueChart />
      </div>

      {/* Alerts */}
      {alerts && alerts.length > 0 && (
        <div className="bg-white rounded-lg border border-tiki-border p-4">
          <h3 className="text-sm font-semibold text-tiki-text mb-3">Cảnh báo</h3>
          <div className="space-y-2">
            {alerts.map((alert) => (
              <div
                key={alert.id}
                className={`flex items-start gap-3 p-3 rounded-lg ${
                  alert.severity === "critical" ? "bg-red-50" :
                  alert.severity === "high" ? "bg-orange-50" :
                  alert.severity === "medium" ? "bg-yellow-50" : "bg-blue-50"
                }`}
              >
                <span className="text-sm">
                  {alert.type === "error" ? "🔴" : alert.type === "warning" ? "🟡" : alert.type === "success" ? "🟢" : "🔵"}
                </span>
                <div>
                  <p className="text-sm font-medium text-tiki-text">{alert.title}</p>
                  <p className="text-xs text-tiki-text-secondary">{alert.message}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
