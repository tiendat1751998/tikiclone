import Link from "next/link";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";
import { ProductGrid } from "@/components/storefront/product/ProductCard";
import { mapProductArray } from "@/lib/api/mapper";
import { Product } from "@/types";
import promotionsData from "@/data/tiki-promotions.json";

const GATEWAY_URL = process.env.GATEWAY_URL || "http://gateway:8080";
const API_BASE = `${GATEWAY_URL}/api/v1`;

async function getDealProducts(): Promise<Product[]> {
  try {
    const res = await fetch(`${API_BASE}/products/deals?limit=20`, { cache: "no-store" });
    const data = await res.json();
    return mapProductArray(Array.isArray(data) ? data : data.data || []);
  } catch {
    return [];
  }
}

export const dynamic = "force-dynamic";

export default async function PromotionsPage() {
  const promo = promotionsData[0] as any;
  const groups = promo?.groups || [];
  const quickLinks = promo?.quick_links || [];
  const products = await getDealProducts();

  return (
    <>
      <Header />
      <main style={{ backgroundColor: "#F5F5FA" }} className="py-4">
        {/* Breadcrumb */}
        <div className="max-w-tiki mx-auto px-6 mb-4">
          <div className="flex items-center h-10">
            <Link href="/" className="text-sm text-tiki-text-secondary hover:text-[#38383D] hover:underline whitespace-nowrap">Trang chủ</Link>
            <svg className="mx-[5.5px] mr-[8.5px]" width="5" height="8" viewBox="0 0 5 8" fill="none">
              <path d="M1 1L4 4L1 7" stroke="#808089" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            <span className="text-sm text-[#38383D]">Khuyến mãi</span>
          </div>
        </div>

        {/* Hero banner */}
        <div className="max-w-tiki mx-auto px-6 mb-5">
          {promo?.image_url && (
            <div className="rounded-xl overflow-hidden">
              <img src={promo.image_url} alt={promo.name} className="w-full h-auto object-cover" />
            </div>
          )}
        </div>

        {/* Quick links */}
        {quickLinks.length > 0 && (
          <div className="max-w-tiki mx-auto px-6 mb-5">
            <div className="grid grid-cols-5 md:grid-cols-10 gap-2">
              {quickLinks.map((link: any, i: number) => (
                <Link
                  key={i}
                  href={`/search?q=${encodeURIComponent(link.name)}`}
                  className="bg-white rounded-lg p-2 flex flex-col items-center gap-1 border border-tiki-border hover:shadow-md transition"
                >
                  {link.image_url && (
                    <img src={link.image_url} alt={link.name} className="w-10 h-10 object-contain" />
                  )}
                  <span className="text-[10px] text-tiki-text text-center leading-tight">{link.name}</span>
                </Link>
              ))}
            </div>
          </div>
        )}

        {/* Promotion groups */}
        {groups.map((group: any) => (
          <div key={group.id} className="max-w-tiki mx-auto px-6 mb-5">
            <div className="bg-white rounded-lg border border-tiki-border overflow-hidden">
              <div className="px-4 py-3 flex items-center justify-between border-b border-tiki-border">
                <h2 className="text-base font-semibold text-tiki-text">{group.name}</h2>
              </div>
              <div className="p-3">
                <div className="grid grid-cols-3 md:grid-cols-6 gap-2">
                  {group.items?.map((item: any, i: number) => (
                    <Link
                      key={i}
                      href={item.url || `/search?q=${encodeURIComponent(item.name)}`}
                      className="bg-[#F5F5FA] rounded-lg p-3 text-center hover:bg-[#EBEBF0] transition"
                    >
                      {item.image_url ? (
                        <img src={item.image_url} alt={item.name} className="w-full h-16 object-contain mb-1" />
                      ) : (
                        <div className="w-full h-16 flex items-center justify-center text-2xl mb-1">🏷️</div>
                      )}
                      <span className="text-xs text-tiki-text">{item.name}</span>
                    </Link>
                  ))}
                </div>
              </div>
            </div>
          </div>
        ))}

        {/* Deal products */}
        <div className="max-w-tiki mx-auto px-6 mb-6">
          <div className="bg-white rounded-lg border border-tiki-border overflow-hidden">
            <div className="px-4 py-3 flex items-center justify-between border-b border-tiki-border">
              <h2 className="text-base font-semibold text-tiki-text">🔥 Giá Sốc Khuyến Mãi</h2>
              <Link href="/products" className="text-sm font-medium text-tiki-blue hover:opacity-80 flex items-center gap-1">
                Xem tất cả <span>→</span>
              </Link>
            </div>
            <div className="p-3">
              {products.length > 0 ? (
                <ProductGrid products={products} />
              ) : (
                <div className="text-center py-12 text-tiki-text-secondary">
                  <p className="text-4xl mb-3">🏷️</p>
                  <p className="font-medium">Chưa có sản phẩm khuyến mãi</p>
                  <p className="text-sm mt-1">Hãy quay lại sau nhé!</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
