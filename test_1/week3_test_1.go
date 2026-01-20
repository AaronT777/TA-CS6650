package main

import (
	"fmt"
	"sync"
)

// QUESTIONS:
// 1. What will the final value of `counter` be after the program exits?
// 2. Are the FIRST THREE printed lines deterministic? If not, list what *must* be true

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	counter := 0

	for i := 0; i < 5; i++ {
		wg.Add(1) // Increment counter before launching goroutine
		go func(id int) {
			defer wg.Done() // Decrement counter when goroutine completes
			for j := 0; j < 3; j++ {
				mu.Lock() // Acquire lock
				counter++
				fmt.Printf("Goroutine %d incremented counter to %d\n", id, counter)
				mu.Unlock() // Release lock
			}
		}(i)
	}

	wg.Wait() // Block until counter becomes zero
	fmt.Println("Final counter:", counter)
}
