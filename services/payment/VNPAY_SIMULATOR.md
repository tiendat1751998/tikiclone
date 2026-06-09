# VNPay Payment Simulator

A payment simulator for testing VNPay integration without requiring actual VNPay credentials.

## Overview

The VNPay simulator mimics the VNPay payment gateway behavior for development and testing purposes. It implements the core VNPay API endpoints as documented in the VNPay technical specification v2.1.0.

## Architecture Flow

```
[Web Frontend] -> [Payment Service] -> [VNPay Simulator] -> [Payment Service Callback] -> [Web Frontend]
       |                |                    |                    |
       | POST /vnpay/create              GET/POST /simulator    GET /vnpay/callback
       | (creates payment)                /vnpay               (processes result)
       |                |                    |
       |                +---- Payment DB ---->
       |                |
       +<-- Payment URL --+
```

## API Flow

### Step 1: Create Payment Request (Frontend -> Payment Service)
```
POST /vnpay/create
Authorization: Bearer <token>
Content-Type: application/json

{
  "order_id": "order-123",
  "amount": 1000000,
  "client_ip": "192.168.1.1",
  "return_url": "https://example.com/checkout/return"
}
```

Response:
```json
{
  "payment_id": "payment-uuid",
  "vnpay_params": {
    "vnp_Version": "2.1.0",
    "vnp_Command": "pay",
    "vnp_TmnCode": "VNPAYDEMO",
    "vnp_Amount": 1000000,
    "vnp_CurrCode": "VND",
    "vnp_TxnRef": "order-123",
    "vnp_OrderInfo": "Order order-123",
    "vnp_ReturnUrl": "https://example.com/checkout/return",
    "vnp_IpAddr": "192.168.1.1",
    "vnp_CreateDate": "20260608120000"
  }
}
```

### Step 2: Redirect to Payment Page (Frontend)
Frontend redirects user to `/simulator/vnpay` with `vnpay_params` as query parameters. The simulator shows a page where user can click "Simulate Success" or "Simulate Failed".

### Step 3: Callback Handler (VNPay -> Payment Service)
After payment processing, VNPay redirects to the return URL, which then calls back to our system:

```
GET /vnpay/callback?vnp_TxnRef=order-123&vnp_ResponseCode=00&vnp_TransactionStatus=00&vnp_TransactionNo=123456
```

This updates the payment status in the database.

## Endpoints

All endpoints are available under `/simulator/vnpay` when running in development/staging mode.

### Payment Flow

1. **Create Payment Request**
   - `GET /simulator/vnpay?...` - Displays payment page with success/failed buttons
   - Parameters: `vnp_Version`, `vnp_Command`, `vnp_TmnCode`, `vnp_Amount`, `vnp_CurrCode`, `vnp_TxnRef`, `vnp_OrderInfo`, `vnp_ReturnUrl`, `vnp_IpAddr`, `vnp_CreateDate`

2. **Process Payment**
   - `POST /simulator/vnpay/process`
   - Parameters: `vnp_TxnRef`, `success` (true/false)
   - Returns: JSON response with transaction result

### API Endpoints

#### Query Transaction
```
POST /simulator/vnpay/query
{
  "vnp_RequestId": "unique-request-id",
  "vnp_Version": "2.1.0",
  "vnp_Command": "querydr",
  "vnp_TmnCode": "TESTMERCHANT",
  "vnp_TxnRef": "your-txn-ref",
  "vnp_CreateDate": "20260608120000",
  "vnp_IpAddr": "127.0.0.1",
  "vnp_OrderInfo": "Query transaction"
}
```

#### Refund Transaction
```
POST /simulator/vnpay/refund
{
  "vnp_RequestId": "unique-request-id",
  "vnp_Version": "2.1.0",
  "vnp_Command": "refund",
  "vnp_TmnCode": "TESTMERCHANT",
  "vnp_TransactionType": "02",
  "vnp_TxnRef": "your-txn-ref",
  "vnp_Amount": 1000000,
  "vnp_OrderInfo": "Refund reason",
  "vnp_CreateBy": "admin@example.com",
  "vnp_CreateDate": "20260608120000",
  "vnp_IpAddr": "127.0.0.1"
}
```

#### Token Operations

##### Create Token
```
POST /simulator/vnpay/token/create
{
  "vnp_Version": "2.1.0",
  "vnp_Command": "token_create",
  "vnp_TmnCode": "TESTMERCHANT",
  "vnp_AppUserId": "user-123",
  "vnp_BankCode": "VIB",
  "vnp_CardType": "01",
  "vnp_TxnRef": "token-ref-123",
  "vnp_TxnDesc": "Create token",
  "vnp_CreateDate": "20260608120000",
  "vnp_IpAddr": "127.0.0.1"
}
```

##### Pay with Token
```
POST /simulator/vnpay/token/pay
{
  "vnp_Version": "2.1.0",
  "vnp_Command": "token_pay",
  "vnp_TmnCode": "TESTMERCHANT",
  "vnp_AppUserId": "user-123",
  "vnp_Amount": 500000,
  "vnp_CurrCode": "VND",
  "vnp_TxnRef": "pay-ref-123",
  "vnp_TxnDesc": "Payment with token",
  "vnp_CreateDate": "20260608120000",
  "vnp_IpAddr": "127.0.0.1"
}
```

##### Remove Token
```
POST /simulator/vnpay/token/remove
{
  "vnp_Version": "2.1.0",
  "vnp_Command": "token_remove",
  "vnp_TmnCode": "TESTMERCHANT",
  "vnp_AppUserId": "user-123",
  "vnp_TxnRef": "remove-ref-123",
  "vnp_CreateDate": "20260608120000",
  "vnp_IpAddr": "127.0.0.1"
}
```

## Configuration

Environment variables:
- `VNPAY_TMNCODE` - Merchant terminal code (default: VNPAYDEMO)
- `VNPAY_HASH_SECRET` - Secret key for HMAC-SHA512 signatures (default: demo-secret-key)

## Response Codes

| Code | Description |
|------|-------------|
| 00 | Success |
| 05 | Insufficient balance |
| 06 | Invalid OTP |
| 07 | Suspected fraud |
| 09 | Card not registered for InternetBanking |
| 10 | Incorrect card/account info (3 times) |
| 11 | Payment expired |
| 12 | Card/account locked |
| 24 | User cancelled transaction |
| 79 | Too many authentication attempts |
| 65 | Daily transaction limit exceeded |
| 97 | Invalid signature |
| 91 | Transaction not found |
| 99 | Other errors |

## Security

The simulator generates HMAC-SHA512 hashes for all responses to match VNPay's signature verification:

```go
signature = HMAC-SHA512(sorted_params + secret_key)
```

All params are sorted alphabetically before hashing.