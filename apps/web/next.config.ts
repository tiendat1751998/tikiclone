import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  poweredByHeader: false,
  compress: true,
  output: "standalone",

  // Enable ISR for pages that can be statically regenerated
  // This avoids full SSR on every request
  images: {
    remotePatterns: [
      { protocol: "https", hostname: "salt.tikicdn.com" },
      { protocol: "https", hostname: "tiki.vn" },
    ],
    formats: ["image/avif", "image/webp"],
    deviceSizes: [640, 750, 828, 1080, 1200],
    minimumCacheTTL: 86400,
  },

  // Cache headers for SSR pages — stale-while-revalidate pattern
  async headers() {
    return [
      {
        source: "/",
        headers: [
          { key: "Cache-Control", value: "public, s-maxage=10, stale-while-revalidate=60" },
        ],
      },
      {
        source: "/api/products",
        headers: [
          { key: "Cache-Control", value: "public, s-maxage=5, stale-while-revalidate=30" },
        ],
      },
      {
        source: "/_next/image(.*)",
        headers: [
          { key: "Cache-Control", value: "public, max-age=86400, immutable" },
        ],
      },
      {
        source: "/_next/static/(.*)",
        headers: [
          { key: "Cache-Control", value: "public, max-age=31536000, immutable" },
        ],
      },
    ];
  },

  experimental: {
    optimizePackageImports: [
      "lucide-react",
      "recharts",
    ],
  },
};

export default nextConfig;
