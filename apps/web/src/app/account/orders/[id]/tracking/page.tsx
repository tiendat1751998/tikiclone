"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";

interface TrackingEvent {
  status: string;
  description: string;
  timestamp: string;
  location: string;
}

export default function TrackingPage({ params }: { params: Promise<{ id: string }> }) {
  const router = useRouter();
  const { id } = useParams_react(params);

  const [tracking, setTracking] = useState<{
    order_id: string;
    status: string;
    carrier: string;
    tracking_number: string;
    estimated_delivery: string;
    events: TrackingEvent[];
  } | null>(null);

  // In production: const { data: tracking } = useQuery({...})
  // Mock data for now
  const mockTracking = {
    order_id: id,
    status: "shipped",
    carrier: "Giao Hàng Nhanh",
    tracking_number: "GHN" + id.slice(0, 8).toUpperCase(),
    estimated_delivery: new Date(Date.now() + 2 * 24 * 60 * 60 * 1000).toLocaleDateString("vi-VN"),
    events: [
      { status: "picked", description: "Đã lấy hàng từ người bán", timestamp: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(), location: "Kho TPHCM" },
      { status: "transit", description: "Đang vận chuyển", timestamp: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(), location: "Trung chuyển Hà Nội" },
      { status: "delivering", description: "Đang giao hàng", timestamp: new Date().toISOString(), location: "Bưu cục Hoàn Kiếm" },
    ],
  };

  const statusLabel: Record<string, string> = {
    pending: "Chờ lấy hàng",
    picked: "Đã lấy hàng",
    transit: "Đang vận chuyển",
    delivering: "Đang giao hàng",
    delivered: "Đã giao hàng",
    failed: "Giao thất bại",
  };

  return (
    <main className="bg-[#F5F5FA] py-4 min-h-[60vh]">
      <div className="max-w-3xl mx-auto px-3">
        <div className="flex items-center gap-2 mb-4">
          <Link href={`/account/orders/${id}`} className="text-tiki-blue text-sm hover:underline">← Chi tiết đơn hàng</Link>
          <span className="text-tiki-text-secondary text-sm">/</span>
          <span className="text-sm text-tiki-text-secondary">Theo dõi đơn hàng</span>
        </div>

        {/* Tracking card */}
        <div className="bg-white rounded-lg border border-tiki-border p-4 mb-4">
          <div className="flex items-center justify-between flex-wrap gap-2">
            <div>
              <p className="text-xs text-tiki-text-secondary">Mã vận đơn</p>
              <p className="text-sm font-bold text-tiki-text">{mockTracking.tracking_number}</p>
            </div>
            <div>
              <p className="text-xs text-tiki-text-secondary">Đơn vị vận chuyển</p>
              <p className="text-sm font-medium text-tiki-text">{mockTracking.carrier}</p>
            </div>
            <div>
              <p className="text-xs text-tiki-text-secondary">Dự kiến giao</p>
              <p className="text-sm font-medium text-tiki-text">{mockTracking.estimated_delivery}</p>
            </div>
            <div>
              <span className="text-xs font-medium px-2 py-0.5 rounded bg-indigo-100 text-indigo-700">
                {statusLabel[mockTracking.status] || mockTracking.status}
              </span>
            </div>
          </div>
        </div>

        {/* Timeline */}
        <div className="bg-white rounded-lg border border-tiki-border p-4">
          <h3 className="text-sm font-semibold text-tiki-text mb-4">Lịch trình giao hàng</h3>
          <div className="space-y-0">
            {mockTracking.events.map((event, i) => (
              <div key={i} className="flex gap-3">
                <div className="flex flex-col items-center">
                  <div className={`w-3 h-3 rounded-full shrink-0 ${i === mockTracking.events.length - 1 ? "bg-green-500" : "bg-green-300"}`} />
                  {i < mockTracking.events.length - 1 && <div className="w-0.5 h-full bg-green-200 my-1" />}
                </div>
                <div className="pb-4 flex-1">
                  <p className="text-sm text-tiki-text font-medium">{event.description}</p>
                  <p className="text-xs text-tiki-text-secondary">{event.location}</p>
                  <p className="text-xs text-tiki-text-secondary">{new Date(event.timestamp).toLocaleString("vi-VN")}</p>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="mt-4">
          <Link href={`/account/orders/${id}`} className="px-4 py-2 border border-tiki-border rounded-lg text-sm text-tiki-text-secondary hover:bg-gray-50 transition">
            ← Quay lại đơn hàng
          </Link>
        </div>
      </div>
    </main>
  );
}

function useParams_react(params: Promise<{ id: string }>): { id: string } {
  const [resolved, setResolved] = useState({ id: "" });
  useEffect(() => { params.then(setResolved); }, []);
  return resolved;
}
