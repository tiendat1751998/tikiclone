// SERVER COMPONENT - Personalized recommendations
import { productsApi } from "@/lib/api/products";
import { ProductCard } from "@/components/product/ProductCard";

interface Product {
  id: string; name: string; image_url: string; price: number;
  original_price?: number | null; discount_percent?: number | null;
  rating_average?: number | null; sold_count?: number;
}

export async function RecommendedProductsServer() {
  const products: Product[] = await productsApi.getFeatured(12);
  return (
    <section className="bg-white rounded-xl shadow-sm overflow-hidden">
      <div className="px-4 py-3 border-b border-[#e8e8e8]">
        <h2 className="text-sm font-bold text-[#222] flex items-center gap-2">⭐ Recommended for You</h2>
      </div>
      <div className="p-4">
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-2 md:gap-3">
          {products.map((p, i) => <ProductCard key={p.id} product={p} priority={i < 4} />)}
        </div>
      </div>
    </section>
  );
}
