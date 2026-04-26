package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// CachedResponse stores the response for an idempotent key
type CachedResponse struct {
	StatusCode int
	Body       []byte
	Completed  bool
}

// MemoryStore is an in-memory store for idempotency keys
type MemoryStore struct {
	mu   sync.Mutex
	data map[string]*CachedResponse
}

// NewMemoryStore creates a new MemoryStore
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]*CachedResponse)}
}

// Get retrieves a cached response by key
func (m *MemoryStore) Get(key string) (*CachedResponse, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	resp, exists := m.data[key]
	return resp, exists
}

// StartProcessing tries to reserve a key for a new request
// Returns true if the key was successfully reserved (new request)
// Returns false if the key already exists
func (m *MemoryStore) StartProcessing(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[key]; exists {
		return false // key already exists
	}
	// Insert an "empty" record, marking that the request is in progress
	m.data[key] = &CachedResponse{Completed: false}
	return true
}

// Finish saves the result and marks the query as completed
func (m *MemoryStore) Finish(key string, status int, body []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if resp, exists := m.data[key]; exists {
		resp.StatusCode = status
		resp.Body = body
		resp.Completed = true
	} else {
		// Just in case there was no entry (shouldn't happen)
		m.data[key] = &CachedResponse{StatusCode: status, Body: body, Completed: true}
	}
}

// IdempotencyMiddleware implements idempotency checking
func IdempotencyMiddleware(store *MemoryStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
			return
		}

		// Check if this key has already been processed
		if cached, exists := store.Get(key); exists {
			if cached.Completed {
				// Return the saved result
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
			} else {
				// A request with this key is still being processed
				http.Error(w, "Duplicate request in progress", http.StatusConflict)
			}
			return
		}

		// Trying to reserve the key (mark as "in progress")
		if !store.StartProcessing(key) {
			// Someone managed to insert the key before us.
			if cached, exists := store.Get(key); exists && cached.Completed {
				// If the request has just completed, return the result
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
			} else {
				// Otherwise, we will report a conflict (duplicate in progress)
				http.Error(w, "Duplicate request in progress", http.StatusConflict)
			}
			return
		}

		// Our first request is to execute the main logic
		recorder := httptest.NewRecorder()
		next.ServeHTTP(recorder, r)

		// Save the result and mark the key as complete
		store.Finish(key, recorder.Code, recorder.Body.Bytes())

		// Return to the client the response received from the main logic
		for k, vals := range recorder.Header() {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(recorder.Code)
		w.Write(recorder.Body.Bytes())
	})
}

// paymentHandler simulates a payment processing endpoint
func paymentHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate heavy operation (payment processing)
	time.Sleep(2 * time.Second)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "paid", "amount": 1000, "transaction_id": "uuid-123456789"}`))
}

func main() {
	store := NewMemoryStore()

	// Create handler with idempotency middleware
	handler := IdempotencyMiddleware(store, http.HandlerFunc(paymentHandler))

	// Create test server
	server := httptest.NewServer(handler)
	defer server.Close()

	client := &http.Client{}

	// Test 1: First request should execute normally
	fmt.Println("=== Test 1: First request (should execute) ===")
	req1, _ := http.NewRequest("POST", server.URL, nil)
	req1.Header.Set("Idempotency-Key", "unique-key-123")
	resp1, err := client.Do(req1)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Status: %d\n", resp1.StatusCode)
		buf := make([]byte, 1024)
		n, _ := resp1.Body.Read(buf)
		fmt.Printf("Body: %s\n", string(buf[:n]))
	}

	// Test 2: Duplicate request should return cached result
	fmt.Println("\n=== Test 2: Duplicate request (should return cached) ===")
	req2, _ := http.NewRequest("POST", server.URL, nil)
	req2.Header.Set("Idempotency-Key", "unique-key-123")
	resp2, err := client.Do(req2)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Status: %d\n", resp2.StatusCode)
		buf := make([]byte, 1024)
		n, _ := resp2.Body.Read(buf)
		fmt.Printf("Body: %s\n", string(buf[:n]))
	}

	// Test 3: Missing Idempotency-Key should return 400
	fmt.Println("\n=== Test 3: Missing Idempotency-Key (should return 400) ===")
	req3, _ := http.NewRequest("POST", server.URL, nil)
	resp3, err := client.Do(req3)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Status: %d\n", resp3.StatusCode)
	}

	// Test 4: Concurrent requests (simulating double-click)
	fmt.Println("\n=== Test 4: Concurrent requests (simulating double-click) ===")
	var wg sync.WaitGroup
	results := make(chan int, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("POST", server.URL, nil)
			req.Header.Set("Idempotency-Key", "concurrent-key-456")
			resp, _ := client.Do(req)
			results <- resp.StatusCode
		}()
	}

	wg.Wait()
	close(results)

	fmt.Println("Results from concurrent requests:")
	for status := range results {
		fmt.Printf("  Status: %d\n", status)
	}
}