import { NextRequest, NextResponse } from "next/server";
import { getCachedPage, setCachedPage } from "@/lib/cache/page-cache";

const CACHE_TTL = 30;

export const config = {
  matcher: ["/", "/products", "/products/:path*", "/categories/:path*"],
};

export async function middleware(request: NextRequest) {
  const path = request.nextUrl.pathname;

  if (request.method !== "GET") {
    return NextResponse.next();
  }

  if (!shouldCachePage(path)) {
    return NextResponse.next();
  }

  const cacheKey = `page:${path}`;
  const cachedHtml = await getCachedPage(cacheKey);

  if (cachedHtml) {
    return new NextResponse(cachedHtml, {
      headers: cacheHeaders("HIT"),
    });
  }

  const response = await NextResponse.next();
  const html = await response.text();

  const contentType = response.headers.get("content-type");
  if (contentType?.includes("text/html")) {
    await setCachedPage(cacheKey, html);
    return new NextResponse(html, {
      status: response.status,
      headers: cacheHeaders("MISS"),
    });
  }

  return response;
}

function cacheHeaders(cacheStatus: "HIT" | "MISS"): Headers {
  const headers = new Headers();
  headers.set("Content-Type", "text/html; charset=utf-8");
  headers.set("Cache-Control", `public, max-age=${CACHE_TTL}, s-maxage=${CACHE_TTL}`);
  headers.set("X-Cache", cacheStatus);
  return headers;
}

function shouldCachePage(path: string): boolean {
  if (path === "/" || path === "/products") {
    return true;
  }
  if (path.startsWith("/products/") || path.startsWith("/categories/")) {
    return true;
  }
  return false;
}