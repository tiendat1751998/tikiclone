# Admin Panel API Contract

Document version: 1.0.0
Last updated: 2025-05-27

This document defines all REST API endpoints that the admin-panel Go backend must expose for the Next.js admin frontend to function.

---

## Authentication

All admin endpoints (except login/refresh) require a Bearer token in the Authorization header.

```
Authorization: Bearer <access_token>
```

The frontend automatically attaches `X-Correlation-ID` to every request for distributed tracing.

---

## Error Response Format

All error responses follow this schema:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error description",
    "details": {} // optional additional context
  },
  "correlation_id": "admin-1234567890-abc123"
}
```

### Common HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 204 | No Content (successful delete) |
| 400 | Bad Request (validation error) |
| 401 | Unauthorized (missing/invalid token) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Not Found |
| 409 | Conflict (duplicate resource) |
| 422 | Unprocessable Entity (business logic error) |
| 429 | Too Many Requests (rate limited) |
| 500 | Internal Server Error |

### Rate Limiting

Rate-limited responses include these headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1716789600
Retry-After: 60
```

---

## Pagination Format

All list endpoints use cursor-based or offset pagination:

### Request Parameters (query string)

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| page | integer | 1 | Page number (1-indexed) |
| per_page | integer | 20 | Items per page (max 100) |
| sort_by | string | created_at | Field to sort by |
| sort_order | string | desc | asc or desc |

### Response Format

```json
{
  "success": true,
  "data": {
    "items": [...],
    "pagination": {
      "current_page": 1,
      "per_page": 20,
      "total_items": 150,
      "total_pages": 8
    }
  },
  "correlation_id": "admin-1234567890-abc123"
}
```

---

## 1. Authentication Endpoints

### POST /api/v1/auth/admin/login

Authenticate an admin user.

**Request:**
```json
{
  "email": "admin@tiki.vn",
  "password": "securepassword123",
  "totp_code": "123456"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "admin@tiki.vn",
      "username": "admin",
      "display_name": "Admin User",
      "phone": "+84123456789",
      "avatar_url": "https://...",
      "role": "super_admin",
      "status": "active",
      "created_at": "2025-01-01T00:00:00Z"
    },
    "tokens": {
      "access_token": "eyJ...",
      "refresh_token": "eyJ...",
      "expires_in": 3600,
      "token_type": "Bearer"
    }
  }
}
```

### POST /api/v1/auth/refresh

Refresh an access token.

**Request:**
```json
{
  "refresh_token": "eyJ..."
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJ...",
    "refresh_token": "eyJ...",
    "expires_in": 3600,
    "token_type": "Bearer"
  }
}
```

### POST /api/v1/auth/logout

Invalidate the current session.

**Request:** (no body, uses Authorization header)

**Response (204):** No content

### GET /api/v1/auth/me

Get current authenticated user info.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "email": "admin@tiki.vn",
    "username": "admin",
    "display_name": "Admin User",
    "role": "super_admin",
    "permissions": ["products.read", "products.write", "orders.read", "orders.write", "users.read", "users.write"]
  }
}
```

---

## 2. Dashboard Endpoints

### GET /api/v1/admin/dashboard/stats

Get overview statistics for the dashboard.

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| period | string | 30d | Time period: 7d, 30d, 90d, 1y |

**Response (200):**
```json
{
  "success": true,
  "data": {
    "total_revenue": 1500000000,
    "revenue_change": 12.5,
    "total_orders": 1250,
    "orders_change": 8.3,
    "total_users": 5420,
    "users_change": 15.2,
    "conversion_rate": 3.45,
    "conversion_change": -0.5,
    "order_status": [
      { "status": "pending", "count": 45 },
      { "status": "confirmed", "count": 120 },
      { "status": "processing", "count": 85 },
      { "status": "shipped", "count": 230 },
      { "status": "delivered", "count": 720 },
      { "status": "cancelled", "count": 35 },
      { "status": "refunded", "count": 15 }
    ],
    "top_products": [
      {
        "id": "uuid",
        "name": "iPhone 15 Pro Max",
        "sales": 450,
        "revenue": 13500000000
      }
    ],
    "recent_orders": [
      {
        "id": "uuid",
        "order_number": "TK-12345",
        "customer": "Nguyen Van A",
        "total": 25000000,
        "status": "pending"
      }
    ]
  }
}
```

### GET /api/v1/admin/dashboard/revenue

Get revenue trend data.

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| days | integer | 30 | Number of days of data |

**Response (200):**
```json
{
  "success": true,
  "data": {
    "revenue": [
      {
        "date": "2025-05-01",
        "revenue": 50000000,
        "orders": 45
      }
    ],
    "new_users": [
      {
        "date": "2025-05-01",
        "users": 120
      }
    ]
  }
}
```

---

## 3. Product Endpoints

### GET /api/v1/admin/products

List products with filtering and pagination.

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| page | integer | Page number |
| per_page | integer | Items per page |
| sort_by | string | Field to sort by |
| sort_order | string | asc or desc |
| search | string | Search by name, SKU |
| category_id | string | Filter by category UUID |
| brand_id | string | Filter by brand UUID |
| status | string | draft, published, archived, all |
| min_price | number | Minimum price filter |
| max_price | number | Maximum price filter |
| low_stock | boolean | Show only low stock items |

**Response (200):**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "uuid",
        "name": "iPhone 15 Pro Max",
        "slug": "iphone-15-pro-max",
        "sku": "IP15PM-256-BLK",
        "price": 34990000,
        "sale_price": 32990000,
        "quantity": 150,
        "reserved_quantity": 12,
        "available_quantity": 138,
        "low_stock_threshold": 20,
        "category_id": "uuid",
        "category_id": "Electronics",
        "brand": "Apple",
        "brand_id": "uuid",
        "status": "published",
        "thumbnail_url": "https://...",
        "created_at": "2025-01-01T00:00:00Z",
        "updated_at": "2025-05-20T00:00:00Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "per_page": 20,
      "total_items": 150,
      "total_pages": 8
    }
  }
}
```

### GET /api/v1/admin/products/stats

Get product statistics.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "total": 150,
    "published": 120,
    "draft": 25,
    "archived": 5,
    "low_stock": 8,
    "out_of_stock": 3
  }
}
```

### GET /api/v1/admin/products/:id

Get a single product by ID.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "iPhone 15 Pro Max",
    "slug": "iphone-15-pro-max",
    "description": "Latest iPhone...",
    "sku": "IP15PM-256-BLK",
    "price": 34990000,
    "sale_price": 32990000,
    "quantity": 150,
    "reserved_quantity": 12,
    "low_stock_threshold": 20,
    "category_id": "uuid",
    "brand_id": "uuid",
    "images": ["https://...", "https://..."],
    "status": "published",
    "attributes": {
      "color": "Black",
      "storage": "256GB"
    },
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-05-20T00:00:00Z"
  }
}
```

### POST /api/v1/admin/products

Create a new product.

**Request:**
```json
{
  "name": "iPhone 15 Pro Max",
  "slug": "iphone-15-pro-max",
  "description": "Latest iPhone...",
  "sku": "IP15PM-256-BLK",
  "price": 34990000,
  "sale_price": 32990000,
  "quantity": 150,
  "low_stock_threshold": 20,
  "category_id": "uuid",
  "brand_id": "uuid",
  "images": ["https://..."],
  "status": "draft",
  "attributes": {
    "color": "Black",
    "storage": "256GB"
  }
}
```

**Response (201):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    // ... full product object
  }
}
```

### PUT /api/v1/admin/products/:id

Update an existing product.

**Request:** (same as POST, all fields optional)

**Response (200):**
```json
{
  "success": true,
  "data": {
    // ... updated product object
  }
}
```

### DELETE /api/v1/admin/products/:id

Delete a product (soft delete, sets status to archived).

**Response (204):** No content

### POST /api/v1/admin/products/bulk-import

Bulk import products from CSV.

**Request:** multipart/form-data
- file: CSV file

**Response (200):**
```json
{
  "success": true,
  "data": {
    "imported": 50,
    "failed": 2,
    "errors": [
      { "row": 5, "error": "Invalid price format" }
    ]
  }
}
```

### GET /api/v1/admin/products/export

Export products to CSV.

**Query Parameters:** (same as list endpoint)

**Response (200):** CSV file download

---

## 4. Category Endpoints

### GET /api/v1/admin/categories

List categories.

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| page | integer | Page number |
| per_page | integer | Items per page |
| search | string | Search by name |
| status | string | active, inactive, all |
| parent_id | string | Filter by parent (empty for root) |

**Response (200):**
```json
{
  "success": true,
  "data": {
    "categories": [
      {
        "id": "uuid",
        "name": "Electronics",
        "slug": "electronics",
        "parent_id": null,
        "parent_name": null,
        "product_count": 150,
        "is_active": true,
        "sort_order": 1,
        "created_at": "2025-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "per_page": 20,
      "total_items": 25,
      "total_pages": 2
    }
  }
}
```

### POST /api/v1/admin/categories

Create a new category.

**Request:**
```json
{
  "name": "Smartphones",
  "slug": "smartphones",
  "parent_id": "uuid",
  "is_active": true,
  "sort_order": 2
}
```

**Response (201):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    // ... full category object
  }
}
```

### PUT /api/v1/admin/categories/:id

Update a category.

### DELETE /api/v1/admin/categories/:id

Delete a category.

---

## 5. Inventory Endpoints

### GET /api/v1/admin/inventory

List inventory items.

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| page | integer | Page number |
| per_page | integer | Items per page |
| search | string | Search by product name or SKU |
| status | string | in_stock, low_stock, out_of_stock, all |
| warehouse | string | Filter by warehouse |
| low_stock | boolean | Show only low stock items |

**Response (200):**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "uuid",
        "product_id": "uuid",
        "product_name": "iPhone 15 Pro Max",
        "product_sku": "IP15PM-256-BLK",
        "quantity": 150,
        "reserved_quantity": 12,
        "available_quantity": 138,
        "low_stock_threshold": 20,
        "warehouse": "Hanoi Warehouse",
        "last_restocked": "2025-05-15T00:00:00Z",
        "status": "in_stock"
      }
    ],
    "pagination": {
      "current_page": 1,
      "per_page": 20,
      "total_items": 200,
      "total_pages": 10
    }
  }
}
```

### GET /api/v1/admin/inventory/stats

Get inventory statistics.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "total_sku": 200,
    "in_stock": 170,
    "low_stock": 15,
    "out_of_stock": 15,
    "total_value": 50000000000
  }
}
```

### POST /api/v1/admin/inventory/adjust

Adjust inventory quantity.

**Request:**
```json
{
  "product_id": "uuid",
  "quantity": 100,
  "reason": "Restocked from supplier",
  "warehouse": "Hanoi Warehouse"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "product_id": "uuid",
    "previous_quantity": 50,
    "new_quantity": 150,
    "adjustment": 100,
    "reason": "Restocked from supplier",
    "adjusted_by": "admin@tiki.vn",
    "adjusted_at": "2025-05-27T00:00:00Z"
  }
}
```

### POST /api/v1/admin/inventory/transfer

Transfer inventory between warehouses.

**Request:**
```json
{
  "product_id": "uuid",
  "from_warehouse": "Hanoi Warehouse",
  "to_warehouse": "HCMC Warehouse",
  "quantity": 50,
  "reason": "Stock rebalancing"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "transfer_id": "uuid",
    "status": "completed"
  }
}
```

---

## 6. Order Endpoints

### GET /api/v1/admin/orders

List orders with filtering.

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| page | integer | Page number |
| per_page | integer | Items per page |
| sort_by | string | Field to sort by |
| sort_order | string | asc or desc |
| search | string | Search by order number, customer name, email |
| status | string | pending, confirmed, processing, shipped, delivered, cancelled, refunded, all |
| date_from | string | Start date (ISO 8601) |
| date_to | string | End date (ISO 8601) |
| min_amount | number | Minimum order total |
| max_amount | number | Maximum order total |

**Response (200):**
```json
{
  "success": true,
  "data": {
    "orders": [
      {
        "id": "uuid",
        "order_number": "TK-12345",
        "customer_id": "uuid",
        "customer_name": "Nguyen Van A",
        "customer_email": "nguyenvana@email.com",
        "total": 25000000,
        "status": "pending",
        "item_count": 3,
        "payment_status": "paid",
        "payment_method": "cod",
        "created_at": "2025-05-27T00:00:00Z",
        "updated_at": "2025-05-27T00:00:00Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "per_page": 20,
      "total_items": 500,
      "total_pages": 25
    }
  }
}
```

### GET /api/v1/admin/orders/:id

Get order details.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "order_number": "TK-12345",
    "status": "pending",
    "customer": {
      "id": "uuid",
      "name": "Nguyen Van A",
      "email": "nguyenvana@email.com",
      "phone": "+84123456789"
    },
    "shipping_address": {
      "street": "123 Nguyen Hue",
      "ward": "Ben Nghe",
      "district": "District 1",
      "city": "Ho Chi Minh City",
      "phone": "+84123456789"
    },
    "billing_address": { },
    "items": [
      {
        "id": "uuid",
        "product_id": "uuid",
        "product_name": "iPhone 15 Pro Max",
        "product_image": "https://...",
        "sku": "IP15PM-256-BLK",
        "quantity": 1,
        "price": 34990000,
        "total": 34990000
      }
    ],
    "subtotal": 34990000,
    "shipping_fee": 30000,
    "discount": 0,
    "tax": 0,
    "total": 35020000,
    "payment_method": "cod",
    "payment_status": "paid",
    "payment_transaction_id": "txn_xxx",
    "timeline": [
      {
        "status": "pending",
        "timestamp": "2025-05-27T00:00:00Z",
        "note": "Order placed"
      },
      {
        "status": "confirmed",
        "timestamp": "2025-05-27T00:05:00Z",
        "note": "Order confirmed"
      }
    ],
    "notes": [],
    "created_at": "2025-05-27T00:00:00Z",
    "updated_at": "2025-05-27T00:05:00Z"
  }
}
```

### PATCH /api/v1/admin/orders/:id/status

Update order status.

**Request:**
```json
{
  "status": "confirmed",
  "note": "Payment verified"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "status": "confirmed",
    "updated_at": "2025-05-27T00:05:00Z"
  }
}
```

### POST /api/v1/admin/orders/:id/refund

Process a refund for an order.

**Request:**
```json
{
  "amount": 25000000,
  "reason": "Customer request",
  "refund_method": "original"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "refund_id": "uuid",
    "amount": 25000000,
    "status": "processing"
  }
}
```

### POST /api/v1/admin/orders/:id/cancel

Cancel an order.

**Request:**
```json
{
  "reason": "Out of stock"
}
```

**Response (200):** Updated order object

### GET /api/v1/admin/orders/export

Export orders to CSV.

**Query Parameters:** (same as list endpoint)

**Response (200):** CSV file download

---

## 7. User Endpoints

### GET /api/v1/admin/users

List users with filtering.

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| page | integer | Page number |
| per_page | integer | Items per page |
| search | string | Search by name, email, phone |
| status | string | active, banned, pending, all |
| role | string | customer, vip, wholesale, all |
| date_from | string | Registration date from |
| date_to | string | Registration date to |

**Response (200):**
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "uuid",
        "name": "Nguyen Van A",
        "email": "nguyenvana@email.com",
        "phone": "+84123456789",
        "avatar_url": "https://...",
        "role": "customer",
        "status": "active",
        "order_count": 15,
        "total_spent": 250000000,
        "created_at": "2025-01-01T00:00:00Z",
        "last_login": "2025-05-27T00:00:00Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "per_page": 20,
      "total_items": 5000,
      "total_pages": 250
    }
  }
}
```

### GET /api/v1/admin/users/:id

Get user details.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Nguyen Van A",
    "email": "nguyenvana@email.com",
    "phone": "+84123456789",
    "avatar_url": "https://...",
    "role": "customer",
    "status": "active",
    "email_verified": true,
    "phone_verified": true,
    "two_factor_enabled": false,
    "address": {
      "street": "123 Nguyen Hue",
      "ward": "Ben Nghe",
      "district": "District 1",
      "city": "Ho Chi Minh City"
    },
    "stats": {
      "total_orders": 15,
      "total_spent": 250000000,
      "avg_order_value": 16666666,
      "last_order_date": "2025-05-25T00:00:00Z",
      "member_since": "2025-01-01T00:00:00Z"
    },
    "recent_orders": [
      {
        "id": "uuid",
        "order_number": "TK-12345",
        "total": 25000000,
        "status": "delivered",
        "date": "2025-05-25T00:00:00Z"
      }
    ],
    "activity_log": [
      {
        "id": "uuid",
        "action": "login",
        "description": "Logged in from Chrome on Ubuntu",
        "ip_address": "192.168.1.1",
        "user_agent": "Mozilla/5.0...",
        "timestamp": "2025-05-27T00:00:00Z"
      }
    ]
  }
}
```

### PATCH /api/v1/admin/users/:id/status

Update user status (ban/unban).

**Request:**
```json
{
  "status": "banned",
  "reason": "Violation of terms"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "status": "banned",
    "updated_at": "2025-05-27T00:00:00Z"
  }
}
```

### GET /api/v1/admin/users/:id/orders

Get order history for a user.

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| page | integer | Page number |
| per_page | integer | Items per page |

**Response (200):** Same format as orders list

---

## 8. Analytics Endpoints

### GET /api/v1/admin/analytics

Get detailed analytics data.

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| period | string | 30d | 7d, 30d, 90d, 1y |

**Response (200):**
```json
{
  "success": true,
  "data": {
    "revenue": {
      "total": 1500000000,
      "growth": 12.5,
      "trend": [
        { "date": "2025-05-01", "value": 50000000 }
      ]
    },
    "orders": {
      "total": 1250,
      "growth": 8.3,
      "by_status": [
        { "status": "delivered", "count": 720 },
        { "status": "pending", "count": 45 }
      ]
    },
    "customers": {
      "new": 450,
      "returning": 800,
      "growth": 15.2,
      "acquisition": [
        { "type": "new", "count": 450 },
        { "type": "returning", "count": 800 }
      ]
    },
    "aov": {
      "value": 1200000,
      "growth": -2.1
    },
    "categories": {
      "top": [
        { "name": "Electronics", "sales": 450 }
      ]
    },
    "funnel": [
      { "stage": "visitors", "count": 100000 },
      { "stage": "product_views", "count": 50000 },
      { "stage": "add_to_cart", "count": 10000 },
      { "stage": "checkout", "count": 3000 },
      { "stage": "purchase", "count": 1250 }
    ]
  }
}
```

---

## 9. Settings Endpoints

### GET /api/v1/admin/settings

Get admin panel settings.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "store": {
      "name": "Tiki Clone",
      "url": "https://tiki.vn",
      "email": "admin@tiki.vn",
      "currency": "VND",
      "timezone": "Asia/Ho_Chi_Minh"
    },
    "orders": {
      "prefix": "TK",
      "low_stock_threshold": 10,
      "auto_cancel_unpaid_hours": 48
    },
    "notifications": {
      "new_order": { "email": true, "push": true },
      "low_stock": { "email": true, "push": false }
    },
    "security": {
      "two_factor_required": false,
      "session_timeout_minutes": 60
    }
  }
}
```

### PUT /api/v1/admin/settings

Update admin panel settings.

**Request:**
```json
{
  "store": {
    "name": "Tiki Clone Updated"
  },
  "orders": {
    "low_stock_threshold": 15
  }
}
```

**Response (200):** Updated settings object

---

## 10. File Upload Endpoints

### POST /api/v1/admin/upload/image

Upload a product image.

**Request:** multipart/form-data
- file: image file (jpg, png, webp, max 5MB)
- type: "product" | "category" | "brand"

**Response (200):**
```json
{
  "success": true,
  "data": {
    "url": "https://cdn.tiki.vn/images/products/xxx.jpg",
    "thumbnail_url": "https://cdn.tiki.vn/images/products/thumb_xxx.jpg",
    "width": 1200,
    "height": 1200,
    "size_bytes": 245760
  }
}
```

### POST /api/v1/admin/upload/images

Upload multiple images.

**Request:** multipart/form-data
- files[]: multiple image files

**Response (200):**
```json
{
  "success": true,
  "data": {
    "images": [
      { "url": "https://...", "thumbnail_url": "https://..." }
    ]
  }
}
```

---

## Data Types Reference

### Money Amounts

All monetary values are represented as integers in the smallest currency unit (Vietnamese Dong).
- Example: 25,000 VND = `25000`
- No decimal places needed for VND

### Timestamps

All timestamps use ISO 8601 format:
- `2025-05-27T10:30:00Z` (UTC)
- `2025-05-27T17:30:00+07:00` (with timezone)

### UUID Format

All IDs use UUID v4 format:
- `550e8400-e29b-41d4-a716-446655440000`

### Status Enums

**Product Status:** `draft` | `published` | `archived`

**Order Status:** `pending` | `confirmed` | `processing` | `shipped` | `delivered` | `cancelled` | `refunded`

**User Status:** `active` | `banned` | `pending`

**Inventory Status:** `in_stock` | `low_stock` | `out_of_stock` | `discontinued`

**Payment Status:** `pending` | `paid` | `failed` | `refunded`

**Admin Roles:** `super_admin` | `product_manager` | `order_manager` | `viewer`

---

## WebSocket Events (Optional)

For real-time notifications, the backend may expose WebSocket connections:

**Endpoint:** `/api/v1/admin/ws`

**Events emitted to client:**

| Event | Payload |
|-------|---------|
| `order.new` | `{ order_id, order_number, total }` |
| `order.status_changed` | `{ order_id, old_status, new_status }` |
| `inventory.low_stock` | `{ product_id, product_name, quantity }` |
| `user.registered` | `{ user_id, email }` |

---

## Environment Variables

The admin-panel Go backend should read these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| ADMIN_PORT | 3001 | Server port |
| API_GATEWAY_URL | http://localhost:8080 | Gateway URL |
| JWT_SECRET | (required) | HMAC secret for JWT |
| CORS_ALLOWED_ORIGINS | http://localhost:3001 | CORS origins |
| RATE_LIMIT_PER_MINUTE | 100 | API rate limit |
| UPLOAD_MAX_SIZE_MB | 5 | Max file upload size |
| CDN_BASE_URL | (optional) | CDN base URL for uploads |

---

## Implementation Notes

1. **Authentication:** The admin-panel BFF proxies auth requests to the auth service. It should validate the JWT signature using the shared secret.

2. **Service Communication:** Internal calls to catalog-product, cart, checkout, etc. should use gRPC for performance. The BFF transforms gRPC responses to REST JSON.

3. **Caching:** Dashboard stats can be cached for 60 seconds. Product listings for 30 seconds. Use Redis for cache storage.

4. **File Storage:** Product images should be stored in object storage (S3/MinIO) and served via CDN.

5. **Audit Logging:** All write operations (POST, PUT, DELETE) should emit audit log events to Kafka for the audit trail.

6. **Health Checks:** Expose `/health` and `/ready` endpoints for Kubernetes probes.

7. **Metrics:** Expose Prometheus metrics at `/metrics` endpoint.
