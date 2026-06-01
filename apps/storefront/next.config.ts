import type { NextConfig } from 'next';

const API_GATEWAY_URL = process.env.API_GATEWAY_URL || 'http://localhost:8080';
const GRAPHQL_BFF_URL = process.env.GRAPHQL_BFF_URL || 'http://localhost:4000';

const nextConfig: NextConfig = {
  output: 'standalone',
  outputFileTracingIncludes: { '*': ['./public/**/*', './src/app/**/*'] },

  // Performance: optimize package imports
  experimental: {
    optimizePackageImports: ['lucide-react', 'date-fns', '@tiki/ui-system'],
  },

  images: {
    remotePatterns: [
      { protocol: 'http', hostname: 'localhost', port: '9000' },
      { protocol: 'http', hostname: 'localhost', port: '8080' },
      { protocol: 'https', hostname: '**.tiki-clone.com' },
      { protocol: 'https', hostname: '**.s3.amazonaws.com' },
      { protocol: 'https', hostname: '**.cloudfront.net' },
    ],
    formats: ['image/avif', 'image/webp'],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920],
    minimumCacheTTL: 86400,
  },

  async rewrites() {
    return [
      { source: '/api/gateway/:path*', destination: `${API_GATEWAY_URL}/api/v1/:path*` },
      { source: '/api/graphql', destination: `${GRAPHQL_BFF_URL}` },
      { source: '/api/graphql/:path*', destination: `${GRAPHQL_BFF_URL}` },
    ];
  },

  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          { key: 'X-Content-Type-Options', value: 'nosniff' },
          { key: 'X-Frame-Options', value: 'DENY' },
          { key: 'X-XSS-Protection', value: '1; mode=block' },
          { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
        ],
      },
      {
        source: '/:path*.:ext(js|css|svg|png|jpg|jpeg|webp|avif|woff2|ico)',
        headers: [{ key: 'Cache-Control', value: 'public, max-age=31536000, immutable' }],
      },
    ];
  },
};

export default nextConfig;
