import Link from "next/link";

interface Product {
  id: string;
  name: string;
  image_url: string;
  price: number;
  original_price?: number | null;
  discount_percent?: number | null;
  rating_average?: number | null;
  sold_count?: number;
}

export function ProductCard({ product, priority, isDeal }: { product: Product; priority?: boolean; isDeal?: boolean }) {
  return (
    <Link href={`/products/${product.id}`} className="block bg-white rounded-lg border border-gray-200 overflow-hidden hover:shadow-md transition-shadow h-full">
      <div className="aspect-square bg-gray-50 relative">
        <img src={product.image_url || "/placeholder.svg"} alt={product.name} className="w-full h-full object-contain" loading={priority ? "eager" : "lazy"} />
        {product.discount_percent && product.discount_percent > 0 && (
          <span className="absolute top-1 left-1 bg-red-500 text-white text-[10px] font-bold px-1.5 py-0.5 rounded">-{product.discount_percent}%</span>
        )}
        {isDeal && (
          <span className="absolute bottom-1 left-1 bg-orange-500 text-white text-[9px] font-bold px-1 py-0.5 rounded">DEAL</span>
        )}
      </div>
      <div className="p-2 space-y-1">
        <h3 className="text-xs line-clamp-2 text-gray-800 leading-tight">{product.name}</h3>
        <div className="flex items-center gap-1">
          <span className="text-sm font-bold text-red-500">{product.price.toLocaleString("vi-VN")}₫</span>
          {product.original_price && product.original_price > product.price && (
            <span className="text-[10px] text-gray-400 line-through">{product.original_price.toLocaleString("vi-VN")}₫</span>
          )}
        </div>
      </div>
    </Link>
  );
}
