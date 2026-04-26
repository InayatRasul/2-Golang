package main

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"time"
)

// PaymentClient handles payment processing with retry logic
type PaymentClient struct {
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

// IsRetryable determines if an error or HTTP response is worth retrying
func IsRetryable(resp *http.Response, err error) bool {
	// Network errors are generally retryable (timeout, connection refused, etc.)
	if err != nil {
		return true
	}

	// Check HTTP status codes
	if resp != nil {
		// Retryable status codes:
		// 429 - Too Many Requests (rate limited)
		// 500 - Internal Server Error
		// 502 - Bad Gateway
		// 503 - Service Unavailable
		// 504 - Gateway Timeout
		switch resp.StatusCode {
		case http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout:
			return true
		// Non-retryable status codes:
		// 401 - Unauthorized (invalid API key)
		// 404 - Not Found
		case http.StatusUnauthorized, http.StatusNotFound:
			return false
		}

		// All other 4xx errors are non-retryable (bad request)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return false
		}
	}

	return false
}

// CalculateBackoff implements exponential backoff with full jitter
func (pc *PaymentClient) CalculateBackoff(attempt int) time.Duration {
	// Exponential backoff: baseDelay * 2^attempt
	backoff := pc.baseDelay * time.Duration(math.Pow(2, float64(attempt)))

	// Cap at maxDelay
	if backoff > pc.maxDelay {
		backoff = pc.maxDelay
	}

	// Add full jitter: random value between 0 and backoff
	jitter := time.Duration(rand.Int63n(int64(backoff)))

	return jitter
}

// ExecutePayment attempts to process a payment with retry logic
func (pc *PaymentClient) ExecutePayment(ctx context.Context, url string, body []byte) (*http.Response, error) {
	var lastErr error
	var resp *http.Response

	for attempt := 0; attempt < pc.maxRetries; attempt++ {
		// Check if context is cancelled or deadline exceeded
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
		if err != nil {
			lastErr = err
			continue
		}

		// Execute request
		resp, err := http.DefaultClient.Do(req)

		// Check if we should retry
		if IsRetryable(resp, err) {
			fmt.Printf("Attempt %d failed: status %d\n", attempt+1, resp.StatusCode)
			lastErr = err

			// Don't sleep after the last attempt
			if attempt < pc.maxRetries-1 {
				waitTime := pc.CalculateBackoff(attempt)
				fmt.Printf("  Waiting %v before next retry...\n", waitTime)

				// Create a timer that can be cancelled
				timer := time.NewTimer(waitTime)
				defer timer.Stop()

				select {
				case <-timer.C:
					// Timer finished, continue to next attempt
				case <-ctx.Done():
					// Context cancelled, stop retrying
					return nil, ctx.Err()
				}
			}
			continue
		}

		// Success or non-retryable error
		if resp != nil && resp.StatusCode == http.StatusOK {
			fmt.Printf("Attempt %d: Success! (Status %d)\n", attempt+1, resp.StatusCode)
		}

		return resp, err
	}

	// All retries exhausted
	fmt.Printf("Failed after %d attempts\n", pc.maxRetries)
	return resp, lastErr
}

// Main function with test scenario
func main() {
	rand.Seed(time.Now().UnixNano())

	// Counter for test server requests
	var requestCount int

	// Create a test server that:
	// - Returns 503 for first 3 requests
	// - Returns 200 on the 4th request
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		fmt.Printf("\n[Server] Received request #%d\n", requestCount)

		if requestCount <= 3 {
			// Simulate service unavailable for first 3 requests
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"error": "service unavailable"}`)
			fmt.Printf("[Server] Returning 503 Service Unavailable\n")
		} else {
			// Success on 4th request
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"status": "success", "transaction_id": "txn_12345", "amount": 1000}`)
			fmt.Printf("[Server] Returning 200 OK with success response\n")
		}
	}))
	defer testServer.Close()

	// Create payment client with retry configuration
	client := &PaymentClient{
		maxRetries: 5,
		baseDelay:  500 * time.Millisecond,
		maxDelay:   5 * time.Second,
	}

	// Create context with 10-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("=== Payment Processing with Retry Mechanism ===")
	fmt.Printf("Starting payment request to: %s\n", testServer.URL)
	fmt.Printf("Configuration: MaxRetries=%d, BaseDelay=%v, MaxDelay=%v\n",
		client.maxRetries, client.baseDelay, client.maxDelay)
	fmt.Println()

	// Execute payment
	resp, err := client.ExecutePayment(ctx, testServer.URL, nil)

	fmt.Println()
	fmt.Println("=== Final Result ===")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else if resp != nil {
		fmt.Printf("Success! Response status: %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("Response body: %s\n", string(body))
	}

	fmt.Printf("Total requests to server: %d\n", requestCount)
}
