package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)
// sync/atomic
func main() {
	var counter int64 // Must be int64 or int32 for atomic
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1) // Atomic increment
		}()
	}

	wg.Wait()
	fmt.Println(atomic.LoadInt64(&counter))
}