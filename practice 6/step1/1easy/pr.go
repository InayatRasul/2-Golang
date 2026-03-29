package main
import (
	"fmt"
)

func goroutineHi(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.PrintLn("Goroutine Hello")
}
func main() {
	var wg sync.WaitGroup
	fmt.Println("main hello")
	wg.Add(1)
	go goroutineHi(&wg)
	wg.Wait()
}
