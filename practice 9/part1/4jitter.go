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
	rand.Seed(time.Now().UnixNano()) // Improtant for randomness

	for attempt := 0; attempt < maxRetries; attempt++ {
		err = doSomethingUnreliable()
		if err == nil {
			break // Success
		}
		if attempt == maxRetries-1 {
			break
		}
		backoffTime := baseDelay * time.Duration(math.Pow(2,float64(attempt)))
		if backoffTime > maxDelay {
			backoffTime = maxDelay
		}
		// Adding Full Jitter
		jitter := time.Duration(rand.Int63n(int64(backoffTime)))
		fmt.Printf("Attempt %d failed, waiting ~%v (backoff %v + jitter) before next retry...\n", attempt+1, jitter, backoffTime)
		time.Sleep(jitter)
	}
// ... (processing the final result) ...
}