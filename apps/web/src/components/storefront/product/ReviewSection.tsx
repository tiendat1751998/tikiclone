"use client";

import { useState } from "react";
import { StarRating } from "@/components/ui";

interface Review {
  id: string;
  user_name: string;
  rating: number;
  content: string;
  images?: string[];
  created_at: string;
  helpful: number;
}

interface ReviewSectionProps {
  productId: string;
  reviewCount?: number;
  ratingAverage?: number | null;
}

export default function ReviewSection({ productId, reviewCount = 0, ratingAverage = null }: ReviewSectionProps) {
  const [filterRating, setFilterRating] = useState(0);

  // In production, fetch from API. Using mock for now.
  const reviews: Review[] = [];

  if (reviewCount === 0) {
    return (
      <div className="text-center py-8">
        <p className="text-3xl mb-2">📝</p>
        <p className="text-sm text-tiki-text-secondary">Chưa có đánh giá nào cho sản phẩm này</p>
        <p className="text-xs text-tiki-text-secondary mt-1">Hãy là người đầu tiên đánh giá!</p>
      </div>
    );
  }

  return (
    <div>
      {/* Rating summary */}
      {ratingAverage && ratingAverage > 0 && (
        <div className="flex items-center gap-4 mb-4 pb-4 border-b border-tiki-border">
          <div className="text-center">
            <div className="text-3xl font-bold text-tiki-text">{ratingAverage.toFixed(1)}</div>
            <StarRating rating={ratingAverage} size="sm" />
            <div className="text-xs text-tiki-text-secondary mt-1">{reviewCount} đánh giá</div>
          </div>
          <div className="flex-1 space-y-1">
            {[5, 4, 3, 2, 1].map((star) => {
              const count = star === 5 ? Math.round(reviewCount * 0.6) : star === 4 ? Math.round(reviewCount * 0.25) : star === 3 ? Math.round(reviewCount * 0.1) : star === 2 ? Math.round(reviewCount * 0.03) : Math.round(reviewCount * 0.02);
              const pct = reviewCount > 0 ? (count / reviewCount) * 100 : 0;
              return (
                <button
                  key={star}
                  onClick={() => setFilterRating(filterRating === star ? 0 : star)}
                  className={`flex items-center gap-2 w-full text-xs hover:bg-gray-50 rounded px-1 py-0.5 transition ${filterRating === star ? "bg-blue-50" : ""}`}
                >
                  <span className="w-8 text-tiki-text-secondary">{star} ★</span>
                  <div className="flex-1 h-2 bg-gray-200 rounded-full overflow-hidden">
                    <div className="h-full bg-yellow-400 rounded-full" style={{ width: `${pct}%` }} />
                  </div>
                  <span className="w-8 text-right text-tiki-text-secondary">{count}</span>
                </button>
              );
            })}
          </div>
        </div>
      )}

      {/* Review list */}
      {reviews.length === 0 ? (
        <p className="text-sm text-tiki-text-secondary text-center py-8">Chưa có đánh giá chi tiết</p>
      ) : (
        <div className="space-y-4">
          {reviews
            .filter((r) => filterRating === 0 || r.rating === filterRating)
            .map((review) => (
              <div key={review.id} className="border-b border-tiki-border pb-4 last:border-0">
                <div className="flex items-center gap-2 mb-1">
                  <div className="w-7 h-7 bg-blue-100 rounded-full flex items-center justify-center text-[10px] font-bold text-tiki-blue">
                    {review.user_name.charAt(0)}
                  </div>
                  <span className="text-sm font-medium text-tiki-text">{review.user_name}</span>
                </div>
                <div className="ml-9">
                  <StarRating rating={review.rating} size="sm" />
                  <p className="text-sm text-tiki-text mt-1">{review.content}</p>
                  <div className="flex items-center gap-4 mt-2 text-xs text-tiki-text-secondary">
                    <span>{new Date(review.created_at).toLocaleDateString("vi-VN")}</span>
                    <button className="hover:text-tiki-blue">👍 Hữu ích ({review.helpful})</button>
                  </div>
                </div>
              </div>
            ))}
        </div>
      )}
    </div>
  );
}
