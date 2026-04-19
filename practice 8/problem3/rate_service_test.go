package problem3

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRate_Success(t *testing.T) {
	tests := []struct {
		name       string
		from       string
		to         string
		mockRate   float64
		expectRate float64
	}{
		{"USD to EUR", "USD", "EUR", 0.92, 0.92},
		{"USD to GBP", "USD", "GBP", 0.79, 0.79},
		{"USD to JPY", "USD", "JPY", 150.50, 150.50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the query parameters
				from := r.URL.Query().Get("from")
				to := r.URL.Query().Get("to")

				if from != tt.from || to != tt.to {
					http.Error(w, "invalid parameters", http.StatusBadRequest)
					return
				}

				// Return successful response
				response := RateResponse{
					Base:   from,
					Target: to,
					Rate:   tt.mockRate,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			service := NewExchangeService(server.URL)
			rate, err := service.GetRate(tt.from, tt.to)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if rate != tt.expectRate {
				t.Errorf("got %.2f, want %.2f", rate, tt.expectRate)
			}
		})
	}
}

func TestGetRate_EmptyInput(t *testing.T) {
	service := NewExchangeService("http://example.com")

	tests := []struct {
		name string
		from string
		to   string
	}{
		{"Empty from", "", "EUR"},
		{"Empty to", "USD", ""},
		{"Both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetRate(tt.from, tt.to)
			if err == nil {
				t.Error("expected error but got none")
			}
		})
	}
}

func TestGetRate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)
	_, err := service.GetRate("USD", "EUR")

	if err == nil {
		t.Error("expected error for API error response")
	}
}

func TestGetRate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)
	_, err := service.GetRate("USD", "EUR")

	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestGetRate_NetworkError(t *testing.T) {
	service := NewExchangeService("http://invalid-host-that-does-not-exist-12345.com")
	_, err := service.GetRate("USD", "EUR")

	if err == nil {
		t.Error("expected error for network failure")
	}
}

func TestGetRate_APIErrorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := RateResponse{
			Base:     "USD",
			Target:   "INVALID",
			ErrorMsg: "currency not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)
	_, err := service.GetRate("USD", "INVALID")

	if err == nil {
		t.Error("expected error for API error message")
	}
}
