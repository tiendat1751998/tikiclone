# Tiki API Gateway

Production-grade API gateway for the Tiki Clone platform, built with Go and Gin.

## Architecture

```
Client → Cloudflare/SSL → NGINX Ingress → Gateway (:8080/:9090) → Upstream Services
                                                │
                          ┌─────────────────────┼─────────────────────┐
                          │                     │                     │
                     Redis Cluster        OTel Collector       Upstreams
                  (rate-limit/session/    (traces/metrics)    (auth, cart,
                   token-blacklist/cache)                      order, ...)
```

The gateway is **stateless**, horizontally scalable, and K8s-native.

## Request Lifecycle

1. **TLS Termination** — NGINX Ingress terminates TLS, forwards to gateway
2. **Correlation ID** — Generate or propagate X-Correlation-ID, X-Request-ID
3. **Security Headers** — Apply HSTS, CSP, XFO, etc.
4. **OpenTelemetry** — Start trace span, propagate W3C Trace Context
5. **Request Logger** — Log method, path, status, latency, trace_id
6. **Sanitizer** — Strip XSS/SQLi from query params and headers
7. **Body Size Limit** — Reject requests exceeding MaxBodySize
8. **CORS** — Validate Origin, set Allow-Origin headers
9. **Device Metadata** — Extract X-Device-ID, User-Agent, platform
10. **Anti-Abuse** — Validate User-Agent, Content-Type, query length
11. **Request Validation** — Reject path traversal, TRACE, missing Host
12. **Global Rate Limit** — Redis sliding-window per IP
13. **IP Rate Limit** — Per-IP rate limiting
14. **Auth** — JWT validation via JWKS, RBAC enforcement
15. **Endpoint Rate Limit** — Per-route rate limit (login: 5/s, checkout: 1/s, etc.)
16. **Reverse Proxy** — Circuit breaker → Retry with backoff → Upstream call

## Middleware Flow

```
Global (applied to all routes):
  Recovery → ErrorHandler → CorrelationID → SecurityHeaders → OTelMiddleware
  → RequestLogger → RequestSanitizer → BodySizeLimiter → CORS → DeviceMetadata
  → AntiAbuse → RequestValidation → GlobalRateLimit → IPRateLimit → Auth

Per-route:
  EndpointRateLimit → Auth (if required) → AuthenticatedRateLimit → Proxy
```

## Tracing Flow

```
Gateway receives request → OTel Middleware starts span → Proxy creates child span
→ Upstream call with trace context → Response received → Span ends → Exported to OTel Collector
→ Collector forwards traces to Tempo, metrics to Prometheus
```

## Scaling Strategy

- **Horizontal**: Stateless design, HPA targets 70% CPU / 80% memory
- **Autoscaling**: 3–20 replicas based on load
- **PDB**: Min 2 available during rolling updates
- **Anti-affinity**: Spread pods across nodes
- **Canary**: Separate deployment with track=canary label for gradual rollouts

## Rate Limiting Strategy

Multi-layered sliding-window rate limiting via Redis (`github.com/go-redis/redis_rate`):

| Layer | Key | Default Rate |
|-------|-----|-------------|
| Global | `global:<client_ip>` | 10,000/s |
| Per-IP | `ip:<client_ip>` | 50/s |
| Per-User | `user:<user_id>` | 200/s |
| Per-Endpoint | `endpoint:<method>:<path>` | Route-specific |

**Fail-closed**: If Redis is down, rate limiter rejects requests (no bypass).

## Redis Strategy

| Purpose | Key Pattern | TTL |
|---------|------------|-----|
| Rate Limiting | `redis_rate:*` | Sliding window |
| Session Data | `session:<session_id>` | Per-session |
| Token Blacklist | `token:blacklist:<sha256>` | 24h |
| Refresh Tokens | `refresh:<user_id>:<session_id>` | 7 days |
| JWKS Cache | In-memory LRU | 1 hour |

## Resilience Strategy

### Circuit Breaker (per service)
- **Closed** → Normal operation
- **Open** → Fail fast when failure ratio > 60% (min 5 samples)
- **Half-Open** → After 30s timeout, allow 5 probes
- Metrics exposed as `tiki_gateway_circuit_breaker_state{service="..."}`

### Retry with Exponential Backoff
- Max 3 attempts (configurable via `UPSTREAM_MAX_RETRIES`)
- Initial interval: 50ms, multiplier: 2x, max: 5s
- Jitter: ±10% to prevent thundering herd
- Only retries transient errors (connection refused, timeout, reset, etc.)

### Timeouts (per service)
| Service | Timeout |
|---------|---------|
| auth | 10s |
| catalog | 15s |
| cart | 5s |
| order | 30s |
| inventory | 5s |
| payment | 30s |
| search | 10s |

## Kubernetes Topology

```
Namespace: tiki-platform
Deployment: 3 replicas (canary: 1)
Service: ClusterIP (:80 → :8080, :9090)
HPA: 3–20 pods @ 70% CPU / 80% memory
PDB: minAvailable: 2
NetworkPolicy: Restrictive ingress/egress
ServiceMonitor: Prometheus scrape every 15s
ConfigMap: Upstream addresses, Redis, OTel
Secrets: Redis password, JWT secrets (via External Secrets)
```

## Security

- **JWT Validation**: JWKS with key rotation, algorithm confusion prevention
- **Token Blacklist**: Fail-closed Redis check before parsing
- **CORS**: Explicit origin whitelist, never wildcard with credentials
- **Headers**: HSTS, XFO, CSP, COOP, COEP, CORP
- **Sanitization**: HTML escape, SQLi/XSS regex blocking
- **CSRF**: Origin/Referer check for state-changing methods
- **Rate Limiting**: Multi-layer sliding window, fail-closed
- **RBAC**: Role-based access control on routes
- **Secrets**: Never hardcoded, pulled from K8s secrets/External Secrets
