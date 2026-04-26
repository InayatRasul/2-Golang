package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// IdempotencyStore stores information about processed idempotency keys
type IdempotencyStore struct {
	mu       sync.RWMutex
	requests map[string]*StoredRequest
}

// StoredRequest holds the status and result of an idempotent request
type StoredRequest struct {
	Status       string // "processing" or "completed"
	StatusCode   int
	ResponseBody string
	CreatedAt    time.Time
}

// NewIdempotencyStore creates a new store
func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{
		requests: make(map[string]*StoredRequest),
	}
}

// CheckAndStoreKey checks if a request with this key already exists
// Returns: (found bool, status string, statusCode int, body string, err string)
func (s *IdempotencyStore) CheckAndStoreKey(key string) (bool, string, int, string, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stored, exists := s.requests[key]; exists {
		// Key already processed
		if stored.Status == "processing" {
			// Still processing - return 409 Conflict
			return true, "processing", http.StatusConflict, "", "Request is still being processed"
		}
		// Completed - return stored result
		return true, "completed", stored.StatusCode, stored.ResponseBody, ""
	}

	// New key - mark as processing
	s.requests[key] = &StoredRequest{
		Status:    "processing",
		CreatedAt: time.Now(),
	}

	return false, "processing", 0, "", ""
}

// UpdateKeyCompleted marks a key as completed and stores the result
func (s *IdempotencyStore) UpdateKeyCompleted(key string, statusCode int, body string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stored, exists := s.requests[key]; exists {
		stored.Status = "completed"
		stored.StatusCode = statusCode
		stored.ResponseBody = body
	}
}

// IdempotencyMiddleware wraps an HTTP handler with idempotency support
func IdempotencyMiddleware(store *IdempotencyStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Idempotency-Key header
		idempotencyKey := r.Header.Get("Idempotency-Key")

		// Check if header is present
		if idempotencyKey == "" {
			fmt.Println("[Middleware] Missing Idempotency-Key header - returning 400")
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, `{"error": "Idempotency-Key header is required"}`)
			return
		}

		fmt.Printf("[Middleware] Processing request with Idempotency-Key: %s\n", idempotencyKey)

		// Check if this key was already processed
		found, status, statusCode, body, errMsg := store.CheckAndStoreKey(idempotencyKey)

		if found {
			if status == "processing" {
				// Request is still being processed
				fmt.Printf("[Middleware] Request with key %s is still processing - returning 409\n", idempotencyKey)
				w.WriteHeader(http.StatusConflict)
				io.WriteString(w, `{"error": "Request is still being processed"}`)
				return
			}

			if status == "completed" {
				// Return the previously saved result
				fmt.Printf("[Middleware] Request with key %s already completed - returning cached result\n", idempotencyKey)
				w.WriteHeader(statusCode)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Cached", "true")
				io.WriteString(w, body)
				return
			}

			if errMsg != "" {
				fmt.Printf("[Middleware] Error: %s\n", errMsg)
				w.WriteHeader(statusCode)
				io.WriteString(w, fmt.Sprintf(`{"error": "%s"}`, errMsg))
				return
			}
		}

		// Key is new - call the next handler and capture the response
		fmt.Printf("[Middleware] Processing new request with key %s\n", idempotencyKey)

		// Create a response writer that captures the response
		recorder := httptest.NewRecorder()

		// Call the next handler
		next.ServeHTTP(recorder, r)

		// Store the completed result
		resultBody := recorder.Body.String()
		store.UpdateKeyCompleted(idempotencyKey, recorder.Code, resultBody)

		fmt.Printf("[Middleware] Stored result for key %s: status %d\n", idempotencyKey, recorder.Code)

		// Copy the recorded response to the actual response writer
		for k, v := range recorder.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(recorder.Code)
		io.WriteString(w, resultBody)
	})
}

// paymentHandler simulates a payment operation
func paymentHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[Handler] Payment processing started")
	// Simulate heavy work
	time.Sleep(2 * time.Second)

	transactionID := fmt.Sprintf("txn_%d", time.Now().UnixNano())
	response := fmt.Sprintf(`{"status": "paid", "amount": 1000, "transaction_id": "%s"}`, transactionID)

	fmt.Println("[Handler] Payment processing completed")
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, response)
}

// Main function demonstrating idempotency with concurrent requests
func main() {
	// Create the idempotency store
	store := NewIdempotencyStore()

	// Create the handler with middleware
	baseHandler := http.HandlerFunc(paymentHandler)
	handlerWithIdempotency := IdempotencyMiddleware(store, baseHandler)

	// Create test server
	testServer := httptest.NewServer(handlerWithIdempotency)
	defer testServer.Close()

	fmt.Println("=== Idempotency Middleware for Loan Repayment ===")
	fmt.Printf("Server running at: %s\n\n", testServer.URL)

	// Simulate 5-10 concurrent requests with the same Idempotency-Key (double-click attack)
	idempotencyKey := "payment_12345_user_777"
	numRequests := 8

	fmt.Printf("Sending %d concurrent requests with the same Idempotency-Key: %s\n", numRequests, idempotencyKey)
	fmt.Println()

	var wg sync.WaitGroup
	results := make([]string, numRequests)
	resultsMutex := sync.Mutex{}

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			fmt.Printf("[Client %d] Sending request with Idempotency-Key: %s\n", index+1, idempotencyKey)

			// Create request
			req, _ := http.NewRequest("POST", testServer.URL, nil)
			req.Header.Set("Idempotency-Key", idempotencyKey)

			// Send request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				resultsMutex.Lock()
				results[index] = fmt.Sprintf("Request %d: Error - %v", index+1, err)
				resultsMutex.Unlock()
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			resultsMutex.Lock()
			results[index] = fmt.Sprintf("Request %d: Status %d, Body: %s", index+1, resp.StatusCode, string(body))
			resultsMutex.Unlock()

			cached := resp.Header.Get("X-Cached")
			if cached == "true" {
				fmt.Printf("[Client %d] Received CACHED response (status %d)\n", index+1, resp.StatusCode)
			} else {
				fmt.Printf("[Client %d] Received FRESH response (status %d)\n", index+1, resp.StatusCode)
			}
		}(i)

		// Stagger the requests slightly to simulate real-world timing
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for all requests to complete
	wg.Wait()

	fmt.Println()
	fmt.Println("=== Request Results ===")
	for _, result := range results {
		fmt.Println(result)
	}

	fmt.Println()
	fmt.Println("=== Summary ===")
	fmt.Printf("Total unique requests processed: 1 (all others were duplicates)\n")
	fmt.Printf("Storage contains %d key(s)\n", len(store.requests))
	for key, stored := range store.requests {
		fmt.Printf("  Key: %s, Status: %s, StatusCode: %d\n", key, stored.Status, stored.StatusCode)
	}
}
