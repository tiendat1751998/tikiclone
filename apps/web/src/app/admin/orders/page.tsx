"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { ordersApi } from "@/lib/api/client";
import type { Order, OrderStatus } from "@/types";

const STATUS_BADGES: Record<OrderStatus, { label: string; color: string }> = {
  pending: { label: "Chờ xác nhận", color: "bg-yellow-100 text-yellow-700" },
  confirmed: { label: "Đã xác nhận", color: "bg-blue-100 text-blue-700" },
  processing: { label: "Đang xử lý", color: "bg-purple-100 text-purple-700" },
  shipped: { label: "Đang giao", color: "bg-indigo-100 text-indigo-700" },
  delivered: { label: "Đã giao", color: "bg-green-100 text-green-700" },
  cancelled: { label: "Đã hủy", color: "bg-red-100 text-red-700" },
  refunded: { label: "Hoàn tiền", color: "bg-gray-100 text-gray-700" },
};

export default function AdminOrdersPage() {
  const [statusFilter, setStatusFilter] = useState("");
  const [page, setPage] = useState(1);

  const { data, isLoading } = useQuery({
    queryKey: ["admin-orders", statusFilter, page],
    queryFn: () => ordersApi.list({ ...(statusFilter ? { status: statusFilter } : {}), page: String(page) }),
  });

  const orders = data?.items || [];
  const totalPages = data?.total_pages || 1;

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold text-tiki-text">Đơn hàng</h2>

      {/* Filters */}
      <div className="bg-white rounded-lg border border-tiki-border p-3 flex gap-2 flex-wrap">
        {["", "pending", "confirmed", "processing", "shipped", "delivered", "cancelled"].map((s) => (
          <button
            key={s}
            onClick={() => { setStatusFilter(s); setPage(1); }}
            className={`px-3 py-1.5 text-xs rounded-lg font-medium transition ${
              statusFilter === s ? "bg-tiki-blue text-white" : "bg-gray-100 text-tiki-text-secondary hover:bg-gray-200"
            }`}
          >
            {s === "" ? "Tất cả" : STATUS_BADGES[s as OrderStatus]?.label || s}
          </button>
        ))}
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg border border-tiki-border overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-50 border-b border-tiki-border">
                <th className="text-left px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Mã đơn</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Khách hàng</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Tổng tiền</th>
                <th className="text-center px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Thanh toán</th>
                <th className="text-center px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Trạng thái</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Ngày đặt</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Thao tác</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className="border-b border-tiki-border animate-pulse">
                    {Array.from({ length: 7 }).map((_, j) => (
                      <td key={j} className="px-4 py-3"><div className="h-4 bg-gray-200 rounded" /></td>
                    ))}
                  </tr>
                ))
              ) : orders.length === 0 ? (
                <tr><td colSpan={7} className="px-4 py-12 text-center text-sm text-tiki-text-secondary">Không có đơn hàng</td></tr>
              ) : (
                orders.map((order: Order) => {
                  const badge = STATUS_BADGES[order.status] || { label: order.status, color: "bg-gray-100 text-gray-700" };
                  return (
                    <tr key={order.id} className="border-b border-tiki-border hover:bg-gray-50">
                      <td className="px-4 py-3 font-mono text-xs">#{order.order_number || order.id.slice(0, 8)}</td>
                      <td className="px-4 py-3 text-tiki-text">{order.shipping_address?.name || order.user_id}</td>
                      <td className="px-4 py-3 text-right font-medium text-tiki-text">{order.total?.toLocaleString("vi-VN")} ₫</td>
                      <td className="px-4 py-3 text-center text-xs capitalize">{order.payment_method || "cod"}</td>
                      <td className="px-4 py-3 text-center">
                        <span className={`text-[10px] font-medium px-2 py-0.5 rounded ${badge.color}`}>{badge.label}</span>
                      </td>
                      <td className="px-4 py-3 text-xs text-tiki-text-secondary">{new Date(order.created_at).toLocaleDateString("vi-VN")}</td>
                      <td className="px-4 py-3 text-right">
                        <button className="text-xs text-tiki-blue hover:underline">Chi tiết</button>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {totalPages > 1 && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-tiki-border">
            <span className="text-xs text-tiki-text-secondary">Trang {page} / {totalPages}</span>
            <div className="flex gap-1">
              <button onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page <= 1} className="px-3 py-1 text-xs border border-tiki-border rounded hover:bg-gray-50 disabled:opacity-50">Trước</button>
              <button onClick={() => setPage((p) => Math.min(totalPages, p + 1))} disabled={page >= totalPages} className="px-3 py-1 text-xs border border-tiki-border rounded hover:bg-gray-50 disabled:opacity-50">Sau</button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
