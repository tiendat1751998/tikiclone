import type { Product, ProductImage, Category } from "@/types";

interface BackendProductImage {
  id: string;
  product_id: string;
  type: string;
  url: string;
  thumbnail_url?: string;
  sort_order: number;
  is_primary: boolean;
}

interface BackendSku {
  id: string;
  product_id: string;
  name: string;
  price: number;
  compare_price: number;
  currency: string;
  stock: number;
  reserved_stock: number;
  status: string;
  sort_order: number;
}

interface BackendProduct {
  id: string;
  name: string;
  description?: string;
  category_id: string;
  brand?: string;
  status: string;
  condition: string;
  weight?: number;
  dimensions?: string;
  shop_id: string;
  skus: BackendSku[];
  media: BackendProductImage[];
  attributes?: Record<string, unknown>;
  sold_count?: number;
  created_at: string;
  updated_at: string;
}

const SELLER_NAMES: Record<string, string> = {
  "usr-001": "Tiki Official",
  "usr-002": "TechStore Vietnam",
  "usr-003": "Fashion Hub",
  "usr-004": "Home & Life",
  "usr-005": "Mẹ & Bé Shop",
};

function getSellerName(shopId: string): string {
  return SELLER_NAMES[shopId] || `Shop ${shopId.slice(-4)}`;
}

function buildImageUrl(p: BackendProduct, media: BackendProductImage): string {
  if (media?.url) return media.url;
  return "/images/placeholder.svg";
}

function buildPrimaryImageUrl(p: BackendProduct): string {
  const img = p.media?.find((m) => m.is_primary) || p.media?.[0];
  return buildImageUrl(p, img as BackendProductImage);
}

export function mapBackendProduct(p: BackendProduct): Product {
  const sku = p.skus?.[0];
  const price = sku?.price || 0;
  const originalPrice = sku?.compare_price || null;
  const discount =
    originalPrice && originalPrice > price
      ? Math.round(((originalPrice - price) / originalPrice) * 100)
      : null;

  return {
    id: p.id,
    shop_id: p.shop_id,
    category_id: p.category_id,
    name: p.name,
    description: p.description || "",
    brand: p.brand || "",
    image_url: buildPrimaryImageUrl(p),
    thumbnail_url: buildPrimaryImageUrl(p),
    images: p.media?.map(
      (m): ProductImage => ({
        id: m.id,
        url: buildImageUrl(p, m),
        thumbnail_url: m.thumbnail_url || buildImageUrl(p, m),
        sort_order: m.sort_order,
        is_primary: m.is_primary,
      })
    ) || [],
    price,
    original_price: originalPrice,
    discount_percent: discount,
    stock: sku?.stock || 0,
    sold_count: p.sold_count || 0,
    quantity_sold_text: p.sold_count ? `Đã bán ${p.sold_count}` : undefined,
    rating_average: null,
    rating_count: 0,
    review_count: 0,
    seller_name: getSellerName(p.shop_id),
    seller_avatar_url: undefined,
    is_tiki_trading: false,
    is_official: false,
    is_sponsored: false,
    status: p.status as Product["status"],
    condition: p.condition,
    weight: p.weight,
    dimensions: p.dimensions,
    created_at: p.created_at,
    updated_at: p.updated_at,
  };
}

export function mapProductArray(products: unknown[]): Product[] {
  if (!Array.isArray(products)) return [];
  return products.map((p) => mapBackendProduct(p as BackendProduct));
}

export function extractProducts(
  response: unknown
): { products: Product[]; total: number; page: number; page_size: number; total_pages: number } {
  if (!response) return { products: [], total: 0, page: 1, page_size: 20, total_pages: 0 };

  if (Array.isArray(response)) {
    const products = mapProductArray(response);
    return {
      products,
      total: products.length,
      page: 1,
      page_size: products.length || 20,
      total_pages: 1,
    };
  }

  const obj = response as Record<string, unknown>;
  if (obj.data && Array.isArray(obj.data)) {
    const products = mapProductArray(obj.data);
    return {
      products,
      total: (obj.total as number) || products.length,
      page: (obj.page as number) || 1,
      page_size: (obj.size as number) || products.length || 20,
      total_pages: Math.ceil((obj.total as number) / ((obj.size as number) || 20)) || 1,
    };
  }

  if (obj.products && Array.isArray(obj.products)) {
    const products = mapProductArray(obj.products);
    return {
      products,
      total: (obj.total as number) || products.length,
      page: (obj.page as number) || 1,
      page_size: (obj.size as number) || products.length || 20,
      total_pages: Math.ceil((obj.total as number) / ((obj.size as number) || 20)) || 1,
    };
  }

  return { products: [], total: 0, page: 1, page_size: 20, total_pages: 0 };
}

export function mapSingleProduct(response: unknown): Product | null {
  if (!response) return null;
  if (response && typeof response === "object" && "id" in (response as Record<string, unknown>) && "skus" in (response as Record<string, unknown>)) {
    return mapBackendProduct(response as BackendProduct);
  }
  if (response && typeof response === "object" && "data" in (response as Record<string, unknown>)) {
    return mapBackendProduct((response as Record<string, unknown>).data as BackendProduct);
  }
  return null;
}
