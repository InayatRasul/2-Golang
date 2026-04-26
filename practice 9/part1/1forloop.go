package main
import (
	"errors"
	"fmt"
	"math/rand"
	// "time"
)
// Simulate an operation that sometimes crashes
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
	for {
		err = doSomethingUnreliable()
		if err == nil {
			break // Success, breaking the loop
		}
		// Error, just try again immediately
	}
	if err != nil {
		fmt.Printf("Failed after multiple retries: %v\n", err)
	}
}