"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { productsApi } from "@/lib/api/client";

interface FlashSaleData {
  end_time: string;
  products: { id: string; name: string; image_url: string; price: number; original_price: number }[];
}

export function FlashSaleTimer() {
  const [flashSale, setFlashSale] = useState<FlashSaleData | null>(null);

  useEffect(() => {
    productsApi.getFlashSale()
      .then((data) => setFlashSale(data as FlashSaleData))
      .catch(() => {
        const endTime = new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString();
        setFlashSale({ end_time: endTime, products: [] });
      });
  }, []);

  const calculateTimeLeft = useCallback(() => {
    if (!flashSale) return { hours: "00", minutes: "00", seconds: "00" };
    const endTime = new Date(flashSale.end_time).getTime();
    const now = Date.now();
    const diff = endTime - now;

    if (diff <= 0) return { hours: "00", minutes: "00", seconds: "00" };

    return {
      hours: String(Math.floor(diff / (1000 * 60 * 60))).padStart(2, "0"),
      minutes: String(Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))).padStart(2, "0"),
      seconds: String(Math.floor((diff % (1000 * 60)) / 1000)).padStart(2, "0"),
    };
  }, [flashSale]);

  const [timeLeft, setTimeLeft] = useState(calculateTimeLeft);

  useEffect(() => {
    const timer = setInterval(() => setTimeLeft(calculateTimeLeft()), 1000);
    return () => clearInterval(timer);
  }, [calculateTimeLeft]);

  return (
    <section className="mb-3">
      <div className="max-w-tiki mx-auto px-3">
        <div className="flash-sale px-4 py-2 flex items-center justify-between text-white rounded-lg">
          <div className="flex items-center gap-2">
            <span className="text-base">⚡</span>
            <span className="font-bold text-sm">FLASH SALE</span>
            <div className="flex items-center gap-1">
              <div className="bg-white/20 rounded px-1 py-0.5 text-xs font-mono font-bold">{timeLeft.hours}</div>
              <span>:</span>
              <div className="bg-white/20 rounded px-1 py-0.5 text-xs font-mono font-bold">{timeLeft.minutes}</div>
              <span>:</span>
              <div className="bg-white/20 rounded px-1 py-0.5 text-xs font-mono font-bold">{timeLeft.seconds}</div>
            </div>
          </div>
          <Link href="/promotions" className="text-xs font-medium text-white/90 hover:text-white">
            Xem tất cả →
          </Link>
        </div>
      </div>
    </section>
  );
}
