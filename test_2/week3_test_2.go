package main

import (
	"fmt"
	"sync"
)

// QUESTION:
// 1. What is gonna be printed on the terminal? Why?

// This program demonstrates OPTIMISTIC CONCURRENCY CONTROL (OCC)
// Key concepts:
// - Version-based conflict detection (like database optimistic locking)
// - Mutex protects internal consistency (prevents race conditions)
// - Version number detects conflicting updates (application-level logic)
// - Only ONE of the two concurrent updates will succeed

type KV struct {
	mu      sync.Mutex // Protects the fields below from concurrent access
	value   string
	version int // Acts like a "timestamp" - increments on each successful write
}

// put implements conditional update: only succeeds if version hasn't changed
// This is similar to: UPDATE table SET value=? WHERE version=? in SQL
func (kv *KV) put(newValue string, expectedVersion int) bool {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if kv.version != expectedVersion {
		return false // Someone else updated it - reject this update
	}
	kv.value = newValue
	kv.version++ // Increment version to invalidate other pending updates
	return true
}

func main() {
	kv := &KV{value: "init", version: 0}
	var wg sync.WaitGroup
	wg.Add(2)

	//
	// Two goroutines try to update the same key with expectedVersion=0
	// Only the FIRST one to acquire the lock will succeed
	// The second one will see version=1 and fail
	go func() {
		defer wg.Done()
		ok_1 := kv.put("Alice", 0) // Expects version=0
		fmt.Println("Alice put result 1:", ok_1)
	}()

	go func() {
		defer wg.Done()
		ok_1 := kv.put("Bob", 0) // Also expects version=0
		fmt.Println("Bob put result 1:", ok_1)
	}()

	wg.Wait()
}
