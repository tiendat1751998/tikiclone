"use client";

import { useQuery } from "@tanstack/react-query";
import { dashboardApi } from "@/lib/api/client";
import type { ChartDataPoint } from "@/types";

function SimpleBarChart({ data }: { data: ChartDataPoint[] }) {
  if (!data || data.length === 0) {
    return <div className="h-48 flex items-center justify-center text-sm text-tiki-text-secondary">Chưa có dữ liệu</div>;
  }

  const maxVal = Math.max(...data.map((d) => d.value));

  return (
    <div className="h-48 flex items-end gap-2">
      {data.map((d, i) => {
        const height = maxVal > 0 ? (d.value / maxVal) * 100 : 0;
        return (
          <div key={i} className="flex-1 flex flex-col items-center gap-1">
            <span className="text-[9px] text-tiki-text-secondary">{d.value >= 1000 ? `${(d.value / 1000).toFixed(1)}k` : d.value}</span>
            <div
              className="w-full bg-tiki-blue rounded-t-sm transition-all duration-300 hover:bg-tiki-blue-dark"
              style={{ height: `${Math.max(height, 2)}%` }}
            />
            <span className="text-[9px] text-tiki-text-secondary">{d.label}</span>
          </div>
        );
      })}
    </div>
  );
}

export default function RevenueChart() {
  const { data: revenueData } = useQuery({
    queryKey: ["dashboard", "revenue", "7d"],
    queryFn: () => dashboardApi.getRevenueChart("7d"),
    initialData: [
      { label: "T2", value: 12500000 },
      { label: "T3", value: 15800000 },
      { label: "T4", value: 11200000 },
      { label: "T5", value: 18900000 },
      { label: "T6", value: 22400000 },
      { label: "T7", value: 28100000 },
      { label: "CN", value: 19500000 },
    ],
  });

  const { data: ordersData } = useQuery({
    queryKey: ["dashboard", "orders", "7d"],
    queryFn: () => dashboardApi.getOrdersChart("7d"),
    initialData: [
      { label: "T2", value: 245 },
      { label: "T3", value: 312 },
      { label: "T4", value: 198 },
      { label: "T5", value: 356 },
      { label: "T6", value: 421 },
      { label: "T7", value: 580 },
      { label: "CN", value: 389 },
    ],
  });

  return (
    <div className="space-y-6">
      <div>
        <h4 className="text-xs font-semibold text-tiki-text-secondary mb-3">Doanh thu (₫)</h4>
        <SimpleBarChart data={revenueData || []} />
      </div>
      <div>
        <h4 className="text-xs font-semibold text-tiki-text-secondary mb-3">Đơn hàng</h4>
        <SimpleBarChart data={ordersData || []} />
      </div>
    </div>
  );
}
