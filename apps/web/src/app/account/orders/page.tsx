"use client";

import { useState } from "react";
import Link from "next/link";
import { useAuthStore } from "@/stores/auth";
import { ordersApi } from "@/lib/api/client";
import { useQuery } from "@tanstack/react-query";
import type { Order, OrderStatus } from "@/types";

const STATUS_FILTERS: { label: string; value: string }[] = [
  { label: "Tất cả", value: "" },
  { label: "Chờ xác nhận", value: "pending" },
  { label: "Đã xác nhận", value: "confirmed" },
  { label: "Đang xử lý", value: "processing" },
  { label: "Đang giao", value: "shipped" },
  { label: "Đã giao", value: "delivered" },
  { label: "Đã hủy", value: "cancelled" },
];

const STATUS_BADGES: Record<OrderStatus, { label: string; color: string }> = {
  pending: { label: "Chờ xác nhận", color: "bg-yellow-100 text-yellow-700" },
  confirmed: { label: "Đã xác nhận", color: "bg-blue-100 text-blue-700" },
  processing: { label: "Đang xử lý", color: "bg-purple-100 text-purple-700" },
  shipped: { label: "Đang giao", color: "bg-indigo-100 text-indigo-700" },
  delivered: { label: "Đã giao", color: "bg-green-100 text-green-700" },
  cancelled: { label: "Đã hủy", color: "bg-red-100 text-red-700" },
  refunded: { label: "Hoàn tiền", color: "bg-gray-100 text-gray-700" },
};

function OrderCard({ order }: { order: Order }) {
  const badge = STATUS_BADGES[order.status] || { label: order.status, color: "bg-gray-100 text-gray-700" };

  return (
    <Link href={`/account/orders/${order.id}`} className="order-card">
      <div className="order-card__header">
        <div className="flex items-center gap-2">
          <span className="text-xs text-tiki-text-secondary">#{order.order_number || order.id.slice(0, 8)}</span>
          <span className={`text-[10px] font-medium px-2 py-0.5 rounded ${badge.color}`}>{badge.label}</span>
        </div>
        <span className="text-xs text-tiki-text-secondary">{new Date(order.created_at).toLocaleDateString("vi-VN")}</span>
      </div>
      <div className="order-card__body">
        {order.items.slice(0, 3).map((item, i) => (
          <div key={i} className="flex items-center gap-2 mb-1.5 last:mb-0">
            <img src={item.image_url || "/images/placeholder.svg"} alt={item.name} className="w-9 h-9 rounded object-cover" />
            <div className="flex-1 min-w-0">
              <p className="text-xs text-tiki-text truncate">{item.name}</p>
              <p className="text-[10px] text-tiki-text-secondary">x{item.quantity}</p>
            </div>
            <span className="text-xs font-medium text-tiki-text whitespace-nowrap">{item.price?.toLocaleString("vi-VN")} ₫</span>
          </div>
        ))}
        {order.items.length > 3 && (
          <p className="text-[10px] text-tiki-text-secondary mt-1">+{order.items.length - 3} sản phẩm khác</p>
        )}
      </div>
      <div className="order-card__footer">
        <span className="text-[10px] text-tiki-text-secondary">Tổng thanh toán:</span>
        <span className="text-sm font-bold text-tiki-red">{order.total?.toLocaleString("vi-VN")} ₫</span>
      </div>
    </Link>
  );
}

export default function OrdersPage() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const [statusFilter, setStatusFilter] = useState("");

  const { data: ordersResponse, isLoading } = useQuery({
    queryKey: ["orders", { status: statusFilter }],
    queryFn: () => ordersApi.list(statusFilter ? { status: statusFilter } : {}),
    enabled: isAuthenticated,
  });

  const orders = ordersResponse?.items || [];

  return (
    <div>
      <h1 className="text-sm font-semibold text-tiki-text mb-3">Đơn hàng của tôi</h1>

      <div className="bg-white rounded-lg border border-tiki-border mb-3 overflow-x-auto">
        <div className="flex">
          {STATUS_FILTERS.map((f) => (
            <button
              key={f.value}
              onClick={() => setStatusFilter(f.value)}
              className={`flex-shrink-0 px-3 py-2.5 text-[11px] font-medium border-b-2 transition ${
                statusFilter === f.value
                  ? "border-tiki-blue text-tiki-blue"
                  : "border-transparent text-tiki-text-secondary hover:text-tiki-text"
              }`}
            >
              {f.label}
            </button>
          ))}
        </div>
      </div>

      {isLoading ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="order-card animate-pulse">
              <div className="order-card__header"><div className="h-3 bg-gray-200 rounded w-24" /></div>
              <div className="order-card__body"><div className="h-8 bg-gray-200 rounded" /></div>
            </div>
          ))}
        </div>
      ) : orders.length === 0 ? (
        <div className="bg-white rounded-lg border border-tiki-border py-14 text-center">
          <p className="text-3xl mb-2">📦</p>
          <p className="text-xs text-tiki-text-secondary">Chưa có đơn hàng nào</p>
          <Link href="/products" className="inline-block mt-3 px-4 py-1.5 bg-tiki-blue text-white rounded-lg text-xs font-medium hover:bg-tiki-blue-dark transition">
            Mua sắm ngay
          </Link>
        </div>
      ) : (
        <div className="space-y-3">
          {orders.map((order: Order) => (
            <OrderCard key={order.id} order={order} />
          ))}
        </div>
      )}
    </div>
  );
}