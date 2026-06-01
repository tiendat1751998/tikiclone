"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { customersApi } from "@/lib/api/client";
import type { Customer } from "@/types";

export default function AdminCustomersPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["admin-customers", page, search],
    queryFn: () => customersApi.list({ page: String(page), ...(search ? { q: search } : {}) }),
  });

  const customers = data?.items || (Array.isArray(data) ? data : []);
  const totalPages = data?.total_pages || 1;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-tiki-text">Khách hàng ({data?.total || 0})</h2>
      </div>

      {/* Search */}
      <div className="bg-white rounded-lg border border-tiki-border p-3">
        <input
          type="text"
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1); }}
          placeholder="Tìm theo tên, email, SĐT..."
          className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-tiki-blue focus:ring-1 focus:ring-tiki-blue outline-none"
        />
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg border border-tiki-border overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-50 border-b border-tiki-border">
                <th className="text-left px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Khách hàng</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Email</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-tiki-text-secondary">SĐT</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Đơn hàng</th>
                <th className="text-right px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Tổng chi</th>
                <th className="text-center px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Trạng thái</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-tiki-text-secondary">Ngày đăng ký</th>
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
              ) : customers.length === 0 ? (
                <tr><td colSpan={7} className="px-4 py-12 text-center text-sm text-tiki-text-secondary">Không có khách hàng</td></tr>
              ) : (
                customers.map((customer: Customer) => (
                  <tr key={customer.id} className="border-b border-tiki-border hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center text-xs font-bold text-tiki-blue">
                          {customer.display_name?.charAt(0) || "?"}
                        </div>
                        <span className="text-sm text-tiki-text">{customer.display_name}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-tiki-text-secondary">{customer.email}</td>
                    <td className="px-4 py-3 text-tiki-text-secondary">{customer.phone || "—"}</td>
                    <td className="px-4 py-3 text-right text-tiki-text">{customer.total_orders}</td>
                    <td className="px-4 py-3 text-right font-medium text-tiki-text">{customer.total_spent?.toLocaleString("vi-VN")} ₫</td>
                    <td className="px-4 py-3 text-center">
                      <span className={`text-[10px] font-medium px-2 py-0.5 rounded ${
                        customer.status === "active" ? "bg-green-100 text-green-700" : "bg-red-100 text-red-700"
                      }`}>
                        {customer.status === "active" ? "Hoạt động" : "Khoá"}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-tiki-text-secondary">{new Date(customer.created_at).toLocaleDateString("vi-VN")}</td>
                  </tr>
                ))
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
