import Link from "next/link";
import type { Product } from "@/types";

interface ProductCardProps {
  product: Product;
  priority?: boolean;
}

export function ProductCard({ product, priority = false }: ProductCardProps) {
  const discount = product.discount_percent ?? null;
  const hasDiscount = discount && discount > 0;
  const hasFreeShipping = product.price && product.price >= 100000; // Tiki.vn style: freeship from 100k

  return (
    <Link
      href={`/products/${product.id}`}
      className="product-card group"
    >
      <div className="product-card__image">
        <img
          src={product.image_url || "/images/placeholder.svg"}
          alt={product.name}
          loading={priority ? "eager" : "lazy"}
        />
        {hasDiscount && (
          <div className="product-card__discount">-{discount}%</div>
        )}
        {product.is_official && (
          <div className="product-card__official">CHÍNH HÃNG</div>
        )}
        {hasFreeShipping && (
          <div className="absolute top-0 right-0 bg-tiki-green text-white text-[9px] font-bold px-1 py-0.5 rounded-bl">
            FREESHIP
          </div>
        )}
      </div>

      <div className="product-card__body">
        <h3 className="product-card__name">{product.name}</h3>

        <div className="flex flex-col gap-0.5">
          <div className="product-card__price">
            <span className="product-card__price-current">
              {product.price?.toLocaleString("vi-VN")} ₫
            </span>
            {hasDiscount && (
              <span className="product-card__price-discount">-{discount}%</span>
            )}
          </div>

          {product.original_price && product.original_price > product.price && (
            <div className="product-card__price-original">
              {product.original_price.toLocaleString("vi-VN")} ₫
            </div>
          )}
        </div>

        {product.rating_average && product.rating_average > 0 && (
          <div className="product-card__rating">
            <span className="text-yellow-400 text-xs">★</span>
            <span className="text-tiki-text text-xs font-medium">{product.rating_average.toFixed(1)}</span>
            {product.quantity_sold_text && (
              <>
                <span className="product-card__rating-sep">|</span>
                <span className="product-card__sold">{product.quantity_sold_text}</span>
              </>
            )}
          </div>
        )}

        {product.seller_name && (
          <div className="product-card__seller">
            {product.is_tiki_trading && (
              <div className="product-card__seller-badge">T</div>
            )}
            <span className="product-card__seller-name">{product.seller_name}</span>
          </div>
        )}
      </div>
    </Link>
  );
}

export function ProductCardSkeleton() {
  return (
    <div className="bg-white rounded-lg border border-tiki-border overflow-hidden">
      <div className="aspect-square bg-gray-200 animate-pulse" />
      <div className="p-3 space-y-2">
        <div className="h-4 bg-gray-200 rounded animate-pulse" />
        <div className="h-4 w-2/3 bg-gray-200 rounded animate-pulse" />
        <div className="h-5 w-1/2 bg-gray-200 rounded animate-pulse" />
      </div>
    </div>
  );
}

interface ProductGridProps {
  products: Product[];
  skeletonCount?: number;
  isLoading?: boolean;
}

export function ProductGrid({ products, skeletonCount = 8, isLoading }: ProductGridProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3">
        {Array.from({ length: skeletonCount }).map((_, i) => (
          <ProductCardSkeleton key={i} />
        ))}
      </div>
    );
  }

  if (!products || products.length === 0) {
    return (
      <div className="text-center py-12 text-tiki-text-secondary">
        <p>Không tìm thấy sản phẩm nào</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3">
      {products.map((product, i) => (
        <ProductCard key={product.id} product={product} priority={i < 4} />
      ))}
    </div>
  );
}
