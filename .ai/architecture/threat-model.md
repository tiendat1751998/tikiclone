# TikiClone — Threat Model (STRIDE)

> **Scope**: Full system — Traefik → Gateway → Microservices → Databases
> **Methodology**: STRIDE per component + Data Flow Diagrams (DFD)
> **Date**: 2026-06-05

---

## 1. DFD Context (Data Flow Summary)

```
[Browser] --HTTPS--> [Traefik] --HTTP--> [Gateway] --gRPC--> [Microservices] --> [MySQL/Redis/MongoDB]
                     [Traefik] --HTTP--> [Next.js Storefront] --fetch--> [Gateway]
```

### Trust Boundaries
1. **Browser ↔ Traefik**: Internet → internal (TLS boundary)
2. **Traefik ↔ Gateway**: Internal network (frontend overlay)
3. **Gateway ↔ Microservices**: Swarm internal (backend overlay)
4. **Microservices ↔ Databases**: Dedicated VLAN (standalone DB VMs)

---

## 2. STRIDE Analysis

### 2.1 Traefik Ingress (Edge Router)

| Threat | Type | Impact | Mitigation | Severity |
|--------|------|--------|------------|----------|
| TLS termination bypass via HTTP downgrade | **S**poofing | Attacker intercepts plaintext | HTTP→HTTPS redirect (301), HSTS headers | **M** |
| Swarm provider socket exposure | **I**nformation Disclosure | Docker socket read | Socket bound read-only (`:ro`), port 8080 enterprise-only | **M** |
| Path traversal to admin routes | **E**levation of Privilege | Dashboard access | `--api.dashboard=true` on internal net, no external exposure | **L** |
| DDoS via connection flood | **D**enial of Service | Resource exhaustion | Rate limiting at Gateway layer (per client-IP) | **H** |

### 2.2 Gateway Service (Go — 6 replicas)

| Threat | Type | Impact | Mitigation | Severity |
|--------|------|--------|------------|----------|
| JWT secret brute force | **S**poofing | Forge tokens | Dev secret only; prod requires Vault-managed 256-bit key | **H** |
| JWT token replay | **S**poofing | Session hijack | 15min access TTL + Redis blacklist | **M** |
| gRPC introspection via reflection | **I**nformation Disclosure | Service discovery | Reflection disabled in production builds | **L** |
| Unlimited request body | **D**enial of Service | Memory exhaustion | Request size limit middleware (default 10MB) | **M** |
| Missing auth gate on routes | **E**levation of Privilege | Unauthenticated write | P0 fixed: delivery route, catalog POST/PUT/DELETE now gated | **H** (fixed) |

### 2.3 identity-auth Service (Java — 3 replicas)

| Threat | Type | Impact | Mitigation | Severity |
|--------|------|--------|------------|----------|
| SQL injection in login | **T**ampering | Credential leak | Parameterized queries (JPA/JDBCTemplate) | **H** |
| Session fixation | **S**poofing | Takeover | Regenerate session ID on login | **M** |
| Brute force login | **D**enial of Service | Account lockout | `RATE_LIMIT_LOGIN_MAX=5` per user, `RATE_LIMIT_IP_MAX=20` | **M** |
| Refresh token theft (XSS) | **I**nformation Disclosure | Persistent access | Refresh token stored in httpOnly cookie, device fingerprint | **H** |
| Concurrent session limit bypass | **S**poofing | Session overflow | Max 10 sessions enforced in Redis | **L** |

### 2.4 Checkout Saga Orchestrator (Go — 3 replicas)

| Threat | Type | Impact | Mitigation | Severity |
|--------|------|--------|------------|----------|
| Double spend (race condition) | **T**ampering | Negative stock | Redis Lua atomic DECRBY (single-threaded) | **H** |
| Saga timeout → inconsistent state | **T**ampering | Orphaned reservation | Compensation step: release.lua on failure | **M** |
| Idempotency key replay | **T**ampering | Duplicate order | Payment idempotency via transaction_ref (UUID) | **M** |
| Missing inventory service | **D**enial of Service | 502 on reserve | Service not deployed → returns 502 (documented) | **M** |

### 2.5 Inventory / Redis Stock Engine

| Threat | Type | Impact | Mitigation | Severity |
|--------|------|--------|------------|----------|
| Lua script injection (via ARGV) | **T**ampering | Arbitrary stock deduction | KEYS/ARGV are typed, all input validated before EVALSHA | **M** |
| Stock desync between Redis → MySQL | **T**ampering | Overselling | Asynchronous reconciliation job (planned) | **H** |
| Redis no-auth access | **I**nformation Disclosure | Session data leak | Redis on backend overlay, no external port; AUTH enabled | **M** |

### 2.6 Databases (MySQL, MongoDB)

| Threat | Type | Impact | Mitigation | Severity |
|--------|------|--------|------------|----------|
| Direct MySQL access from worker nodes | **I**nformation Disclosure | Data exfiltration | DB VLAN isolated from swarm; connection via ProxySQL | **M** |
| MySQL master failure | **D**enial of Service | Write unavailable | No automated failover; manual promotion of slave | **H** |
| ProxySQL single point of access | **D**enial of Service | All DB traffic blocked | ProxySQL is a global service, single instance (host mode) | **M** |
| MongoDB replica set no auth | **S**poofing | Unauthenticated reads | Dev only; prod requires X.509 + TLS | **M** |

### 2.7 Kafka Event Bus

| Threat | Type | Impact | Mitigation | Severity |
|--------|------|--------|------------|----------|
| PLAINTEXT listener (no SASL) | **T**ampering | Message injection | Dev only; prod requires SASL/SSL | **H** |
| Topic auto-creation | **T**ampering | Topic squatting | `auto.create.topics.enable=true` (dev); disable in prod | **M** |
| No message signing | **T**ampering | Message forgery | Internal trust boundary; future: message signing | **L** |

### 2.8 Next.js Storefront (6 replicas)

| Threat | Type | Impact | Mitigation | Severity |
|--------|------|--------|------------|----------|
| SSRF via fetch to internal services | **I**nformation Disclosure | Internal network scan | Gateway URL fixed via env var, no user-controlled fetch | **M** |
| RSC data leakage | **I**nformation Disclosure | Unauthorized product data | `revalidate: 60` safe; no user-specific data in RSC | **L** |
| XSS via product description | **T**ampering | Client-side injection | React 19 auto-escaping; CSP headers pending | **M** |

---

## 3. Risk Matrix

```
High     │ • JWT secret exposure        • SQL injection
         │ • Double spend / stock race   • Redis→MySQL desync
         │ • Auth gate missing (P0 fixed) • MySQL master SPOF
         │
Medium   │ • TLS downgrade              • Saga timeout inconsistency
         │ • XSS via product desc       • Brute force login
         │ • Redis no-auth (dev)        • Kafka PLAINTEXT
         │ • Lua script injection       • SSRF via fetch
         │
Low      │ • Traefik dashboard          • gRPC reflection
         │ • Topic auto-creation        • Message signing
         │
         └─────────────────────────────────────────────
           Likelihood →            
```

---

## 4. Recommended Hardening (Priority Order)

| Priority | Action | Component | Effort |
|----------|--------|-----------|--------|
| **P0** | Enable MySQL automated failover | MySQL | 2d |
| **P0** | Implement Redis→MySQL stock reconciliation job | Inventory | 3d |
| **P1** | Deploy inventory, payment, shipment services | Missing services | 5d |
| **P1** | Move JWT secrets to Vault / Docker secrets | All services | 1d |
| **P1** | Enable SASL/SSL for Kafka | Kafka | 2d |
| **P2** | Add CSP headers to Next.js responses | Storefront | 0.5d |
| **P2** | Rate limiting enable (`RATE_LIMIT_ENABLED=true`) | Gateway | 0.5d |
| **P2** | Circuit breaker enable | Gateway | 0.5d |
| **P3** | Implement request signing for service→service | All services | 5d |
| **P3** | MongoDB X.509 auth | MongoDB | 2d |
| **P3** | Harden Traefik dashboard (disable external) | Traefik | 0.5d |

---

## 5. Attack Tree: Checkout → Payment (Critical Path)

```
Checkout Flow Attack Surface
├── 1. Manipulate cart prices before checkout
│   ├── 1.1 Intercept cart API (XSS/mitm) → [Mitigated: HTTPS + JWT]
│   └── 1.2 Direct cart DB access → [Mitigated: internal network only]
├── 2. Race condition on stock reservation
│   ├── 2.1 Send 1000 concurrent requests → [Mitigated: Lua atomic DECRBY]
│   └── 2.2 Bypass user lock → [Mitigated: per-user lock key in Lua]
├── 3. Payment failure after stock reserved
│   ├── 3.1 Timeout → [Mitigated: compensation release.lua]
│   └── 3.2 Malicious cancellation → [Mitigated: idempotency + lock TTL]
└── 4. Order status tampering
    ├── 4.1 Direct order DB update → [Mitigated: Auth gate + internal net]
    └── 4.2 Kafka message injection → [Mitigated: PLAINTEXT → SASL (planned)]
```

---

## 6. Dev vs. Prod Gap Analysis

| Feature | Development | Production Required | Status |
|---------|-------------|-------------------|--------|
| TLS | Self-signed | Let's Encrypt / Paid CA | Done |
| JWT Secret | Hardcoded in stack | Vault/Docker Secrets | **Missing** |
| Kafka Auth | PLAINTEXT | SASL/SSL | **Missing** |
| Rate Limiting | Disabled | Enabled (per-IP + per-user) | **Missing** |
| Circuit Breaker | Disabled | Enabled with fallback | **Missing** |
| MySQL HA | Single master | ProxySQL + automated failover | **Missing** |
| Monitoring Alerts | Prometheus + Grafana | PagerDuty/opsgenie integration | Partial |
| Backup | None | Automated daily + point-in-time | **Missing** |
| Audit Logging | None | Structured audit trail | **Missing** |

---

*Generated: 2026-06-05 · STRIDE framework · Next review: after inventory/payment service deployment*
