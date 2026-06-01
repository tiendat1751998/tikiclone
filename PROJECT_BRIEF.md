# TIKI CLONE — MASTER PROJECT BRIEF

> Synthesized from ALL 149 .md files. Last updated: 2026-05-27.

---

## 1. PROJECT VISION

Production-grade Tiki clone — Vietnam-scale e-commerce:
- 10M users, 1M DAU, 100K concurrent, 20K RPS burst
- p95 < 200ms internal APIs, p99 < 500ms
- 99.9% uptime. Kubernetes-native, horizontally scalable.

---

## 2. MONOREPO STRUCTURE

```
tiki-clone/
├── .ai/                    # Architecture docs, rules, planning
├── .github/                # GitHub agents, prompts, OpenSpec
├── apps/web/               # Next.js 15 storefront (Tiki clone), output:"standalone"
├── services/               # Core Go microservices
│   ├── gateway/            # API gateway, rate limiting, routing
│   ├── auth/               # JWT, sessions, RBAC
│   ├── product/            # Product CRUD, search
│   ├── product-catalog/    # Catalog management
│   ├── inventory/          # Stock management, Redis cache
│   ├── cart/               # Cart management
│   ├── promotion/          # Discounts, vouchers
│   ├── checkout/           # Checkout flow
│   ├── order/              # Order management
│   ├── payment/            # Payment processing
│   ├── shipment/           # Shipping, tracking
│   └── identity-auth/      # Java Spring Boot identity service
├── packages/               # Shared Go libraries
├── k8s/                    # Kubernetes manifests
├── database/               # Schema, migrations
└── scripts/                # Build, deploy scripts
```

---

## 3. TECH STACK

- **Backend**: Go (microservices), Java Spring Boot (identity-auth)
- **Frontend**: Next.js 15 App Router, Zustand, React Query
- **Databases**: MySQL (primary), Redis (cache/sessions)
- **Message Queue**: Kafka
- **Infrastructure**: Docker, Kubernetes, gRPC + REST

---

## 4. KEY CONVENTIONS

- Prices: int64 (BIGINT), weight: int, currency: VND
- DB enum: UPPERCASE
- Domain errors: *DomainError pointers
- MySQL: host=mysql-primary, port=3306, user=tiki, pass=tiki_dev
- K8s YAML: single-document only (no --- separators)
- SQL escaping: '' not \'
- Next.js: --legacy-peer-deps, rm -rf .next before rebuild
- Go: sonic v1.15.1 (not v1.12.4, incompatible with Go 1.26.3)

---

## 5. SERVICE COMMUNICATION

- Gateway → services: gRPC (internal), REST (external)
- Service → service: gRPC via service discovery
- Async: Kafka for event-driven (order events, inventory updates)
- Auth: JWT access + refresh tokens, Redis blacklist

---

## 6. KNOWN ISSUES

| Issue | Severity | Status |
|-------|----------|--------|
| Wrong args to NewPaymentService | ERROR | BLOCKED (pre-existing) |
| TestResetPasswordValidation expects 422 gets 400 | ERROR | BLOCKED (pre-existing) |

---

## 7. ARCHITECTURE DETAILS

Read sections on-demand. Don't load entire brief for simple tasks.

### 7.1 Auth Flow
Client → Next.js proxy → Gateway (rate limit + JWT verify) → Auth service → MySQL + Redis

### 7.2 Rate Limiting
Gateway-level, per-client-IP. Uses X-Forwarded-For + X-Real-IP (not c.ClientIP() in Docker).

### 7.3 Checkout Flow
Cart → Checkout (price calc + validation) → Payment → Order → Shipment. Saga pattern for distributed tx.

### 7.4 Frontend State
Zustand for client state (cart, user, UI). React Query for server state. Server Components for static content.
