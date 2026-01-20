package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// QUESTIONS:
// 1. Which node becomes leader in a single run?
// 2. What is the maximum possible election time in this setup (timeouts [150ms, 400ms))? Explain.

//
// This program simulates RAFT LEADER ELECTION (simplified)
// Key concepts:
// - Randomized timeouts prevent split votes in distributed consensus
// - Compare-And-Swap (CAS) is a lock-free atomic operation
// - First node to timeout becomes leader (winner-takes-all)

func main() {

	const N = 5 // number of nodes
	var wg sync.WaitGroup
	var leader int32 = -1 // -1 means no leader yet

	wg.Add(N)
	for id := 0; id < N; id++ {
		go func(id int) {
			defer wg.Done()
			// Each node waits a random election timeout (like Raft followers).
			// waits between 150ms to 400ms
			sleep := time.Duration(150+rand.Intn(250)) * time.Millisecond
			time.Sleep(sleep)

			// ok := atomic.CompareAndSwapInt32(addr *int32, old, new)
			// does this atomically (as one indivisible hardware instruction):
			// Looks at the value stored at *addr.
			// If it equals old, it replaces it with new and returns true.
			// If it does not equal old, it leaves it unchanged and returns false.
			if atomic.CompareAndSwapInt32(&leader, -1, int32(id)) {
				fmt.Printf("[Node %d] timeout=%v -> ELECTED leader\n", id, sleep)
			} else {
				fmt.Printf("[Node %d] timeout=%v -> lost (leader=%d)\n", id, sleep, atomic.LoadInt32(&leader))
			}
		}(id)
	}

	wg.Wait()
	fmt.Printf("Final leader: %d\n", leader)
}
