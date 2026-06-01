"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuthStore } from "@/stores/auth";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ordersApi } from "@/lib/api/client";
import type { OrderStatus } from "@/types";

const STATUS_BADGES: Record<OrderStatus, { label: string; color: string }> = {
  pending: { label: "Chờ xác nhận", color: "bg-yellow-100 text-yellow-700" },
  confirmed: { label: "Đã xác nhận", color: "bg-blue-100 text-blue-700" },
  processing: { label: "Đang xử lý", color: "bg-purple-100 text-purple-700" },
  shipped: { label: "Đang giao", color: "bg-indigo-100 text-indigo-700" },
  delivered: { label: "Đã giao", color: "bg-green-100 text-green-700" },
  cancelled: { label: "Đã hủy", color: "bg-red-100 text-red-700" },
  refunded: { label: "Hoàn tiền", color: "bg-gray-100 text-gray-700" },
};

const TIMELINE_STEPS: OrderStatus[] = ["pending", "confirmed", "processing", "shipped", "delivered"];

export default function OrderDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = useParams_from_promise(params);
  const router = useRouter();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const queryClient = useQueryClient();
  const [showCancelConfirm, setShowCancelConfirm] = useState(false);

  const { data: order, isLoading } = useQuery({
    queryKey: ["orders", id],
    queryFn: () => ordersApi.getById(id),
    enabled: !!id && isAuthenticated,
  });

  const cancelMutation = useMutation({
    mutationFn: () => ordersApi.cancel(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["orders", id] });
      setShowCancelConfirm(false);
    },
  });

  if (!isAuthenticated) {
    router.push("/login");
    return null;
  }

  if (isLoading) {
    return (
      <main className="bg-[#F5F5FA] py-4 min-h-[60vh]">
        <div className="max-w-3xl mx-auto px-3 animate-pulse space-y-4">
          <div className="h-6 bg-gray-200 rounded w-1/3" />
          <div className="bg-white rounded-lg border border-tiki-border h-48" />
          <div className="bg-white rounded-lg border border-tiki-border h-32" />
        </div>
      </main>
    );
  }

  if (!order) {
    return (
      <main className="bg-[#F5F5FA] py-16 min-h-[60vh] text-center">
        <p className="text-4xl mb-3">❌</p>
        <p className="text-sm text-tiki-text-secondary mb-4">Không tìm thấy đơn hàng</p>
        <Link href="/account/orders" className="text-tiki-blue text-sm hover:underline">← Quay lại đơn hàng</Link>
      </main>
    );
  }

  const badge = STATUS_BADGES[order.status] || { label: order.status, color: "bg-gray-100 text-gray-700" };
  const currentStep = TIMELINE_STEPS.indexOf(order.status);

  return (
    <main className="bg-[#F5F5FA] py-4 min-h-[60vh]">
      <div className="max-w-3xl mx-auto px-3">
        <div className="flex items-center gap-2 mb-4">
          <Link href="/account/orders" className="text-tiki-blue text-sm hover:underline">← Đơn hàng của tôi</Link>
          <span className="text-tiki-text-secondary text-sm">/</span>
          <span className="text-sm text-tiki-text-secondary">#{order.order_number || order.id.slice(0, 8)}</span>
        </div>

        <div className="bg-white rounded-lg border border-tiki-border p-4 mb-4">
          <div className="flex items-center justify-between mb-4">
            <span className="text-sm font-semibold text-tiki-text">Trạng thái đơn hàng</span>
            <span className={`text-xs font-medium px-2 py-0.5 rounded ${badge.color}`}>{badge.label}</span>
          </div>
          <div className="flex items-center gap-1">
            {TIMELINE_STEPS.map((step, i) => {
              const done = i <= currentStep && order.status !== "cancelled";
              const cancelled = order.status === "cancelled";
              return (
                <div key={step} className="flex-1 flex items-center">
                  <div className={`w-full h-1.5 rounded ${cancelled && i > 0 ? "bg-red-200" : done ? "bg-green-500" : "bg-gray-200"}`} />
                </div>
              );
            })}
          </div>
          <div className="flex justify-between mt-2">
            {TIMELINE_STEPS.map((step) => (
              <span key={step} className="text-[9px] text-tiki-text-secondary text-center flex-1">
                {STATUS_BADGES[step]?.label}
              </span>
            ))}
          </div>
          {order.timeline && order.timeline.length > 0 && (
            <div className="mt-4 pt-4 border-t border-tiki-border space-y-3">
              {order.timeline.map((t, i) => (
                <div key={i} className="flex items-start gap-3">
                  <div className="w-2 h-2 rounded-full bg-green-500 mt-1.5 shrink-0" />
                  <div>
                    <p className="text-sm text-tiki-text">{t.description}</p>
                    <p className="text-xs text-tiki-text-secondary">{new Date(t.timestamp).toLocaleString("vi-VN")}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="bg-white rounded-lg border border-tiki-border p-4 mb-4">
          <h3 className="text-sm font-semibold text-tiki-text mb-3">Địa chỉ nhận hàng</h3>
          <p className="text-sm text-tiki-text font-medium">{order.shipping_address?.name}</p>
          <p className="text-sm text-tiki-text-secondary">{order.shipping_address?.phone}</p>
          <p className="text-sm text-tiki-text-secondary">
            {[order.shipping_address?.address_line1, order.shipping_address?.city, order.shipping_address?.state]
              .filter(Boolean)
              .join(", ")}
          </p>
        </div>

        <div className="bg-white rounded-lg border border-tiki-border p-4 mb-4">
          <h3 className="text-sm font-semibold text-tiki-text mb-3">Sản phẩm</h3>
          {order.items?.map((item, i) => (
            <div key={i} className="flex items-center gap-3 py-3 border-b border-tiki-border last:border-0 first:pt-0 last:pb-0">
              <img src={item.image_url || "/images/placeholder.svg"} alt={item.name} className="w-14 h-14 rounded object-cover" />
              <div className="flex-1 min-w-0">
                <p className="text-sm text-tiki-text">{item.name}</p>
                <p className="text-xs text-tiki-text-secondary">x{item.quantity}</p>
              </div>
              <span className="text-sm font-medium text-tiki-text">{item.price?.toLocaleString("vi-VN")} ₫</span>
            </div>
          ))}
        </div>

        <div className="bg-white rounded-lg border border-tiki-border p-4">
          <h3 className="text-sm font-semibold text-tiki-text mb-3">Thanh toán</h3>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-tiki-text-secondary">Tạm tính</span>
              <span className="text-tiki-text">{order.subtotal?.toLocaleString("vi-VN")} ₫</span>
            </div>
            <div className="flex justify-between">
              <span className="text-tiki-text-secondary">Phí vận chuyển</span>
              <span className="text-tiki-text">{order.shipping_fee?.toLocaleString("vi-VN")} ₫</span>
            </div>
            {(order.discount ?? 0) > 0 && (
              <div className="flex justify-between">
                <span className="text-tiki-text-secondary">Giảm giá</span>
                <span className="text-green-600">-{(order.discount ?? 0).toLocaleString("vi-VN")} ₫</span>
              </div>
            )}
            <div className="flex justify-between border-t border-tiki-border pt-2 mt-2">
              <span className="font-semibold text-tiki-text">Tổng cộng</span>
              <span className="text-lg font-bold text-tiki-red">{order.total?.toLocaleString("vi-VN")} ₫</span>
            </div>
          </div>
        </div>

        <div className="flex gap-3 mt-4">
          <Link href="/account/orders" className="px-4 py-2 border border-tiki-border rounded-lg text-sm text-tiki-text-secondary hover:bg-gray-50 transition">
            ← Quay lại
          </Link>
          {order.status === "pending" && (
            <button
              onClick={() => setShowCancelConfirm(true)}
              className="px-4 py-2 bg-red-50 text-red-600 border border-red-200 rounded-lg text-sm font-medium hover:bg-red-100 transition"
            >
              Hủy đơn hàng
            </button>
          )}
        </div>
      </div>

      {showCancelConfirm && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-sm mx-3">
            <h3 className="text-sm font-semibold text-tiki-text mb-2">Xác nhận hủy đơn hàng</h3>
            <p className="text-sm text-tiki-text-secondary mb-4">Bạn có chắc chắn muốn hủy đơn hàng này? Hành động này không thể hoàn tác.</p>
            <div className="flex gap-3 justify-end">
              <button
                onClick={() => setShowCancelConfirm(false)}
                className="px-4 py-2 border border-tiki-border rounded-lg text-sm text-tiki-text-secondary hover:bg-gray-50 transition"
                disabled={cancelMutation.isPending}
              >
                Không
              </button>
              <button
                onClick={() => cancelMutation.mutate()}
                className="px-4 py-2 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700 transition disabled:opacity-50"
                disabled={cancelMutation.isPending}
              >
                {cancelMutation.isPending ? "Đang hủy..." : "Xác nhận hủy"}
              </button>
            </div>
            {cancelMutation.isError && (
              <p className="text-xs text-red-600 mt-2">{(cancelMutation.error as any)?.message || "Hủy đơn hàng thất bại"}</p>
            )}
          </div>
        </div>
      )}
    </main>
  );
}

import { useEffect as _useEffect, useState as _useState } from "react";

function useParams_from_promise(params: Promise<{ id: string }>): { id: string } {
  const [resolved, setResolved] = _useState({ id: "" });
  _useEffect(() => { params.then(setResolved); }, []);
  return resolved;
}
