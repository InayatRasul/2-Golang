package main
import (
	"fmt"
	"sync"
)
func main() {
	var counter int

	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++
		}()
	}

	wg.Wait()

	fmt.Println(counter)
}

// 1. Explanation for Team Lead
	// "The final count is inconsistent because multiple goroutines perform non-atomic read-modify-write operations on the shared counter variable simultaneously, leading to lost updates."