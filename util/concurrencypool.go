package util

import (
    "sync"
)

// Pool of concurrency slots. Can be used, for example, to limit asynchronous
// processing of items in a queue. Warning: If you mistakenly do more Release()
// calls than Get() calls, the extra Release() call will block, as well as a
// Close() call.
type ConcurrencyPool struct {
    pool chan int
    closed bool
    mutex sync.Mutex
}

// Create a new ConcurrencyPool
func NewConcurrencyPool(slots int) *ConcurrencyPool {
    slots = MaxInt(1, slots)
    cp := new(ConcurrencyPool)
    cp.pool = make(chan int, slots)
    for i := 0; i < slots; i++ {
        cp.pool <- 0
    }
    return cp
}

// Get one slot from the ConcurrencyPool
func (this *ConcurrencyPool) Get() {
    <- this.pool
}

// Release one slot back to the ConcurrencyPool
func (this *ConcurrencyPool) Release() {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    // avoid writing to a closed channel, which would panic
    if !this.closed {
        this.pool <- 0
    }
}

// Close the ConcurrencyPool.
// After closing, all Get() will unblock and all future Get() and Release()
// will just return without blocking.
func (this *ConcurrencyPool) Close() {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    if !this.closed {
        close(this.pool)
        this.closed = true
    }
}
