package main
import (
	"fmt"
	"sync"
)


// func goroutineHello() {
// 	fmt.Println("goroutine hello")	
// }
// func main() {
// 	fmt.Println("main hello")

// 	go goroutineHello()
// }



//Q
// Hey, where’s the goroutine output?
// Why don’t we see “goroutine hello”?

//A
// The program starts the goroutine with go goroutineHello()
// But then main() immediately exits
// When the main goroutine finishes, the entire program terminates
// The spawned goroutine never gets scheduled to run before the program ends

//waitGroup
// It starts at 0
// While it’s greater than 0, we tell the main function to wait before
// continuing its execution and ending the program
// When it returns to 0, the program can finish

func goroutineHello(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("goroutine hello")
}

func main() {
	var wg sync.WaitGroup
	fmt.Println("main hello")

	wg.Add(1)
	go goroutineHello(&wg)
	wg.Wait()
}

// - wg.Add(1) adds 1 to the counter, to signal that we’re creating a
// goroutine
// - wg.Done() decreases the counter by 1, signaling that the
// goroutine has finished
// - wg.Wait() asks the main function to wait until this counter
// reaches zero

// Q:
// What if we call wg.Done() when the counter is zero? - panic: sync: negative WaitGroup counter
// Why we used wg parameter as a pointer? - Because WaitGroup must be shared by value between goroutines, and you need them all to modify the same counter.