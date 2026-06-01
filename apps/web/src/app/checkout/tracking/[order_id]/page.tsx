"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";
import { deliveryApi } from "@/lib/api/client";
import { useUIStore } from "@/stores/ui";
import {
  Package,
  MapPin,
  CheckCircle,
  Clock,
  Truck,
  User,
  Phone,
  CircleDot,
  ArrowRight,
} from "lucide-react";

const WS_BASE = (process.env.NEXT_PUBLIC_WSS_URL || "").replace(/^ws/, "ws") || "";


interface TrackingState {
  id: string;
  customer_id: string;
  driver_id?: string;
  status: string;
  distance_meters: number;
  duration_seconds: number;
  polyline: string;
  pickup: { lat: number; lng: number; address?: string };
  dropoff: { lat: number; lng: number; address?: string };
  created_at: string;
  updated_at: string;
  assigned_at?: string;
  delivered_at?: string;
}

const STATUS_STEPS = [
  { key: "pending", label: "Đã đặt hàng", icon: Package },
  { key: "searching_driver", label: "Tìm tài xế", icon: User },
  { key: "driver_assigned", label: "Tài xế nhận đơn", icon: CheckCircle },
  { key: "picked_up", label: "Đã lấy hàng", icon: Truck },
  { key: "delivering", label: "Đang giao", icon: MapPin },
  { key: "completed", label: "Đã giao", icon: CheckCircle },
];

function getStatusIndex(status: string): number {
  const idx = STATUS_STEPS.findIndex((s) => s.key === status);
  return idx >= 0 ? idx : 0;
}

function formatTime(seconds: number): string {
  if (seconds >= 3600) {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    return `${h}h ${m}p`;
  }
  return `${Math.ceil(seconds / 60)} phút`;
}

function formatDate(dateStr: string): string {
  try {
    const d = new Date(dateStr);
    return d.toLocaleString("vi-VN", {
      hour: "2-digit",
      minute: "2-digit",
      day: "2-digit",
      month: "2-digit",
    });
  } catch {
    return dateStr;
  }
}

export default function DeliveryTrackingPage() {
  const params = useParams();
  const orderId = params?.order_id as string;
  const addToast = useUIStore((s) => s.addToast);

  const [order, setOrder] = useState<TrackingState | null>(null);
  const [loading, setLoading] = useState(true);
  const [driverLocation, setDriverLocation] = useState<{ lat: number; lng: number } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>(undefined);

  // Fetch order details
  const fetchOrder = useCallback(async () => {
    if (!orderId) return;
    try {
      const data = await deliveryApi.getOrder(orderId);
      setOrder(data as TrackingState);
      setError(null);
    } catch (err: any) {
      setError(err?.message || "Không thể tải thông tin đơn hàng");
    } finally {
      setLoading(false);
    }
  }, [orderId]);

  useEffect(() => {
    fetchOrder();
  }, [fetchOrder]);

  // WebSocket connection for realtime tracking
  useEffect(() => {
    if (!orderId) return;

    function connect() {
      const wsUrl = WS_BASE || `${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.host}`;
      const ws = new WebSocket(`${wsUrl}/ws?user_id=${orderId}&user_type=customer&room_id=order:${orderId}`);

      ws.onopen = () => {
        console.log("[WS] Connected for order tracking:", orderId);
        // Join the order room
        ws.send(JSON.stringify({ action: "join", room: `order:${orderId}` }));
      };

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          if (msg.type === "order:tracking:update") {
            const payload = msg.payload;
            setOrder((prev) =>
              prev
                ? {
                    ...prev,
                    status: payload.status || prev.status,
                    updated_at: payload.timestamp || prev.updated_at,
                  }
                : prev
            );
            // If driver assigned, store location
            if (payload.lat && payload.lng) {
              setDriverLocation({ lat: payload.lat, lng: payload.lng });
            }
          }
          if (msg.type === "driver:location:update") {
            const payload = msg.payload;
            if (payload.lat && payload.lng) {
              setDriverLocation({ lat: payload.lat, lng: payload.lng });
            }
          }
          if (msg.type === "driver:assigned") {
            const payload = msg.payload;
            setOrder((prev) =>
              prev
                ? {
                    ...prev,
                    driver_id: payload.driver_id,
                    status: "driver_assigned",
                    assigned_at: payload.timestamp,
                    updated_at: payload.timestamp,
                  }
                : prev
            );
            addToast({
              type: "info",
              title: "Tài xế đã nhận đơn",
              message: `Tài xế ${payload.driver_id} đang đến lấy hàng`,
            });
          }
        } catch {
          // ignore parse errors
        }
      };

      ws.onclose = () => {
        console.log("[WS] Disconnected, reconnecting in 5s...");
        reconnectTimeoutRef.current = setTimeout(connect, 5000);
      };

      ws.onerror = () => {
        ws.close();
      };

      wsRef.current = ws;
    }

    connect();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [orderId, addToast]);

  // Poll order status every 30s as fallback
  useEffect(() => {
    const interval = setInterval(fetchOrder, 30000);
    return () => clearInterval(interval);
  }, [fetchOrder]);

  if (loading) {
    return (
      <>
        <Header />
        <main className="py-16 text-center">
          <div className="w-8 h-8 border-3 border-tiki-blue border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-sm text-tiki-text-secondary">Đang tải thông tin đơn hàng...</p>
        </main>
        <Footer />
      </>
    );
  }

  if (error || !order) {
    return (
      <>
        <Header />
        <main className="py-16 text-center">
          <div className="text-5xl mb-4">📦</div>
          <h1 className="text-lg font-semibold text-tiki-text mb-2">
            {error || "Không tìm thấy đơn hàng"}
          </h1>
          <Link
            href="/"
            className="inline-block mt-4 px-6 py-2.5 bg-tiki-blue text-white rounded-lg font-semibold text-sm hover:bg-tiki-blue-dark transition"
          >
            Về trang chủ
          </Link>
        </main>
        <Footer />
      </>
    );
  }

  const statusIndex = getStatusIndex(order.status);

  return (
    <>
      <Header />
      <main className="py-4" style={{ backgroundColor: "#F5F5FA" }}>
        <div className="max-w-[800px] mx-auto px-4">
          {/* Order header */}
          <div className="bg-white rounded-lg border border-tiki-border p-4 mb-4">
            <div className="flex items-center justify-between mb-3">
              <div>
                <h1 className="text-base font-semibold text-tiki-text">
                  Đơn hàng #{order.id.slice(0, 8)}
                </h1>
                <p className="text-xs text-tiki-text-secondary">
                  Đặt lúc {formatDate(order.created_at)}
                </p>
              </div>
              <span
                className={`px-3 py-1 rounded-full text-xs font-medium ${
                  order.status === "completed"
                    ? "bg-green-100 text-green-700"
                    : order.status === "cancelled"
                    ? "bg-red-100 text-red-700"
                    : "bg-blue-100 text-blue-700"
                }`}
              >
                {STATUS_STEPS.find((s) => s.key === order.status)?.label || order.status}
              </span>
            </div>

            {/* ETA */}
            {order.status !== "completed" && order.status !== "cancelled" && order.duration_seconds > 0 && (
              <div className="flex items-center gap-4 p-3 bg-orange-50 border border-orange-200 rounded-lg mb-3">
                <Clock className="w-5 h-5 text-orange-500" />
                <div>
                  <p className="text-sm font-medium text-orange-700">
                    Dự kiến giao: {formatTime(order.duration_seconds)}
                  </p>
                  <p className="text-xs text-orange-600">
                    Khoảng cách: {order.distance_meters >= 1000
                      ? `${(order.distance_meters / 1000).toFixed(1)} km`
                      : `${order.distance_meters} m`}
                  </p>
                </div>
                {driverLocation && (
                  <div className="ml-auto flex items-center gap-1">
                    <span className="relative flex h-2.5 w-2.5">
                      <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75" />
                      <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-green-500" />
                    </span>
                    <span className="text-xs text-green-600 font-medium">Trực tiếp</span>
                  </div>
                )}
              </div>
            )}

            {/* Delivery locations */}
            <div className="space-y-3">
              <div className="flex items-start gap-3">
                <div className="w-5 h-5 rounded-full bg-green-100 flex items-center justify-center shrink-0 mt-0.5">
                  <div className="w-2 h-2 rounded-full bg-green-500" />
                </div>
                <div>
                  <p className="text-xs text-tiki-text-secondary">Lấy hàng</p>
                  <p className="text-sm text-tiki-text">
                    {order.pickup.address || `${order.pickup.lat.toFixed(4)}, ${order.pickup.lng.toFixed(4)}`}
                  </p>
                </div>
              </div>
              <div className="ml-2.5 h-4 border-l-2 border-dashed border-gray-300" />
              <div className="flex items-start gap-3">
                <div className="w-5 h-5 rounded-full bg-red-100 flex items-center justify-center shrink-0 mt-0.5">
                  <MapPin className="w-3 h-3 text-red-500" />
                </div>
                <div>
                  <p className="text-xs text-tiki-text-secondary">Giao hàng</p>
                  <p className="text-sm text-tiki-text font-medium">
                    {order.dropoff.address || `${order.dropoff.lat.toFixed(4)}, ${order.dropoff.lng.toFixed(4)}`}
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Status timeline */}
          <div className="bg-white rounded-lg border border-tiki-border p-4 mb-4">
            <h2 className="text-sm font-semibold text-tiki-text mb-4">Trạng thái đơn hàng</h2>
            <div className="space-y-0">
              {STATUS_STEPS.map((step, i) => {
                const Icon = step.icon;
                const isActive = i === statusIndex;
                const isDone = i < statusIndex;
                const isCancelled = order.status === "cancelled" && i > 0;

                return (
                  <div key={step.key} className="flex items-start gap-3">
                    <div className="flex flex-col items-center">
                      <div
                        className={`w-8 h-8 rounded-full flex items-center justify-center ${
                          isDone
                            ? "bg-green-500 text-white"
                            : isActive
                            ? "bg-tiki-blue text-white"
                            : isCancelled
                            ? "bg-gray-200 text-gray-400"
                            : "bg-gray-100 text-gray-400"
                        }`}
                      >
                        {isDone ? (
                          <CheckCircle className="w-4 h-4" />
                        ) : (
                          <Icon className="w-4 h-4" />
                        )}
                      </div>
                      {i < STATUS_STEPS.length - 1 && (
                        <div
                          className={`w-0.5 h-8 ${
                            isDone ? "bg-green-500" : "bg-gray-200"
                          }`}
                        />
                      )}
                    </div>
                    <div className="pt-1.5 pb-2">
                      <p
                        className={`text-sm ${
                          isActive
                            ? "font-semibold text-tiki-text"
                            : isDone
                            ? "text-tiki-text"
                            : "text-gray-400"
                        }`}
                      >
                        {step.label}
                      </p>
                      {isActive && (
                        <p className="text-xs text-tiki-text-secondary mt-0.5">Đang thực hiện</p>
                      )}
                      {isDone && i === statusIndex - 1 && order.assigned_at && step.key === "driver_assigned" && (
                        <p className="text-xs text-tiki-text-secondary mt-0.5">
                          {formatDate(order.assigned_at)}
                        </p>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Driver info */}
          {order.driver_id && (
            <div className="bg-white rounded-lg border border-tiki-border p-4 mb-4">
              <h2 className="text-sm font-semibold text-tiki-text mb-3">Tài xế</h2>
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-full bg-tiki-blue/10 flex items-center justify-center">
                  <User className="w-5 h-5 text-tiki-blue" />
                </div>
                <div className="flex-1">
                  <p className="text-sm font-medium text-tiki-text">{order.driver_id}</p>
                  <p className="text-xs text-tiki-text-secondary">Tài xế giao hàng</p>
                </div>
                {driverLocation && (
                  <span className="flex items-center gap-1 text-xs text-green-600">
                    <CircleDot className="w-3 h-3" />
                    Đang trên đường
                  </span>
                )}
              </div>
            </div>
          )}

          {/* Actions */}
          <div className="flex gap-3 mb-6">
            <Link
              href="/"
              className="flex-1 py-2.5 bg-tiki-blue text-white rounded-lg font-semibold text-sm text-center hover:bg-tiki-blue-dark transition"
            >
              Tiếp tục mua sắm
            </Link>
            <button
              onClick={fetchOrder}
              className="px-4 py-2.5 border border-tiki-border rounded-lg text-sm text-tiki-text hover:bg-gray-50 transition"
            >
              Làm mới
            </button>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
