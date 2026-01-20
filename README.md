# CS6650 Week 3 Quiz - Go Concurrency & Distributed Systems

## TA 使用说明 (中文)

### 测试流程
1. **先让学生回答问题** - 不要运行代码
2. **收集学生答案** - 记录他们的理解
3. **运行代码展示实际结果** - 使用下面的命令
4. **讲解知识点** - 使用本文档中的解释

### 运行命令
```bash
# Test 1 - 并发计数器
go run test_1/week3_test_1.go

# Test 2 - 乐观并发控制
go run test_2/week3_test_2.go

# Test 3 - Raft 领导选举
go run test_3/week3_test_3.go

# 多次运行观察不同结果（对 test_2 和 test_3 特别有用）
for i in {1..5}; do echo "=== Run $i ==="; go run test_2/week3_test_2.go; done
```

Great! Now let's move on to a hands-on exercise.

I've prepared three Go programs that test different concurrency concepts - things like race conditions, synchronization primitives, and distributed consensus.

Here's how this works:
- I'll show you the code and questions
- Think it through and tell me what you expect to happen
- Don't worry about getting it perfect - the discussion is more important
- After your answer, we'll run it together and discuss

Sound good? Let's look at the first one.

---

## Test 1: Concurrent Counter with Mutex

### Question for Students
```
1. What will the final value of `counter` be after the program exits?
2. Are the FIRST THREE printed lines deterministic? If not, list what *must* be true.
```

### Answer Key (For TA Reference)

**Question 1:** 
- Final counter value: **15**
- Explanation: 5 goroutines, each loops 3 times, total 5 × 3 = 15

**Question 2:**
- **No**, the first three lines are NOT deterministic
- What must be true:
  - All three lines will show counter values from the set {1, 2, 3}
  - The goroutine IDs can be any combination of {0, 1, 2, 3, 4}
  - Each printed counter value must be unique (no duplicates in first 3 lines)
  - Each line must show: counter(line N) = N

### 给学生的知识点讲解 (English)

#### Key Concepts

**1. sync.WaitGroup**
```go
var wg sync.WaitGroup
wg.Add(1)    // Increment counter before launching goroutine
wg.Done()    // Decrement counter when goroutine completes
wg.Wait()    // Block until counter becomes zero
```
- Used to wait for multiple goroutines to finish
- Must call `Add()` before starting goroutine to avoid race condition
- `Done()` is typically called with `defer` to ensure it runs even if panic occurs

**2. sync.Mutex (Mutual Exclusion Lock)**
```go
var mu sync.Mutex
mu.Lock()    // Acquire lock (blocks if already locked)
// Critical section - only one goroutine can execute this at a time
mu.Unlock()  // Release lock
```
- Protects shared resources from concurrent access
- Without mutex: race condition → unpredictable counter value
- With mutex: operations are **serialized** (one at a time)

**3. Why the Output Order is Non-Deterministic**
- The Go scheduler can run goroutines in any order
- Even with mutex, we only guarantee:
  - **Atomicity**: Each increment is atomic
  - **Sequential consistency**: Counter values are 1, 2, 3, ..., 15
- We do NOT guarantee which goroutine runs first

**4. Common Student Mistakes**
- Thinking the output will always show goroutines in order (0, 1, 2, 3, 4)
- Forgetting that mutex only protects the critical section, not execution order
- Thinking they can predict which goroutine will print first

---

## Test 2: Optimistic Concurrency Control

### Question for Students
```
1. What is gonna be printed on the terminal? Why?
```

### Answer Key (For TA Reference)

**Possible outputs (two cases):**

**Case 1:**
```
Alice put result 1: true
Bob put result 1: false
```

**Case 2:**
```
Bob put result 1: true
Alice put result 1: false
```

**Key points:**
- One always returns `true`, one returns `false`
- Order is non-deterministic (depends on scheduling)
- Only the first goroutine to acquire lock succeeds (version matches)
- Second goroutine sees version=1, fails because expectedVersion=0

### TA Deep Dive (Anticipated Student Questions)

**Questions students might ask:**

1. **Why can't both succeed?**
   - First one succeeds → version changes from 0 to 1
   - Second one checks and sees version=1, which doesn't match expectedVersion=0, so it fails

2. **How does this relate to database optimistic locking?**
   - This IS optimistic concurrency control (OCC)
   - Unlike mutex which "pessimistically" blocks concurrency
   - OCC "optimistically" allows concurrent reads, only checks version on write

3. **Why need mutex if we already have version?**
   - Version only detects conflicts, doesn't prevent race conditions
   - Without mutex, there's a race between `if kv.version != expectedVersion` and `kv.version++`
   - Both goroutines could see version=0 and both think they can update

### 给学生的知识点讲解 (English)

#### Key Concepts

**1. Optimistic Concurrency Control (OCC)**
- **Philosophy**: Assume conflicts are rare, detect them when they happen
- **Contrast with Pessimistic Locking**: 
  - Pessimistic: Lock before read/write (like `sync.Mutex`)
  - Optimistic: Read freely, check version before write

**2. Version-Based Conflict Detection**
```go
type KV struct {
    mu      sync.Mutex
    value   string
    version int  // Increments on every successful write
}
```
- Version acts as a "timestamp" or "sequence number"
- Client reads data + version
- Client sends update with expected version
- Server rejects if version changed (someone else wrote)

**3. Why Both Mutex AND Version?**
- **Mutex**: Protects internal consistency of the data structure
- **Version**: Implements application-level conflict detection
- They serve different purposes:
  - Mutex: Low-level race condition prevention
  - Version: High-level business logic (detect conflicting updates)

**4. Real-World Examples**
- **Etcd/ZooKeeper**: Key-value stores with version numbers
- **Databases**: `UPDATE ... WHERE version = ?`
- **HTTP**: `ETag` headers and `If-Match` conditions
- **Git**: Merge conflicts are version conflicts!

**5. Execution Timeline**
```
Time →
Goroutine Alice:                  Goroutine Bob:
----------------                  ---------------
mu.Lock()
  check version = 0 ✓
  set value = "Alice"
  version = 1
mu.Unlock()
                                  mu.Lock()
                                    check version = 1 ✗ (expected 0)
                                    return false
                                  mu.Unlock()
```

---

## Test 3: Raft Leader Election Simulation

### Question for Students
```
1. Which node becomes leader in a single run?
2. What is the maximum possible election time in this setup (timeouts [150ms, 400ms))? Explain.
```

### Answer Key (For TA Reference)

**Question 1:**
- **Non-deterministic** - depends on random timeouts
- The node with shortest timeout becomes leader
- Result varies on each run

**Question 2:**
- Maximum election time: **just under 400ms** (technically 399ms)
- Explanation:
  - All nodes start simultaneously
  - The node with shortest timeout wins
  - Worst case: winner has timeout just under 400ms (max of range)
  - Even if winner has 399ms, others don't matter (already lost)
  - Election completes when the FIRST node times out

**Common incorrect answers:**
- Wrong: 800ms (longest + second longest)
  - After first node becomes leader, other timeouts don't affect election
- Wrong: 250ms (average)
  - We're asking for worst case, not average case

### TA Deep Dive (Advanced Concepts)

**Deep understanding:**

1. **Why use atomic.CompareAndSwapInt32?**
   - Lock-free primitive
   - More efficient than mutex
   - Hardware-level atomic operation (CPU instruction)
   - On x86, this is the `CMPXCHG` instruction

2. **Actual Raft election process (simplified):**
   - Followers wait for heartbeat
   - On timeout → become Candidate, start election
   - Request votes from other nodes
   - Become Leader after receiving majority votes
   - Our code simplifies this: CAS models "first to timeout wins"

3. **Why randomized timeouts?**
   - Avoid split votes
   - If all nodes timeout simultaneously, they might all vote for themselves
   - Randomization reduces conflict probability

### 给学生的知识点讲解 (English)

#### Key Concepts

**1. Atomic Operations**
```go
atomic.CompareAndSwapInt32(&leader, old, new)
```
- **Atomic** = indivisible = happens as single CPU instruction
- **CAS (Compare-And-Swap)**: 
  - Check if value equals `old`
  - If yes → set to `new`, return `true`
  - If no → leave unchanged, return `false`
  - All in ONE atomic step (no interruption possible)

**2. Why CAS Instead of Mutex?**
```go
// With mutex (more code, slightly slower):
mu.Lock()
if leader == -1 {
    leader = id
    won = true
}
mu.Unlock()

// With CAS (one atomic operation):
won = atomic.CompareAndSwapInt32(&leader, -1, id)
```
- CAS is **lock-free**: no blocking, no waiting
- Better performance in low-contention scenarios
- Used in high-performance concurrent data structures

**3. Raft Leader Election (Simplified)**

**Real Raft Algorithm:**
1. Each node has an **election timeout** (randomized)
2. When timeout expires → become **Candidate**
3. Candidate increments its term, votes for itself
4. Requests votes from other nodes
5. If gets majority → become **Leader**
6. Leader sends heartbeats to prevent new elections

**Our Simulation:**
- We skip voting phase
- First node to timeout "wins" via CAS
- Models the key insight: **randomized timeouts prevent conflicts**

**4. Why Random Timeouts?**
```
Without randomization:          With randomization:
All timeout at same time    →   One times out first
All become candidates       →   Becomes leader via CAS
Split vote problem          →   Others lose, see leader already elected
Need retry                  →   Election completes immediately
```

**5. Maximum Election Time Analysis**
```
Timeout range: [150ms, 400ms)

Best case:  Leader times out at 150ms → election done at 150ms
Worst case: Leader times out at 399ms → election done at 399ms

Key insight: Election completes when FIRST node times out
            (because CAS ensures only one winner)

NOT 150ms + 400ms = 550ms (this would be if we needed TWO nodes)
NOT 250ms (this is average, not maximum)
```

**6. Comparing with Real Distributed Systems**
| System | Leader Election Method |
|--------|------------------------|
| **Raft** | Randomized timeouts + voting |
| **Paxos** | Proposer competition with promises |
| **Zab (ZooKeeper)** | Similar to Raft, epoch-based |
| **Bully Algorithm** | Highest ID wins, simpler but less fault-tolerant |

---

## Additional Tips for TA

### 如何引导讨论

1. **Test 1 - 从简单开始**
   - 先问：没有 mutex 会怎样？
   - 让学生画时间线图
   - 强调：mutex 保证正确性，但不保证顺序

2. **Test 2 - 连接到实际应用**
   - 问：数据库的乐观锁和这个有什么关系？
   - 举例：购物车同时结账、文档同时编辑
   - 讨论：什么时候用乐观锁 vs 悲观锁？

3. **Test 3 - 扩展到分布式系统**
   - 问：如果有网络延迟呢？
   - 讨论：如何处理脑裂（split brain）？
   - 提及：CAP 定理、一致性问题

### 常见学生误区

| 误区 | 正确理解 |
|------|---------|
| Mutex 让程序按顺序执行 | Mutex 只保证临界区互斥，不控制执行顺序 |
| Version 可以替代 mutex | Version 检测冲突，mutex 防止 race condition |
| CAS 总是比 mutex 好 | CAS 适合低竞争场景，高竞争时 mutex 可能更好 |
| 第一个启动的 goroutine 先执行 | Go 调度器决定执行顺序，不可预测 |

### 扩展问题（如果时间充足）

1. **Test 1**: 如果不用 mutex，如何用 channel 实现？
2. **Test 2**: 如何实现 Compare-And-Set 语义？
3. **Test 3**: 如果两个节点同时超时怎么办？

---

## Quick Reference: Go Concurrency Primitives

| Primitive | Use Case | Example |
|-----------|----------|---------|
| `goroutine` | Concurrent execution | `go func() { ... }()` |
| `sync.WaitGroup` | Wait for multiple goroutines | `wg.Add(1)`, `wg.Done()`, `wg.Wait()` |
| `sync.Mutex` | Protect shared data | `mu.Lock()`, `mu.Unlock()` |
| `sync.RWMutex` | Multiple readers, single writer | `rw.RLock()`, `rw.RUnlock()` |
| `atomic.*` | Lock-free operations | `atomic.AddInt32()`, `atomic.CompareAndSwapInt32()` |
| `channel` | Communication & synchronization | `ch := make(chan int)` |

---

## 最后提醒

- **运行代码前一定要先让学生回答**
- **多运行几次展示不确定性**（特别是 test_2 和 test_3）
- **鼓励学生画时间线图**
- **连接到课程内容**：分布式系统、一致性、Raft 算法

Good luck with your interviews! 🎓
