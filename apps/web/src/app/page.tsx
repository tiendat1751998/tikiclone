import { Suspense } from "react";
import Link from "next/link";
import { FlashSaleTimer } from "@/components/storefront/FlashSaleTimer";
import { mapProductArray } from "@/lib/api/mapper";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";
import type { Product } from "@/types";
import categoriesData from "@/data/tiki-categories.json";

const GATEWAY_URL = process.env.GATEWAY_URL || "http://gateway:8080";
const API_BASE = `${GATEWAY_URL}/api/v1`;

async function fetchFromAPI<T>(path: string): Promise<T | null> {
  try {
    const res = await fetch(`${API_BASE}${path}`, {
      next: { revalidate: 60 },
      signal: AbortSignal.timeout(5000),
    });
    if (!res.ok) return null;
    const json = await res.json();
    return (json?.data ?? json) as T;
  } catch {
    return null;
  }
}

async function fetchProducts(endpoint: string): Promise<Product[]> {
  const apiData = await fetchFromAPI<Product[]>(endpoint);
  if (apiData && Array.isArray(apiData) && apiData.length > 0) {
    return mapProductArray(apiData);
  }
  return [];
}

async function fetchFeatured(): Promise<Product[]> {
  return fetchProducts("/products/featured?limit=10");
}

async function fetchDeals(): Promise<Product[]> {
  const apiData = await fetchFromAPI<Product[]>("/products/deals?limit=10");
  if (apiData && Array.isArray(apiData) && apiData.length > 0) {
    return mapProductArray(apiData);
  }
  return [];
}

async function fetchCategoryProducts(slug: string): Promise<Product[]> {
  const apiData = await fetchFromAPI<Product[]>(`/products?category_slug=${slug}&limit=10`);
  if (apiData && Array.isArray(apiData) && apiData.length > 0) {
    return mapProductArray(apiData);
  }
  return [];
}

const CATEGORY_ICONS: Record<string, string> = {
  "nha-sach-tiki": "📚",
  "nha-cua-doi-song": "🏠",
  "dien-thoai-may-tinh-bang": "📱",
  "thoi-trang-nu": "👗",
  "thoi-trang-nam": "👔",
  "lam-dep-suc-khoe": "💄",
  "giay-dep-nu": "👟",
  "giay-dep-nam": "👞",
  "bach-hoa-online": "🛒",
  "the-thao-da-ngoai": "⚽",
  "dien-tu-dien-lanh": "❄️",
  "laptop-may-vi-tinh-linh-kien": "💻",
  "do-choi-me-be": "🧸",
  "thiet-bi-so-phu-kien-so": "🔌",
  "dong-ho-va-trang-suc": "⌚",
  "tui-thoi-trang-nu": "👛",
  "tui-thoi-trang-nam": "👝",
  "dien-gia-dung": "🔌",
  "oto-xe-may-xe-dap": "🏍️",
  "cross-border-hang-quoc-te": "🌐",
  "may-anh-may-quay-phim": "📷",
  "phu-kien-thoi-trang": "💍",
  "ngon": "🍜",
  "balo-va-vali": "🎒",
  "voucher-dich-vu": "🎫",
  "cham-soc-nha-cua": "🧹",
  "vat-pham-van-phong": "📎",
  "thuc-pham-do-uong": "🍎",
  "thuoc-dong-y": "🌿",
};

function getCategoryEmoji(slug: string): string {
  for (const [key, icon] of Object.entries(CATEGORY_ICONS)) {
    if (slug.includes(key)) return icon;
  }
  return "📦";
}

const QUICK_SERVICE_LINKS = [
  { icon: "🔥", title: "6.6 Sale", subtitle: "Hè Bùng Nổ" },
  { icon: "👤", title: "Khách Hàng", subtitle: "Thân Thiết" },
  { icon: "🛒", title: "Tiki", subtitle: "Trading" },
  { icon: "🎫", title: "Hot Coupon", subtitle: "Mỗi Ngày" },
  { icon: "🛍️", title: "Đại Siêu Thị", subtitle: "Online" },
  { icon: "🎮", title: "Tiki", subtitle: "Game" },
  { icon: "🌐", title: "Quốc Tế", subtitle: "Thiếu Nhi" },
  { icon: "💻", title: "Thiết Bị", subtitle: "Văn Phòng" },
  { icon: "📚", title: "Sách Hay", subtitle: "Ngày Hè" },
  { icon: "🏷️", title: "Xả Kho", subtitle: "Nửa Giá" },
];

const UTILITY_LINKS = [
  { icon: "💳", label: "Ưu đãi thẻ, ví" },
  { icon: "📱", label: "Đóng tiền, nạp thẻ" },
  { icon: "💸", label: "Mua trước trả sau" },
];

function MobileCategoryChips() {
  const cats = categoriesData.slice(0, 12);
  return (
    <div className="lg:hidden bg-white border-b border-tiki-border px-3 py-2">
      <div className="flex overflow-x-auto scrollbar-none gap-2">
        {cats.map((cat) => (
          <Link
            key={cat.id}
            href={`/categories/${cat.slug}`}
            className="flex-shrink-0 flex items-center gap-1.5 px-3 py-1.5 text-xs bg-tiki-bg rounded-full hover:bg-gray-200 transition"
          >
            <span className="text-sm">{getCategoryEmoji(cat.slug)}</span>
            <span className="whitespace-nowrap text-tiki-text">{cat.name}</span>
          </Link>
        ))}
      </div>
    </div>
  );
}

function DesktopSidebar() {
  const cats = categoriesData.slice(0, 12);
  return (
    <aside className="w-1/5 min-w-0 hidden lg:block">
      <div className="bg-white rounded-lg border border-tiki-border overflow-hidden sticky top-4 flex flex-col" style={{ maxHeight: "calc(100vh - 120px)" }}>
        <div className="px-3 py-2 bg-tiki-bg border-b border-tiki-border">
          <span className="text-xs font-semibold text-tiki-text uppercase">Danh mục</span>
        </div>
        <div className="flex-1 overflow-y-auto">
          {cats.map((cat) => (
            <Link
              key={cat.id}
              href={`/categories/${cat.slug}`}
              className="flex items-center gap-2 px-3 py-2 text-xs text-tiki-text-secondary hover:bg-tiki-bg transition border-b border-tiki-border last:border-b-0"
            >
              <span className="text-sm w-4">{getCategoryEmoji(cat.slug)}</span>
              <span className="truncate">{cat.name}</span>
            </Link>
          ))}
        </div>
        <div className="border-t border-tiki-border p-2">
          <p className="text-[10px] font-semibold text-tiki-text mb-2 px-1">Tiện ích</p>
          {UTILITY_LINKS.map((item) => (
            <button key={item.label} className="flex items-center gap-2 px-2 py-1.5 text-xs text-tiki-text-secondary hover:bg-tiki-bg rounded transition w-full">
              <span className="text-sm">{item.icon}</span>
              <span>{item.label}</span>
            </button>
          ))}
        </div>
        <div className="border-t border-tiki-border p-2">
          <button className="w-full py-2 text-xs font-medium text-tiki-blue border border-tiki-blue rounded-lg hover:bg-tiki-bg transition">
            Bán hàng cùng Tiki
          </button>
        </div>
      </div>
    </aside>
  );
}

function MobileBanner() {
  return (
    <div className="lg:hidden mb-3">
      <div className="relative rounded-xl overflow-hidden bg-white border border-tiki-border" style={{ aspectRatio: "16/9" }}>
        <img src="/images/placeholder.svg" alt="Banner" className="w-full h-full object-cover" />
        <div className="absolute bottom-2 left-1/2 -translate-x-1/2 flex gap-1">
          <div className="w-1.5 h-1.5 rounded-full bg-white"></div>
          <div className="w-1.5 h-1.5 rounded-full bg-white/50"></div>
          <div className="w-1.5 h-1.5 rounded-full bg-white/50"></div>
        </div>
      </div>
    </div>
  );
}

function DesktopBanner() {
  return (
    <div className="hidden lg:grid grid-cols-[60%_40%] gap-2 mb-4">
      <div className="relative rounded-xl overflow-hidden bg-white border border-tiki-border" style={{ aspectRatio: "21/9" }}>
        <img src="/images/placeholder.svg" alt="Children's Day Promo" className="w-full h-full object-cover" />
        <div className="absolute bottom-2 left-1/2 -translate-x-1/2 flex gap-1">
          <div className="w-1.5 h-1.5 rounded-full bg-white"></div>
          <div className="w-1.5 h-1.5 rounded-full bg-white/50"></div>
          <div className="w-1.5 h-1.5 rounded-full bg-white/50"></div>
        </div>
      </div>
      <div className="grid grid-rows-2 gap-2">
        <div className="relative rounded-xl overflow-hidden bg-white border border-tiki-border" style={{ aspectRatio: "16/9" }}>
          <img src="/images/placeholder.svg" alt="Deal Siêu Rẻ" className="w-full h-full object-cover" />
          <div className="absolute bottom-1.5 left-1/2 -translate-x-1/2 flex gap-1">
            <div className="w-1 h-1 rounded-full bg-white"></div>
            <div className="w-1 h-1 rounded-full bg-white/50"></div>
          </div>
        </div>
        <div className="relative rounded-xl overflow-hidden bg-white border border-tiki-border" style={{ aspectRatio: "16/9" }}>
          <img src="/images/placeholder.svg" alt="Deal Siêu Rẻ 2" className="w-full h-full object-cover" />
          <div className="absolute bottom-1.5 left-1/2 -translate-x-1/2 flex gap-1">
            <div className="w-1 h-1 rounded-full bg-white"></div>
            <div className="w-1 h-1 rounded-full bg-white/50"></div>
          </div>
        </div>
      </div>
    </div>
  );
}

function QuickServiceMobile() {
  return (
    <div className="lg:hidden bg-white rounded-lg border border-tiki-border p-2 mb-3">
      <div className="grid grid-flow-col grid-rows-2 overflow-x-auto scrollbar-none gap-3">
        {QUICK_SERVICE_LINKS.map((link) => (
          <Link
            key={link.title}
            href="/products"
            className="flex flex-col items-center justify-center p-2 rounded-lg hover:bg-tiki-bg transition w-16"
          >
            <div className="w-10 h-10 rounded-full bg-tiki-bg flex items-center justify-center mb-1">
              <span className="text-lg">{link.icon}</span>
            </div>
            <span className="text-[10px] font-medium text-tiki-text text-center leading-tight">
              {link.title}
              <br />
              <span className="text-[9px] text-tiki-text-secondary">{link.subtitle}</span>
            </span>
          </Link>
        ))}
      </div>
    </div>
  );
}

function QuickServiceDesktop() {
  return (
    <div className="hidden lg:block bg-white rounded-lg border border-tiki-border p-3 mb-4">
      <div className="flex justify-between">
        {QUICK_SERVICE_LINKS.map((link) => (
          <Link
            key={link.title}
            href="/products"
            className="flex flex-col items-center justify-center p-2 rounded-lg hover:bg-tiki-bg transition"
          >
            <div className="w-10 h-10 rounded-full bg-tiki-bg flex items-center justify-center mb-1">
              <span className="text-lg">{link.icon}</span>
            </div>
            <span className="text-[10px] font-medium text-tiki-text text-center leading-tight">
              {link.title}
              <br />
              <span className="text-[9px] text-tiki-text-secondary">{link.subtitle}</span>
            </span>
          </Link>
        ))}
      </div>
    </div>
  );
}

function StarIcon({ filled = true }: { filled?: boolean }) {
  return (
    <svg width="10" height="10" viewBox="0 0 24 24" fill={filled ? "#FADB14" : "none"} stroke="#FADB14" strokeWidth="1.5">
      <polygon points="12 2 15 8.5 22 9.3 17 14.2 18.5 21 12 17.5 5.5 21 7 14.2 2 9.3 9 8.5" />
    </svg>
  );
}

function ProductCardMobile({ product }: { product: Product }) {
  const discount = product.discount_percent ?? null;
  const hasDiscount = discount && discount > 0;
  const hasFreeShipping = product.price && product.price >= 100000;

  return (
    <Link href={`/products/${product.id}`} className="block">
      <div className="bg-white rounded-lg border border-tiki-border p-1.5 flex flex-col">
        <div className="aspect-square bg-gray-50 rounded relative">
          <img
            src={product.image_url || "/images/placeholder.svg"}
            alt={product.name}
            className="w-full h-full object-contain rounded"
            loading="lazy"
          />
          {hasDiscount && (
            <span className="absolute top-1 left-1 bg-tiki-red text-white text-[9px] font-bold px-1 py-0.5 rounded">
              -{discount}%
            </span>
          )}
        </div>
        <h3 className="text-[11px] leading-tight line-clamp-2 mt-1 text-tiki-text">{product.name}</h3>
        <div className="flex items-center gap-0.5 mt-0.5">
          {Array(5).fill(0).map((_, i) => (
            <StarIcon key={i} filled={i < Math.floor((product.rating_average || 0) / 20)} />
          ))}
          <span className="text-[9px] text-tiki-text-secondary ml-0.5">({product.rating_count || 0})</span>
        </div>
        <div className="flex items-baseline gap-1 mt-0.5">
          <span className="text-[13px] font-semibold text-tiki-red">{product.price?.toLocaleString("vi-VN")} ₫</span>
          {hasDiscount && (
            <span className="text-[10px] text-gray-500 line-through">{product.original_price?.toLocaleString("vi-VN")} ₫</span>
          )}
        </div>
        <p className="text-[9px] text-tiki-text-secondary mt-0.5">Giao ngày mai</p>
      </div>
    </Link>
  );
}

function ProductCardDesktop({ product, showProgress = false }: { product: Product; showProgress?: boolean }) {
  const discount = product.discount_percent ?? null;
  const hasDiscount = discount && discount > 0;

  return (
    <Link href={`/products/${product.id}`} className="product-card group block">
      <div className="product-card__image aspect-square bg-gray-50">
        <img
          src={product.image_url || "/images/placeholder.svg"}
          alt={product.name}
          className="w-full h-full object-contain"
          loading="lazy"
        />
        {hasDiscount && (
          <span className="absolute top-2 right-2 bg-tiki-red text-white text-[10px] font-bold px-1.5 py-0.5 rounded">
            -{discount}%
          </span>
        )}
      </div>
      <div className="product-card__body p-3 flex flex-col gap-1">
        <h3 className="product-card__name text-[11px] leading-tight line-clamp-2">{product.name}</h3>
        <div className="flex items-center gap-0.5">
          {Array(5).fill(0).map((_, i) => (
            <StarIcon key={i} filled={i < Math.floor((product.rating_average || 0) / 20)} />
          ))}
          <span className="text-[9px] text-tiki-text-secondary ml-0.5">({product.rating_count || 0})</span>
        </div>
        <div className="flex items-baseline gap-1">
          <span className="text-sm font-semibold text-tiki-red">{product.price?.toLocaleString("vi-VN")} ₫</span>
          {hasDiscount && (
            <span className="text-[10px] text-gray-500 line-through">{product.original_price?.toLocaleString("vi-VN")} ₫</span>
          )}
        </div>
        <p className="text-[9px] text-tiki-text-secondary mt-auto">Giao ngày mai</p>
        {showProgress && (
          <div className="mt-1">
            <div className="text-[9px] text-orange-500 flex items-center gap-1">
              <span>🔥</span>
              <span>Vừa mở bán</span>
            </div>
            <div className="w-full h-1 bg-gray-200 rounded-full mt-1 overflow-hidden">
              <div className="w-1/3 h-full bg-orange-400 rounded-full"></div>
            </div>
          </div>
        )}
      </div>
    </Link>
  );
}

function TopDealSection({ products }: { products: Product[] }) {
  return (
    <section className="mb-4">
      <div className="bg-white rounded-lg border border-tiki-border overflow-hidden">
        <div className="px-3 py-2 flex items-center justify-between border-b border-tiki-border">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-tiki-red">TOP DEAL • SIÊU RẺ</span>
            <span className="text-lg">👍</span>
          </div>
          <Link
            href="/products"
            className="text-xs font-medium text-tiki-blue hover:opacity-80 flex items-center gap-1"
          >
            Xem tất cả <span>→</span>
          </Link>
        </div>
        <div className="lg:p-2">
          <div className="grid grid-cols-2 md:grid-cols-6 gap-2 p-2 lg:p-0">
            {products.slice(0, 6).map((product, i) => (
              <div key={product.id} className="lg:block hidden">
                <ProductCardDesktop product={product} showProgress={i === 0} />
              </div>
            ))}
            {products.slice(0, 4).map((product) => (
              <div key={product.id + "-mobile"} className="lg:hidden">
                <ProductCardMobile product={product} />
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

function RecommendationsSection({ products }: { products: Product[] }) {
  return (
    <section className="mb-4">
      <div className="bg-white rounded-lg border border-tiki-border overflow-hidden">
        <div className="px-3 py-2 border-b border-tiki-border">
          <h2 className="text-sm font-semibold text-tiki-text">Sản phẩm bạn quan tâm</h2>
        </div>
        <div className="lg:p-2">
          <div className="grid grid-cols-2 md:grid-cols-6 gap-2 p-2 lg:p-0">
            {products.slice(0, 6).map((product) => (
              <div key={product.id}>
                <div className="hidden lg:block">
                  <ProductCardDesktop product={product} />
                </div>
                <div className="lg:hidden">
                  <ProductCardMobile product={product} />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

function CategoryProductSection({ slug, name }: { slug: string; name: string }) {
  return (
    <section className="mb-4">
      <div className="bg-white rounded-lg border border-tiki-border overflow-hidden">
        <div className="px-3 py-2 flex items-center justify-between border-b border-tiki-border">
          <h2 className="text-sm font-semibold text-tiki-text">{name}</h2>
          <Link
            href={`/categories/${slug}`}
            className="text-xs font-medium text-tiki-blue hover:opacity-80 flex items-center gap-1"
          >
            Xem tất cả <span>→</span>
          </Link>
        </div>
        <div className="lg:p-2">
          <Suspense fallback={<div className="grid grid-cols-2 gap-2 p-2"><div className="bg-gray-200 rounded-lg h-48 animate-pulse" /><div className="bg-gray-200 rounded-lg h-48 animate-pulse" /></div>}>
            <ProductSectionContent slug={slug} />
          </Suspense>
        </div>
      </div>
    </section>
  );
}

async function ProductSectionContent({ slug }: { slug: string }) {
  const categoryProducts = await fetchCategoryProducts(slug);
  const products = categoryProducts.length > 0 ? categoryProducts : (await fetchProducts("/products?limit=6"));
  const displayProducts = products.slice(0, 6);
  return (
    <div className="grid grid-cols-2 md:grid-cols-6 gap-2 p-2 lg:p-0">
      {displayProducts.map((product) => (
        <div key={product.id}>
          <div className="hidden lg:block">
            <ProductCardDesktop product={product} />
          </div>
          <div className="lg:hidden">
            <ProductCardMobile product={product} />
          </div>
        </div>
      ))}
    </div>
  );
}

function MobileBottomNav() {
  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 bg-white border-t border-tiki-border flex justify-around py-1.5 lg:hidden pb-safe">
      <Link href="/" className="flex flex-col items-center gap-0.5 px-2">
        <span className="text-lg">🏠</span>
        <span className="text-[10px] text-tiki-text-secondary">Trang chủ</span>
      </Link>
      <Link href="/categories" className="flex flex-col items-center gap-0.5 px-2">
        <span className="text-lg">📋</span>
        <span className="text-[10px] text-tiki-text-secondary">Danh mục</span>
      </Link>
      <Link href="/products" className="flex flex-col items-center gap-0.5 px-2">
        <span className="text-lg">🎯</span>
        <span className="text-[10px] text-tiki-text-secondary">Lướt Tiki</span>
      </Link>
      <Link href="/account/orders" className="flex flex-col items-center gap-0.5 px-2">
        <span className="text-lg">🔔</span>
        <span className="text-[10px] text-tiki-text-secondary">Thông báo</span>
      </Link>
      <Link href="/account" className="flex flex-col items-center gap-0.5 px-2">
        <span className="text-lg">👤</span>
        <span className="text-[10px] text-tiki-text-secondary">Cá nhân</span>
      </Link>
    </nav>
  );
}

function FloatingWidgets() {
  return (
    <div className="fixed bottom-6 right-6 z-50 flex flex-col gap-2">
      <button className="w-12 h-12 rounded-full bg-tiki-blue text-white flex items-center justify-center shadow-lg hover:opacity-90 transition">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
          <path d="M12 2C6.48 2 2 6.48 2 12c0 1.66.48 3.22 1.29 4.52L2 22l4.48-1.29A9.96 9.96 0 0 0 12 22c5.52 0 10-4.48 10-10S17.52 2 12 2z" stroke="currentColor" strokeWidth="1.5"/>
        </svg>
        <span className="sr-only">Trợ lý</span>
      </button>
      <button className="w-12 h-12 rounded-full bg-tiki-blue text-white flex items-center justify-center shadow-lg hover:opacity-90 transition relative">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
          <path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2z" stroke="currentColor" strokeWidth="1.5"/>
        </svg>
        <span className="absolute -top-1 -right-1 w-4 h-4 bg-tiki-red rounded-full flex items-center justify-center text-[8px] font-bold">3</span>
        <span className="sr-only">Tin mới</span>
      </button>
    </div>
  );
}

export const revalidate = 60;

export default async function HomePage() {
  const [featured, deals] = await Promise.all([fetchFeatured(), fetchDeals()]);
  const recommendations = [...(deals.length > 0 ? deals : featured)].slice(6, 12);

  return (
    <>
      <Header />
      <main className="py-2 sm:py-3" style={{ backgroundColor: "#F5F5FA" }}>
        <div className="max-w-[1272px] mx-auto px-3">
          <div className="flex gap-3">
            <DesktopSidebar />
            <div className="flex-1 min-w-0">
              <MobileBanner />
              <MobileCategoryChips />
              <QuickServiceMobile />
              <DesktopBanner />
              <QuickServiceDesktop />
              <TopDealSection products={deals.length > 0 ? deals : featured} />
              <RecommendationsSection products={recommendations.length > 0 ? recommendations : deals.slice(0, 6)} />
              <CategoryProductSection slug="dien-thoai-may-tinh-bang" name="📱 Điện Thoại - Máy Tính Bảng" />
              <CategoryProductSection slug="laptop-may-vi-tinh-linh-kien" name="💻 Laptop - Máy Vi Tính" />
              <CategoryProductSection slug="thoi-trang-nu" name="👗 Thời Trang Nữ" />
              <CategoryProductSection slug="thoi-trang-nam" name="👔 Thời Trang Nam" />
              <CategoryProductSection slug="giay-dep-nu" name="👟 Giày - Dép Nữ" />
              <CategoryProductSection slug="giay-dep-nam" name="👞 Giày - Dép Nam" />
              <CategoryProductSection slug="bach-hoa-online" name="🛒 Bách Hóa Online" />
              <CategoryProductSection slug="the-thao-da-ngoai" name="⚽ Thể Thao - Dã Ngoại" />
              <FlashSaleTimer />
            </div>
          </div>
        </div>
        <MobileBottomNav />
        <FloatingWidgets />
      </main>
      <Footer />
    </>
  );
}