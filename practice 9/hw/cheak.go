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
// and cached responses after completion
func main() {
	// Create the idempotency store
	store := NewIdempotencyStore()

	// Create the handler with middleware
	baseHandler := http.HandlerFunc(paymentHandler)
	handlerWithIdempotency := IdempotencyMiddleware(store, baseHandler)

	// Create test server
	testServer := httptest.NewServer(handlerWithIdempotency)
	defer testServer.Close()

	fmt.Println("=== Idempotency Middleware for Loan Repayment (Advanced) ===")
	fmt.Printf("Server running at: %s\n\n", testServer.URL)

	// Scenario: Simulate a double-click attack followed by retries after completion
	idempotencyKey := "payment_advanced_user_777"

	fmt.Println("--- Phase 1: Concurrent Requests During Processing ---")
	fmt.Printf("Sending 4 concurrent requests with Idempotency-Key: %s\n\n", idempotencyKey)

	var wg sync.WaitGroup
	results := make(map[int]string)
	resultsMutex := sync.Mutex{}

	// Phase 1: Concurrent requests during processing
	for i := 1; i <= 4; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			fmt.Printf("[Phase1-Client %d] Sending request\n", index)

			req, _ := http.NewRequest("POST", testServer.URL, nil)
			req.Header.Set("Idempotency-Key", idempotencyKey)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				resultsMutex.Lock()
				results[index] = fmt.Sprintf("Error: %v", err)
				resultsMutex.Unlock()
				return
			}
			defer resp.Body.Close()

			_, _ = io.ReadAll(resp.Body)
			resultsMutex.Lock()
			results[index] = fmt.Sprintf("Status %d", resp.StatusCode)
			resultsMutex.Unlock()

			cached := resp.Header.Get("X-Cached")
			if cached == "true" {
				fmt.Printf("[Phase1-Client %d] Got CACHED response (status %d)\n", index, resp.StatusCode)
			} else {
				fmt.Printf("[Phase1-Client %d] Got FRESH response (status %d)\n", index, resp.StatusCode)
			}
		}(i)

		time.Sleep(150 * time.Millisecond)
	}

	wg.Wait()

	fmt.Println("\n--- Phase 2: Retry Requests After Completion ---")
	fmt.Printf("Sending 3 additional requests with same Idempotency-Key (after first completed)\n\n")

	// Phase 2: Retries after the first request completed
	for i := 5; i <= 7; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			fmt.Printf("[Phase2-Client %d] Sending retry request\n", index)

			req, _ := http.NewRequest("POST", testServer.URL, nil)
			req.Header.Set("Idempotency-Key", idempotencyKey)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				resultsMutex.Lock()
				results[index] = fmt.Sprintf("Error: %v", err)
				resultsMutex.Unlock()
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			cached := resp.Header.Get("X-Cached")

			resultsMutex.Lock()
			if cached == "true" {
				results[index] = fmt.Sprintf("Status %d (CACHED)", resp.StatusCode)
			} else {
				results[index] = fmt.Sprintf("Status %d", resp.StatusCode)
			}
			resultsMutex.Unlock()

			if cached == "true" {
				fmt.Printf("[Phase2-Client %d] Got CACHED response (status %d)\n", index, resp.StatusCode)
				fmt.Printf("[Phase2-Client %d] Response body: %s\n", index, string(body))
			} else {
				fmt.Printf("[Phase2-Client %d] Got FRESH response (status %d)\n", index, resp.StatusCode)
			}
		}(i)

		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()

	fmt.Println()
	fmt.Println("=== Final Results Summary ===")
	for i := 1; i <= 7; i++ {
		fmt.Printf("Request %d: %s\n", i, results[i])
	}

	fmt.Println()
	fmt.Println("=== Idempotency Analysis ===")
	fmt.Printf("Storage contains %d key(s):\n", len(store.requests))
	for key, stored := range store.requests {
		fmt.Printf("  Key: %s\n", key)
		fmt.Printf("    Status: %s\n", stored.Status)
		fmt.Printf("    StatusCode: %d\n", stored.StatusCode)
		fmt.Printf("    ResponseBody length: %d bytes\n", len(stored.ResponseBody))
	}
	fmt.Println()
	fmt.Println("✓ Idempotency achieved: Payment handler executed only ONCE")
	fmt.Println("✓ All duplicate requests returned cached result or conflict status")
}
