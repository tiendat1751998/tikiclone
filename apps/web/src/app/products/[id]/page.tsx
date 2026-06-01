import Link from "next/link";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";
import ProductDetailClient from "./ProductDetailClient";

const GATEWAY_URL = process.env.GATEWAY_URL || "http://gateway:8080";
const API_BASE = `${GATEWAY_URL}/api/v1`;

interface ProductImage { id: string; url: string; is_primary: boolean }
interface ProductDetail {
  id: string; name: string; description?: string; short_description?: string;
  image_url: string; images?: ProductImage[]; price: number; original_price?: number | null;
  discount_percent?: number | null; stock: number; sold_count: number; quantity_sold_text?: string;
  rating_average?: number | null; rating_count?: number; review_count?: number;
  brand?: string; seller_name?: string; seller_avatar_url?: string; is_official?: boolean;
  attributes?: { name: string; value: string }[]; category_name?: string; category_id: string;
  weight?: number; dimensions?: string; status: string; shop_id?: string; shop_name?: string;
}

async function getProduct(id: string): Promise<ProductDetail | null> {
   try {
     const res = await fetch(`${API_BASE}/products/${id}`, { next: { revalidate: 60 } });
    if (!res.ok) return null;
    const data = await res.json();
    const p = data?.data || data;
    if (!p?.id) return null;
    const sku = p.skus?.[0] || {};
    const mediaImgs = (p.media || []).map((m: { url: string; id?: string; is_primary?: boolean }) => ({
      id: m.id || "main",
      url: m.url || "/images/placeholder.svg",
      is_primary: m.is_primary || false,
    }));
    return {
      ...p,
      price: sku.price || 0,
      original_price: sku.compare_price > 0 ? sku.compare_price : null,
      discount_percent: sku.compare_price > sku.price ? Math.round(((sku.compare_price - sku.price) / sku.compare_price) * 100) : null,
      stock: sku.stock || 0,
      brand: p.attributes?.brand || p.brand || "",
      image_url: mediaImgs[0]?.url || "/images/placeholder.svg",
      images: mediaImgs.length > 0 ? mediaImgs : [{ id: "main", url: p.image_url || "/images/placeholder.svg", is_primary: true }],
    };
  } catch { return null; }
}

export default async function ProductPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const p = await getProduct(id);

  if (!p) {
    return (
      <>
        <Header />
        <main className="py-16 text-center">
          <p className="text-5xl mb-4">😕</p>
          <h1 className="text-lg font-semibold text-tiki-text mb-2">Sản phẩm không tồn tại</h1>
          <Link href="/products" className="text-tiki-blue hover:underline text-sm">← Quay lại</Link>
        </main>
        <Footer />
      </>
    );
  }

  const allImages = p.images?.length ? p.images : [{ id: "main", url: p.image_url || "/images/placeholder.svg", is_primary: true }];

  return (
    <>
      <Header />
      <main className="py-2 sm:py-3" style={{ backgroundColor: "#F5F5FA" }}>
        <ProductDetailClient product={p} allImages={allImages} />
      </main>
      <Footer />
    </>
  );
}
