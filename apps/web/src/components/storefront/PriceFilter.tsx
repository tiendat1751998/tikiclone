"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useState, useEffect, useCallback } from "react";

const PRICE_RANGES = [
  { label: "Dưới 500.000", value: "0-500000" },
  { label: "500.000 - 1.000.000", value: "500000-1000000" },
  { label: "1.000.000 - 3.000.000", value: "1000000-3000000" },
  { label: "Trên 3.000.000", value: "3000000-" },
];

interface PriceFilterProps {
  basePath: string;
}

export function PriceFilter({ basePath }: PriceFilterProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [selectedPrices, setSelectedPrices] = useState<string[]>([]);

  useEffect(() => {
    const prices = searchParams.get("price");
    if (prices) {
      setSelectedPrices(prices.split(","));
    } else {
      setSelectedPrices([]);
    }
  }, [searchParams]);

  const togglePrice = useCallback((value: string) => {
    setSelectedPrices((prev) => {
      const next = prev.includes(value)
        ? prev.filter((v) => v !== value)
        : [...prev, value];
      
      const sp = new URLSearchParams(searchParams.toString());
      if (next.length > 0) {
        sp.set("price", next.join(","));
      } else {
        sp.delete("price");
      }
      sp.set("page", "1");
      router.push(`${basePath}?${sp.toString()}`, { scroll: false });
      
      return next;
    });
  }, [router, searchParams, basePath]);

  return (
    <div className="px-3 py-2">
      <div className="text-[10px] font-semibold text-tiki-text mb-1.5">GIÁ BÁN</div>
      <div className="flex flex-col gap-1">
        {PRICE_RANGES.map((r) => (
          <label key={r.value} className="flex items-center gap-2 text-[11px] text-tiki-text-secondary cursor-pointer">
            <input
              type="checkbox"
              checked={selectedPrices.includes(r.value)}
              onChange={() => togglePrice(r.value)}
              className="w-3 h-3 rounded border-gray-300 accent-tiki-blue"
            />
            <span>{r.label}</span>
          </label>
        ))}
      </div>
    </div>
  );
}
