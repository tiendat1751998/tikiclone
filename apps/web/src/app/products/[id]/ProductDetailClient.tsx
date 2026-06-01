"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { StarRating } from "@/components/ui";
import { AddToCartButton } from "./AddToCartButton";
import { useCartStore } from "@/stores/cart";
import { useUIStore } from "@/stores/ui";
import ProductDetailWithReviews from "@/components/storefront/product/ProductReview";
import RelatedProducts from "@/components/storefront/product/RelatedProducts";

interface ProductImage {
  id: string;
  url: string;
  is_primary: boolean;
}

interface ProductDetail {
  id: string;
  name: string;
  description?: string;
  short_description?: string;
  image_url: string;
  images?: ProductImage[];
  price: number;
  original_price?: number | null;
  discount_percent?: number | null;
  stock: number;
  sold_count: number;
  quantity_sold_text?: string;
  rating_average?: number | null;
  rating_count?: number;
  review_count?: number;
  brand?: string;
  seller_name?: string;
  seller_avatar_url?: string;
  is_official?: boolean;
  attributes?: { name: string; value: string }[];
  category_name?: string;
  category_id: string;
  weight?: number;
  dimensions?: string;
  status: string;
  shop_id?: string;
  shop_name?: string;
}

const VARIANTS = [
  { id: "v1", name: "Xanh Navy", image: "/images/placeholder.svg" },
  { id: "v2", name: "Đen", image: "/images/placeholder.svg" },
  { id: "v3", name: "Trắng", image: "/images/placeholder.svg" },
];

export default function ProductDetailClient({
  product,
  allImages,
}: {
  product: ProductDetail;
  allImages: ProductImage[];
}) {
  const router = useRouter();
  const [selectedImage, setSelectedImage] = useState(0);
  const [quantity, setQuantity] = useState(1);
  const [selectedVariant, setSelectedVariant] = useState(VARIANTS[0]?.id || "v1");
  const addItem = useCartStore((s) => s.addItem);
  const addToast = useUIStore((s) => s.addToast);
  const [isBuying, setIsBuying] = useState(false);

  const currentImage = allImages[selectedImage]?.url || product.image_url;
  const subtotal = product.price * quantity;
  const hasDiscount = product.discount_percent != null && product.discount_percent > 0;

  const handleBuyNow = async () => {
    setIsBuying(true);
    try {
      await addItem({
        product_id: product.id, sku_id: product.id, name: product.name,
        image_url: product.image_url, price: product.price, quantity,
        stock: product.stock || 0, shop_id: product.shop_id, shop_name: product.shop_name,
      });
      router.push("/checkout");
    } catch {
      addToast({ type: "error", title: "Có lỗi xảy ra", message: "Không thể thêm sản phẩm" });
    } finally {
      setIsBuying(false);
    }
  };

  return (
    <>
      <div className="max-w-tiki mx-auto px-3 py-3" style={{ backgroundColor: "#F5F5FA" }}>
        <div className="flex items-center h-8 text-xs text-tiki-text-secondary mb-3">
          <Link href="/" className="hover:text-tiki-blue">Trang chủ</Link>
          <svg className="mx-1.5" width="5" height="8" viewBox="0 0 5 8" fill="none"><path d="M1 1L4 4L1 7" stroke="#787880" strokeWidth="1.5"/></svg>
          {product.category_name && (
            <>
              <Link href={`/categories/${product.category_id}`} className="hover:text-tiki-blue">{product.category_name}</Link>
              <svg className="mx-1.5" width="5" height="8" viewBox="0 0 5 8" fill="none"><path d="M1 1L4 4L1 7" stroke="#787880" strokeWidth="1.5"/></svg>
            </>
          )}
          <span className="text-tiki-text truncate max-w-xs">{product.name}</span>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-[30%_45%_25%] gap-4">
          <div>
            <div className="bg-white rounded-lg border border-tiki-border p-4">
              <div className="aspect-square bg-gray-50 rounded-lg overflow-hidden mb-3">
                <img src={currentImage} alt={product.name} className="w-full h-full object-contain" />
              </div>
              {allImages.length > 1 && (
                <div className="flex gap-2">
                  {allImages.map((img, idx) => (
                    <button
                      key={img.id}
                      onClick={() => setSelectedImage(idx)}
                      className={`w-12 h-12 shrink-0 rounded border-2 overflow-hidden transition ${
                        idx === selectedImage ? "border-tiki-blue" : "border-gray-200 hover:border-gray-300"
                      }`}
                    >
                      <img src={img.url} alt={`Thumbnail ${idx + 1} of ${product.name}`} className="w-full h-full object-cover" />
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          <div className="lg:col-span-1 flex flex-col gap-4">
            <div className="bg-white rounded-lg border border-tiki-border p-4">
              <div className="flex items-center gap-2 mb-2">
                <span className="text-[10px] bg-orange-500 text-white px-1.5 py-0.5 rounded font-medium">30 NGÀY ĐỔI TRẢ</span>
                {product.is_official && (
                  <span className="text-[10px] bg-tiki-blue text-white px-1.5 py-0.5 rounded font-medium">Chính hãng</span>
                )}
              </div>

              <h1 className="text-lg font-bold text-tiki-text mb-2">{product.name}</h1>

              <div className="flex items-center gap-3 text-xs text-tiki-text-secondary mb-3">
                {product.rating_average != null && product.rating_average > 0 && (
                  <div className="flex items-center gap-1">
                    <StarRating rating={product.rating_average} size="sm" />
                    <span className="font-medium">{product.rating_average.toFixed(1)}</span>
                  </div>
                )}
                {product.review_count != null && product.review_count > 0 && (
                  <a href="#reviews" className="text-tiki-blue hover:underline">{product.review_count} đánh giá</a>
                )}
                {product.sold_count > 0 && (
                  <span className="text-tiki-text-secondary">Đã bán {product.quantity_sold_text || product.sold_count.toLocaleString("vi-VN")}</span>
                )}
              </div>

              <div className="text-2xl font-bold text-tiki-red mb-4">
                {product.price.toLocaleString("vi-VN")} ₫
                {hasDiscount && product.original_price && (
                  <span className="text-sm text-tiki-text-secondary line-through ml-2 font-normal">
                    {product.original_price.toLocaleString("vi-VN")} ₫
                  </span>
                )}
              </div>

              <div className="mb-4">
                <span className="text-xs font-semibold text-tiki-text mb-2 block">Màu sắc</span>
                <div className="flex gap-2 flex-wrap">
                  {VARIANTS.map((variant) => (
                    <button
                      key={variant.id}
                      onClick={() => setSelectedVariant(variant.id)}
                      className={`flex items-center gap-1.5 px-2 py-1 text-xs border rounded transition ${
                        selectedVariant === variant.id
                          ? "border-tiki-blue ring-1 ring-tiki-blue"
                          : "border-gray-300 hover:border-tiki-blue"
                      }`}
                    >
                      <img src={variant.image} alt={variant.name} className="w-4 h-4 rounded" />
                      <span className={selectedVariant === variant.id ? "font-medium" : ""}>{variant.name}</span>
                    </button>
                  ))}
                </div>
              </div>
            </div>

            <div className="bg-white rounded-lg border border-tiki-border p-4">
              <div className="flex items-center gap-2 mb-2">
                <svg className="w-4 h-4 text-tiki-text-secondary" viewBox="0 0 24 24" fill="none"><path d="M12 2C8.13 2 5 5.13 5 9c0 5.55 7 13 7 13s7-7.45 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5S10.62 6.5 12 6.5s2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z" fill="currentColor"/></svg>
                <span className="text-xs font-semibold text-tiki-text">Thông tin vận chuyển</span>
              </div>
              <div className="flex items-center gap-1 text-xs text-tiki-text-secondary mb-1">
                <span>Giao đến:</span>
                <span className="text-tiki-text font-medium">Q. Hoàn Kiếm, Hà Nội</span>
                <button className="text-tiki-blue hover:underline">Đổi</button>
              </div>
              <div className="text-xs text-tiki-text mb-1">Giao ngày mai • Freeship 16.500đ</div>
            </div>

            <div className="bg-white rounded-lg border border-tiki-border p-4">
              <h3 className="text-xs font-semibold text-tiki-text mb-3">Dịch vụ bổ sung</h3>
              <div className="space-y-2 text-xs">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" className="w-4 h-4" />
                  <span className="text-tiki-text">Đăng ký TikiCARD nhận ngay ưu đãi</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" className="w-4 h-4" />
                  <span className="text-tiki-text">Trả góp từ 0% qua thẻ tín dụng</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" className="w-4 h-4" />
                  <span className="text-tiki-text">Mua trước trả sau (BNPL)</span>
                </label>
              </div>
            </div>
          </div>

          <div>
            <div className="sticky top-4 bg-white rounded-lg border border-tiki-border p-4">
              <div className="flex items-center gap-2 mb-3">
                <span className="text-xs font-semibold text-tiki-blue">Tiki Trading</span>
                <div className="flex items-center">
                  <span className="text-yellow-400 text-xs">★★★★★</span>
                </div>
                <span className="text-[10px] bg-tiki-blue text-white px-1 rounded">Official</span>
              </div>

              <div className="flex items-center gap-2 mb-3">
                <img src={product.image_url || "/images/placeholder.svg"} alt={product.name} className="w-12 h-12 rounded border border-tiki-border object-contain" />
                <div className="flex-1 min-w-0">
                  <p className="text-xs text-tiki-text truncate">{product.name}</p>
                  <p className="text-xs text-tiki-text-secondary">Còn hàng</p>
                </div>
              </div>

              <div className="flex items-center gap-2 mb-3">
                <span className="text-xs text-tiki-text-secondary">Số lượng</span>
                <div className="flex items-center border border-gray-300 rounded">
                  <button className="w-8 h-8 flex items-center justify-center text-sm text-tiki-text border-r border-gray-300 hover:bg-gray-50" onClick={() => setQuantity(Math.max(1, quantity - 1))}>−</button>
                  <input className="w-10 h-8 text-center text-xs border-none outline-none" type="text" value={quantity} readOnly />
                  <button className="w-8 h-8 flex items-center justify-center text-sm text-tiki-text border-l border-gray-300 hover:bg-gray-50" onClick={() => setQuantity(Math.min(product.stock || 999, quantity + 1))}>+</button>
                </div>
              </div>

              <div className="border-t border-tiki-border pt-3 mb-3">
                <div className="flex justify-between text-xs">
                  <span className="text-tiki-text-secondary">Tạm tính</span>
                  <span className="text-lg font-bold text-tiki-red">{subtotal.toLocaleString("vi-VN")} ₫</span>
                </div>
              </div>

              <div className="space-y-2">
                <button onClick={handleBuyNow} disabled={isBuying} className="w-full py-3 bg-tiki-red text-white rounded-lg font-semibold text-sm hover:bg-red-600 transition disabled:opacity-50">
                  {isBuying ? "Đang xử lý..." : "Mua ngay"}
                </button>
                <AddToCartButton product={product} quantity={quantity} />
                <button className="w-full py-2.5 text-tiki-blue border border-tiki-blue rounded-lg font-semibold text-sm hover:bg-blue-50 transition">
                  Mua trả góp - trả sau
                </button>
              </div>

              <div className="mt-4">
                <img src="/images/placeholder.svg" alt="Promotion" className="w-full h-16 rounded object-cover" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-tiki mx-auto px-3">
        <ProductDetailWithReviews product={product} />
        <RelatedProducts productId={product.id} />
      </div>

      <div className="fixed right-4 bottom-20 flex flex-col gap-2 z-40">
        <button className="w-12 h-12 bg-white rounded-full shadow-lg flex items-center justify-center hover:bg-tiki-bg transition">
          <span className="text-xl">🤖</span>
        </button>
        <button className="w-12 h-12 bg-white rounded-full shadow-lg flex items-center justify-center hover:bg-tiki-bg transition">
          <span className="text-xl">🔔</span>
        </button>
      </div>
    </>
  );
}