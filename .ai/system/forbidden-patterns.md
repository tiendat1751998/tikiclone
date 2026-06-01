# Forbidden Patterns: Anti-Patterns vs Standard Patterns

Avoid these architectural pitfalls. Follow the standardized solutions instead.

## 1. SQL Injection Vectors
### 🚫 Forbidden (Anti-Pattern)
```go
// Direct query concatenation allows severe SQL injection
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", emailInput)
db.Raw(query).Scan(&user)
```
### ✅ Standard (Best Practice)
```go
// Parameterized queries are safe and pre-compiled
db.Raw("SELECT * FROM users WHERE email = ?", emailInput).Scan(&user)
```

## 2. Direct Cross-Service Database Access
### 🚫 Forbidden (Anti-Pattern)
```java
// Inside OrderService, calling Product database directly
@Autowired
private MongoTemplate productMongoTemplate; // Accessing catalog DB from Order service
```
### ✅ Standard (Best Practice)
```java
// Communicate via gRPC client or consume replicated Kafka events
@GrpcClient("catalog-service")
private ProductServiceGrpc.ProductServiceBlockingStub productService;
```

## 3. Storage of Cryptographic Keys in Configuration
### 🚫 Forbidden (Anti-Pattern)
```yaml
jwt:
  secret: "my-super-secret-tiki-key-12345-that-is-hardcoded"
```
### ✅ Standard (Best Practice)
```yaml
jwt:
  secret: ${JWT_SIGNING_KEY} # Passed securely via Kubernetes Secrets or HashiCorp Vault
```

NEVER:
- use fully-qualified class names inline
- generate IDE fallback syntax
- generate unreadable generic signatures
- generate compiler-style code
- generate decompiled-looking code