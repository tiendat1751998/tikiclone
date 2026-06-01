"use client";

import { useState, useCallback } from "react";
import { ProductCard } from "@/components/storefront/product/ProductCard";
import type { Product } from "@/types";
import { mapProductArray } from "@/lib/api/mapper";

function LoadMoreButton({ onClick, loading, hasMore }: { onClick: () => void; loading: boolean; hasMore: boolean }) {
  if (!hasMore) {
    return (
      <div className="text-center py-6 text-xs text-tiki-text-secondary">
        Đã hiển thị tất cả sản phẩm
      </div>
    );
  }
  return (
    <div className="text-center py-6">
      <button
        onClick={onClick}
        disabled={loading}
        className="px-8 py-2.5 text-xs font-medium rounded-lg border border-tiki-border bg-white text-tiki-text hover:border-tiki-blue hover:text-tiki-blue transition disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {loading ? "Đang tải..." : "Xem thêm sản phẩm"}
      </button>
    </div>
  );
}

export function CategoryListingClient({
  slug,
  initialTotalPages,
}: {
  slug: string;
  initialTotalPages: number;
}) {
  const [extraProducts, setExtraProducts] = useState<Product[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [hasMore, setHasMore] = useState(initialTotalPages > 1);
  const [loading, setLoading] = useState(false);

  const handleLoadMore = useCallback(async () => {
    if (loading || !hasMore) return;
    setLoading(true);
    try {
      const nextPage = currentPage + 1;
      const sp = new URLSearchParams({ category_slug: slug, page: String(nextPage), size: "20" });
      const res = await fetch(`/api/v1/products?${sp}`, { cache: "no-store" });
      const data = await res.json();
      const rawProducts: Product[] = data.products || data.data || [];
      const products = mapProductArray(rawProducts);
      const total: number = data.total || 0;
      const size: number = data.size || data.page_size || 20;
      const totalPages = Math.ceil(total / size) || 1;
      setExtraProducts((prev) => [...prev, ...products]);
      setCurrentPage(nextPage);
      setHasMore(nextPage < totalPages);
    } catch {}
    setLoading(false);
  }, [loading, hasMore, currentPage, slug]);

  return (
    <>
      {extraProducts.length > 0 && (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3 mt-3">
          {extraProducts.map((product, i) => (
            <ProductCard key={product.id + "-extra-" + i} product={product} priority={false} />
          ))}
        </div>
      )}
      <LoadMoreButton onClick={handleLoadMore} loading={loading} hasMore={hasMore} />
    </>
  );
}
