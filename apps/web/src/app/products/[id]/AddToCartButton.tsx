"use client";
import { useState } from "react";
import { useCartStore } from "@/stores/cart";
import { useUIStore } from "@/stores/ui";
import { useAuthStore } from "@/stores/auth";
import Link from "next/link";

export function AddToCartButton({ product, quantity: qty = 1 }: { product: { id: string; name: string; image_url: string; price: number; stock?: number; shop_id?: string; shop_name?: string }; quantity?: number }) {
  const [isAdding, setIsAdding] = useState(false);
  const addItem = useCartStore((s) => s.addItem);
  const addToast = useUIStore((s) => s.addToast);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const items = useCartStore((s) => s.items);
  const guestItemsCount = isAuthenticated ? 0 : items.length;

  const handleAdd = async () => {
    setIsAdding(true);
    try {
      await addItem({
        product_id: product.id, sku_id: product.id, name: product.name,
        image_url: product.image_url, price: product.price, quantity: qty,
        stock: product.stock || 0, shop_id: product.shop_id, shop_name: product.shop_name,
      });
      addToast({ type: "success", title: "Đã thêm vào giỏ hàng", message: product.name });
    } catch {
      addToast({ type: "error", title: "Có lỗi xảy ra", message: "Không thể thêm sản phẩm" });
    } finally {
      setIsAdding(false);
    }
  };

  return (
    <div className="flex-1 space-y-2">
      <button onClick={handleAdd} disabled={isAdding}
        style={{ width: "100%" }}
        className="py-3 border border-tiki-blue text-tiki-blue rounded font-semibold text-sm hover:bg-blue-50 transition disabled:opacity-50">
        {isAdding ? "Đang thêm..." : "Thêm vào giỏ"}
      </button>
      {guestItemsCount > 0 && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3 text-center">
          <p className="text-xs text-yellow-800 mb-1">
            Bạn có {guestItemsCount} sản phẩm trong giỏ hàng (chưa đăng nhập)
          </p>
          <Link
            href="/login?redirect=/cart"
            className="text-xs text-tiki-blue hover:underline font-medium"
          >
            Đăng nhập để đồng bộ giỏ hàng
          </Link>
        </div>
      )}
    </div>
  );
}
