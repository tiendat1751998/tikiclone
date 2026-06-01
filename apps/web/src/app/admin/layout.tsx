"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuthStore } from "@/stores/auth";

const NAV_ITEMS = [
  { href: "/admin", label: "Tổng quan", icon: "📊" },
  { href: "/admin/analytics", label: "Phân tích", icon: "📈" },
  { href: "/admin/orders", label: "Đơn hàng", icon: "📦" },
  { href: "/admin/products", label: "Sản phẩm", icon: "🛍️" },
  { href: "/admin/customers", label: "Khách hàng", icon: "👥" },
  { href: "/admin/inventory", label: "Kho hàng", icon: "🏪" },
];

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const user = useAuthStore((s) => s.user);

  useEffect(() => {
    if (!isAuthenticated) {
      router.push("/login");
      return;
    }
    if (user && user.role !== "admin") {
      router.push("/");
    }
  }, [isAuthenticated, user, router]);

  return (
    <div className="min-h-screen bg-tiki-bg flex">
      <aside className="w-56 admin-sidebar shrink-0 fixed inset-y-0 left-0 overflow-y-auto">
        <div className="p-4 border-b border-white/10">
          <Link href="/" className="flex items-center gap-2">
            <svg width="32" height="16" viewBox="0 0 64 22" fill="none">
              <rect width="64" height="22" rx="4" fill="#1A94FF" />
              <text x="10" y="16" fill="white" fontSize="12" fontWeight="700" fontFamily="Inter, sans-serif">Tiki</text>
            </svg>
            <span className="text-[10px] text-white/60 ml-1 uppercase tracking-wider">Admin</span>
          </Link>
        </div>
        <nav className="p-3 space-y-0.5">
          {NAV_ITEMS.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="admin-sidebar__nav-link"
            >
              <span className="text-base">{item.icon}</span>
              {item.label}
            </Link>
          ))}
        </nav>
      </aside>

      <div className="flex-1 ml-56">
        <header className="bg-white border-b border-tiki-border px-6 py-2.5 flex items-center justify-between sticky top-0 z-10">
          <h1 className="text-xs font-semibold text-tiki-text">Dashboard</h1>
          <div className="flex items-center gap-3">
            <Link href="/" className="text-[11px] text-tiki-text-secondary hover:text-tiki-blue">Xem trang chủ</Link>
            <span className="text-[11px] text-tiki-text-secondary">{user?.display_name || user?.email}</span>
          </div>
        </header>
        <main className="p-6">{children}</main>
      </div>
    </div>
  );
}
