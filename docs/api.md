# API Reference

Base URL: `http://localhost:8080`

All endpoints return JSON. Error responses follow a consistent structure:

```json
{
  "code": "ERROR_CODE",
  "messages": ["human-readable error description"]
}
```

A `X-Trace-Id` header is returned on every response. If the client sends an `X-Trace-Id` header, it is echoed back; otherwise, the server generates a UUID v4.

---

## POST /v1/payments

Create a new payment. This endpoint is idempotency-protected.

### Headers

| Header              | Required | Description                                          |
|---------------------|----------|------------------------------------------------------|
| `X-Idempotency-Key` | Yes      | Unique key for idempotent requests. Max 64 characters. |
| `Content-Type`      | Yes      | Must be `application/json`.                          |

### Request Body

| Field         | Type   | Required | Description                                    |
|---------------|--------|----------|------------------------------------------------|
| `amount`      | float  | Yes      | Payment amount. Must be greater than 0.        |
| `currency`    | string | Yes      | ISO currency code. Supported: IDR, THB, VND, PHP. |
| `customer_id` | string | Yes      | Identifier for the customer.                   |
| `ride_id`     | string | Yes      | Identifier for the ride.                       |
| `card_number` | string | Yes      | Card number. Only the last 4 digits are stored.|
| `description` | string | No       | Optional payment description.                  |

**Example request body:**

```json
{
  "amount": 150000,
  "currency": "IDR",
  "customer_id": "cust_abc123",
  "ride_id": "ride_xyz789",
  "card_number": "4242424242424242",
  "description": "Ride from Airport to Downtown"
}
```

### Response 201 Created

Returned when the payment is successfully created, or when a duplicate request with the same idempotency key and identical payload is received (cached response).

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "amount": 150000,
  "currency": "IDR",
  "customer_id": "cust_abc123",
  "ride_id": "ride_xyz789",
  "status": "SUCCEEDED",
  "card_last_4": "4242",
  "description": "Ride from Airport to Downtown",
  "created_at": "2026-02-24T10:30:00Z"
}
```

Possible `status` values: `SUCCEEDED`, `FAILED`, `PENDING`.

When `status` is `FAILED`, an additional `fail_reason` field is included (e.g., `"insufficient_funds"`, `"expired_card"`, `"processing_error"`).

### Error Responses

**400 Bad Request -- Missing idempotency key:**

```json
{
  "code": "IDEMPOTENCY_KEY_MISSING",
  "messages": ["X-Idempotency-Key header is required"]
}
```

**400 Bad Request -- Key too long:**

```json
{
  "code": "IDEMPOTENCY_KEY_TOO_LONG",
  "messages": ["X-Idempotency-Key must be at most 64 characters"]
}
```

**400 Bad Request -- Invalid request body:**

```json
{
  "code": "INVALID_PAYMENT_REQUEST",
  "messages": [
    "amount must be greater than 0",
    "customer_id is required"
  ]
}
```

**400 Bad Request -- Unsupported currency:**

```json
{
  "code": "INVALID_CURRENCY",
  "messages": ["currency 'EUR' is not supported; valid currencies: IDR, THB, VND, PHP"]
}
```

**409 Conflict -- Same key, different payload:**

```json
{
  "code": "IDEMPOTENCY_KEY_CONFLICT",
  "messages": ["idempotency key 'my-key-123' already used with different request payload"]
}
```

**409 Conflict -- Payment currently processing:**

```json
{
  "code": "PAYMENT_PROCESSING",
  "messages": ["a payment with this idempotency key is currently being processed"]
}
```

### curl Example

```bash
curl -X POST http://localhost:8080/v1/payments \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: ride-payment-xyz789-001" \
  -d '{
    "amount": 150000,
    "currency": "IDR",
    "customer_id": "cust_abc123",
    "ride_id": "ride_xyz789",
    "card_number": "4242424242424242",
    "description": "Ride from Airport to Downtown"
  }'
```

---

## GET /v1/payments/:id

Retrieve a payment by its ID.

### Path Parameters

| Parameter | Description                              |
|-----------|------------------------------------------|
| `id`      | The UUID of the payment to retrieve.     |

### Response 200 OK

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "amount": 150000,
  "currency": "IDR",
  "customer_id": "cust_abc123",
  "ride_id": "ride_xyz789",
  "status": "SUCCEEDED",
  "card_last_4": "4242",
  "description": "Ride from Airport to Downtown",
  "created_at": "2026-02-24T10:30:00Z"
}
```

### Error Responses

**404 Not Found:**

```json
{
  "code": "PAYMENT_NOT_FOUND",
  "messages": ["payment 'nonexistent-id' not found"]
}
```

### curl Example

```bash
curl http://localhost:8080/v1/payments/a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

---

## GET /v1/idempotency/:key

Look up an idempotency record by its key. Useful for debugging and inspecting the state of a previous request.

### Path Parameters

| Parameter | Description                                  |
|-----------|----------------------------------------------|
| `key`     | The idempotency key to look up.              |

### Response 200 OK

```json
{
  "key": "ride-payment-xyz789-001",
  "request_fingerprint": "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
  "payment_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "COMPLETED",
  "created_at": "2026-02-24T10:30:00Z",
  "expires_at": "2026-02-25T10:30:00Z"
}
```

Possible `status` values: `PROCESSING`, `COMPLETED`.

### Error Responses

**404 Not Found:**

```json
{
  "code": "IDEMPOTENCY_KEY_NOT_FOUND",
  "messages": ["idempotency key 'nonexistent-key' not found"]
}
```

### curl Example

```bash
curl http://localhost:8080/v1/idempotency/ride-payment-xyz789-001
```

---

## GET /health

Health check endpoint. Returns a simple status to confirm the server is running.

### Response 200 OK

```json
{
  "status": "ok"
}
```

### curl Example

```bash
curl http://localhost:8080/health
```

---

## Test Card Numbers

The payment processor simulator uses specific card numbers to produce deterministic outcomes:

| Card Number        | Result     | Fail Reason       |
|--------------------|------------|-------------------|
| `4242424242424242`  | SUCCEEDED  | --                |
| `4000000000000002`  | FAILED     | insufficient_funds|
| `4000000000000069`  | FAILED     | expired_card      |
| `4000000000000119`  | FAILED     | processing_error  |
| `4000000000000259`  | PENDING    | --                |
| Any other number    | SUCCEEDED  | --                |
