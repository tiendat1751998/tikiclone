# Swarm Architecture Diagram

```mermaid
flowchart TB
    subgraph INTERNET["🌐 Internet"]
        client[Internet Users]
    end

    subgraph TRAEFIK["Traefik Load Balancer\n(Entry Point)"]
        lb[Traefik LB<br/>Port 80/443]
    end

    subgraph MANAGER["Manager Node\nControl Plane"]
        mgr[Manager API<br/>Orchestrator<br/>Scheduler]
    end

    subgraph W1["Worker Node 1\n(8GB RAM)"]
        w1_g1[gateway]
        w1_p1[product]
        w1_p2[product]
        w1_cart1[cart]
        w1_cart2[cart]
        w1_auth1[auth]
    end

    subgraph W2["Worker Node 2\n(8GB RAM)"]
        w2_g2[gateway]
        w2_p3[product]
        w2_p4[product]
        w2_p5[product]
        w2_cart3[cart]
        w2_cart4[cart]
        w2_search1[search]
        w2_cat1[category]
    end

    subgraph W3["Worker Node 3\n(8GB RAM)"]
        w3_g3[gateway]
        w3_g4[gateway]
        w3_g5[gateway]
        w3_p6[product]
        w3_p7[product]
        w3_cart5[cart]
        w3_auth2[auth]
        w3_auth3[auth]
        w3_search2[search]
        w3_search3[search]
        w3_cat2[category]
        w3_cat3[category]
        w3_check1[checkout]
        w3_check2[checkout]
        w3_check3[checkout]
        w3_ord1[order]
        w3_ord2[order]
        w3_ord3[order]
        w3_notif1[notification]
        w3_notif2[notification]
        w3_notif3[notification]
    end

    subgraph INFRA1["Infrastructure Node\n(Stateful Services)"]
        infra_mysql[(MySQL)]
        infra_redis[(Redis)]
        infra_mongodb[(MongoDB)]
        infra_kafka[(Kafka)]
    end

    subgraph OVERLAY_FRONTEND["Overlay Network: frontend\n(Gateway Ingress)"]
        front_desc[Gateway Services<br/>Total: 5 replicas]
    end

    subgraph OVERLAY_BACKEND["Overlay Network: backend\n(Service Communication)"]
        back_desc[Backend Services:<br/>Product(5) · Cart(5) · Auth(3)<br/>Search(3) · Category(3)<br/>Checkout(3) · Order(3) · Notification(3)]
    end

    subgraph OVERLAY_MONITORING["Overlay Network: monitoring\n(Observability)"]
        mon_desc[Monitoring<br/>Prometheus · Grafana<br/>Loki · cAdvisor]
    end

    client --> lb
    lb --> mgr

    mgr -- manages --> W1
    mgr -- manages --> W2
    mgr -- manages --> W3

    W1 --- w1_g1 & w1_p1 & w1_p2 & w1_cart1 & w1_cart2 & w1_auth1
    W2 --- w2_g2 & w2_p3 & w2_p4 & w2_p5 & w2_cart3 & w2_cart4 & w2_search1 & w2_cat1
    W3 --- w3_g3 & w3_g4 & w3_g5 & w3_p6 & w3_p7 & w3_cart5 & w3_auth2 & w3_auth3 & w3_search2 & w3_search3 & w3_cat2 & w3_cat3 & w3_check1 & w3_check2 & w3_check3 & w3_ord1 & w3_ord2 & w3_ord3 & w3_notif1 & w3_notif2 & w3_notif3
    W1 --- infra_mysql
    W2 --- infra_redis
    W3 --- infra_mongodb & infra_kafka

    style MANAGER fill:#fff3e0,stroke:#e65100
    style W1 fill:#e3f2fd,stroke:#1565c0
    style W2 fill:#e3f2fd,stroke:#1565c0
    style W3 fill:#e3f2fd,stroke:#1565c0
    style OVERLAY_FRONTEND fill:#e8f5e9,stroke:#2e7d32,stroke-dasharray:5 5
    style OVERLAY_BACKEND fill:#e8f5e9,stroke:#2e7d32,stroke-dasharray:5 5
    style OVERLAY_MONITORING fill:#e8f5e9,stroke:#2e7d32,stroke-dasharray:5 5
    style w1_mysql fill:#ffe0b2,stroke:#ef6c00
    style w2_redis fill:#ffe0b2,stroke:#ef6c00
    style w3_mongodb fill:#ffe0b2,stroke:#ef6c00
    style w3_kafka fill:#ffe0b2,stroke:#ef6c00
```