package main
import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

func doSomethingUnreliable() error {
	if rand.Intn(10) < 7 {
		fmt.Println("Operation failed, retrying...")
		return errors.New("temporary failure")
	}
	fmt.Println("Operation succeeded!")
	return nil
}

func main() {
	var err error
	const maxRetries = 5
	baseDelay := 100 * time.Millisecond
	maxDelay := 5 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		err = doSomethingUnreliable()
		if err == nil {
			break // Success
		}
		if attempt == maxRetries-1 {
			break // We do not count the delay after the last attempt
		}
		// Calculating exponential delay
		backoffTime := baseDelay * time.Duration(math.Pow(2,float64(attempt)))
		// Limiting the maximum delay
		if backoffTime > maxDelay {
			backoffTime = maxDelay
		}
		fmt.Printf("Attempt %d failed, waiting %v before next retry...\n", attempt+1, backoffTime)
		time.Sleep(backoffTime)
	}
// further logic
}