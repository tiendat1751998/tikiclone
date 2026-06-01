"use client";

import { useState, useMemo } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";
import { useCartStore } from "@/stores/cart";
import { useAuthStore } from "@/stores/auth";
import { useUIStore } from "@/stores/ui";
import type { CartItem } from "@/types";

function groupBySeller(items: CartItem[]): Record<string, CartItem[]> {
  const groups: Record<string, CartItem[]> = {};
  for (const item of items) {
    const seller = item.shop_id || "default";
    if (!groups[seller]) groups[seller] = [];
    groups[seller].push(item);
  }
  return groups;
}

export default function CartPage() {
  const items = useCartStore((s) => s.items);
  const removeItem = useCartStore((s) => s.removeItem);
  const updateQuantity = useCartStore((s) => s.updateQuantity);
  const toggleSelect = useCartStore((s) => s.toggleSelect);
  const selectAll = useCartStore((s) => s.selectAll);
  const clearCart = useCartStore((s) => s.clearCart);
  const getSubtotal = useCartStore((s) => s.getSubtotal);
  const addToast = useUIStore((s) => s.addToast);
  const selectedItems = useMemo(() => items.filter((i) => i.is_selected), [items]);
  const allSelected = items.length > 0 && items.every((i) => i.is_selected);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const sellerGroups = useMemo(() => groupBySeller(items), [items]);

  const handleClearCart = () => {
    if (items.length === 0) return;
    clearCart();
    addToast({ type: "success", title: "Đã xóa giỏ hàng", message: "Tất cả sản phẩm đã được xóa khỏi giỏ hàng" });
  };

  const router = useRouter();

  const handleCheckout = () => {
    router.push("/checkout");
  };

  if (items.length === 0) {
    return (
      <>
        <Header />
        <main className="py-4">
          <div className="max-w-tiki mx-auto px-3">
            <div className="flex items-center gap-2 text-[11px] text-tiki-text-secondary mb-3">
              <Link href="/" className="hover:text-tiki-blue">Trang chủ</Link>
              <span>/</span>
              <span className="text-tiki-text">Giỏ hàng</span>
            </div>
            <div className="bg-white rounded-lg border border-tiki-border py-14 text-center">
              <div className="text-4xl mb-3">🛒</div>
              <h1 className="text-sm font-semibold text-tiki-text mb-1">Giỏ hàng trống</h1>
              <p className="text-xs text-tiki-text-secondary mb-4">Hãy thêm sản phẩm để mua sắm nhé!</p>
              <Link href="/products" className="inline-block px-5 py-2 bg-tiki-blue text-white rounded-lg text-xs font-medium hover:bg-tiki-blue-dark transition">
                Mua sắm ngay
              </Link>
            </div>
          </div>
        </main>
        <Footer />
      </>
    );
  }

  return (
    <>
      <Header />
      <main className="py-2 bg-tiki-bg">
        <div className="max-w-[1270px] mx-auto px-3">
          <div className="flex items-center h-8 text-[11px] text-tiki-text-secondary">
            <Link href="/" className="hover:text-tiki-blue">Trang chủ</Link>
            <span className="mx-1.5">›</span>
            <span className="text-tiki-text">Giỏ hàng ({items.length})</span>
          </div>

          {!isAuthenticated && (
            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3 mb-3 flex items-center justify-between">
              <div>
                <p className="text-xs text-yellow-800 font-medium">Đăng nhập để đồng bộ giỏ hàng</p>
                <p className="text-[10px] text-yellow-600 mt-0.5">Giỏ hàng đang lưu cục bộ. Đăng nhập để truy cập trên mọi thiết bị.</p>
              </div>
              <Link href="/login?redirect=/cart" className="shrink-0 ml-3 px-3 py-1.5 bg-tiki-blue text-white rounded-lg text-[11px] font-medium hover:bg-tiki-blue-dark transition">
                Đăng nhập
              </Link>
            </div>
          )}

          <div className="flex gap-4">
            <div className="flex-1 min-w-0 space-y-3">
              <div className="bg-white rounded-lg border border-tiki-border px-4 py-2.5 flex items-center justify-between">
                <label className="flex items-center gap-2 text-xs text-tiki-text cursor-pointer">
                  <input
                    type="checkbox"
                    checked={allSelected}
                    onChange={selectAll}
                    className="w-3.5 h-3.5 rounded border-gray-300"
                  />
                  Chọn tất cả ({items.length} sản phẩm)
                </label>
                <button onClick={handleClearCart} className="text-[10px] text-tiki-text-secondary hover:text-tiki-red">Xóa tất cả</button>
              </div>

              {Object.entries(sellerGroups).map(([sellerId, sellerItems]) => (
                <div key={sellerId} className="bg-white rounded-lg border border-tiki-border">
                  <div className="px-4 py-2.5 flex items-center gap-2 border-b border-tiki-border">
                    <input
                      type="checkbox"
                      checked={sellerItems.every((i) => i.is_selected)}
                      onChange={() => {
                        const allSel = sellerItems.every((i) => i.is_selected);
                        sellerItems.forEach((i) => {
                          if (i.is_selected === allSel) toggleSelect(i.id);
                        });
                      }}
                      className="w-3.5 h-3.5 rounded border-gray-300"
                    />
                    <span className="text-[11px] font-medium text-tiki-text">
                      {sellerItems[0]?.shop_name || (sellerId !== "default" ? `Shop ${sellerId.slice(-4)}` : "Sản phẩm")}
                    </span>
                  </div>

                  {sellerItems.map((item) => (
                    <div key={item.id} className="flex items-center gap-3 px-4 py-3 border-b border-tiki-border last:border-b-0">
                      <input
                        type="checkbox"
                        checked={item.is_selected}
                        onChange={() => toggleSelect(item.id)}
                        className="w-3.5 h-3.5 rounded border-gray-300 shrink-0"
                      />
                      <img
                        src={item.image_url || "/images/placeholder.svg"}
                        alt={item.name}
                        className="w-16 h-16 object-cover rounded border border-tiki-border shrink-0"
                      />
                      <div className="flex-1 min-w-0">
                        <p className="text-xs text-tiki-text line-clamp-2">{item.name}</p>
                        {item.sku_name && item.sku_name !== "default" && (
                          <p className="text-[10px] text-tiki-text-secondary mt-0.5">{item.sku_name}</p>
                        )}
                        <div className="flex items-center gap-1.5 mt-1">
                          <span className="text-sm font-medium text-tiki-text">{item.price?.toLocaleString("vi-VN")} ₫</span>
                          {item.original_price && item.original_price > item.price && (
                            <span className="text-[10px] text-tiki-text-secondary line-through">
                              {item.original_price.toLocaleString("vi-VN")} ₫
                            </span>
                          )}
                        </div>
                      </div>

                      <div className="flex items-center gap-2">
                        <div className="flex items-center border border-gray-300 rounded">
                          <button
                            className="w-7 h-7 flex items-center justify-center text-sm text-tiki-text border-r border-gray-300 hover:bg-gray-50"
                            onClick={() => updateQuantity(item.id, Math.max(1, item.quantity - 1))}
                          >
                            −
                          </button>
                          <input
                            type="text"
                            value={item.quantity}
                            readOnly
                            className="w-9 h-7 text-center text-xs border-none outline-none"
                          />
                          <button
                            className="w-7 h-7 flex items-center justify-center text-sm text-tiki-text border-l border-gray-300 hover:bg-gray-50"
                            onClick={() => updateQuantity(item.id, Math.min(item.stock || 99, item.quantity + 1))}
                          >
                            +
                          </button>
                        </div>
                        <button
                          onClick={() => removeItem(item.id)}
                          className="text-[10px] text-tiki-text-secondary hover:text-tiki-red shrink-0 px-2"
                        >
                          Xóa
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              ))}
            </div>

            <div className="w-[300px] shrink-0">
              <div className="bg-white rounded-lg border border-tiki-border p-4 sticky top-4">
                <h2 className="text-xs font-semibold text-tiki-text mb-3">Đơn hàng</h2>
                <div className="flex justify-between text-xs mb-2">
                  <span className="text-tiki-text-secondary">Tạm tính ({selectedItems.length} sản phẩm)</span>
                  <span className="text-tiki-text font-medium">{getSubtotal().toLocaleString("vi-VN")} ₫</span>
                </div>
                <div className="flex justify-between text-xs mb-2">
                  <span className="text-tiki-text-secondary">Phí vận chuyển</span>
                  <span className="text-tiki-text-secondary">—</span>
                </div>
                <div className="flex justify-between text-sm font-semibold text-tiki-text pt-2 border-t border-tiki-border">
                  <span>Tổng tiền</span>
                  <span>{getSubtotal().toLocaleString("vi-VN")} ₫</span>
                </div>
                <p className="text-[10px] text-tiki-text-secondary mt-1">(Đã bao gồm VAT nếu có)</p>
                <button
                  disabled={selectedItems.length === 0}
                  onClick={handleCheckout}
                  className="w-full mt-3 py-2.5 bg-tiki-red text-white rounded-lg font-semibold text-xs hover:bg-tiki-red-dark transition disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Mua hàng ({selectedItems.length})
                </button>
                <Link href="/products" className="block text-center mt-2 text-[11px] text-tiki-blue hover:underline">
                  ← Tiếp tục mua sắm
                </Link>
              </div>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
