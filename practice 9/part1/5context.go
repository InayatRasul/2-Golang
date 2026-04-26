package main
import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type RetryConfig struct {
	maxRetries int
	baseDelay time.Duration
	maxDelay time.Duration
}

func doSomethingUnreliable() error {
	if rand.Intn(10) < 7 {
		fmt.Println("Operation failed, retrying...")
		return errors.New("temporary failure")
	}
	fmt.Println("Operation succeeded!")
	return nil
}

func Retry(ctx context.Context, cfg RetryConfig) error {
	var err error
	for attempt := 0; attempt < cfg.maxRetries; attempt++ {
		// check for cancellation before each attempt
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = doSomethingUnreliable()
		if err == nil {
			return nil
		}
		// if it's the last attempt, we exit with an error.
		if attempt == cfg.maxRetries-1 {
			return err
		}
		// exponential backoff с Full Jitter
		backoff := cfg.baseDelay * time.Duration(math.Pow(2,float64(attempt)))
		if backoff > cfg.maxDelay {
			backoff = cfg.maxDelay
		}
		jitter := time.Duration(rand.Int63n(int64(backoff)))
		fmt.Printf("Attempt %d failed, waiting ~%v (max backoff:%v)...\n", attempt+1, jitter, backoff)
		time.Sleep(jitter)
	}
	return err
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := RetryConfig{
		maxRetries: 5,
		baseDelay:  100 * time.Millisecond,
		maxDelay:   2 * time.Second,
	}

	err := Retry(ctx, cfg)
	if err != nil {
		fmt.Printf("Final error: %v\n", err)
	}
}