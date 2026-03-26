package main
import(
	"fmt"
	"sync"
)

// What if we want the goroutines to execute in sequence? For this,
// we’ll use channels — a special Go feature that allows passing
// values between goroutines.
func goroutineCh(num int, wg *sync.WaitGroup, ch chan bool) {
	defer wg.Done()
	fmt.Println("Waiting for permission to start...")
	<-ch
	fmt.Printf("Goroutine: %d\n", num)
	ch <- true
}

func main() {
	var wg sync.WaitGroup

	ch := make(chan bool, 1)
	ch <- true
	fmt.Println("hello world from main")

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go goroutineCh(i, &wg, ch)
		ch <- true
		fmt.Println("waiting for previous goroutine to finish...")
		<-ch
	}

	wg.Wait()
}

// Conclusion for part 1:

// Use go to create goroutines
// Use waitGroup to synchronize with the main function
// Pass waitGroup by reference
// Channels (chan < type >) help control flow between goroutines
// Go’s scheduler is non-deterministic by default