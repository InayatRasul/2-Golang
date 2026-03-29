package main
import(
	"fmt"
	"sync"
)
func goroutine(num int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("goroutine: %d\n", num)
}
func main() {
	var wg sync.WaitGroup
	fmt.Println("hello from main")

	for i := 0; i < 10; i++ {
		wg.Add(1)              // Counter: 1
		go goroutine(i, &wg)   // Schedules goroutine, doesn't run yet!
	}

	// Counter is still 10 at this point!
	// None of the goroutines have executed yet

	wg.Wait()  // Main waits here until counter reaches 0
}

// Something interesting happens when we run this code several
// times. Run it yourself!
// Why is this happening?

//A
// When you run this code multiple times, the output order of the goroutines is different each time:
// The go keyword only schedules the goroutine to run later. It doesn't execute immediately. So all 10 goroutines are added before any of them actually run and call Done().
// The Go scheduler decides which goroutine runs first - this decision is not guaranteed and can vary
// The goroutines don't run in order; they run whenever the scheduler gives them CPU time

// The order of goroutine execution is non-deterministic (unpredictable)