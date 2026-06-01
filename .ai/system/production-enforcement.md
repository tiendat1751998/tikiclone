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