package util

import (
    "container/list"
    "sync"
)

// Simple Thread-Safe FIFO Queue
type Queue struct {
    list *list.List
    mutex sync.Mutex
    cond *sync.Cond
    closed bool
    wg sync.WaitGroup
}

// Create a new empty Queue
func NewQueue() *Queue {
    q := new(Queue)
    q.list = list.New()
    q.cond = sync.NewCond(&q.mutex)
    return q
}

// Add an item to the end of the Queue
func (this *Queue) Add(v interface{}) {
    this.mutex.Lock()
    if this.closed {
        return
    }
    this.list.PushBack(v)
    this.cond.Signal()
    this.mutex.Unlock()
}

// Remove the first item in the Queue
func (this *Queue) Remove() interface{} {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    e := this.list.Front()
    if e == nil {
        return nil
    }
    return this.list.Remove(e)
}

// Retrieve (but do not remove) the first item in the Queue
func (this *Queue) Peek() interface{} {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    e := this.list.Front()
    if e == nil {
        return nil
    }
    return e.Value
}

// Remove the first item in the Queue.
// Blocks until an item is available.
func (this *Queue) RemoveWait() interface{} {
    this.mutex.Lock()
    if this.closed {
        this.mutex.Unlock()
        return nil
    }
    this.wg.Add(1)
    defer this.wg.Done()
    e := this.list.Front()
    for e == nil {
        this.cond.Wait()
        if this.closed {
            this.mutex.Unlock()
            return nil
        }
        e = this.list.Front()
    }
    defer this.mutex.Unlock()
    return this.list.Remove(e)
}

// Retrieve (but do not remove) the first item in the Queue.
// Blocks until an item is available.
func (this *Queue) PeekWait() interface{} {
    this.mutex.Lock()
    if this.closed {
        this.mutex.Unlock()
        return nil
    }
    this.wg.Add(1)
    defer this.wg.Done()
    e := this.list.Front()
    for e == nil {
        this.cond.Wait()
        if this.closed {
            this.mutex.Unlock()
            return nil
        }
        e = this.list.Front()
    }
    defer this.mutex.Unlock()
    return e.Value
}

// Get the number of items in the Queue
func (this *Queue) Len() int {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    return this.list.Len()
}

// Discard all items in the Queue
func (this *Queue) Clear() {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    this.list.Init()
}

// Close the Queue.
// All blocked *Wait() functions will unblock and return nil.
func (this *Queue) Close() {
    this.mutex.Lock()
    this.list.Init()
    this.closed = true
    this.cond.Broadcast()
    this.mutex.Unlock()
    this.wg.Wait()
}

const (
    LimitStrategyReject = iota
    LimitStrategyCycle
)

type LimitQueue struct {
    Queue
    maxItems int
    strategy int
}

// Create a new empty LimitQueue
func NewLimitQueue(maxItems int, strategy int) *LimitQueue {
    // validate params
    if maxItems <= 0 {
        return nil
    }
    switch strategy {
    case LimitStrategyReject, LimitStrategyCycle:
    default:
        return nil
    }

    q := new(LimitQueue)
    q.list     = list.New()
    q.maxItems = maxItems
    q.strategy = strategy
    q.cond     = sync.NewCond(&q.mutex)

    return q
}

// Add an item to the end of the Queue.
// If the queue is full, use the limit strategy to determine whether to reject
// the new item or to remove the oldest item to make room.
// For LimitStrategyReject, returns whether the item was added.
// For LimitStrategyCycle, returns whether the queue was NOT full.
func (this *LimitQueue) Add(v interface{}) bool {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    if this.closed {
        // closed queue is considered full
        return false
    }
    if this.list.Len() >= this.maxItems && this.strategy == LimitStrategyReject {
        return false
    }
    space := true
    for this.list.Len() >= this.maxItems {
        space = false
        e := this.list.Front()
        if e == nil {
            panic("empty queue. this is a bug.")
        }
        this.list.Remove(e)
    }
    this.list.PushBack(v)
    this.cond.Signal()
    return space
}

// Get the number of available spaces remaining.
func (this *LimitQueue) Available() int {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    if this.closed {
        return 0
    }
    return MaxInt(0, this.maxItems - this.list.Len())
}
