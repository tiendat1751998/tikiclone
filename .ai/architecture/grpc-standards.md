# gRPC Standards & mTLS Specifications

Internal service-to-service communication requires maximum performance and security.

## 1. Protobuf Styling & Validation
- **Package Names**: Follow `tiki.service_name.v1` convention.
- **Validation Annotations**: Use `protoc-gen-validate` (PGV) annotations to enforce parameters checking at compilation/gRPC receiver layer:
  ```protobuf
  import "validate/validate.proto";

  message CreateCartItemRequest {
    string sku_id = 1 [(validate.rules).string.min_len = 8];
    int32 quantity = 2 [(validate.rules).int32.gt = 0];
  }
  ```

## 2. Connection Management & mTLS
- **Keep-Alives**: Configure HTTP/2 ping keep-alives every **30 seconds** to prevent cloud firewalls from abruptly terminating idle connections.
- **mTLS Encryptions**: Enforce TLS 1.3 between all service pods. Gateway Kong terminates HTTPS and acts as the secure mTLS proxy.
