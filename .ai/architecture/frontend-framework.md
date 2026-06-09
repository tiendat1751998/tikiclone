# Frontend Architecture & Framework Selection Guide

> **Project**: TikiClone · Next.js 15 Storefront
> **Status**: Production (Docker Swarm, 6 replicas)
> **Updated**: 2026-06-05

---

## 1. TikiClone Frontend Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Framework | **Next.js 15.5.19** (App Router) | SSR-first e-commerce: SEO-critical, dynamic pricing, personalized content |
| State (Client) | **Zustand** | Lightweight, no boilerplate, DevTools support |
| State (Server) | **fetch + revalidate** (no React Query) | SSR page 1 + "Load more" client pattern avoids re-fetch duplication |
| Styling | **Tailwind CSS** + custom `tiki-*` tokens | Design system tokens: `signature-blue`, `signature-orange`, spacing/border tokens |
| Images | `next/image` (remotePatterns: `salt.tikicdn.com`) | AVIF/WebP, 86,400s cache TTL |
| Caching | `stale-while-revalidate` = 60s (HTML), 86,400s (static assets) | Balance freshness vs. latency |
| Deployment | `output: "standalone"` on `node:20-alpine` | Docker Swarm compatible, minimal image size |

### Critical Patterns
- **SSR page 1** (Server Component) + **client "Load more"** (plain fetch, NOT React Query's `useInfiniteQuery` — that always re-fetches on mount causing product duplication)
- `"use client"` + `async function` = re-fetches on every render → 429 flood
- Cart API requires all fields — no partial updates

---

## 2. Framework Decision Matrix

### When to Use What

| Criterion | **Next.js** | **Astro** | **SvelteKit** | **Remix** |
|-----------|-------------|-----------|---------------|-----------|
| **SEO-critical SSR** | ✅ Best choice | ✅ Islands | ⚠️ Needs config | ✅ Great |
| **Dynamic user-specific content** | ✅ RSC + Server Actions | ❌ Static-first | ✅ Stores | ✅ Loaders |
| **Real-time / WebSocket heavy** | ⚠️ Polling/SSE | ❌ | ✅ First-class | ⚠️ Resource routes |
| **Content-heavy / marketing** | ⚠️ ISR possible | ✅ **Best choice** | ⚠️ | ⚠️ |
| **E-commerce (browse→cart→checkout)** | ✅ **Best choice** | ❌ Checkout needs JS | ⚠️ | ✅ Good fit |
| **Dashboard / admin SPA** | ✅ App Router | ❌ | ✅ Reactive | ✅ Progressive |
| **Team skill availability** | ✅ Common | ⚠️ Niche | ⚠️ Growing | ⚠️ Niche |
| **API integration (multiple backends)** | ✅ fetch / Server Actions | ❌ Static gen | ✅ Hooks | ✅ Loaders + Actions |
| **Partial Prerendering (PPR)** | ✅ Experimental | ✅ Astro static by default | ❌ | ❌ |
| **Streaming SSR** | ✅ `loading.tsx` + Suspense | ❌ | ✅ | ✅ Remix |

### 2.1 Next.js 15+ — E-commerce / Full-Stack Apps

**Choose when**:
- SSR with ISR for product/category pages
- Server Actions for form handling (checkout, add-to-cart)
- Mixed static + dynamic content per route
- Team already knows React ecosystem

**Available in TikiClone**:
- App Router: `/products/[id]`, `/categories/[slug]`, `/checkout`
- RSC for initial product listings → client hydrates with "Load more"
- ISR: `revalidate: 60` for homepage, `revalidate: 5` for `/api/products`
- `optimizePackageImports`: `lucide-react`, `date-fns`

```typescript
// Recommended: SSR initial page + client pagination
// pages/page.tsx (Server Component)
async function HomePage() {
  const products = await fetchProducts("/products/featured?limit=10");
  return <ProductGrid initialProducts={products} />;
}

// ProductGrid (Client Component) — fetch only extra pages
"use client";
function ProductGrid({ initialProducts }: { initialProducts: Product[] }) {
  const [products, setProducts] = useState(initialProducts);
  const [page, setPage] = useState(1);

  async function loadMore() {
    const res = await fetch(`/api/products?page=${page + 1}`);
    const data = await res.json();
    setProducts(prev => [...prev, ...data]);
    setPage(p => p + 1);
  }

  return (
    <div>
      {products.map(p => <ProductCard key={p.id} product={p} />)}
      <button onClick={loadMore}>Xem thêm</button>
    </div>
  );
}
```

**Pitfalls to avoid**:
- ❌ `"use client"` + `async function` in page.tsx = infinite re-fetch loop
- ❌ React Query `useInfiniteQuery` on already-SSR'd data = duplicate products
- ❌ Mixing Zustand + React Query for same data = stale/loading mismatch

### 2.2 Astro — Content-Heavy / Marketing Sites

**Choose when**:
- Mostly static content: blog, landing pages, docs, marketing
- Minimal dynamic user interaction
- Maximum performance via 0 JS by default
- Content from MDX, CMS, or headless API

**Do NOT use for**:
- Complex state management (cart, checkout flow)
- Real-time features
- Heavy client-side interactivity

```astro
---
// Astro example — zero JS shipped
import ProductCard from "../components/ProductCard.astro";
const products = await fetch("http://gateway:8080/api/v1/products").then(r => r.json());
---
<main>
  {products.map(p => <ProductCard product={p} />)}
</main>
```

### 2.3 SvelteKit — Real-Time / Reactive Apps

**Choose when**:
- Real-time dashboards (WebSocket-heavy)
- Highly reactive UIs (collaborative tools, monitoring)
- Team prefers Svelte's minimal syntax
- Offline-first PWA

**Do NOT use for**:
- SEO-heavy e-commerce (SSR possible but less mature ecosystem)
- Large team with React expertise

### 2.4 Remix — Web Fundamentals / Progressive Enhancement

**Choose when**:
- Focus on web standards (FormData, progressive enhancement)
- Server-side mutations without custom API routes
- Strong emphasis on data loading patterns
- Nested routing with parent/child data dependency

---

## 3. Current Architecture (TikiClone Next.js)

```
pages/
├── page.tsx                  ← RSC: SSR featured + deals + categories
├── layout.tsx                ← RSC: Header + Footer
├── loading.tsx               ← Streaming fallback
├── error.tsx                 ← Error boundary
├── products/
│   ├── page.tsx              ← RSC: product listing SSR
│   ├── [id]/
│   │   ├── page.tsx          ← RSC: product detail
│   │   ├── ProductDetailClient.tsx  ← Client: interactions, images
│   │   ├── AddToCartButton.tsx      ← Client: add to cart
│   │   └── BuyNowButton.tsx        ← Client: buy now
│   └── ProductsListingClient.tsx   ← Client: filters + pagination
├── categories/[slug]/
│   ├── page.tsx              ← RSC: category SSR
│   └── CategoryListingClient.tsx   ← Client: load more, sort
├── cart/page.tsx             ← Client: cart management
├── checkout/
│   ├── page.tsx              ← Client: checkout form
│   └── success/page.tsx      ← RSC: order confirmation
└── search/
    ├── page.tsx              ← RSC: search SSR
    └── SearchListingClient.tsx     ← Client: pagination
```

### Component Tree

```
<Header>
  ├── Logo, SearchBar, CartIcon, UserMenu (client interactivity)
<main>
  ├── FlashSaleTimer (client — countdown)
  ├── QuickServiceDesktop (static)
  ├── CategoryProductSection × N (RSC — products from API)
  │   └── ProductCard (RSC → client on scroll)
  └── Pagination / Load More (client — Zustand page state)
<Footer>
```

### Data Flow

```
Browser
  ↓ HTTPS
Traefik (port 443 TLS)
  ↓ PathPrefix("/")      ↓ PathPrefix("/api/")
Next.js (6 replicas)    Gateway (6 replicas)
  ↓ fetch                    ↓ gRPC
  └── GW:8080            └── Services (product, cart, etc.)
                             ↓
                           MySQL / Redis / MongoDB
```

---

## 4. Build & Deploy

```bash
# Development
cd apps/web
pnpm dev                # Turbopack (HMR ~50ms)

# Production build
pnpm build              # output: "standalone" → .next/standalone/
docker build -t tikiclone-web:latest .

# Deploy to Swarm
docker service update --image tikiclone-web:latest web
```

### Docker
```dockerfile
FROM node:20-alpine AS base
FROM base AS deps     # pnpm fetch
FROM base AS builder  # pnpm build
FROM base AS runner   # node:20-alpine — no health check
COPY --from=builder /app/.next/standalone ./
EXPOSE 3000
CMD ["node", "server.js"]
```

---

## 5. Migration Guide: Switching Framework

If requirements shift, map concepts:

| Concept | Next.js → | Astro | SvelteKit | Remix |
|---------|-----------|-------|-----------|-------|
| Server Component | → `.astro` template | → `+page.server.js` | → `loader()` |
| Client Component | → `client:load` island | → `+page.js` | → `clientLoader()` |
| API Route | → API endpoint + fetch | → `+server.js` | → `action()` |
| Layout | → `Layout.astro` | → `+layout.svelte` | → `root.tsx` |
| Middleware | → `middleware.ts` | → `hooks.server.js` | → `entry.server.tsx` |
| ISR | → N/A (all static by default) | → `revalidate` config | → `headers()` + CDN |

---

## 6. Performance Targets (TikiClone)

| Metric | Target | Current | Method |
|--------|--------|---------|--------|
| LCP (homepage) | < 1.5s | ~800ms | RSC + ISR + streaming |
| FCP (any page) | < 800ms | ~400ms | No client JS for initial render |
| TTI | < 2s | ~1.2s | Selective hydration |
| API → GW → Service | < 200ms p95 | ~20ms | Multi-replica, keepalive conns |
| Static asset cache | 100% cache hit | Next.js 86,400s | CDN / immutable headers |

---

*Framework selection reviewed 2026-06-05 · Current: Next.js 15 App Router with RSC + Server Actions*
