package main

import (
	"fmt"
	"sync"
)

// QUESTIONS:
// 1. What will the final value of `counter` be after the program exits?
// 2. Are the FIRST THREE printed lines deterministic? If not, list what *must* be true

//
// This program demonstrates:
// - Race conditions and why we need synchronization
// - Mutex protects shared data but does NOT control execution order
// - Go scheduler determines goroutine execution order non-deterministically

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	counter := 0

	for i := 0; i < 5; i++ {
		wg.Add(1) // Must add before goroutine starts (avoid race)
		go func(id int) {
			defer wg.Done() // Ensure Done() is called even if panic occurs
			for j := 0; j < 3; j++ {
				mu.Lock() // Critical section starts - only ONE goroutine can enter
				counter++ // Without mutex, this would be a DATA RACE
				fmt.Printf("Goroutine %d incremented counter to %d\n", id, counter)
				mu.Unlock() // Critical section ends - allow other goroutines to enter
			}
		}(i)
	}

	wg.Wait() // Block main until all goroutines finish
	fmt.Println("Final counter:", counter)
}
