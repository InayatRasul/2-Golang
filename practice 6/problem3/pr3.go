package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// startServer is provided in the assignment [cite: 150-165]
// It sends metrics at random intervals until the context is canceled.
func startServer(ctx context.Context, name string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(rand.Intn(500)) * time.Millisecond):
				out <- fmt.Sprintf("[%s] metric: %d", name, rand.Intn(100))
			}
		}
	}()
	return out
}

// FanIn merges multiple channels into one [cite: 167-172]
func FanIn(ctx context.Context, channels ...<-chan string) <-chan string {
	out := make(chan string)
	var wg sync.WaitGroup

	// Function to pipe data from one channel to the 'out' channel
	multiplex := func(c <-chan string) {
		defer wg.Done()
		for val := range c {
			select {
			case <-ctx.Done():
				return
			case out <- val:
			}
		}
	}

	// Add the number of channels to the WaitGroup 
	wg.Add(len(channels))
	for _, ch := range channels {
		go multiplex(ch)
	}

	// Goroutine to close 'out' only after all source channels are closed 
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	// Initialize context with a 2-second timeout [cite: 174, 180]
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start 3 servers [cite: 175, 183-188]
	ch1 := startServer(ctx, "Alpha")
	ch2 := startServer(ctx, "Beta")
	ch3 := startServer(ctx, "Gamma")

	// Pass channels to FanIn [cite: 176, 189]
	ch4 := FanIn(ctx, ch1, ch2, ch3)

	// Read and display data [cite: 177, 191-193]
	for val := range ch4 {
		fmt.Println(val)
	}
}