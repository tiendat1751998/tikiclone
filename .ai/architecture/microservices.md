# Core gRPC Interfaces (Protobuf Specifications)

Internal microservices communicate using high-performance, strictly-typed gRPC interfaces.

## 1. Inventory Service Definition (`inventory.proto`)
```protobuf
syntax = "proto3";

package tiki.inventory;
option go_package = "services/inventory/pb";

service InventoryService {
  rpc ReserveStock (ReserveStockRequest) returns (ReserveStockResponse);
  rpc ReleaseStockCompensate (ReleaseStockRequest) returns (ReleaseStockResponse);
}

message ReserveStockRequest {
  string order_id = 1;
  string sku_id = 2;
  int32 quantity = 3;
}

message ReserveStockResponse {
  bool success = 1;
  string message = 2;
  int64 reserved_at = 3;
}

message ReleaseStockRequest {
  string order_id = 1;
  string sku_id = 2;
  int32 quantity = 3;
}
```

## 2. Order Service Definition (`order.proto`)
```protobuf
syntax = "proto3";

package tiki.order;
option java_package = "com.tiki.order.grpc";

service OrderService {
  rpc GetOrderStatus (GetOrderStatusRequest) returns (GetOrderStatusResponse);
}

message GetOrderStatusRequest {
  string order_id = 1;
}

message GetOrderStatusResponse {
  string order_id = 1;
  string status = 2;
  int64 created_at = 3;
  int64 updated_at = 4;
}
```
