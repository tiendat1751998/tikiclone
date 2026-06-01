import { Suspense } from "react";
import Link from "next/link";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";
import { ProductGrid } from "@/components/storefront/product/ProductCard";
import { PriceFilter } from "@/components/storefront/PriceFilter";
import { ProductsListingClient } from "./ProductsListingClient";
import { Product } from "@/types";
import { extractProducts, mapProductArray } from "@/lib/api/mapper";

const SORT_OPTIONS = [
  { label: "Phổ biến", value: "popular", sort_by: "popularity", sort_order: "DESC" },
  { label: "Mới nhất", value: "newest", sort_by: "created_at", sort_order: "DESC" },
  { label: "Bán chạy", value: "best_selling", sort_by: "sales_count", sort_order: "DESC" },
  { label: "Giá thấp → cao", value: "price_asc", sort_by: "price", sort_order: "ASC" },
  { label: "Giá cao → thấp", value: "price_desc", sort_by: "price", sort_order: "DESC" },
];

type CategoryItem = { id: string; name: string; slug: string };

export default async function ProductsPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string>>;
}) {
  const sp = await searchParams;
  const API_BASE = process.env.GATEWAY_URL || "http://gateway:8080";
  const apiBase = `${API_BASE}/api/v1`;

  // Build query params
  const params = new URLSearchParams();
  for (const [k, v] of Object.entries(sp)) {
    params.set(k, v);
  }
  params.set("page", "1");
  params.set("size", "20");

  // Resolve price filter
  const price = sp.price;
  if (price) {
    const ranges = price.split(",");
    const minMax = ranges.reduce(
      (acc, r) => {
        const [min, max] = r.split("-");
        if (min && (!acc.min || Number(min) < acc.min)) acc.min = Number(min);
        if (max && (!acc.max || Number(max) > acc.max)) acc.max = Number(max);
        return acc;
      },
      { min: undefined as number | undefined, max: undefined as number | undefined }
    );
    if (minMax.min !== undefined) { params.set("min_price", String(minMax.min)); params.delete("price"); }
    if (minMax.max !== undefined) { params.set("max_price", String(minMax.max)); params.delete("price"); }
  }

  // Fetch page 1 + categories in parallel
  let initialProducts: Product[] = [];
  let total = 0;
  let totalPages = 0;
  let categories: CategoryItem[] = [];

  try {
    const [prodRes, catRes] = await Promise.all([
      fetch(`${apiBase}/products?${params}`, { next: { revalidate: 30 } }).then((r) => r.json()),
      fetch(`${apiBase}/categories`, { next: { revalidate: 300 } }).then((r) => r.json()),
    ]);

    if (Array.isArray(prodRes)) {
      initialProducts = mapProductArray(prodRes);
      total = prodRes.length;
      totalPages = 1;
    } else {
      const extracted = extractProducts(prodRes);
      initialProducts = extracted.products;
      total = extracted.total;
      totalPages = extracted.total_pages;
    }

    if (Array.isArray(catRes)) {
      categories = catRes.flatMap((c: Record<string, unknown>) =>
        Array.isArray(c.children) ? (c.children as CategoryItem[]) : []
      );
    }
  } catch {}

  const currentSortBy = sp.sort_by || "";
  const currentSortOrder = sp.sort_order || "DESC";

  return (
    <>
      <Header />
      <main style={{ backgroundColor: "#F5F5FA" }}>
        <div className="max-w-tiki mx-auto">
          {/* Breadcrumb */}
          <div className="flex items-center h-[36px] text-xs">
            <Link href="/" className="text-tiki-text-secondary hover:text-tiki-blue hover:underline">Trang chủ</Link>
            <svg className="mx-[5px]" width="5" height="8" viewBox="0 0 5 8" fill="none"><path d="M1 1L4 4L1 7" stroke="#808089" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/></svg>
            <span className="text-tiki-text">Tất cả sản phẩm</span>
          </div>

          <div className="flex gap-3">
            {/* Sidebar */}
            <aside className="w-[200px] shrink-0 hidden md:block">
              <div className="bg-white rounded-lg border border-tiki-border overflow-hidden sticky top-[60px]">
                <div className="px-3 py-2 border-b border-tiki-border">
                  <span className="text-[11px] font-semibold text-tiki-text">BỘ LỌC</span>
                </div>
                <div className="px-3 py-2 border-b border-tiki-border">
                  <div className="text-[10px] font-semibold text-tiki-text mb-1.5">DANH MỤC</div>
                  <div className="flex flex-col gap-1">
                    {categories.map((cat) => (
                      <Link key={cat.id} href={`/categories/${cat.slug}`} className="text-[11px] text-tiki-text-secondary hover:text-tiki-blue transition">{cat.name}</Link>
                    ))}
                  </div>
                </div>
                <PriceFilter basePath="/products" />
              </div>
            </aside>

            {/* Main */}
            <div className="flex-1 min-w-0">
              {/* Sort bar + count */}
              <div className="bg-white rounded-lg border border-tiki-border mb-2">
                <div className="px-3 py-2 flex items-center justify-between">
                  <span className="text-xs text-tiki-text-secondary">
                    {total > 0 ? `${total} sản phẩm` : "Đang tải..."}
                  </span>
                  <div className="flex items-center gap-1">
                    <span className="text-[10px] text-tiki-text-secondary mr-1">Sắp xếp:</span>
                    {SORT_OPTIONS.map((opt) => (
                      <Link
                        key={opt.value}
                        href={`/products?sort_by=${opt.sort_by}&sort_order=${opt.sort_order}`}
                        className={`px-2 py-1 text-[10px] rounded transition ${
                          currentSortBy === opt.sort_by && currentSortOrder === opt.sort_order
                            ? "bg-tiki-blue text-white"
                            : "text-tiki-text-secondary hover:bg-gray-50"
                        }`}
                      >
                        {opt.label}
                      </Link>
                    ))}
                  </div>
                </div>
              </div>

              {/* SSR product grid */}
              <ProductGrid products={initialProducts} />

              {/* Client-side load-more (renders extra products + button) */}
              <Suspense fallback={<div className="text-center py-6 text-xs text-tiki-text-secondary">Đang tải...</div>}>
                <ProductsListingClient
                  initialTotal={total}
                  initialTotalPages={totalPages}
                />
              </Suspense>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
