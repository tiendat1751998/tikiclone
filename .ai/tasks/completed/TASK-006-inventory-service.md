# TASK-006 — INVENTORY SERVICE

## Goal

Build a REAL production-grade Inventory Service.

This service is responsible for:
- stock management
- inventory reservation
- anti-oversell protection
- warehouse inventory
- inventory synchronization
- stock movement tracking
- flash-sale inventory handling
- distributed stock consistency
- inventory locking
- stock allocation

This is NOT a toy CRUD inventory service.

The service must support:
- massive concurrency
- flash-sale traffic spikes
- distributed deployments
- Kubernetes-native deployment
- fault tolerance
- eventual consistency
- observability-first architecture

The architecture MUST prioritize:
- anti-oversell guarantees
- concurrency correctness
- resiliency
- operational stability

---

## Tech Stack

Use:
- Golang
- Gin/Fiber
- gRPC
- MySQL
- Redis Cluster
- Kafka or NATS JetStream
- OpenTelemetry
- Prometheus
- Kubernetes
- Helm

Optional:
- Lua scripting for Redis atomic operations

---

## Core Responsibilities

The Inventory Service MUST support:

### Stock Management
- increase stock
- decrease stock
- stock adjustment
- warehouse stock
- available stock
- reserved stock

### Reservation System
- reserve stock
- release reservation
- reservation expiration
- reservation confirmation
- reservation rollback

### Anti-Oversell Protection
- distributed locking
- atomic stock deduction
- reservation-first workflow
- idempotent operations
- concurrency-safe updates

### Warehouse Management
- multiple warehouses
- warehouse allocation
- warehouse priority
- warehouse stock sync

### Flash Sale Handling
- high concurrency inventory
- stock pre-deduction
- burst traffic handling
- anti-thundering-herd protection

### Inventory Movement
- stock movement logs
- inventory audit trail
- inventory event history

---

## Architecture Requirements

The service MUST:
- follow clean architecture
- separate domain/application/infrastructure
- support distributed deployments
- support event-driven workflows
- support eventual consistency

The service MUST:
- support horizontal scaling
- support concurrency correctness
- support distributed locking
- support resilience

Use:
- CQRS where appropriate
- dependency injection
- modular architecture
- resilience patterns

The inventory system MUST be:
- replay-safe
- idempotent
- concurrency-safe

---

## Folder Structure

Generate:

services/inventory/
├── cmd/
├── internal/
│   ├── config/
│   ├── domain/
│   ├── application/
│   ├── infrastructure/
│   ├── transport/
│   ├── middleware/
│   ├── stock/
│   ├── reservation/
│   ├── warehouse/
│   ├── locking/
│   ├── movement/
│   ├── flashsale/
│   ├── consistency/
│   ├── events/
│   ├── cache/
│   ├── metrics/
│   ├── tracing/
│   ├── logging/
│   ├── validation/
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
- stock records
- reservations
- warehouse mappings
- movement logs
- audit history

Generate:
- optimized schemas
- indexes
- constraints
- transactional flows

Requirements:
- optimistic locking where appropriate
- transactional integrity
- high-write optimization
- pagination everywhere

Support:
- eventual consistency
- replay safety
- inventory reconciliation

Never:
- use naive stock updates
- use SELECT *
- ignore locking correctness
- ignore transaction isolation

---

## Redis Requirements

Use Redis for:
- distributed locking
- reservation caching
- flash-sale stock handling
- hot inventory cache
- rate limiting

Generate:
- Lua atomic operations
- retry handling
- TTL strategy
- lock expiration handling
- stale lock recovery

Support:
- distributed deployments
- lock contention handling
- retry storms
- stale reservation cleanup

The locking system MUST be production-grade.

No fake distributed locks.

---

## Event-Driven Requirements

Generate events for:
- stock reserved
- stock released
- stock deducted
- reservation expired
- warehouse updated
- inventory adjusted

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
- eventual consistency
- distributed inventory synchronization

The async system MUST be realistic.

---

## Reservation Requirements

Implement:
- reservation lifecycle
- expiration handling
- rollback support
- retry safety
- idempotent reservation

Support:
- cart reservation
- checkout reservation
- order confirmation
- cancellation rollback

Generate:
- reservation state machine
- timeout handling
- cleanup workers

The reservation system MUST tolerate:
- retries
- duplicate requests
- partial failures

---

## Flash Sale Requirements

Support:
- massive burst traffic
- anti-thundering herd
- stock pre-deduction
- queue-based processing
- traffic smoothing

Generate:
- flash-sale inventory strategy
- Redis burst handling
- async stock synchronization
- rate limiting strategy

The flash-sale architecture MUST resemble:
- Tiki
- Lazada
- TikTok Shop

No naive locking approaches.

---

## API Requirements

Generate:
- REST APIs
- gRPC APIs
- OpenAPI specs
- proto files

Endpoints:
- /inventory
- /stock
- /reservations
- /warehouses
- /flashsale

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
- validate stock operations
- sanitize input
- prevent unauthorized adjustments

Never:
- trust client stock values
- expose internal locking metadata
- expose warehouse internals improperly

Generate:
- authorization middleware
- idempotency validation
- operation validation

---

## Observability Requirements

Generate:
- OpenTelemetry tracing
- Prometheus metrics
- structured logging
- distributed tracing
- correlation IDs

Metrics:
- reservation latency
- stock deduction latency
- lock contention
- Redis latency
- DB latency
- event retry count
- oversell prevention metrics

Logs:
- JSON structured logs
- trace IDs
- correlation IDs

Never log sensitive inventory internals.

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
- retry policies
- resilience middleware
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
- reservation tests
- locking tests
- Redis tests
- replay tests
- flash-sale tests

Test:
- oversell prevention
- duplicate requests
- retry storms
- reservation expiration
- lock contention
- node failures
- high concurrency

---

## Output Requirements

Explain:
- inventory architecture
- anti-oversell strategy
- distributed locking strategy
- reservation workflow
- flash-sale strategy
- Redis locking flow
- event-driven consistency
- scaling strategy
- resilience strategy
- Kubernetes deployment strategy

Generate production-grade code only.

No toy inventory service.
No fake distributed locks.
No naive stock deduction.

---

## Acceptance Criteria

The Inventory Service must support future integration with:
- Product Service
- Cart Service
- Checkout Service
- Order Service
- Flash Sale System

without major future refactors.

The service MUST realistically tolerate:
- flash-sale spikes
- retry storms
- duplicate events
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