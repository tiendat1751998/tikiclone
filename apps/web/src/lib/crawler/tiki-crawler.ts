// Server-side Tiki.vn crawler
// Crawls tiki.vn when data is missing from the database

export interface CrawledProduct {
  id: string;
  name: string;
  price: number;
  original_price: number | null;
  discount_percent: number | null;
  image_url: string;
  rating_average: number;
  review_count: number;
  quantity_sold: number;
  seller_name: string;
  category_slug: string;
  is_official: boolean;
}

const TIKI_API = "https://tiki.vn/api/v2";

const USER_AGENT =
  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36";

async function fetchFromTiki<T>(url: string): Promise<T | null> {
  try {
    const res = await fetch(url, {
      headers: {
        "User-Agent": USER_AGENT,
        Accept: "application/json",
      },
      next: { revalidate: 3600 }, // cache 1h
    });
    if (!res.ok) return null;
    return (await res.json()) as T;
  } catch {
    return null;
  }
}

export interface TikiProductResponse {
  data: Array<{
    id: number;
    name: string;
    price: number;
    original_price: number;
    discount: number;
    thumbnail_url: string;
    rating_average: number;
    review_count: number;
    quantity_sold: { value: number; text: string } | null;
    seller_name?: string;
    seller_product_id?: string;
    is_official?: boolean;
    is_authentic?: boolean;
  }>;
  paging: { total: number; total_pages: number };
}

export interface TikiCategoryResponse {
  id: number;
  name: string;
  slug: string;
  children?: TikiCategoryResponse[];
  product_count?: number;
}

// Map tiki.vn category slugs to our DB slugs
const CATEGORY_SLUG_MAP: Record<string, string> = {
  "nha-sach-tiki": "nha-sach-tiki",
  "nha-cua-doi-song": "nha-cua-doi-song",
  "dien-thoai-may-tinh-bang": "dien-thoai-may-tinh-bang",
  "do-choi-me-be": "do-choi-me-be",
  "thiet-bi-so-phu-kien-so": "thiet-bi-so-phu-kien-so",
  "thoi-trang-nu": "thoi-trang-nu",
  "thoi-trang-nam": "thoi-trang-nam",
  "giay-dep-nu": "giay-dep-nu",
  "giay-dep-nam": "giay-dep-nam",
  "dien-tu-dien-lanh": "dien-tu-dien-lanh",
  "bach-hoa-online": "bach-hoa-online",
  "the-thao-da-ngoai": "the-thao-da-ngoai",
  "lam-dep-suc-khoe": "lam-dep-suc-khoe",
  "suc-khoe": "suc-khoe",
  "vat-pham-van-phong": "vat-pham-van-phong",
  "tui-vi": "tui-vi",
  "laptop-may-vi-tinh-linh-kien": "laptop-may-vi-tinh-linh-kien",
  "thiet-bi-gia-dinh": "thiet-bi-gia-dinh",
  "thuoc-dong-y": "thuoc-dong-y",
  "thuc-pham-do-uong": "thuc-pham-do-uong",
  "dong-ho": "dong-ho",
};

const TIKI_CATEGORY_IDS: Record<string, number> = {
  "nha-sach-tiki": 8322,
  "nha-cua-doi-song": 1883,
  "dien-thoai-may-tinh-bang": 2549,
  "do-choi-me-be": 2541,
  "thiet-bi-so-phu-kien-so": 1789,
  "thoi-trang-nu": 2761,
  "thoi-trang-nam": 2760,
  "giay-dep-nu": 2770,
  "giay-dep-nam": 2771,
  "dien-tu-dien-lanh": 4221,
  "bach-hoa-online": 4384,
  "the-thao-da-ngoai": 2561,
  "lam-dep-suc-khoe": 1520,
  "laptop-may-vi-tinh-linh-kien": 1846,
  "thiet-bi-gia-dinh": 1882,
  "dong-ho": 3922,
  "tui-vi": 2774,
  "vat-pham-van-phong": 2807,
  "thuc-pham-do-uong": 4385,
  "suc-khoe": 2663,
  "thuoc-dong-y": 3483,
};

export function mapTikiSlug(tikiSlug: string): string | null {
  return CATEGORY_SLUG_MAP[tikiSlug] || null;
}

export function getTikiCategoryId(dbSlug: string): number | null {
  return TIKI_CATEGORY_IDS[dbSlug] || null;
}

export async function crawlTikiProducts(
  categorySlug: string,
  page: number = 1,
  limit: number = 20
): Promise<CrawledProduct[]> {
  const tikiCategoryId = getTikiCategoryId(categorySlug);
  if (!tikiCategoryId) return [];

  const url = `${TIKI_API}/products?category=${tikiCategoryId}&page=${page}&limit=${limit}&sort=top_sellers&aggregations=2`;
  const data = await fetchFromTiki<TikiProductResponse>(url);
  if (!data?.data) return [];

  return data.data.map((item) => ({
    id: `crawled-${item.id}`,
    name: item.name,
    price: item.price,
    original_price: item.original_price || null,
    discount_percent: item.discount || null,
    image_url: item.thumbnail_url || "",
    rating_average: item.rating_average || 0,
    review_count: item.review_count || 0,
    quantity_sold: item.quantity_sold?.value || 0,
    seller_name: item.seller_name || "Tiki Trading",
    category_slug: categorySlug,
    is_official: item.is_official || item.is_authentic || false,
  }));
}

export async function crawlTikiCategories(): Promise<
  Array<{ id: number; name: string; slug: string; children: Array<{ id: number; name: string; slug: string }> }>
> {
  const url = `${TIKI_API}/categories?parent_id=0`;
  const data = await fetchFromTiki<{ data: TikiCategoryResponse[] }>(url);
  if (!data?.data) return [];

  const result: Array<{
    id: number;
    name: string;
    slug: string;
    children: Array<{ id: number; name: string; slug: string }>;
  }> = [];

  for (const cat of data.data) {
    const children: Array<{ id: number; name: string; slug: string }> = [];
    if (cat.children) {
      for (const child of cat.children) {
        children.push({ id: child.id, name: child.name, slug: child.slug });
      }
    }
    result.push({ id: cat.id, name: cat.name, slug: cat.slug, children });
  }

  return result;
}

export function crawledToProduct(
  crawled: CrawledProduct
): import("@/types").Product {
  return {
    id: crawled.id,
    shop_id: "crawled-shop",
    category_id: crawled.category_slug,
    name: crawled.name,
    description: "",
    image_url: crawled.image_url,
    price: crawled.price,
    original_price: crawled.original_price,
    discount_percent: crawled.discount_percent,
    stock: 50,
    sold_count: crawled.quantity_sold,
    quantity_sold_text: crawled.quantity_sold
      ? `Đã bán ${crawled.quantity_sold}`
      : undefined,
    rating_average: crawled.rating_average,
    rating_count: crawled.review_count,
    review_count: crawled.review_count,
    seller_name: crawled.seller_name,
    is_official: crawled.is_official,
    is_tiki_trading: !crawled.is_official,
    is_sponsored: false,
    status: "active",
    condition: "new",
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };
}

export async function crawlDeals(limit: number = 10): Promise<CrawledProduct[]> {
  const url = `${TIKI_API}/products?limit=${limit}&sort=top_sellers&aggregations=2`;
  const data = await fetchFromTiki<TikiProductResponse>(url);
  if (!data?.data) return [];

  return data.data.map((item) => ({
    id: `crawled-deal-${item.id}`,
    name: item.name,
    price: item.price,
    original_price: item.original_price || null,
    discount_percent: item.discount || null,
    image_url: item.thumbnail_url || "",
    rating_average: item.rating_average || 0,
    review_count: item.review_count || 0,
    quantity_sold: item.quantity_sold?.value || 0,
    seller_name: item.seller_name || "Tiki Trading",
    category_slug: "all",
    is_official: item.is_official || item.is_authentic || false,
  }));
}

export async function seedAllFromTiki(): Promise<Record<string, number>> {
  const results: Record<string, number> = {};

  for (const [dbSlug, tikiId] of Object.entries(TIKI_CATEGORY_IDS)) {
    const products = await crawlTikiProducts(dbSlug, 1, 20);
    results[dbSlug] = products.length;
    // In production, would insert into database
    // await db.insert(products).into('products')
  }

  return results;
}
