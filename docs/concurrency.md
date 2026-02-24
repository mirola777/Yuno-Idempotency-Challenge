# Concurrency Strategy

## The Problem

In a ride-hailing payment system, network instability and client retries can cause two (or more) requests with the **same idempotency key** to arrive at the server simultaneously. Without proper synchronization, both requests could pass the "does this key exist?" check at the same time, leading to duplicate payments -- charging a rider twice for the same ride.

This is a classic race condition:

```
Request A: reads idempotency key --> not found
Request B: reads idempotency key --> not found  (A hasn't written yet)
Request A: creates payment, inserts record
Request B: creates payment, inserts record      (DUPLICATE)
```

---

## The Solution

This project uses **PostgreSQL row-level locking** via `SELECT ... FOR UPDATE` within a GORM database transaction. The lock is acquired at the database level, which means it works correctly across multiple application instances and survives process crashes.

The locking is implemented in the `FindByKeyForUpdate` repository method:

```go
func (r *IdempotencyRepo) FindByKeyForUpdate(ctx context.Context, tx *gorm.DB, key string) (*IdempotencyRecord, error) {
    var record domain.IdempotencyRecord
    err := tx.WithContext(ctx).
        Clauses(clause.Locking{Strength: "UPDATE"}).
        Where("key = ? AND expires_at > ?", key, time.Now()).
        First(&record).Error
    // ...
}
```

This translates to the SQL:

```sql
SELECT * FROM idempotency_records
WHERE key = $1 AND expires_at > NOW()
FOR UPDATE;
```

---

## Step-by-Step Flow

When a `POST /v1/payments` request arrives, the service executes the following sequence inside a single database transaction:

### Step 1: Begin Transaction

A GORM transaction is opened with `db.Transaction(func(tx *gorm.DB) error { ... })`. All subsequent database operations use this transaction handle.

### Step 2: SELECT ... FOR UPDATE

The service calls `FindByKeyForUpdate(ctx, tx, idempotencyKey)`. This query does two things:
- Looks up the idempotency record by key.
- Acquires a **row-level exclusive lock** if the record exists.

If the record does not exist, no lock is acquired, but the transaction still holds a gap lock that prevents concurrent inserts of the same key (depending on isolation level and index).

### Step 3: Evaluate the Record

**If the record exists and status is PROCESSING:**
Return a `409 Conflict` with code `PAYMENT_PROCESSING`. This means another request is actively processing this key.

**If the record exists and status is COMPLETED:**
Compare the request fingerprint (SHA-256 hash of the request body).
- Fingerprint matches: deserialize and return the cached response (idempotent replay).
- Fingerprint differs: return a `409 Conflict` with code `IDEMPOTENCY_KEY_CONFLICT`.

**If the record does not exist:**
Proceed to step 4.

### Step 4: Insert PROCESSING Record

Create a new `IdempotencyRecord` with status `PROCESSING`, the computed request fingerprint, and an expiration timestamp (default: 24 hours). This insert happens within the transaction.

### Step 5: Process the Payment

Call the payment processor (simulator) to create the payment. The processor returns a `Payment` object with a UUID, status, and card metadata.

### Step 6: Persist the Payment

Insert the `Payment` record into the database within the same transaction.

### Step 7: Update to COMPLETED

Update the idempotency record: set status to `COMPLETED`, attach the `payment_id`, and store the serialized payment response as `response_body` (JSONB).

### Step 8: Commit Transaction

The transaction commits, which:
- Makes the idempotency record and payment visible to other transactions.
- Releases the row-level lock.

---

## Concurrent Request Timeline

Below is a timeline showing how two simultaneous requests with the same idempotency key are handled:

```
Time    Request A                           Request B
----    ---------                           ---------
T0      BEGIN TRANSACTION
T1      SELECT ... FOR UPDATE
        (no record found)
T2      INSERT idempotency record
        (status = PROCESSING)
T3      Process payment...                  BEGIN TRANSACTION
T4      |                                   SELECT ... FOR UPDATE
        |                                   (BLOCKED - waiting for A's lock)
T5      Insert payment record               |
T6      Update idempotency record           |
        (status = COMPLETED)                |
T7      COMMIT TRANSACTION                  |
        (lock released)                     |
T8                                          (lock acquired)
                                            Record found: COMPLETED
T9                                          Fingerprint matches ->
                                            return cached response
T10                                         COMMIT TRANSACTION
```

**Key observations:**

- At T4, Request B attempts `SELECT ... FOR UPDATE` on the same key. Because Request A holds a lock on that row (or the transaction has an exclusive intent on the key space), Request B **blocks**.
- At T7, Request A commits, releasing the lock.
- At T8, Request B's query completes and finds the now-COMPLETED record.
- At T9, Request B compares fingerprints. If they match, it returns the cached response without creating a second payment.

---

## Why Not an In-Memory Mutex?

An application-level mutex (such as `sync.Mutex` in Go) would fail in several scenarios:

| Scenario                        | In-Memory Mutex | PostgreSQL FOR UPDATE |
|---------------------------------|-----------------|-----------------------|
| Multiple application instances  | Not synchronized| Synchronized via DB   |
| Process crash mid-transaction   | Lock lost       | Transaction rolled back, lock released |
| Horizontal scaling              | Requires distributed lock (Redis, etcd) | Works out of the box |
| Deadlock detection              | Manual          | PostgreSQL built-in deadlock detector |

PostgreSQL's `SELECT ... FOR UPDATE` provides **distributed mutual exclusion** without additional infrastructure. The database is already part of the architecture, so no external coordination service is needed.

---

## Fingerprint Comparison

To distinguish "same request retried" from "different request reusing a key," the system computes a SHA-256 hash of the entire request body:

```go
func Compute(req domain.PaymentRequest) string {
    data, _ := json.Marshal(req)
    hash := sha256.Sum256(data)
    return fmt.Sprintf("%x", hash)
}
```

This fingerprint is stored alongside the idempotency record. On subsequent requests:
- **Same fingerprint** = safe retry, return cached response.
- **Different fingerprint** = misuse, return 409 Conflict.

---

## Edge Cases

**Request arrives while another is PROCESSING:**
If Request A is still inside the transaction (has not committed), Request B will block at `SELECT ... FOR UPDATE`. If Request A takes too long and the database connection times out, PostgreSQL rolls back Request A's transaction and releases the lock.

**Expired idempotency records:**
The `WHERE expires_at > NOW()` clause in the query ensures that expired records are treated as if they do not exist. A background goroutine periodically cleans up expired records based on the `CLEANUP_INTERVAL` configuration.

**Crash recovery:**
If the application crashes between inserting a PROCESSING record and committing the transaction, PostgreSQL automatically rolls back the uncommitted transaction. The PROCESSING record is never persisted, so subsequent retries start fresh.
