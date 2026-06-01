"use client";

import { useState } from "react";
import ReviewSection from "./ReviewSection";

export default function ProductDetailWithReviews({ product }: { product: any }) {
  const [activeTab, setActiveTab] = useState<"detail" | "reviews">("detail");

  return (
    <div>
      {/* Tabs */}
      <div className="mt-6 bg-white rounded-lg border border-tiki-border">
        <div className="flex border-b border-tiki-border">
          <button
            onClick={() => setActiveTab("detail")}
            className={`flex-1 py-3 text-sm font-medium border-b-2 transition ${
              activeTab === "detail" ? "border-tiki-blue text-tiki-blue" : "border-transparent text-tiki-text-secondary hover:text-tiki-text"
            }`}
          >
            Chi tiết sản phẩm
          </button>
          <button
            onClick={() => setActiveTab("reviews")}
            className={`flex-1 py-3 text-sm font-medium border-b-2 transition ${
              activeTab === "reviews" ? "border-tiki-blue text-tiki-blue" : "border-transparent text-tiki-text-secondary hover:text-tiki-text"
            }`}
          >
            Đánh giá {product.review_count ? `(${product.review_count})` : ""}
          </button>
        </div>

        <div className="p-4">
          {activeTab === "detail" ? (
            <div className="text-sm text-tiki-text-secondary leading-relaxed whitespace-pre-wrap">
              {product.description || "Chưa có mô tả sản phẩm"}
            </div>
          ) : (
            <ReviewSection productId={product.id} reviewCount={product.review_count} ratingAverage={product.rating_average} />
          )}
        </div>
      </div>
    </div>
  );
}
