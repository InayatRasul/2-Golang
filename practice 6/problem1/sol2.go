package main

import (
	"fmt"
	"sync"
)

// sync.rwmutex

func main() {
	unsafeMap := make(map[string]int)
	var mu sync.RWMutex
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			mu.Lock() // Lock for writing
			unsafeMap["key"] = val
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	mu.RLock() // Lock for reading
	fmt.Printf("Value (RWMutex): %d\n", unsafeMap["key"])
	mu.RUnlock()
}