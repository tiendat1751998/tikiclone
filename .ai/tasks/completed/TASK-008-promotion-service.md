# TASK-008 — PROMOTION SERVICE

## Goal

Build a REAL production-grade Promotion Service.

This service is responsible for:
- voucher engine
- promotion engine
- campaign engine
- flash-sale campaigns
- pricing rules
- promotion eligibility
- seller promotions
- platform promotions
- stacking rules
- campaign scheduling
- promotion validation

This is NOT a toy CRUD promotion service.

The Promotion Service must support:
- millions of users
- massive flash-sale traffic
- distributed deployments
- Kubernetes-native deployment
- observability-first architecture
- eventual consistency
- high concurrency

The architecture MUST prioritize:
- low-latency validation
- pricing correctness
- abuse prevention
- resiliency
- operational scalability

---

## Tech Stack

Use:
- Golang
- Gin/Fiber
- gRPC
- Redis Cluster
- MySQL
- Kafka or NATS JetStream
- OpenTelemetry
- Prometheus
- Kubernetes
- Helm

Optional:
- Lua scripting for atomic rule validation

---

## Core Responsibilities

The Promotion Service MUST support:

### Voucher Engine
- voucher creation
- voucher redemption
- usage limits
- user limits
- seller vouchers
- platform vouchers
- voucher expiration

### Campaign Engine
- flash-sale campaigns
- scheduled campaigns
- seasonal campaigns
- seller campaigns
- category campaigns

### Pricing Rules
- percentage discounts
- fixed discounts
- shipping discounts
- minimum spend validation
- SKU-specific rules
- category rules

### Eligibility Engine
- user eligibility
- seller eligibility
- region eligibility
- payment eligibility
- campaign eligibility

### Stacking Rules
- promotion stacking
- mutually exclusive promotions
- priority rules
- conflict resolution

### Abuse Prevention
- duplicate redemption prevention
- rate limiting
- fraud detection hooks
- replay-safe redemption

---

## Architecture Requirements

The service MUST:
- follow clean architecture
- separate domain/application/infrastructure
- support distributed deployments
- support eventual consistency
- support event-driven workflows

The Promotion Service MUST:
- support horizontal scaling
- support async recalculation
- support distributed cache
- support low-latency validation

Use:
- CQRS where appropriate
- dependency injection
- modular architecture
- resilience patterns

The promotion system MUST tolerate:
- duplicate requests
- retry storms
- stale cache
- flash-sale bursts
- distributed deployments

---

## Folder Structure

Generate:

services/promotion/
├── cmd/
├── internal/
│   ├── config/
│   ├── domain/
│   ├── application/
│   ├── infrastructure/
│   ├── transport/
│   ├── middleware/
│   ├── voucher/
│   ├── campaigns/
│   ├── pricing/
│   ├── eligibility/
│   ├── stacking/
│   ├── validation/
│   ├── abuse/
│   ├── scheduling/
│   ├── cache/
│   ├── events/
│   ├── metrics/
│   ├── tracing/
│   ├── logging/
│   └── health/
│
├── migrations/
├── deployments/
├── charts/
├── tests/
├── configs/
└── Dockerfile

---

## Database Requirements

Use MySQL for:
- voucher metadata
- campaign metadata
- redemption history
- pricing rules
- stacking rules
- eligibility rules

Generate:
- optimized schemas
- indexes
- constraints
- audit tables

Requirements:
- pagination everywhere
- high read optimization
- transactional correctness
- replay safety

Never:
- use naive redemption logic
- ignore concurrency
- ignore indexing
- ignore replay protection

---

## Redis Requirements

Use Redis for:
- hot promotion cache
- redemption counters
- eligibility cache
- flash-sale campaigns
- distributed rate limiting

Generate:
- TTL strategy
- cache invalidation strategy
- atomic redemption operations
- retry handling
- stale cache handling

Support:
- distributed deployments
- flash-sale traffic
- high concurrency

The cache layer MUST be production-grade.

---

## Event-Driven Requirements

Generate events for:
- voucher created
- voucher redeemed
- campaign started
- campaign ended
- promotion updated
- pricing recalculated

Use:
- Kafka or NATS JetStream

Requirements:
- retries
- DLQ
- idempotent consumers
- replay-safe processing
- event versioning
- consumer groups

Support:
- async propagation
- eventual consistency
- distributed recalculation

No fake async architecture.

---

## Flash Sale Requirements

Support:
- massive burst traffic
- flash-sale campaigns
- high-concurrency redemption
- async recalculation
- distributed rate limiting

Generate:
- flash-sale validation strategy
- Redis burst handling
- redemption throttling
- anti-thundering-herd protection

The flash-sale architecture MUST resemble:
- Tiki
- Lazada
- TikTok Shop

No naive redemption logic.

---

## Pricing Rule Requirements

Support:
- dynamic pricing rules
- campaign pricing
- SKU pricing
- seller pricing
- shipping discounts
- conditional discounts

Generate:
- pricing evaluation flow
- rule priority system
- rule conflict resolution
- async recalculation strategy

The service MUST NOT own checkout execution.

---

## Eligibility Requirements

Support:
- user targeting
- regional targeting
- payment targeting
- seller targeting
- product targeting

Generate:
- eligibility evaluation engine
- distributed cache strategy
- replay-safe validation

---

## Abuse Prevention Requirements

Implement:
- duplicate redemption prevention
- replay-safe redemption
- distributed rate limiting
- suspicious redemption detection
- idempotency validation

Support:
- retry storms
- duplicate requests
- concurrent redemption attempts

---

## API Requirements

Generate:
- REST APIs
- gRPC APIs
- OpenAPI specs
- proto files

Endpoints:
- /vouchers
- /campaigns
- /promotions
- /eligibility
- /pricing-preview
- /redeem

Support:
- pagination
- filtering
- validation
- idempotency keys

---

## Security Requirements

The service MUST:
- validate ownership
- enforce RBAC
- sanitize input
- validate redemption requests
- prevent abuse

Never:
- trust client pricing
- trust client eligibility
- expose internal rule engine metadata

Generate:
- authorization middleware
- idempotency validation
- anti-abuse middleware

---

## Observability Requirements

Generate:
- OpenTelemetry tracing
- Prometheus metrics
- structured logging
- distributed tracing
- correlation IDs

Metrics:
- redemption latency
- validation latency
- cache hit ratio
- Redis latency
- campaign traffic
- rule evaluation latency
- abuse detection count

Logs:
- JSON structured logs
- trace IDs
- correlation IDs

Never log sensitive redemption metadata.

---

## Reliability Requirements

Implement:
- retries
- timeout handling
- graceful shutdown
- panic recovery
- circuit breakers
- backoff strategies

Support:
- rolling deployment
- autoscaling
- distributed deployments

Generate:
- resilience middleware
- retry policies
- failure isolation

---

## Kubernetes Requirements

Generate:
- Deployment
- Service
- ConfigMap
- Secret integration
- HPA
- PodDisruptionBudget
- ServiceMonitor
- NetworkPolicy
- Helm chart

Support:
- readiness/liveness probes
- autoscaling
- rolling deployment
- canary deployment

---

## CI/CD Requirements

Generate:
- GitHub Actions or Drone pipeline
- linting
- testing
- vulnerability scanning
- Docker build
- Helm validation

---

## Testing Requirements

Generate:
- unit tests
- integration tests
- concurrency tests
- redemption tests
- replay tests
- cache tests
- abuse prevention tests
- flash-sale tests

Test:
- duplicate redemption
- retry storms
- flash-sale spikes
- replay attacks
- stale cache
- distributed validation
- high concurrency

---

## Output Requirements

Explain:
- promotion architecture
- voucher engine strategy
- flash-sale strategy
- pricing evaluation flow
- stacking rule strategy
- eligibility engine
- abuse prevention strategy
- cache strategy
- scaling strategy
- resilience strategy

Generate production-grade code only.

No toy voucher CRUD service.
No fake pricing engine.
No fake async architecture.

---

## Acceptance Criteria

The Promotion Service must support future integration with:
- Cart Service
- Checkout Service
- Order Service
- Payment Service
- Recommendation System

without major future refactors.

The service MUST realistically tolerate:
- flash-sale traffic
- duplicate redemption
- retry storms
- stale cache
- distributed deployments

---

## Constraints

Follow ALL:
- .ai/system/*
- .ai/architecture/*
- .ai/planning/*
- .ai/context/*
- .ai/prompts/*

Production-grade only.
If a real production system would require operational complexity,
YOU MUST model that complexity realistically instead of simplifying it away.