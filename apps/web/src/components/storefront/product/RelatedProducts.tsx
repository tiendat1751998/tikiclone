"use client";

import { useQuery } from "@tanstack/react-query";
import { recommendationsApi } from "@/lib/api/client";
import { ProductCard } from "./ProductCard";
import type { Product } from "@/types";

interface RelatedProductsProps {
  productId: string;
}

export default function RelatedProducts({ productId }: RelatedProductsProps) {
  const { data: products } = useQuery({
    queryKey: ["recommendations", productId],
    queryFn: () => recommendationsApi.getRelated(productId, 8),
    initialData: [] as Product[],
    staleTime: 5 * 60 * 1000, // 5 minutes
    refetchOnWindowFocus: false,
  });

  if (!products || products.length === 0) return null;

  return (
    <section className="mt-6 bg-white rounded-lg border border-tiki-border p-4">
      <h3 className="text-sm font-semibold text-tiki-text mb-3">Sản phẩm tương tự</h3>
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
        {products.slice(0, 8).map((product) => (
          <ProductCard key={product.id} product={product} />
        ))}
      </div>
    </section>
  );
}
