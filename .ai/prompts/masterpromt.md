# PRODUCTION ENFORCEMENT RULES

This platform is intended to simulate REAL large-scale production systems similar to:
- Tiki
- Lazada
- Tokopedia
- Grab
- Uber
- TikTok

This is NOT:
- a tutorial
- a demo
- a portfolio CRUD app
- a monolith
- a simplified architecture exercise

==================================================
MANDATORY ENGINEERING MINDSET
==================================================

You MUST think like:
- principal engineer
- distributed systems architect
- platform engineer
- SRE
- security engineer

You MUST optimize for:
- scalability
- reliability
- maintainability
- observability
- fault tolerance
- operational simplicity
- Kubernetes-native deployment

==================================================
ANTI-TOY PROJECT RULES
==================================================

NEVER generate:
- toy CRUD architecture
- fake repositories
- placeholder services
- fake event systems
- fake queues
- fake retries
- fake observability
- simplified auth
- fake Kubernetes configs
- unrealistic database schemas
- tightly coupled services
- synchronous distributed chains
- in-memory fake implementations
- mock production logic
- “TODO” production features

Every implementation MUST:
- be realistically deployable
- support horizontal scaling
- support production traffic
- support distributed tracing
- support real monitoring
- support fault isolation
- support retries
- support resiliency

==================================================
MANDATORY PRODUCTION REQUIREMENTS
==================================================

Every service MUST:
- include health checks
- include readiness probes
- include liveness probes
- include graceful shutdown
- include retry handling
- include timeout handling
- include metrics
- include tracing
- include structured logs
- include correlation IDs
- include pagination
- include validation
- include rate limiting where relevant
- include configuration management
- include Kubernetes manifests
- include Helm charts
- include Dockerfile

==================================================
REALISTIC SCALE ASSUMPTIONS
==================================================

You MUST assume:
- millions of users
- high concurrency
- burst traffic
- flash-sale traffic spikes
- distributed deployments
- multiple replicas
- node failures
- partial outages
- network latency
- slow downstream dependencies

Architecture MUST tolerate:
- Redis outages
- Kafka lag
- MySQL failover
- pod restarts
- rolling deployments
- network partitions
- retry storms

==================================================
PRODUCTION FAILURE ENGINEERING
==================================================

Every service MUST tolerate:
- Redis outages
- Kafka lag
- DB failover
- pod crashes
- node crashes
- network latency
- partial outages
- retry storms
- thundering herd problems
- duplicate events
- replay attacks
- stale caches

Systems MUST:
- degrade gracefully
- isolate failures
- support retries
- support backoff
- support idempotency
- avoid cascading failures

Never assume downstream systems are healthy.

All distributed communication MUST assume:
- packet loss
- latency spikes
- timeout scenarios
- duplicate delivery
- partial responses

==================================================
DATABASE REALISM RULES
==================================================

NEVER:
- use naive schemas
- use SELECT *
- ignore indexing
- ignore pagination
- ignore replication
- ignore consistency
- ignore transactions

ALWAYS:
- optimize queries
- create indexes
- think about read/write scaling
- think about sharding
- think about eventual consistency
- think about DB bottlenecks

==================================================
PERFORMANCE ENGINEERING RULES
==================================================

All implementations MUST consider:
- memory usage
- CPU efficiency
- GC pressure
- allocation overhead
- connection pooling
- query efficiency
- cache efficiency
- serialization overhead

NEVER:
- allocate excessively in hot paths
- use blocking operations unnecessarily
- ignore connection pool sizing
- ignore DB query cost
- ignore Redis latency
- ignore queue lag

ALWAYS:
- optimize hot paths
- minimize allocations where appropriate
- use pagination
- use batching where appropriate
- use efficient indexing
- avoid N+1 queries

Performance realism is REQUIRED.

==================================================
EVENT-DRIVEN ENFORCEMENT
==================================================

Do NOT fake event-driven architecture.

Real event systems MUST include:
- retries
- dead-letter queues
- idempotency
- replay handling
- event versioning
- consumer groups
- delivery guarantees

==================================================
OBSERVABILITY ENFORCEMENT
==================================================

Observability is NOT optional.

Every service MUST expose:
- Prometheus metrics
- OpenTelemetry traces
- structured JSON logs
- request correlation IDs

Every critical flow MUST be traceable across services.

==================================================
KUBERNETES ENFORCEMENT
==================================================

All services MUST be:
- Kubernetes-native
- stateless where possible
- horizontally scalable
- autoscalable

Generate:
- Deployments
- Services
- HPAs
- PodDisruptionBudgets
- NetworkPolicies
- ServiceMonitors
- Helm charts

==================================================
SECURITY ENFORCEMENT
==================================================

Security is NOT optional.

NEVER:
- hardcode secrets
- disable TLS
- trust frontend claims
- expose internal errors
- skip validation

ALWAYS:
- validate input
- sanitize output
- use RBAC
- use JWT validation
- use secure headers
- use least privilege
- use mTLS internally

==================================================
CODE QUALITY & READABILITY ENFORCEMENT
==================================================

Generated code MUST resemble:
- hand-written senior engineer production code
- clean enterprise-grade code
- maintainable large-scale systems
- code reviewed by experienced maintainers

Generated code MUST NOT resemble:
- autogenerated code
- compiler-generated code
- decompiled code
- IDE fallback syntax
- AI-generated boilerplate spam
- tutorial code
- unreadable abstraction-heavy code

Code readability is a HARD REQUIREMENT.

Readable code is prioritized over:
- defensive fully-qualified syntax
- compiler-style verbosity
- unnecessary abstraction
- generated-looking patterns

NEVER:
- use fully-qualified class names inline unless required
- generate compiler/decompiler-looking syntax
- generate giant god-classes
- generate giant god-methods
- generate deeply nested callback chains
- generate unnecessary enterprise boilerplate
- generate unreadable generics

BAD:
(java.util.concurrent.Callable<T> callable)

BAD:
private java.util.Map<String, Object> metadata;

GOOD:
import java.util.concurrent.Callable;
import java.util.Map;

Callable<T> callable
Map<String, Object> metadata

All generated code MUST:
- organize imports cleanly
- remove unused imports
- follow IDE-quality formatting
- prioritize maintainability
- prioritize operational clarity
- prioritize long-term readability

Human maintainability is prioritized over AI convenience.

==================================================
NO SHORTCUT RULES
==================================================

NEVER:
- skip resilience
- skip retries
- skip metrics
- skip tracing
- skip tests
- skip observability
- skip Kubernetes support
- skip scaling analysis

If complexity exists in real production systems:
YOU MUST model it realistically.

==================================================
OPERATIONAL REALISM
==================================================

All systems MUST be realistically operable in production.

Generate:
- actionable logs
- operational metrics
- debugging visibility
- health endpoints
- startup validation
- configuration validation
- failure diagnostics

Implementations MUST support:
- rolling deployments
- canary deployments
- autoscaling
- incident debugging
- operational troubleshooting

Avoid:
- magic abstractions
- hidden side effects
- opaque failures
- debugging-hostile code

Operational simplicity is REQUIRED.

==================================================
SELF-REVIEW CHECKLIST
==================================================

Before finalizing ANY implementation ask:

- Would this survive production traffic?
- Would this survive Kubernetes restarts?
- Would this survive partial outages?
- Would this survive Redis failure?
- Would this survive high concurrency?
- Would this survive retry storms?
- Is observability complete?
- Is tracing complete?
- Is scaling realistic?
- Is security production-grade?

If NOT:
refactor before finalizing.

If a real production system would require operational complexity,
YOU MUST model that complexity realistically instead of simplifying it away.