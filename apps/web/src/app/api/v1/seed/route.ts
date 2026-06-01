// API Route: Seed data from Tiki.vn crawler
import { NextRequest, NextResponse } from "next/server";
import {
  crawlTikiProducts,
  crawlDeals,
  crawledToProduct,
} from "@/lib/crawler/tiki-crawler";
import type { Product } from "@/types";

export async function GET(request: NextRequest) {
  const { searchParams } = request.nextUrl;
  const category = searchParams.get("category");
  const page = parseInt(searchParams.get("page") || "1", 10);
  const limit = parseInt(searchParams.get("limit") || "20", 10);

  try {
    if (category) {
      const crawled = await crawlTikiProducts(category, page, limit);
      const products: Product[] = crawled.map(crawledToProduct);
      return NextResponse.json({ success: true, data: products, count: products.length, source: "crawled" });
    }

    const deals = await crawlDeals(limit);
    const products: Product[] = deals.map(crawledToProduct);
    return NextResponse.json({ success: true, data: products, count: products.length, source: "crawled" });
  } catch (error) {
    console.error("Seed error:", error);
    return NextResponse.json({ success: false, error: "Crawl failed" }, { status: 500 });
  }
}
