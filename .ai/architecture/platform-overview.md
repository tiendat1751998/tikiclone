# System Topology - Tiki Clone Enterprise

This system leverages high-concurrency microservices designed to scale independently.

```mermaid
graph TD
    Client[Next.js Client] -->|HTTPS/WSS| WAF[Cloudflare CDN & WAF]
    WAF -->|DNS Routing| Gateway[Kong API Gateway]
    
    subgraph Core Clusters
        Gateway -->|JWT Validate| Auth[Identity Service: Spring Boot]
        Gateway -->|mTLS| Catalog[Catalog Service: Go]
        Gateway -->|mTLS| Cart[Cart Service: Go]
        Gateway -->|mTLS| Order[Order Service: Spring Boot]
        Gateway -->|mTLS| Inventory[Inventory Service: Go]
        Gateway -->|mTLS| Payment[Payment Service: Spring Boot]
        Gateway -->|REST| Search[Search Service: Go]
    end

    subgraph Messaging & Storage
        Order -->|1. Write Outbox| DB_Order[(PostgreSQL)]
        DB_Order -->|CDC via Debezium| Kafka[Kafka Event Bus]
        Kafka -->|Async Consumer| SearchSync[Search Indexer]
        SearchSync --> DB_Search[(Elasticsearch)]
        
        Inventory -->|Lua stock reservation| Redis[Redis Cluster]
        Catalog --> DB_Catalog[(MongoDB)]
        Payment --> DB_Pay[(PostgreSQL)]
    end
```

## Transaction Lifecycle (Search to Order Complete)
1. **Browse**: User searches products via `/api/v1/search` resolved instantly via **Elasticsearch** (read-optimized replica).
2. **Add to Cart**: `/api/v1/cart` checks variations and locks active items in **Redis memory store** (latency < 2ms).
3. **Checkout Submit**:
   - `/api/v1/checkout` coordinates with **Inventory** using high-concurrency **Redis Lua Stock reservation**.
   - If stock is locked successfully, Order transitions to `PENDING_PAYMENT` and triggers Payment transaction.
   - Payment webhooks complete, raising `payment.completed` event to finalize order shipping.
