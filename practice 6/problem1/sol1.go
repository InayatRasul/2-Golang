package main

import (
	"fmt"
	"sync"
)

// sync.map

func main() {
	// sync.Map is safe for concurrent use
	var safeMap sync.Map
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			// Store uses key-value pairs
			safeMap.Store("key", val)
		}(i)
	}

	wg.Wait()
	
	// Load retrieves the value; it returns (interface{}, bool)
	if value, ok := safeMap.Load("key"); ok {
		fmt.Printf("Value (sync.Map): %d\n", value)
	}
}
