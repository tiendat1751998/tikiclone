// SERVER COMPONENT with streaming - fetches deals
import { Suspense } from "react";
import { productsApi } from "@/lib/api/products";
import { ProductCard } from "@/components/product/ProductCard";
import { ProductGridSkeleton } from "@/components/ui/Skeleton";
import { CountdownTimer } from "./CountdownTimer";

interface Product {
  id: string; name: string; image_url: string; price: number;
  original_price?: number | null; discount_percent?: number | null;
  rating_average?: number | null; sold_count?: number;
}

async function FlashSaleProducts() {
  const products: Product[] = await productsApi.getDeals(20);
  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-2 md:gap-3">
      {products.map((p, i) => <ProductCard key={p.id} product={p} priority={i < 4} isDeal />)}
    </div>
  );
}

export async function FlashSaleServer() {
  return (
    <section className="bg-white rounded-xl shadow-sm overflow-hidden">
      <div className="px-4 py-3 border-b border-[#e8e8e8] flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-lg">⚡</span>
          <h2 className="text-sm font-bold text-[#222]">Flash Sale</h2>
          <CountdownTimer />
        </div>
        <a href="/products?sort=deals" className="text-xs text-[#189eff] hover:underline font-medium">See All</a>
      </div>
      <div className="p-4">
        <Suspense fallback={<ProductGridSkeleton count={10} />}>
          <FlashSaleProducts />
        </Suspense>
      </div>
    </section>
  );
}
