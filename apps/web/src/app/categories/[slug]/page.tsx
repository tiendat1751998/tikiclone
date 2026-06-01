import { Suspense } from "react";
import Link from "next/link";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";
import { ProductGrid } from "@/components/storefront/product/ProductCard";
import { CategoryListingClient } from "./CategoryListingClient";
import { Product, Category } from "@/types";
import { mapProductArray } from "@/lib/api/mapper";
import categoriesData from "@/data/tiki-categories.json";

const SORT_OPTIONS = [
  { label: "Phổ biến", value: "popular", by: "popularity", order: "DESC" },
  { label: "Bán chạy", value: "best_selling", by: "sales_count", order: "DESC" },
  { label: "Giá thấp → cao", value: "price_asc", by: "price", order: "ASC" },
  { label: "Giá cao → thấp", value: "price_desc", by: "price", order: "DESC" },
];

const CATEGORY_ICONS: Record<string, string> = {
  "dien-thoai-may-tinh-bang": "📱",
  "laptop-may-vi-tinh-linh-kien": "💻",
  "thoi-trang-nu": "👗",
  "thoi-trang-nam": "👔",
  "lam-dep-suc-khoe": "💄",
  "giay-dep-nu": "👟",
  "giay-dep-nam": "👞",
  "bach-hoa-online": "🛒",
  "the-thao-da-ngoai": "⚽",
};

function getCategoryEmoji(slug: string): string {
  return CATEGORY_ICONS[slug] || "📦";
}

const SUBCATEGORY_ICONS: Record<string, string> = {
  "smartphone": "📱",
  "tablet": "📱",
  "ereader": "📖",
  "phu-kien": "🎧",
};

const FILTER_BRANDS = ["Samsung", "Apple", "Xiaomi", "OPPO", "Vivo", "Realme", "Nokia"];

const PROMO_BADGES = [
  { label: "Giao siêu tốc 2H", color: "bg-orange-500" },
  { label: "TOP DEAL Siêu rẻ", color: "bg-red-500" },
  { label: "Freeship XTRA", color: "bg-green-500" },
  { label: "Từ 4 sao", color: "bg-yellow-500" },
];

function findCategory(cats: Category[], slug: string): Category | null {
  for (const cat of cats) {
    if (cat.slug === slug) return cat;
    if (cat.children) {
      const found = findCategory(cat.children, slug);
      if (found) return found;
    }
  }
  return null;
}

async function fetchCategoryProducts(slug: string, page = 1): Promise<Product[]> {
   try {
     const res = await fetch(
       `${process.env.GATEWAY_URL || "http://gateway:8080"}/api/v1/products?category_slug=${slug}&page=${page}&size=20`,
       { next: { revalidate: 30 } }
     );
    const data = await res.json();
    const rawProducts = Array.isArray(data) ? data : (data.products || data.data || []);
    return mapProductArray(rawProducts as unknown[]);
  } catch {
    return [];
  }
}

export default async function CategoryPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;

  const category = findCategory(categoriesData, slug);
  const categoryName = category?.name || slug.replace(/-/g, " ");

  const subcategories = [
    { name: "Điện thoại Smartphone", slug: "smartphone" },
    { name: "Máy tính bảng", slug: "tablet" },
    { name: "Máy đọc sách", slug: "ereader" },
    { name: "Phụ kiện điện thoại", slug: "phu-kien" },
  ];

  const featuredProducts = await fetchCategoryProducts(slug, 1);

  return (
    <>
      <Header />
      <main style={{ backgroundColor: "#F5F5FA" }}>
        <div className="max-w-tiki mx-auto px-3">
          <div className="flex items-center h-8 text-xs text-tiki-text-secondary mb-3">
            <Link href="/" className="hover:text-tiki-blue hover:underline">Trang chủ</Link>
            <svg className="mx-1.5" width="5" height="8" viewBox="0 0 5 8" fill="none"><path d="M1 1L4 4L1 7" stroke="#787880" strokeWidth="1.5"/></svg>
            <span className="text-tiki-text">{categoryName}</span>
          </div>

          <div className="flex gap-3">
            <aside className="w-1/5 min-w-0 hidden md:block">
              <div className="bg-white rounded-lg border border-tiki-border p-4 mb-3">
                <h3 className="text-[13px] font-semibold text-tiki-text mb-3">Khám phá theo danh mục</h3>
                <div className="space-y-1">
                  {subcategories.map((sub) => (
                    <Link
                      key={sub.slug}
                      href={`/categories/${slug}?subcategory=${sub.slug}`}
                      className="block text-xs text-tiki-text-secondary hover:text-tiki-blue py-1"
                    >
                      {sub.name}
                    </Link>
                  ))}
                </div>
              </div>

              <div className="bg-white rounded-lg border border-tiki-border p-4">
                <h3 className="text-[13px] font-semibold text-tiki-text mb-3">Tiki Trading</h3>
                <div className="space-y-3">
                  {featuredProducts.slice(0, 2).map((p) => (
                    <div key={p.id} className="border border-tiki-border rounded-lg p-2">
                      <img src={p.image_url || "/images/placeholder.svg"} alt={p.name} className="w-full h-24 object-contain mb-2" />
                      <p className="text-[10px] text-tiki-text line-clamp-2 mb-1">{p.name}</p>
                      <p className="text-xs font-bold text-tiki-red mb-2">{p.price?.toLocaleString("vi-VN")} ₫</p>
                      <Link href={`/products/${p.id}`} className="block text-center text-[10px] text-tiki-blue border border-tiki-blue rounded py-1">Xem thêm</Link>
                    </div>
                  ))}
                </div>
              </div>
            </aside>

            <div className="flex-1 min-w-0">
              <h1 className="text-xl font-bold text-tiki-text mb-3">{categoryName}</h1>

              <div className="grid grid-cols-2 gap-3 mb-4">
                <Link href="/promotions" className="hero-banner rounded-lg relative overflow-hidden">
                  <div className="hero-fallback">Samsung Galaxy S26 Series</div>
                </Link>
                <Link href="/promotions" className="hero-banner rounded-lg relative overflow-hidden">
                  <div className="hero-fallback">Điện thoại Hiệu Dưới 5 Triệu</div>
                </Link>
              </div>

              <div className="bg-white rounded-lg border border-tiki-border p-3 mb-4">
                <h3 className="text-[13px] font-semibold text-tiki-text mb-2">Khám phá theo danh mục</h3>
                <div className="flex gap-3 overflow-x-auto pb-2">
                  {subcategories.map((sub) => (
                    <Link
                      key={sub.slug}
                      href={`/categories/${slug}?subcategory=${sub.slug}`}
                      className="flex flex-col items-center w-20 flex-shrink-0"
                    >
                      <div className="w-12 h-12 rounded-full bg-tiki-bg flex items-center justify-center mb-1">
                        <span className="text-lg">{SUBCATEGORY_ICONS[sub.slug] || "📱"}</span>
                      </div>
                      <span className="text-[10px] text-tiki-text text-center">{sub.name.split(" ")[0]}</span>
                    </Link>
                  ))}
                </div>
              </div>

              <div className="bg-white rounded-lg border border-tiki-border p-3 mb-4">
                <div className="flex flex-wrap items-center gap-2 mb-3">
                  <span className="text-xs font-semibold text-tiki-text mr-2">Tất cả sản phẩm</span>
                  <div className="flex items-center gap-2 overflow-x-auto">
                    {FILTER_BRANDS.map((brand) => (
                      <button
                        key={brand}
                        className="px-3 py-1 text-xs border border-tiki-border rounded-full text-tiki-text-secondary hover:border-tiki-blue hover:text-tiki-blue transition whitespace-nowrap"
                      >
                        {brand}
                      </button>
                    ))}
                  </div>
                </div>

                <div className="flex flex-wrap gap-2 mb-3">
                  {PROMO_BADGES.map((badge) => (
                    <button
                      key={badge.label}
                      className={`px-2 py-0.5 text-xs text-white rounded-full ${badge.color} hover:opacity-80 transition`}
                    >
                      {badge.label}
                    </button>
                  ))}
                </div>

                <div className="flex justify-end">
                  <select className="text-xs border border-tiki-border rounded px-2 py-1 text-tiki-text">
                    <option>Phổ biến</option>
                    <option>Bán chạy</option>
                    <option>Giá thấp → cao</option>
                    <option>Giá cao → thấp</option>
                  </select>
                </div>
              </div>

              <div className="bg-white rounded-lg border border-tiki-border p-3">
                <div className="grid grid-cols-4 gap-3">
                  <Suspense fallback={<div className="col-span-4"><ProductGridSkeleton count={8} /></div>}>
                    <CategoryProductGrid slug={slug} />
                  </Suspense>
                </div>
                <CategoryListingClient slug={slug} initialTotalPages={1} />
              </div>
            </div>
          </div>
        </div>

        <div className="fixed right-4 bottom-20 flex flex-col gap-2 z-40">
          <button className="w-12 h-12 bg-white rounded-full shadow-lg flex items-center justify-center hover:bg-tiki-bg transition">
            <span className="text-xl">🤖</span>
          </button>
          <button className="w-12 h-12 bg-white rounded-full shadow-lg flex items-center justify-center hover:bg-tiki-bg transition">
            <span className="text-xl">🔔</span>
          </button>
        </div>
      </main>
      <Footer />
    </>
  );
}

async function CategoryProductGrid({ slug }: { slug: string }) {
  const products = await fetchCategoryProducts(slug, 1);
  if (products.length === 0) {
    return (
      <div className="col-span-4 text-center py-8 text-tiki-text-secondary">
        Đang tải sản phẩm...
      </div>
    );
  }
  return (
    <>
      {products.map((product) => (
        <Link key={product.id} href={`/products/${product.id}`} className="product-card group block">
          <div className="product-card__image aspect-square bg-gray-50">
            <img
              src={product.image_url || "/images/placeholder.svg"}
              alt={product.name}
              className="w-full h-full object-contain"
              loading="lazy"
            />
            {product.discount_percent && product.discount_percent > 0 && (
              <div className="product-card__discount">-{product.discount_percent}%</div>
            )}
            {product.is_official && (
              <div className="product-card__official">CHÍNH HÃNG</div>
            )}
            {product.price && product.price >= 100000 && (
              <div className="absolute top-0 right-0 bg-tiki-green text-white text-[9px] font-bold px-1 rounded-bl">
                FREESHIP
              </div>
            )}
          </div>
          <div className="product-card__body p-2 flex flex-col gap-1">
            <h3 className="product-card__name text-[11px] text-tiki-text line-clamp-2 leading-tight">
              {product.name}
            </h3>
            <div className="flex items-center gap-0.5">
              {Array.from({ length: 5 }).map((_, i) => (
                <span key={i} className="text-[10px] text-yellow-400">★</span>
              ))}
              {product.quantity_sold_text && (
                <span className="text-[9px] text-tiki-text-secondary ml-1">{product.quantity_sold_text}</span>
              )}
            </div>
            <div className="flex items-center justify-between mt-auto">
              <span className="text-sm font-bold text-tiki-red">{product.price?.toLocaleString("vi-VN")} ₫</span>
              {product.discount_percent && product.discount_percent > 0 && (
                <span className="text-[10px] font-bold text-tiki-red bg-red-50 px-1 rounded">
                  -{product.discount_percent}%
                </span>
              )}
            </div>
            <span className="inline-block text-[9px] text-tiki-text-secondary bg-tiki-bg px-1 rounded">
              Giao siêu tốc 2H
            </span>
          </div>
        </Link>
      ))}
    </>
  );
}

function ProductGridSkeleton({ count }: { count: number }) {
  return (
    <>
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="bg-white rounded-lg border border-tiki-border overflow-hidden">
          <div className="aspect-square bg-gray-200 animate-pulse" />
          <div className="p-2 space-y-1">
            <div className="h-3 bg-gray-200 rounded animate-pulse" />
            <div className="h-3 w-2/3 bg-gray-200 rounded animate-pulse" />
          </div>
        </div>
      ))}
    </>
  );
}