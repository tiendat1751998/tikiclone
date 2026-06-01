"use client";

import { useState } from "react";
import Link from "next/link";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";
import { ProductCard } from "@/components/storefront/product/ProductCard";
import { Button } from "@/components/ui";
import { Product } from "@/types";

// Mock seller data - in production this comes from API
const SELLER_INFO = {
  id: "",
  name: "",
  avatar: "",
  rating: 4.8,
  totalProducts: 0,
  totalSales: 0,
  joinedDate: "",
  responseTime: "trong vài giờ",
  followers: 0,
};

function SellerRating({ rating }: { rating: number }) {
  return (
    <div className="flex items-center gap-1">
      {[1, 2, 3, 4, 5].map((i) => (
        <span key={i} className={`text-sm ${i <= Math.round(rating) ? "text-yellow-400" : "text-gray-300"}`}>★</span>
      ))}
      <span className="text-sm font-medium text-tiki-text ml-1">{rating}</span>
    </div>
  );
}

export default function SellerPage({ params }: { params: Promise<{ id: string }> }) {
  const [tab, setTab] = useState<"products" | "info">("products");
  const [products, setProducts] = useState<Product[]>([]);
  const sellerId = "";
  const seller = { ...SELLER_INFO, id: sellerId };

  return (
    <>
      <Header />
      <main className="bg-[#F5F5FA] py-4 min-h-[60vh]">
        <div className="max-w-5xl mx-auto px-3">
          {/* Seller info card */}
          <div className="bg-white rounded-lg border border-tiki-border p-4 mb-4">
            <div className="flex items-start gap-4">
              <div className="w-16 h-16 bg-blue-100 rounded-lg flex items-center justify-center text-2xl font-bold text-tiki-blue shrink-0">
                {seller.name?.charAt(0) || "S"}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-3 flex-wrap">
                  <h1 className="text-lg font-bold text-tiki-text">{seller.name || `Shop #${sellerId.slice(-6)}`}</h1>
                  <span className="text-[10px] bg-blue-100 text-tiki-blue px-2 py-0.5 rounded font-medium">OFFICIAL</span>
                </div>
                <div className="flex items-center gap-4 mt-2 text-xs text-tiki-text-secondary flex-wrap">
                  <SellerRating rating={seller.rating} />
                  <span>{seller.totalProducts} sản phẩm</span>
                  <span>{seller.totalSales > 0 ? `${seller.totalSales.toLocaleString("vi-VN")} đã bán` : "Mới tham gia"}</span>
                  <span>Phản hồi: {seller.responseTime}</span>
                </div>
              </div>
              <button className="px-4 py-1.5 border border-tiki-blue text-tiki-blue rounded-lg text-xs font-medium hover:bg-blue-50 transition shrink-0">
                + Theo dõi
              </button>
            </div>
          </div>

          {/* Tabs */}
          <div className="bg-white rounded-lg border border-tiki-border mb-4">
            <div className="flex">
              <button
                onClick={() => setTab("products")}
                className={`flex-1 py-3 text-sm font-medium border-b-2 transition ${
                  tab === "products" ? "border-tiki-blue text-tiki-blue" : "border-transparent text-tiki-text-secondary hover:text-tiki-text"
                }`}
              >
                Sản phẩm ({products.length})
              </button>
              <button
                onClick={() => setTab("info")}
                className={`flex-1 py-3 text-sm font-medium border-b-2 transition ${
                  tab === "info" ? "border-tiki-blue text-tiki-blue" : "border-transparent text-tiki-text-secondary hover:text-tiki-text"
                }`}
              >
                Giới thiệu shop
              </button>
            </div>
          </div>

          {tab === "products" ? (
            products.length === 0 ? (
              <div className="bg-white rounded-lg border border-tiki-border py-16 text-center">
                <p className="text-4xl mb-3">🏪</p>
                <p className="text-sm text-tiki-text-secondary">Shop chưa có sản phẩm nào</p>
              </div>
            ) : (
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3">
                {products.map((product) => (
                  <ProductCard key={product.id} product={product} />
                ))}
              </div>
            )
          ) : (
            <div className="bg-white rounded-lg border border-tiki-border p-6">
              <h3 className="text-sm font-semibold text-tiki-text mb-4">Thông tin shop</h3>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <p className="text-tiki-text-secondary text-xs">Tên shop</p>
                  <p className="text-tiki-text font-medium">{seller.name || `Shop #${sellerId.slice(-6)}`}</p>
                </div>
                <div>
                  <p className="text-tiki-text-secondary text-xs">Đánh giá</p>
                  <SellerRating rating={seller.rating} />
                </div>
                <div>
                  <p className="text-tiki-text-secondary text-xs">Ngày tham gia</p>
                  <p className="text-tiki-text font-medium">{seller.joinedDate || "—"}</p>
                </div>
                <div>
                  <p className="text-tiki-text-secondary text-xs">Người theo dõi</p>
                  <p className="text-tiki-text font-medium">{seller.followers?.toLocaleString("vi-VN") || "—"}</p>
                </div>
              </div>
            </div>
          )}
        </div>
      </main>
      <Footer />
    </>
  );
}
