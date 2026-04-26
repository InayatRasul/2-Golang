# Practice 9: Part 2 - Advanced Retry Mechanisms & Idempotency

This directory contains implementations for Part 2 of Practice 9, demonstrating production-grade retry mechanisms and idempotency patterns in Go.

## Overview

### Part 2 includes two main tasks:

1. **Task 1: Payment Processing with Fault-Tolerant Retry Mechanism**
   - Demonstrates exponential backoff with jitter
   - Filters retryable vs non-retryable errors
   - Implements context-aware cancellation
   
2. **Task 2: Idempotency Middleware for Loan Repayment**
   - Prevents duplicate processing with idempotency keys
   - Handles concurrent duplicate requests  
   - Returns cached results for completed operations

---

## Task 1: Payment Processing with Retry Mechanism

### File: `task1_payment_retry.go`

#### Key Components

**IsRetryable Function**
```go
func IsRetryable(resp *http.Response, err error) bool
```
- Returns `true` for temporary errors (network timeouts, 429, 500, 502, 503, 504)
- Returns `false` for permanent errors (401 Unauthorized, 404 Not Found)
- Classifies all other 4xx errors as non-retryable (client errors)

**CalculateBackoff Function**
```go
func (pc *PaymentClient) CalculateBackoff(attempt int) time.Duration
```
- Implements exponential backoff: `baseDelay × 2^attempt`
- Caps maximum delay at `maxDelay`
- Adds full jitter: random value between 0 and calculated backoff
- Distributes load across time to prevent thundering herd

**ExecutePayment Method**
```go
func (pc *PaymentClient) ExecutePayment(ctx context.Context, url string, body []byte) (*http.Response, error)
```
- Respects context deadlines and cancellations
- Retries only retriable errors up to `maxRetries`
- Supports graceful shutdown via context

#### Configuration
```
MaxRetries: 5
BaseDelay: 500ms
MaxDelay: 5s
Global Timeout: 10s (via context)
```

#### Expected Output Example
```
Attempt 1 failed: status 503
  Waiting 370ms before next retry...
Attempt 2 failed: status 503
  Waiting 838ms before next retry...
Attempt 3 failed: status 503
  Waiting 82ms before next retry...
Attempt 4: Success! (Status 200)
Response body: {"status": "success", "transaction_id": "txn_12345", "amount": 1000}
```

#### Running Task 1
```bash
go run task1_payment_retry.go
go run -race task1_payment_retry.go  # With race condition detector
```

---

## Task 2: Idempotency Middleware for Loan Repayment

### Files
- `task2_idempotency.go` - Basic concurrent request scenario
- `task2_idempotency_advanced.go` - Advanced scenario with phases

#### Key Components

**IdempotencyStore Structure**
- Thread-safe map using `sync.RWMutex`
- Tracks request status: "processing" or "completed"
- Stores HTTP response status and body

**Request States**
1. **New Request**: First occurrence of idempotency key
   - Marked as "processing"
   - Handler executes business logic
   - Result stored upon completion

2. **Concurrent Duplicate**: Key still processing
   - Returns HTTP 409 Conflict
   - Prevents duplicate execution

3. **Cached Response**: Key already completed
   - Returns HTTP 200 OK with cached response
   - No handler execution
   - Header `X-Cached: true` indicates cached response

**IdempotencyMiddleware**
```go
func IdempotencyMiddleware(store *IdempotencyStore, next http.Handler) http.Handler
```
- Intercepts all requests before business logic
- Validates Idempotency-Key header presence
- Manages request states atomically
- Returns appropriate HTTP status codes

#### Request Flow

```
┌─────────────────────────────────────────────────────┐
│ Incoming Request with Idempotency-Key               │
└────────────────┬────────────────────────────────────┘
                 │
         ┌───────▼────────┐
         │ Key exists?    │
         └───┬──────────┬─┘
             │          │
        YES  │          │ NO
             │          └─────────────────────┐
         ┌───▼──────────────┐                 │
         │ Check Status     │                 │
         └───┬──────────┬───┘                 │
             │          │                    │
      PROCESSING COMPLETED         ┌─────────▼──────────┐
             │          │          │ Mark as PROCESSING │
             │          │          │ Execute Handler    │
        409  │         200│         │ Store Result       │
        Conflict Cached   │          │ Return 200         │
             │     Result │          └────────────────────┘
             └──────┬─────┘
                    │
            ┌───────▼────────┐
            │ Send Response  │
            └────────────────┘
```

#### Concurrent Request Behavior

**Scenario**: 8 concurrent requests with same Idempotency-Key

```
Request 1: Processing starts     ──→ Status 200 (Fresh)
Request 2: Still processing     ──→ Status 409 (Conflict)
Request 3: Still processing     ──→ Status 409 (Conflict)
...
Request N: Still processing     ──→ Status 409 (Conflict)
Request 1: Processing completes ──→ Result cached
Request 8: After completion     ──→ Status 200 (Cached)
```

#### Running Task 2

**Basic scenario (concurrent during processing)**:
```bash
go run task2_idempotency.go
go run -race task2_idempotency.go  # With race detector
```

**Advanced scenario (with phases)**:
```bash
go run task2_idempotency_advanced.go
go run -race task2_idempotency_advanced.go  # With race detector
```

#### Expected Output (Basic)
```
[Middleware] Processing new request with key payment_12345_user_777
[Handler] Payment processing started
[Middleware] Request still processing - returning 409
[Handler] Payment processing completed
[Middleware] Stored result for key payment_12345_user_777: status 200

Request 1: Status 200 (Fresh execution)
Request 2: Status 409 (Still processing)
Request 3: Status 409 (Still processing)
...
Total unique requests processed: 1 (all others were duplicates)
```

#### Expected Output (Advanced)
```
--- Phase 1: Concurrent During Processing ---
Request 1: Status 200 (executed)
Request 2: Status 409 (conflict)
Request 3: Status 409 (conflict)
Request 4: Status 409 (conflict)

--- Phase 2: Requests After Completion ---
Request 5: Status 200 (cached)
Request 6: Status 200 (cached)
Request 7: Status 200 (cached)

✓ Idempotency achieved: Handler executed only ONCE
```

---

## Key Concepts Demonstrated

### Exponential Backoff with Jitter
- Prevents "thundering herd" by distributing retry attempts
- Adapts wait time based on attempt number
- Formula: `min(baseDelay × 2^attempt, maxDelay)`
- Jitter: `random(0, calculatedBackoff)`

### Error Classification
- **Retriable**: Temporary/transient errors
  - Network timeouts
  - 429 Too Many Requests
  - 5xx Server Errors (500, 502, 503, 504)
  
- **Non-retriable**: Permanent errors
  - 401 Unauthorized (bad credentials)
  - 404 Not Found
  - All other 4xx errors (client problems)

### Idempotency Guarantees
- **At-most-once execution**: Business logic runs ≤ 1 time
- **Deduplication**: Duplicate requests detected and rejected
- **State consistency**: No double-charging or duplicate records
- **Safe retries**: Client can safely retry without side effects

### Context Handling
- Respects parent context cancellation
- Implements timeout checking on each retry
- Allows graceful shutdown of long-running operations
- Uses `ctx.Err()` to detect cancellation

---

## Thread Safety

Both implementations are **fully thread-safe**:

### Task 1
- HTTP client is thread-safe (Go standard library)
- Context cancellation is atomic
- No shared mutable state

### Task 2
- `IdempotencyStore` uses `sync.RWMutex`
- All map operations are protected
- Passes Go's `-race` detector
- Safe for concurrent goroutines

**Verify with**:
```bash
go run -race task2_idempotency.go
```

---

## Testing & Validation

All implementations include:
- ✓ Proper error handling
- ✓ Logging/tracing for debugging
- ✓ Test servers that simulate real scenarios
- ✓ Configurable parameters
- ✓ Race condition detection passed

### Run Tests
```bash
# Without race detector
go run task1_payment_retry.go
go run task2_idempotency.go
go run task2_idempotency_advanced.go

# With race detector (recommended)
go run -race task1_payment_retry.go
go run -race task2_idempotency.go
go run -race task2_idempotency_advanced.go
```

---

## Summary of Requirements Met

### Task 1 ✓
- [x] `IsRetryable()` function filters temporary vs permanent errors
- [x] `CalculateBackoff()` implements exponential backoff with jitter
- [x] `ExecutePayment()` respects context with global timeout
- [x] Test server returns 503 for first 3 attempts, then 200
- [x] Proper logging shows retry progress and wait times
- [x] Executes with `go run` and `go run -race`

### Task 2 ✓
- [x] Idempotency middleware checks Idempotency-Key header
- [x] Returns 400 if header missing
- [x] Returns 409 if request still processing
- [x] Returns 200 with cached result if completed
- [x] Business logic executes only once per unique key
- [x] Handles concurrent "double-click" requests
- [x] Thread-safe with sync.RWMutex
- [x] Executes with `go run` and `go run -race`

---

## Real-World Patterns

These implementations follow production patterns used by:
- **Stripe**: Idempotency keys for payment processing
- **AWS**: Exponential backoff for API calls
- **Google Cloud**: Retry policies with jitter
- **Netflix**: Hystrix circuit breaker patterns

---

## Related Files

- Part 1 examples (retry patterns):
  - `../Part1/1forloop.go` - Simple loop retry
  - `../Part1/2pause.go` - Fixed delay retry
  - `../Part1/3expbackoff.go` - Exponential backoff
  - `../Part1/4jitter.go` - Jitter addition
  - `../Part1/5context.go` - Context support
