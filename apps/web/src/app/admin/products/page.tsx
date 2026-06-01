"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { productsApi } from "@/lib/api/client";
import type { Product } from "@/types";

export default function AdminProductsPage() {
  const [page, setPage] = useState(1);
  const { data, isLoading } = useQuery({
    queryKey: ["admin-products", page],
    queryFn: () => productsApi.list({ page: String(page), limit: "20" }),
  });

  const products = data?.products || (Array.isArray(data) ? data : []);
  const totalPages = data?.total_pages || 1;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-tiki-text">Sản phẩm ({data?.total || 0})</h2>
        <button className="px-4 py-2 bg-tiki-blue text-white rounded-lg text-sm font-medium hover:bg-tiki-blue-dark transition">
          + Thêm sản phẩm
        </button>
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg border border-tiki-border overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-50 border-b border-tiki-border">
                <th className="text-left px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Sản phẩm</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Danh mục</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Giá</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Kho</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Đã bán</th>
                <th className="text-center px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Trạng thái</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Thao tác</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className="border-b border-tiki-border animate-pulse">
                    <td className="px-4 py-3"><div className="h-4 bg-gray-200 rounded w-40" /></td>
                    <td className="px-4 py-3"><div className="h-4 bg-gray-200 rounded w-20" /></td>
                    <td className="px-4 py-3"><div className="h-4 bg-gray-200 rounded w-16 ml-auto" /></td>
                    <td className="px-4 py-3"><div className="h-4 bg-gray-200 rounded w-10 ml-auto" /></td>
                    <td className="px-4 py-3"><div className="h-4 bg-gray-200 rounded w-10 ml-auto" /></td>
                    <td className="px-4 py-3"><div className="h-4 bg-gray-200 rounded w-16 mx-auto" /></td>
                    <td className="px-4 py-3"><div className="h-4 bg-gray-200 rounded w-16 ml-auto" /></td>
                  </tr>
                ))
              ) : products.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center text-sm text-tiki-text-secondary">
                    Chưa có sản phẩm nào
                  </td>
                </tr>
              ) : (
                products.map((product: Product) => (
                  <tr key={product.id} className="border-b border-tiki-border hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-3">
                        <img src={product.image_url || "/images/placeholder.svg"} alt={product.name} className="w-10 h-10 rounded object-cover" />
                        <span className="text-sm text-tiki-text max-w-xs truncate">{product.name}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-tiki-text-secondary">{product.category_name || product.category_id}</td>
                    <td className="px-4 py-3 text-right text-tiki-text font-medium">{product.price?.toLocaleString("vi-VN")} ₫</td>
                    <td className="px-4 py-3 text-right text-tiki-text-secondary">{product.stock}</td>
                    <td className="px-4 py-3 text-right text-tiki-text-secondary">{product.sold_count}</td>
                    <td className="px-4 py-3 text-center">
                      <span className={`text-[10px] font-medium px-2 py-0.5 rounded ${
                        product.status === "active" ? "bg-green-100 text-green-700" :
                        product.status === "inactive" ? "bg-gray-100 text-gray-600" :
                        product.status === "out_of_stock" ? "bg-red-100 text-red-700" :
                        "bg-yellow-100 text-yellow-700"
                      }`}>
                        {product.status === "active" ? "Hoạt động" :
                         product.status === "inactive" ? "Tạm ẩn" :
                         product.status === "out_of_stock" ? "Hết hàng" : "Nháp"}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button className="text-xs text-tiki-blue hover:underline mr-2">Sửa</button>
                      <button className="text-xs text-red-500 hover:underline">Xóa</button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-tiki-border">
            <span className="text-xs text-tiki-text-secondary">Trang {page} / {totalPages}</span>
            <div className="flex gap-1">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
                className="px-3 py-1 text-xs border border-tiki-border rounded hover:bg-gray-50 disabled:opacity-50"
              >
                Trước
              </button>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
                className="px-3 py-1 text-xs border border-tiki-border rounded hover:bg-gray-50 disabled:opacity-50"
              >
                Sau
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
