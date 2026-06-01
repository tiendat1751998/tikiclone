"use client";

import { useState, useEffect, useCallback, Suspense, use } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";
import { ProductGrid } from "@/components/storefront/product/ProductCard";
import { StarRating } from "@/components/ui";
import { useInfiniteProducts, ProductListPage } from "@/hooks/useApi";
import type { Product } from "@/types";

const SORT_OPTIONS = [
  { label: "Phù hợp", value: "" },
  { label: "Mới nhất", value: "created_at" },
  { label: "Bán chạy", value: "sales" },
  { label: "Giá thấp → cao", value: "price_asc" },
  { label: "Giá cao → thấp", value: "price_desc" },
];

const PRICE_RANGES = [
  { label: "Dưới 50.000", min: "", max: "50000" },
  { label: "50.000 - 200.000", min: "50000", max: "200000" },
  { label: "200.000 - 500.000", min: "200000", max: "500000" },
  { label: "500.000 - 1.000.000", min: "500000", max: "1000000" },
  { label: "1.000.000 - 3.000.000", min: "1000000", max: "3000000" },
  { label: "Trên 3.000.000", min: "3000000", max: "" },
];

const BRANDS = ["Samsung", "Apple", "Xiaomi", "OPPO", "Vivo", "Realme", "Nokia", "Asus"];

function buildSearchUrl(sp: URLSearchParams, overrides: Record<string, string | undefined>) {
  const p = new URLSearchParams(sp);
  for (const [k, v] of Object.entries(overrides)) {
    if (v === undefined || v === "") p.delete(k);
    else p.set(k, v);
  }
  return `/search?${p.toString()}`;
}

function LoadMoreButton({ onClick, loading, hasMore }: { onClick: () => void; loading: boolean; hasMore: boolean }) {
  if (!hasMore) {
    return (
      <div className="text-center py-6 text-xs text-tiki-text-secondary">
        Đã hiển thị tất cả kết quả
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

function SearchContent({ searchParams }: { searchParams: Promise<Record<string, string | undefined>> }) {
  const params = use(searchParams) as Record<string, string | undefined>;
  const sp = new URLSearchParams(params as any);
  const q = params.q || "";

  const currentSort = SORT_OPTIONS.find((o) => o.value === (params.sort_by || "")) || SORT_OPTIONS[0];

  const extraParams: Record<string, string> = {};
  if (params.sort_by) extraParams.sort_by = params.sort_by;
  if (params.sort_order) extraParams.sort_order = params.sort_order;
  if (params.min_price) extraParams.min_price = params.min_price;
  if (params.max_price) extraParams.max_price = params.max_price;
  if (params.brand) extraParams.brand = params.brand;
  if (params.rating) extraParams.rating = params.rating;

  const searchPath = q ? "/products/search" : "/products";
  if (q) extraParams.q = q;

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
  } = useInfiniteProducts(searchPath, extraParams);

  const pages: ProductListPage[] = data?.pages || [];
  const allProducts = pages.flatMap((p) => p.products);
  const total = pages[0]?.total ?? 0;
  const totalPages = pages[pages.length - 1]?.total_pages ?? 0;
  const currentPage = pages[pages.length - 1]?.page ?? 1;

  const handleLoadMore = useCallback(() => {
    if (hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  if (!q) {
    return (
      <>
        <Header />
        <main className="bg-tiki-bg py-16 text-center">
          <div className="max-w-tiki mx-auto px-3">
            <div className="bg-white rounded-lg border border-tiki-border py-14">
              <p className="text-4xl mb-3">🔍</p>
              <p className="text-xs text-tiki-text-secondary">Nhập từ khóa để tìm kiếm sản phẩm</p>
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
      <main className="bg-tiki-bg">
        <div className="max-w-tiki mx-auto px-3">
          <div className="flex items-center h-8 text-[11px] text-tiki-text-secondary">
            <Link href="/" className="hover:text-tiki-blue">Trang chủ</Link>
            <span className="mx-1.5">›</span>
            <span className="text-tiki-text">Tìm kiếm: {q}</span>
          </div>

          <div className="flex gap-4">
            <aside className="w-[200px] shrink-0 hidden md:block">
              <div className="filter-sidebar">
                <div className="filter-sidebar__group">
                  <div className="filter-sidebar__title">Khoảng giá</div>
                  <div className="space-y-1">
                    {PRICE_RANGES.map((range) => {
                      const isActive = params.min_price === range.min && params.max_price === range.max;
                      return (
                        <Link
                          key={range.label}
                          href={buildSearchUrl(sp, { min_price: range.min || undefined, max_price: range.max || undefined })}
                          className={`filter-sidebar__option ${isActive ? "text-tiki-blue font-medium" : ""}`}
                        >
                          <span>{range.label}</span>
                        </Link>
                      );
                    })}
                  </div>
                </div>
                <div className="filter-sidebar__group">
                  <div className="filter-sidebar__title">Đánh giá</div>
                  <div className="space-y-1">
                    {[5, 4, 3, 2].map((star) => {
                      const isActive = params.rating === String(star);
                      return (
                        <Link
                          key={star}
                          href={buildSearchUrl(sp, { rating: isActive ? undefined : String(star) })}
                          className={`filter-sidebar__option ${isActive ? "text-tiki-blue font-medium" : ""}`}
                        >
                          <StarRating rating={star} size="sm" />
                          <span className="text-[10px] text-tiki-text-secondary ml-1">trở lên</span>
                        </Link>
                      );
                    })}
                  </div>
                </div>
                <div className="filter-sidebar__group">
                  <div className="filter-sidebar__title">Thương hiệu</div>
                  <div className="space-y-1">
                    {BRANDS.map((brand) => {
                      const isActive = params.brand === brand;
                      return (
                        <Link
                          key={brand}
                          href={buildSearchUrl(sp, { brand: isActive ? undefined : brand })}
                          className={`filter-sidebar__option ${isActive ? "text-tiki-blue font-medium" : ""}`}
                        >
                          <span>{brand}</span>
                        </Link>
                      );
                    })}
                  </div>
                </div>
              </div>
            </aside>

            <div className="flex-1 min-w-0">
              <div className="bg-white rounded-lg border border-tiki-border mb-2 px-3 py-2 flex items-center justify-between">
                <span className="text-[11px] text-tiki-text-secondary">
                  {pages.length === 0 && isLoading ? "Đang tìm..." : `${total} kết quả`} cho "<strong className="text-tiki-text">{q}</strong>"
                </span>
                <div className="flex items-center gap-1">
                  <span className="text-[10px] text-tiki-text-secondary mr-1">Sắp xếp:</span>
                  {SORT_OPTIONS.map((opt) => (
                    <Link
                      key={opt.value}
                      href={buildSearchUrl(sp, {
                        sort_by: opt.value === "price_asc" || opt.value === "price_desc" ? "price" : opt.value || undefined,
                        sort_order: opt.value === "price_asc" ? "ASC" : opt.value === "price_desc" ? "DESC" : undefined,
                      })}
                      className={`px-2 py-1 text-[10px] rounded transition ${
                        currentSort.value === opt.value
                          ? "bg-tiki-blue text-white"
                          : "text-tiki-text-secondary hover:bg-gray-50"
                      }`}
                    >
                      {opt.label}
                    </Link>
                  ))}
                </div>
              </div>

              <ProductGrid products={allProducts as Product[]} isLoading={pages.length === 0 && isLoading} />

              {allProducts.length > 0 && (
                <LoadMoreButton
                  onClick={handleLoadMore}
                  loading={isFetchingNextPage}
                  hasMore={currentPage < totalPages}
                />
              )}
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}

export default function SearchPage({ searchParams }: { searchParams: Promise<{ q?: string }> }) {
  return (
    <Suspense fallback={<>
      <Header />
      <main className="bg-tiki-bg py-8 text-center"><p className="text-xs text-tiki-text-secondary">Đang tải...</p></main>
      <Footer />
    </>}>
      <SearchContent searchParams={searchParams} />
    </Suspense>
  );
}
