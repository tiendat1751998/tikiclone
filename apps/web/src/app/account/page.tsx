"use client";

import { useState } from "react";
import { useAuthStore } from "@/stores/auth";
import { authApi } from "@/lib/api/client";
import Link from "next/link";

const NAV_ITEMS = [
  { href: "/account", label: "Thông tin tài khoản", icon: "👤" },
  { href: "/account/orders", label: "Quản lý đơn hàng", icon: "📦" },
  { href: "/account/orders", label: "Quản lý đổi trả", icon: "🔄" },
  { href: "/account/addresses", label: "Sổ địa chỉ", icon: "📍" },
  { href: "/account", label: "Thông tin thanh toán", icon: "💳" },
  { href: "/account", label: "Đánh giá sản phẩm", icon: "⭐" },
  { href: "/account", label: "Sản phẩm đã xem", icon: "👁️" },
  { href: "/account/wishlist", label: "Sản phẩm yêu thích", icon: "❤️" },
  { href: "/account", label: "Tiki VIP (Thành viên)", icon: "👑" },
  { href: "/account", label: "Mã giảm giá", icon: "🎟️" },
  { href: "/account", label: "Hỗ trợ khách hàng", icon: "❓" },
];

const DAYS = Array.from({ length: 31 }, (_, i) => i + 1);
const MONTHS = Array.from({ length: 12 }, (_, i) => i + 1);
const YEARS = Array.from({ length: 80 }, (_, i) => 1950 + i).reverse();

export default function AccountPage() {
  const user = useAuthStore((s) => s.user);
  const refreshUser = useAuthStore((s) => s.refreshUser);
  const [displayName, setDisplayName] = useState(user?.display_name || "");
  const [nickname, setNickname] = useState(user?.username || "");
  const [savingPersonal, setSavingPersonal] = useState(false);
  const [gender, setGender] = useState<"male" | "female" | "other">("male");
  const [birthday, setBirthday] = useState({ day: "1", month: "1", year: "1990" });
  const [nationality, setNationality] = useState("");

  async function handleSavePersonal(e: React.FormEvent) {
    e.preventDefault();
    setSavingPersonal(true);
    try {
      await authApi.put("/auth/profile", { display_name: displayName });
      await refreshUser();
    } finally {
      setSavingPersonal(false);
    }
  }

  if (!user) return null;

  return (
    <div className="grid grid-cols-12 gap-5">
      {/* Left Sidebar - User Navigation */}
      <div className="col-span-3">
        <div className="bg-white rounded-lg border border-tiki-border overflow-hidden">
          {/* Profile Header */}
          <div className="flex items-center gap-3 p-4 border-b border-tiki-border">
            <img
              src="/images/placeholder.svg"
              alt="Avatar"
              className="w-12 h-12 rounded-full object-cover"
            />
            <div>
              <p className="text-xs text-tiki-text-secondary">Tài khoản của</p>
              <p className="text-sm font-semibold text-tiki-text">{user.display_name || user.username}</p>
            </div>
          </div>

          {/* Navigation Menu */}
          <nav className="py-2">
            {NAV_ITEMS.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className={`flex items-center gap-3 px-4 py-2.5 text-sm transition ${
                  item.href === "/account"
                    ? "text-tiki-blue bg-tiki-bg"
                    : "text-tiki-text hover:bg-tiki-bg"
                }`}
              >
                <span className="text-base w-5">{item.icon}</span>
                <span>{item.label}</span>
              </Link>
            ))}
          </nav>
        </div>
      </div>

      {/* Right Main Content - Profile Dashboard */}
      <div className="col-span-9">
        <div className="bg-white rounded-lg border border-tiki-border p-6">
          <h3 className="text-lg font-semibold text-tiki-text mb-5">Thông tin tài khoản</h3>

          <div className="grid grid-cols-2 gap-8">
            {/* LEFT INTERNAL PANE: Personal Identification */}
            <div>
              <h4 className="text-xs font-semibold text-tiki-text mb-4">Thông tin cá nhân</h4>

              {/* Avatar Upload Widget */}
              <div className="flex items-center gap-4 mb-5">
                <img
                  src="/images/placeholder.svg"
                  alt="Avatar"
                  className="w-20 h-20 rounded-full object-cover border border-tiki-border"
                />
                <button className="text-xs text-tiki-blue border border-tiki-blue px-3 py-1.5 rounded hover:bg-blue-50 transition">
                  Chọn ảnh
                </button>
              </div>

              {/* Form Fields */}
              <form onSubmit={handleSavePersonal} className="space-y-4">
                <div>
                  <label className="block text-xs font-medium text-tiki-text-secondary mb-1.5">Họ & Tên</label>
                  <input
                    type="text"
                    value={displayName}
                    onChange={(e) => setDisplayName(e.target.value)}
                    className="w-full rounded border border-gray-300 px-3 py-1.5 text-sm focus:border-tiki-blue outline-none"
                    placeholder="Nhập họ và tên"
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-tiki-text-secondary mb-1.5">Nickname</label>
                  <input
                    type="text"
                    value={nickname}
                    onChange={(e) => setNickname(e.target.value)}
                    className="w-full rounded border border-gray-300 px-3 py-1.5 text-sm focus:border-tiki-blue outline-none"
                    placeholder="Nhập nickname"
                  />
                </div>

                {/* Birthday Selector */}
                <div>
                  <label className="block text-xs font-medium text-tiki-text-secondary mb-1.5">Ngày sinh</label>
                  <div className="grid grid-cols-3 gap-2">
                    <select
                      value={birthday.day}
                      onChange={(e) => setBirthday({ ...birthday, day: e.target.value })}
                      className="rounded border border-gray-300 px-2 py-1.5 text-sm focus:border-tiki-blue outline-none"
                    >
                      <option value="">Ngày</option>
                      {DAYS.map((d) => (
                        <option key={d} value={d}>{d}</option>
                      ))}
                    </select>
                    <select
                      value={birthday.month}
                      onChange={(e) => setBirthday({ ...birthday, month: e.target.value })}
                      className="rounded border border-gray-300 px-2 py-1.5 text-sm focus:border-tiki-blue outline-none"
                    >
                      <option value="">Tháng</option>
                      {MONTHS.map((m) => (
                        <option key={m} value={m}>{m}</option>
                      ))}
                    </select>
                    <select
                      value={birthday.year}
                      onChange={(e) => setBirthday({ ...birthday, year: e.target.value })}
                      className="rounded border border-gray-300 px-2 py-1.5 text-sm focus:border-tiki-blue outline-none"
                    >
                      <option value="">Năm</option>
                      {YEARS.map((y) => (
                        <option key={y} value={y}>{y}</option>
                      ))}
                    </select>
                  </div>
                </div>

                {/* Gender Selection */}
                <div>
                  <label className="block text-xs font-medium text-tiki-text-secondary mb-1.5">Giới tính</label>
                  <div className="flex items-center gap-6">
                    <label className="flex items-center gap-1.5 cursor-pointer">
                      <input
                        type="radio"
                        name="gender"
                        checked={gender === "male"}
                        onChange={() => setGender("male")}
                        className="w-4 h-4"
                      />
                      <span className="text-sm text-tiki-text">Nam</span>
                    </label>
                    <label className="flex items-center gap-1.5 cursor-pointer">
                      <input
                        type="radio"
                        name="gender"
                        checked={gender === "female"}
                        onChange={() => setGender("female")}
                        className="w-4 h-4"
                      />
                      <span className="text-sm text-tiki-text">Nữ</span>
                    </label>
                    <label className="flex items-center gap-1.5 cursor-pointer">
                      <input
                        type="radio"
                        name="gender"
                        checked={gender === "other"}
                        onChange={() => setGender("other")}
                        className="w-4 h-4"
                      />
                      <span className="text-sm text-tiki-text">Khác</span>
                    </label>
                  </div>
                </div>

                {/* Nationality */}
                <div>
                  <label className="block text-xs font-medium text-tiki-text-secondary mb-1.5">Quốc tịch</label>
                  <select
                    value={nationality}
                    onChange={(e) => setNationality(e.target.value)}
                    className="w-full rounded border border-gray-300 px-3 py-1.5 text-sm focus:border-tiki-blue outline-none"
                  >
                    <option value="">Chọn quốc tịch</option>
                    <option value="Việt Nam">Việt Nam</option>
                    <option value="USA">USA</option>
                    <option value="China">China</option>
                  </select>
                </div>

                <button
                  type="submit"
                  disabled={savingPersonal}
                  className="mt-2 bg-tiki-blue text-white px-8 py-2 rounded-md text-sm font-medium hover:bg-tiki-blue-dark transition disabled:opacity-50"
                >
                  {savingPersonal ? "Đang lưu..." : "Lưu thay đổi"}
                </button>
              </form>
            </div>

            {/* RIGHT INTERNAL PANE: Security, Contacts & Integrations */}
            <div className="flex flex-col gap-6">
              {/* Contacts Section */}
              <div>
                <h4 className="text-xs font-semibold text-tiki-text mb-3">Số điện thoại và Email</h4>

                <div className="flex items-center justify-between py-3 border-b border-tiki-border">
                  <div className="flex items-center gap-3">
                    <span className="text-lg">📞</span>
                    <div>
                      <p className="text-xs text-tiki-text-secondary">Số điện thoại</p>
                      <p className="text-sm text-tiki-text">(+84) ********89</p>
                    </div>
                  </div>
                  <button className="text-xs border border-tiki-blue text-tiki-blue px-3 py-1 rounded hover:bg-blue-50 transition">
                    Cập nhật
                  </button>
                </div>

                <div className="flex items-center justify-between py-3">
                  <div className="flex items-center gap-3">
                    <span className="text-lg">✉️</span>
                    <div>
                      <p className="text-xs text-tiki-text-secondary">Địa chỉ email</p>
                      <p className="text-sm text-tiki-text">{user.email}</p>
                    </div>
                  </div>
                  <button className="text-xs border border-tiki-blue text-tiki-blue px-3 py-1 rounded hover:bg-blue-50 transition">
                    Cập nhật
                  </button>
                </div>
              </div>

              {/* Security Section */}
              <div>
                <h4 className="text-xs font-semibold text-tiki-text mb-3">Bảo mật</h4>

                <div className="flex items-center justify-between py-3 border-b border-tiki-border">
                  <div>
                    <p className="text-sm text-tiki-text">Thiết lập mật khẩu</p>
                    <p className="text-xs text-tiki-text-secondary mt-0.5">Bảo vệ tài khoản của bạn</p>
                  </div>
                  <button className="text-xs border border-tiki-blue text-tiki-blue px-3 py-1 rounded hover:bg-blue-50 transition">
                    Cập nhật
                  </button>
                </div>

                <div className="flex items-center justify-between py-3 border-b border-tiki-border">
                  <div>
                    <p className="text-sm text-tiki-text">Thiết lập mã PIN</p>
                    <p className="text-xs text-tiki-text-secondary mt-0.5">Yêu cầu đối với các dịch vụ</p>
                  </div>
                  <button className="text-xs border border-tiki-blue text-tiki-blue px-3 py-1 rounded hover:bg-blue-50 transition">
                    Thiết lập
                  </button>
                </div>

                <div className="flex items-center justify-between py-3">
                  <div>
                    <p className="text-sm text-tiki-text">Yêu cầu xóa tài khoản</p>
                    <p className="text-xs text-tiki-text-secondary mt-0.5">Thao tác không thể hoàn tác</p>
                  </div>
                  <button className="text-xs border border-red-500 text-red-500 px-3 py-1 rounded hover:bg-red-50 transition">
                    Yêu cầu
                  </button>
                </div>
              </div>

              {/* Social Integrations */}
              <div>
                <h4 className="text-xs font-semibold text-tiki-text mb-3">Liên kết mạng xã hội</h4>

                <div className="flex items-center justify-between py-3 border-b border-tiki-border">
                  <div className="flex items-center gap-3">
                    <span className="text-blue-600 font-bold text-sm">f</span>
                    <div>
                      <p className="text-sm text-tiki-text">Facebook</p>
                    </div>
                  </div>
                  <button className="text-xs text-tiki-text-secondary hover:text-tiki-blue transition">
                    Liên kết
                  </button>
                </div>

                <div className="flex items-center justify-between py-3">
                  <div className="flex items-center gap-3">
                    <span className="text-red-500 font-bold">G</span>
                    <div>
                      <p className="text-sm text-tiki-text">Google</p>
                    </div>
                  </div>
                  <span className="text-xs bg-gray-100 text-tiki-text-secondary px-2 py-0.5 rounded">
                    Đã liên kết
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}