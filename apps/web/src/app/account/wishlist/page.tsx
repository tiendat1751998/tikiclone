"use client";

import { useState } from "react";
import Link from "next/link";
import { ProductCard } from "@/components/storefront/product/ProductCard";
import type { Product } from "@/types";

export default function WishlistPage() {
  const [items, setItems] = useState<Product[]>([]);

  if (items.length === 0) {
    return (
      <div className="max-w-3xl">
        <h1 className="text-lg font-semibold text-tiki-text mb-4">Sản phẩm yêu thích</h1>
        <div className="bg-white rounded-lg border border-tiki-border py-16 text-center">
          <p className="text-4xl mb-3">❤️</p>
          <p className="text-sm text-tiki-text-secondary mb-1">Danh sách yêu thích trống</p>
          <p className="text-xs text-tiki-text-secondary">Nhấn ❤️ trên sản phẩm để thêm vào danh sách</p>
          <Link href="/products" className="inline-block mt-4 px-4 py-2 bg-tiki-blue text-white rounded-lg text-sm font-medium hover:bg-tiki-blue-dark transition">
            Khám phá sản phẩm
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-5xl">
      <h1 className="text-lg font-semibold text-tiki-text mb-4">Sản phẩm yêu thích ({items.length})</h1>
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3">
        {items.map((product) => (
          <div key={product.id} className="relative">
            <ProductCard product={product} />
            <button
              onClick={() => setItems((prev) => prev.filter((p) => p.id !== product.id))}
              className="absolute top-2 right-2 w-7 h-7 bg-white rounded-full shadow flex items-center justify-center text-red-500 text-xs hover:bg-red-50 transition"
              title="Bỏ yêu thích"
            >
              ✕
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}