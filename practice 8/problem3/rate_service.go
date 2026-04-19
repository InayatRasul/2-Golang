package problem3

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RateResponse struct {
	Base     string  `json:"base"`
	Target   string  `json:"target"`
	Rate     float64 `json:"rate"`
	ErrorMsg string  `json:"error,omitempty"`
}

type ExchangeService struct {
	BaseURL string
	Client  *http.Client
}

func NewExchangeService(baseURL string) *ExchangeService {
	return &ExchangeService{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetRate requests the rate. Example URL: /convert?from=USD&to=EUR
func (s *ExchangeService) GetRate(from, to string) (float64, error) {
	if from == "" || to == "" {
		return 0, fmt.Errorf("from and to currencies cannot be empty")
	}

	url := fmt.Sprintf("%s/convert?from=%s&to=%s", s.BaseURL, from, to)
	resp, err := s.Client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var result RateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("decode error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if result.ErrorMsg != "" {
			return 0, fmt.Errorf("api error: %s", result.ErrorMsg)
		}
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return result.Rate, nil
}
