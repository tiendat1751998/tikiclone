"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";

interface InventoryItem {
  id: string;
  product_id: string;
  product_name: string;
  sku: string;
  stock: number;
  reserved: number;
  available: number;
  low_stock: boolean;
}

export default function AdminInventoryPage() {
  const [filter, setFilter] = useState<"all" | "low" | "out">("all");

  const { data: inventory, isLoading } = useQuery({
    queryKey: ["admin-inventory"],
    queryFn: () => api.get<InventoryItem[]>("/inventory"),
    initialData: [
      { id: "inv-1", product_id: "prod-1", product_name: "iPhone 15 Pro Max 256GB", sku: "IP15PM-256", stock: 45, reserved: 5, available: 40, low_stock: false },
      { id: "inv-2", product_id: "prod-2", product_name: "Samsung Galaxy S24 Ultra", sku: "SS24U-512", stock: 8, reserved: 3, available: 5, low_stock: true },
      { id: "inv-3", product_id: "prod-3", product_name: "MacBook Air M3 13\"", sku: "MBA-M3-256", stock: 0, reserved: 0, available: 0, low_stock: true },
      { id: "inv-4", product_id: "prod-4", product_name: "AirPods Pro 2 USB-C", sku: "APP2-USBC", stock: 120, reserved: 15, available: 105, low_stock: false },
      { id: "inv-5", product_id: "prod-5", product_name: "iPad Air M2 11\"", sku: "IPA-M2-128", stock: 3, reserved: 1, available: 2, low_stock: true },
    ],
  });

  const filtered = (inventory || []).filter((item) => {
    if (filter === "low") return item.low_stock && item.stock > 0;
    if (filter === "out") return item.stock === 0;
    return true;
  });

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold text-tiki-text">Kho hàng</h2>

      {/* Summary cards */}
      <div className="grid grid-cols-3 gap-4">
        <div className="bg-white rounded-lg border border-tiki-border p-4">
          <p className="text-xs text-tiki-text-secondary">Tổng SKU</p>
          <p className="text-2xl font-bold text-tiki-text">{inventory?.length || 0}</p>
        </div>
        <div className="bg-white rounded-lg border border-tiki-border p-4">
          <p className="text-xs text-tiki-text-secondary">Sắp hết hàng</p>
          <p className="text-2xl font-bold text-orange-600">{(inventory || []).filter((i) => i.low_stock && i.stock > 0).length}</p>
        </div>
        <div className="bg-white rounded-lg border border-tiki-border p-4">
          <p className="text-xs text-tiki-text-secondary">Hết hàng</p>
          <p className="text-2xl font-bold text-red-600">{(inventory || []).filter((i) => i.stock === 0).length}</p>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg border border-tiki-border p-3 flex gap-2">
        {(["all", "low", "out"] as const).map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-3 py-1.5 text-xs rounded-lg font-medium transition ${
              filter === f ? "bg-tiki-blue text-white" : "bg-gray-100 text-tiki-text-secondary hover:bg-gray-200"
            }`}
          >
            {f === "all" ? "Tất cả" : f === "low" ? "Sắp hết" : "Hết hàng"}
          </button>
        ))}
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg border border-tiki-border overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-50 border-b border-tiki-border">
                <th className="text-left px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Sản phẩm</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-tiki-text-secondary">SKU</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Tồn kho</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Đã đặt</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Khả dụng</th>
                <th className="text-center px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Trạng thái</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className="border-b border-tiki-border animate-pulse">
                    {Array.from({ length: 6 }).map((_, j) => (
                      <td key={j} className="px-4 py-3"><div className="h-4 bg-gray-200 rounded" /></td>
                    ))}
                  </tr>
                ))
              ) : filtered.length === 0 ? (
                <tr><td colSpan={6} className="px-4 py-12 text-center text-sm text-tiki-text-secondary">Không có dữ liệu</td></tr>
              ) : (
                filtered.map((item) => (
                  <tr key={item.id} className="border-b border-tiki-border hover:bg-gray-50">
                    <td className="px-4 py-3 text-tiki-text">{item.product_name}</td>
                    <td className="px-4 py-3 text-tiki-text-secondary font-mono text-xs">{item.sku}</td>
                    <td className="px-4 py-3 text-right text-tiki-text font-medium">{item.stock}</td>
                    <td className="px-4 py-3 text-right text-tiki-text-secondary">{item.reserved}</td>
                    <td className="px-4 py-3 text-right text-tiki-text font-medium">{item.available}</td>
                    <td className="px-4 py-3 text-center">
                      <span className={`text-[10px] font-medium px-2 py-0.5 rounded ${
                        item.stock === 0 ? "bg-red-100 text-red-700" :
                        item.low_stock ? "bg-orange-100 text-orange-700" :
                        "bg-green-100 text-green-700"
                      }`}>
                        {item.stock === 0 ? "Hết hàng" : item.low_stock ? "Sắp hết" : "Bình thường"}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
